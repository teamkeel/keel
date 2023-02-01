package node

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

type GeneratedFile struct {
	Contents string
	Path     string
}

type GeneratedFiles []*GeneratedFile

func (files GeneratedFiles) Write() error {
	for _, f := range files {
		err := os.MkdirAll(filepath.Dir(f.Path), 0777)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		err = os.WriteFile(f.Path, []byte(f.Contents), 0777)
		if err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}
	return nil
}

type generateOptions struct {
	developmentServer bool
}

// WithDevelopmentServer enables or disables the generation of the development
// server entry point. By default this is disabled.
func WithDevelopmentServer(b bool) func(o *generateOptions) {
	return func(o *generateOptions) {
		o.developmentServer = b
	}
}

// Generate generates and returns a list of objects that represent files to be written
// to a project. Calling .Write() on the result will cause those files be written to disk.
func Generate(ctx context.Context, dir string, opts ...func(o *generateOptions)) (GeneratedFiles, error) {
	options := &generateOptions{}
	for _, o := range opts {
		o(options)
	}

	builder := schema.Builder{}

	schema, err := builder.MakeFromDirectory(dir)
	if err != nil {
		return nil, err
	}

	if !IsEnabled(dir, schema) {
		return GeneratedFiles{}, nil
	}

	files := generateSdkPackage(dir, schema)
	files = append(files, generateTestingPackage(dir, schema)...)

	if options.developmentServer {
		files = append(files, generateDevelopmentServer(dir, schema)...)
	}

	return files, nil
}

func generateSdkPackage(dir string, schema *proto.Schema) GeneratedFiles {
	sdk := &Writer{}
	sdk.Writeln(`const runtime = require("@teamkeel/functions-runtime")`)
	sdk.Writeln("")

	sdkTypes := &Writer{}
	sdkTypes.Writeln(`import { Kysely, Generated } from "kysely"`)
	sdkTypes.Writeln(`import * as runtime from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln("")

	for _, enum := range schema.Enums {
		writeEnum(sdkTypes, enum)
		writeEnumWhereCondition(sdkTypes, enum)

		writeEnumObject(sdk, enum)
	}

	for _, model := range schema.Models {
		writeTableInterface(sdkTypes, model)
		writeModelInterface(sdkTypes, model)
		writeCreateValuesInterface(sdkTypes, model)
		writeWhereConditionsInterface(sdkTypes, model)
		writeUniqueConditionsInterface(sdkTypes, model)
		writeModelAPIDeclaration(sdkTypes, model)
		writeModelQueryBuilderDeclaration(sdkTypes, model)

		writeModelDefaultValuesFunction(sdk, model)

		for _, op := range model.Operations {
			// We only care about custom functions for the SDK
			if op.Implementation != proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
				continue
			}

			writeActionInputTypes(sdkTypes, schema, op, false)
			writeCustomFunctionWrapperType(sdkTypes, model, op)

			sdk.Writef("module.exports.%s = (fn) => fn;", strcase.ToCamel(op.Name))
			sdk.Writeln("")
		}
	}

	writeAPIDeclarations(sdkTypes, schema.Models)
	writeAPIFactory(sdk, schema.Models)

	writeDatabaseInterface(sdkTypes, schema)
	sdk.Writeln("module.exports.getDatabase = runtime.getDatabase")

	return []*GeneratedFile{
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/sdk/index.js"),
			Contents: sdk.String(),
		},
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/sdk/index.d.ts"),
			Contents: sdkTypes.String(),
		},
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/sdk/package.json"),
			Contents: `{"name": "@teamkeel/sdk"}`,
		},
	}
}

func writeTableInterface(w *Writer, model *proto.Model) {
	w.Writef("export interface %sTable {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		w.Write(strcase.ToSnake(field.Name))
		w.Write(": ")
		t := toTypeScriptType(field.Type)
		if field.DefaultValue != nil {
			t = fmt.Sprintf("Generated<%s>", t)
		}
		w.Write(t)
		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeModelInterface(w *Writer, model *proto.Model) {
	w.Writef("export interface %s {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		w.Write(field.Name)
		w.Write(": ")
		t := toTypeScriptType(field.Type)
		w.Write(t)
		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeCreateValuesInterface(w *Writer, model *proto.Model) {
	w.Writef("export interface %sCreateValues {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		// For now you can't create related models when creating a record
		if field.Type.Type == proto.Type_TYPE_MODEL {
			continue
		}
		w.Write(field.Name)
		if field.Optional || field.DefaultValue != nil {
			w.Write("?")
		}
		w.Write(": ")
		t := toTypeScriptType(field.Type)
		w.Write(t)
		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeWhereConditionsInterface(w *Writer, model *proto.Model) {
	w.Writef("export interface %sWhereConditions {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		w.Write(field.Name)
		w.Write("?")
		w.Write(": ")
		w.Write(toTypeScriptType(field.Type))
		w.Write(" | ")
		w.Write(toWhereConditionType(field))
		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeUniqueConditionsInterface(w *Writer, model *proto.Model) {
	w.Writef("export type %sUniqueConditions = \n", model.Name)
	w.Indent()
	for _, f := range model.Fields {
		if f.Unique || f.PrimaryKey {
			w.Writef("| {%s: %s}\n", f.Name, toTypeScriptType(f.Type))
			continue
		}

		// TODO: support f.UniqueWith for compound unique constraints
	}
	w.Dedent()
}

func writeModelAPIDeclaration(w *Writer, model *proto.Model) {
	w.Writef("export type %sAPI = {\n", model.Name)
	w.Indent()
	w.Writef("create(values: %sCreateValues): Promise<%s>;\n", model.Name, model.Name)
	w.Writef("update(where: %sUniqueConditions, values: Partial<%s>): Promise<%s>;\n", model.Name, model.Name, model.Name)
	w.Writef("delete(where: %sUniqueConditions): Promise<string>;\n", model.Name)
	w.Writef("findOne(where: %sUniqueConditions): Promise<%s | null>;\n", model.Name, model.Name)
	w.Writef("findMany(where: %sWhereConditions): Promise<%s[]>;\n", model.Name, model.Name)
	w.Writef("where(where: %sWhereConditions): %sQueryBuilder;\n", model.Name, model.Name)
	w.Dedent()
	w.Writeln("}")
}

func writeModelQueryBuilderDeclaration(w *Writer, model *proto.Model) {
	w.Writef("export type %sQueryBuilder = {\n", model.Name)
	w.Indent()
	w.Writef("where(where: %sWhereConditions): %sQueryBuilder;\n", model.Name, model.Name)
	w.Writef("orWhere(where: %sWhereConditions): %sQueryBuilder;\n", model.Name, model.Name)
	w.Writef("findMany(): Promise<%s[]>;\n", model.Name)
	w.Dedent()
	w.Writeln("}")
}

func writeEnumObject(w *Writer, enum *proto.Enum) {
	w.Writef("module.exports.%s = {\n", enum.Name)
	w.Indent()
	for _, v := range enum.Values {
		w.Write(v.Name)
		w.Write(": ")
		w.Writef(`"%s"`, v.Name)
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("};")
}

func writeEnum(w *Writer, enum *proto.Enum) {
	w.Writef("export enum %s {\n", enum.Name)
	w.Indent()
	for _, v := range enum.Values {
		w.Write(v.Name)
		w.Write(" = ")
		w.Writef(`"%s"`, v.Name)
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeEnumWhereCondition(w *Writer, enum *proto.Enum) {
	w.Writef("export interface %sWhereCondition {\n", enum.Name)
	w.Indent()
	w.Write("equals?: ")
	w.Writeln(enum.Name)
	w.Write("oneOf?: ")
	w.Write(enum.Name)
	w.Writeln("[]")
	w.Dedent()
	w.Writeln("}")
}

func writeDatabaseInterface(w *Writer, schema *proto.Schema) {
	w.Writeln("interface database {")
	w.Indent()
	for _, model := range schema.Models {
		w.Writef("%s: %sTable;", strcase.ToSnake(model.Name), model.Name)
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
	w.Write("export declare function getDatabase(): Kysely<database>;")
}

func writeAPIDeclarations(w *Writer, models []*proto.Model) {
	w.Writeln("export type ModelsAPI = {")
	w.Indent()
	for _, model := range models {
		w.Write(strcase.ToLowerCamel(model.Name))
		w.Write(": ")
		w.Writef(`%sAPI`, model.Name)
		w.Writeln(";")
	}
	w.Dedent()
	w.Writeln("}")

	w.Writeln("export type FunctionAPI = {")
	w.Indent()
	w.Writeln("models: ModelsAPI;")
	w.Dedent()
	w.Writeln("}")
}

func writeAPIFactory(w *Writer, models []*proto.Model) {
	w.Writeln("function createFunctionAPI() {")
	w.Indent()

	w.Writeln("const models = {")
	w.Indent()
	for _, model := range models {
		w.Write(strcase.ToLowerCamel(model.Name))
		w.Write(": ")
		w.Writef(`new runtime.ModelAPI("%s", %sDefaultValues)`, strcase.ToSnake(model.Name), strcase.ToLowerCamel(model.Name))
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("};")
	w.Writeln("return {models};")

	w.Dedent()
	w.Writeln("}")
	w.Writeln("module.exports.createFunctionAPI = createFunctionAPI;")
}

func writeModelDefaultValuesFunction(w *Writer, model *proto.Model) {
	w.Writef("function %sDefaultValues() {", strcase.ToLowerCamel(model.Name))
	w.Writeln("")
	w.Indent()
	w.Writeln("const r = {};")
	for _, field := range model.Fields {
		if field.DefaultValue == nil {
			continue
		}
		if field.DefaultValue.UseZeroValue {
			w.Writef("r.%s = ", field.Name)
			switch field.Type.Type {
			case proto.Type_TYPE_ID:
				w.Write("runtime.ksuid()")
			case proto.Type_TYPE_STRING:
				w.Write(`""`)
			case proto.Type_TYPE_BOOL:
				w.Write(`false`)
			case proto.Type_TYPE_INT:
				w.Write(`0`)
			case proto.Type_TYPE_DATETIME, proto.Type_TYPE_DATE, proto.Type_TYPE_TIMESTAMP:
				w.Write("new Date()")
			}
			w.Writeln(";")
			continue
		}
		// TODO: support expressions
	}
	w.Writeln("return r;")
	w.Dedent()
	w.Writeln("}")
}

func writeActionInputTypes(w *Writer, schema *proto.Schema, op *proto.Operation, isTestingPackage bool) {
	hasWhere := false
	hasValues := false
	for _, i := range op.Inputs {
		if i.Mode == proto.InputMode_INPUT_MODE_READ {
			hasWhere = true
		}
		if i.Mode == proto.InputMode_INPUT_MODE_WRITE {
			hasValues = true
		}
	}
	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_UPDATE, proto.OperationType_OPERATION_TYPE_LIST:
		if hasWhere {
			w.Writef("export interface %sInputWhere ", strcase.ToCamel(op.Name))
			w.Writeln("{")
			w.Indent()
			writeActionInputInterfaceFields(w, schema, op, proto.InputMode_INPUT_MODE_READ, isTestingPackage)
			w.Dedent()
			w.Writeln("}")
		}
	}

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		if hasValues {
			w.Writef("export interface %sInputValues ", strcase.ToCamel(op.Name))
			w.Writeln("{")
			w.Indent()
			writeActionInputInterfaceFields(w, schema, op, proto.InputMode_INPUT_MODE_WRITE, isTestingPackage)
			w.Dedent()
			w.Writeln("}")
		}
	}

	w.Writef("export interface %sInput ", strcase.ToCamel(op.Name))
	w.Writeln("{")
	w.Indent()

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		writeActionInputInterfaceFields(w, schema, op, proto.InputMode_INPUT_MODE_WRITE, isTestingPackage)
	case proto.OperationType_OPERATION_TYPE_GET, proto.OperationType_OPERATION_TYPE_DELETE:
		writeActionInputInterfaceFields(w, schema, op, proto.InputMode_INPUT_MODE_READ, isTestingPackage)
	case proto.OperationType_OPERATION_TYPE_LIST:
		if hasWhere {
			w.Write("where: ")
			w.Writef("%sInputWhere", strcase.ToCamel(op.Name))
			w.Writeln(";")
		}
		// TODO: pagination params e.g. first, after etc...
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		if hasWhere {
			w.Write("where: ")
			w.Writef("%sInputWhere", strcase.ToCamel(op.Name))
			w.Writeln(";")
		}
		if hasValues {
			w.Write("values: ")
			w.Writef("%sInputValues", strcase.ToCamel(op.Name))
			w.Writeln(";")
		}
	}

	w.Dedent()
	w.Writeln("}")
}

func writeActionInputInterfaceFields(w *Writer, schema *proto.Schema, op *proto.Operation, mode proto.InputMode, isTestingPackage bool) {
	for _, input := range op.Inputs {
		if input.Mode != mode {
			continue
		}
		w.Write(input.Name)

		// An optional input doesn't need to be provided at all
		if input.Optional {
			w.Write("?")
		}

		w.Write(": ")

		sdkPrefix := ""
		if isTestingPackage {
			sdkPrefix = "sdk."
		}

		if op.Type == proto.OperationType_OPERATION_TYPE_LIST && input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			switch input.Type.Type {
			case proto.Type_TYPE_DATE:
				w.Write("runtime.DateQueryInput")
			case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
				w.Write("runtime.TimestampQueryInput")
			case proto.Type_TYPE_STRING:
				w.Write("runtime.StringWhereCondition")
			case proto.Type_TYPE_ID:
				w.Write("runtime.IDWhereCondition")
			case proto.Type_TYPE_BOOL:
				w.Write("runtime.BooleanWhereCondition")
			case proto.Type_TYPE_INT:
				w.Write("runtime.NumberWhereCondition")
			case proto.Type_TYPE_ENUM:
				w.Writef("%s%sWhereCondition", sdkPrefix, input.Type.EnumName.Value)
			}
		} else {
			if input.Type.Type == proto.Type_TYPE_ENUM && isTestingPackage {
				w.Write("sdk.")
			}

			w.Write(toTypeScriptType(input.Type))
		}

		nullable := false

		// If an input isn't tied to a model field and it's optional then it's allowed to be null
		if input.Type.FieldName == nil && input.Optional {
			nullable = true
		}

		// If an input is tied to a model field and that field is nullable then the input is also nullable
		if input.Type.FieldName != nil {
			f := proto.FindField(schema.Models, input.Type.ModelName.Value, input.Type.FieldName.Value)
			if f.Optional {
				nullable = true
			}
		}

		if nullable {
			w.Write(" | null")
		}

		w.Writeln(";")
	}
}

func writeCustomFunctionWrapperType(w *Writer, model *proto.Model, op *proto.Operation) {
	w.Writef("export declare function %s", strcase.ToCamel(op.Name))
	w.Writef("(fn: (inputs: %sInput, api: FunctionAPI) => ", strcase.ToCamel(op.Name))
	w.Write(toCustomFunctionReturnType(model, op, false))
	w.Write("): ")
	w.Write(toCustomFunctionReturnType(model, op, false))
	w.Writeln(";")
}

func toCustomFunctionReturnType(model *proto.Model, op *proto.Operation, isTestingPackage bool) string {
	returnType := "Promise<"
	sdkPrefix := ""
	if isTestingPackage {
		sdkPrefix = "sdk."
	}
	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		returnType += sdkPrefix + model.Name
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		returnType += sdkPrefix + model.Name
	case proto.OperationType_OPERATION_TYPE_GET:
		returnType += sdkPrefix + model.Name + " | null"
	case proto.OperationType_OPERATION_TYPE_LIST:
		returnType += sdkPrefix + model.Name + "[]"
	case proto.OperationType_OPERATION_TYPE_DELETE:
		returnType += "string"
	}
	returnType += ">"
	return returnType
}

func toActionReturnType(model *proto.Model, op *proto.Operation) string {
	returnType := "Promise<"
	sdkPrefix := "sdk."

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		returnType += sdkPrefix + model.Name
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		returnType += sdkPrefix + model.Name
	case proto.OperationType_OPERATION_TYPE_GET:
		returnType += sdkPrefix + model.Name + " | null"
	case proto.OperationType_OPERATION_TYPE_LIST:
		returnType += "{results: " + sdkPrefix + model.Name + "[], hasNextPage: boolean}"
	case proto.OperationType_OPERATION_TYPE_DELETE:
		// todo: create ID type
		returnType += "string"
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		// TODO: fix this when authenticate has been re-worked following Arbitrary Functions
		// https://www.notion.so/keelhq/Arbitrary-Functions-428c199902cf4353b18838434c8910d1
		returnType += "any"
	}

	returnType += ">"
	return returnType
}

func GenerateDevelopmentServerImportsAndFunctions(schema *proto.Schema) string {
	w := &Writer{}
	w.Writeln(`import { handleRequest } from '@teamkeel/functions-runtime';`)
	w.Writeln(`import { createFunctionAPI } from '@teamkeel/sdk';`)
	w.Writeln(`import { createServer } from "http";`)

	functionNames := []string{}
	for _, model := range schema.Models {
		for _, op := range model.Operations {
			if op.Implementation != proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
				continue
			}
			functionNames = append(functionNames, op.Name)
			// namespace import to avoid naming clashes
			w.Writef(`import function_%s from "../functions/%s.ts"`, op.Name, op.Name)
			w.Writeln(";")
		}
	}

	w.Writeln("const functions = {")
	w.Indent()
	for _, name := range functionNames {
		w.Writef("%s: function_%s,", name, name)
	}
	w.Dedent()
	w.Writeln("}")
	return w.String()
}

func generateDevelopmentServer(dir string, schema *proto.Schema) GeneratedFiles {
	w := &Writer{}
	w.Writeln(GenerateDevelopmentServerImportsAndFunctions(schema))

	w.Writeln(`
const listener = async (req, res) => {
	const u = new URL(req.url, "http://" + req.headers.host);
	if (req.method === "GET" && u.pathname === "/_health") {
		res.statusCode = 200;
		res.end();
		return;
	}

	if (req.method === "POST") {
		const buffers = [];
		for await (const chunk of req) {
			buffers.push(chunk);
		}
		const data = Buffer.concat(buffers).toString();
		const json = JSON.parse(data);

		const rpcResponse = await handleRequest(json, {
			functions,
			createFunctionAPI,
		});

		res.statusCode = 200;
		res.setHeader('Content-Type', 'application/json');
		res.write(JSON.stringify(rpcResponse));
		res.end();
		return;
	}

	res.statusCode = 400;
	res.end();
};

const server = createServer(listener);
const port = (process.env.PORT && parseInt(process.env.PORT, 10)) || 3001;
server.listen(port);`)

	return []*GeneratedFile{
		{
			Path:     filepath.Join(dir, ".build/server.js"),
			Contents: w.String(),
		},
	}
}

func generateTestingPackage(dir string, schema *proto.Schema) GeneratedFiles {
	js := &Writer{}
	types := &Writer{}

	// The testing package uses ES modules as it only used in the context of running tests
	// with Vitest
	js.Writeln(`import crypto from "node:crypto";`)
	js.Writeln(`import { getDatabase, createFunctionAPI } from "@teamkeel/sdk"`)
	js.Writeln(`import { ActionExecutor, sql } from "@teamkeel/testing-runtime";`)
	js.Writeln("")

	js.Writeln(`export const actions = new ActionExecutor({});`)
	js.Writeln("export const models = createFunctionAPI().models;")

	js.Writeln("export async function resetDatabase() {")
	js.Indent()
	js.Write("await sql`TRUNCATE TABLE ")
	tableNames := []string{}
	for _, model := range schema.Models {
		tableNames = append(tableNames, strcase.ToSnake(model.Name))
	}
	js.Writef("%s CASCADE", strings.Join(tableNames, ","))
	js.Writeln("`.execute(getDatabase());")
	js.Dedent()
	js.Writeln("}")

	writeTestingTypes(types, schema)

	return GeneratedFiles{
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/testing/index.mjs"),
			Contents: js.String(),
		},
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/testing/index.d.ts"),
			Contents: types.String(),
		},
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/testing/package.json"),
			Contents: `{"name": "@teamkeel/testing", "type": "module", "exports": "./index.mjs"}`,
		},
	}
}

func writeTestingTypes(w *Writer, schema *proto.Schema) {
	w.Writeln(`import * as sdk from "@teamkeel/sdk";`)
	w.Writeln(`import * as runtime from "@teamkeel/functions-runtime";`)

	// We need to import the testing-runtime package to get
	// the types for the extended vitest matchers e.g. expect(v).toHaveAuthorizationError()
	w.Writeln(`import "@teamkeel/testing-runtime";`)
	w.Writeln("")

	// For the testing package we need input types for all actions
	for _, model := range schema.Models {
		for _, op := range model.Operations {
			writeActionInputTypes(w, schema, op, true)
		}
	}

	w.Writeln("declare class ActionExecutor {")
	w.Indent()
	w.Writeln("withIdentity(identity: sdk.Identity): ActionExecutor;")
	w.Writeln("withAuthToken(token: string): ActionExecutor;")
	for _, model := range schema.Models {
		for _, op := range model.Operations {
			w.Writef(`%s(i: %sInput): %s`, op.Name, strcase.ToCamel(op.Name), toActionReturnType(model, op))
			w.Writeln(";")
		}
	}
	w.Dedent()
	w.Writeln("}")
	w.Writeln("export declare const actions: ActionExecutor;")
	w.Writeln("export declare const models: sdk.ModelsAPI;")
	w.Writeln("export declare function resetDatabase(): Promise<void>;")
}

func toTypeScriptType(t *proto.TypeInfo) string {
	switch t.Type {
	case proto.Type_TYPE_ID:
		return "string"
	case proto.Type_TYPE_STRING:
		return "string"
	case proto.Type_TYPE_BOOL:
		return "boolean"
	case proto.Type_TYPE_INT:
		return "number"
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		return "Date"
	case proto.Type_TYPE_ENUM:
		return t.EnumName.Value
	default:
		return "any"
	}
}

func toWhereConditionType(f *proto.Field) string {
	switch f.Type.Type {
	case proto.Type_TYPE_ID:
		return "runtime.IDWhereCondition"
	case proto.Type_TYPE_STRING:
		return "runtime.StringWhereCondition"
	case proto.Type_TYPE_BOOL:
		return "runtime.BooleanWhereCondition"
	case proto.Type_TYPE_INT:
		return "runtime.NumberWhereCondition"
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		return "runtime.DateWhereCondition"
	case proto.Type_TYPE_ENUM:
		return fmt.Sprintf("%sWhereCondition", f.Type.EnumName.Value)
	default:
		return "any"
	}
}

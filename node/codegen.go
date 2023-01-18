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
	sdkTypes.Writeln(`import { Generated } from "kysely"`)
	sdkTypes.Writeln(`import * as runtime from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln("")

	for _, enum := range schema.Enums {
		writeEnum(sdkTypes, enum)
		writeEnumWhereCondition(sdkTypes, enum)
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
			if op.Implementation != proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
				continue
			}

			writeCustomFunctionInputTypes(sdkTypes, op)
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
	w.Write("oneOf?: [")
	w.Write(enum.Name)
	w.Writeln("]")
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
				w.Write("KSUID.randomSync().string")
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

func writeCustomFunctionInputTypes(w *Writer, op *proto.Operation) {
	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_UPDATE, proto.OperationType_OPERATION_TYPE_LIST:
		w.Writef("export interface %sInputWhere ", strcase.ToCamel(op.Name))
		w.Writeln("{")
		w.Indent()
		writeInputInterfaceFields(w, op, proto.InputMode_INPUT_MODE_READ)
		w.Dedent()
		w.Writeln("}")
	}

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		w.Writef("export interface %sInputValues ", strcase.ToCamel(op.Name))
		w.Writeln("{")
		w.Indent()
		writeInputInterfaceFields(w, op, proto.InputMode_INPUT_MODE_WRITE)
		w.Dedent()
		w.Writeln("}")
	}

	w.Writef("export interface %sInput ", strcase.ToCamel(op.Name))
	w.Writeln("{")
	w.Indent()

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		writeInputInterfaceFields(w, op, proto.InputMode_INPUT_MODE_WRITE)
	case proto.OperationType_OPERATION_TYPE_GET, proto.OperationType_OPERATION_TYPE_DELETE:
		writeInputInterfaceFields(w, op, proto.InputMode_INPUT_MODE_READ)
	case proto.OperationType_OPERATION_TYPE_LIST:
		w.Write("where: ")
		w.Writef("%sInputWhere", strcase.ToCamel(op.Name))
		w.Writeln(";")
		// TODO: pagination params e.g. first, after etc...
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		w.Write("where: ")
		w.Writef("%sInputWhere", strcase.ToCamel(op.Name))
		w.Writeln(";")
		w.Write("values: ")
		w.Writef("%sInputValues", strcase.ToCamel(op.Name))
		w.Writeln(";")
	}

	w.Dedent()
	w.Writeln("}")
}

func writeInputInterfaceFields(w *Writer, op *proto.Operation, mode proto.InputMode) {
	for _, input := range op.Inputs {
		if input.Mode != mode {
			continue
		}
		w.Write(input.Name)
		if input.Optional {
			w.Write("?")
		}
		w.Write(": ")
		w.Write(toTypeScriptType(input.Type))
		w.Writeln(";")
	}
}

func writeCustomFunctionWrapperType(w *Writer, model *proto.Model, op *proto.Operation) {
	w.Writef("export declare function %s", strcase.ToCamel(op.Name))
	w.Writef("(fn: (inputs: %sInput, api: FunctionAPI) => ", strcase.ToCamel(op.Name))
	w.Write(toOperationReturnType(model, op, false))
	w.Write("): ")
	w.Write(toOperationReturnType(model, op, false))
	w.Writeln(";")
}

func toOperationReturnType(model *proto.Model, op *proto.Operation, isTestingPackage bool) string {
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

func generateDevelopmentServer(dir string, schema *proto.Schema) GeneratedFiles {
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

	js.Writeln(`const crypto = require("crypto")`)
	js.Writeln(`const { getDatabase, createFunctionAPI } = require("@teamkeel/sdk");`)
	js.Writeln(`const { sql } = require("kysely");`)
	js.Writeln("")

	// If there ends up being significantly more generic code for this package then
	// it would probably make sense to move it into a @teamkeel/testing-runtime package.
	// But for now it's just this one class, so it can live here.
	js.Writeln(`
class ActionExecutor {
	constructor() {
		return new Proxy(this, {
			get(target, prop, receiver) {
				if (["withIdentity", "_execute"].includes(prop)) {
					return Reflect.get(...arguments);
				}
				return target._execute.bind(target, prop);
			},
		});
	}
	withIdentity(i) {
		this._identity = i;
		return this;
	}
	_execute(method, params) {
		const headers = { "Content-Type": "application/json" };
		if (this._identity) {
			headers["X-Testing-Identity-ID"] = this._identity.id;
		}
		return fetch(process.env.KEEL_ACTIONS_RPC_URL, {
			method: "POST",
			body: JSON.stringify({
				jsonrpc: "2.0",
				method,
				params,
				id: crypto.randomBytes(16).toString('hex'),
			}),
			headers,
		}).then((r) => r.json());
	}
}

module.exports.actions = new ActionExecutor()`)

	js.Writeln("module.exports.models = createFunctionAPI().models;")

	js.Writeln("async function resetDatabase() {")
	js.Indent()
	js.Write("await sql`TRUNCATE TABLE ")
	tableNames := []string{}
	for _, model := range schema.Models {
		tableNames = append(tableNames, strcase.ToLowerCamel(model.Name))
	}
	js.Writef("%s CASCADE", strings.Join(tableNames, ","))
	js.Writeln("`.execute(getDatabase());")
	js.Dedent()
	js.Writeln("}")
	js.Writeln("module.exports.resetDatabase = resetDatabase;")

	writeTestingTypes(types, schema)

	return GeneratedFiles{
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/testing/index.js"),
			Contents: js.String(),
		},
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/testing/index.d.ts"),
			Contents: types.String(),
		},
		{
			Path:     filepath.Join(dir, "node_modules/@teamkeel/testing/package.json"),
			Contents: `{"name": "@teamkeel/testing"}`,
		},
	}
}

func writeTestingTypes(w *Writer, schema *proto.Schema) {
	w.Writeln(`import * as sdk from "@teamkeel/sdk"`)
	w.Writeln("")

	w.Writeln("declare class ActionExecutor {")
	w.Indent()
	w.Writeln("withIdentity(identity: sdk.Identity): ActionExecutor;")
	for _, model := range schema.Models {
		for _, op := range model.Operations {
			w.Writef(`async %s(i: sdk.%sInput): %s`, op.Name, strcase.ToCamel(op.Name), toOperationReturnType(model, op, true))
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

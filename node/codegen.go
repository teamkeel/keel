package node

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

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
// This function should not interact with the file system so it can be used in a backend
// context.
func Generate(ctx context.Context, schema *proto.Schema, opts ...func(o *generateOptions)) (codegen.GeneratedFiles, error) {
	options := &generateOptions{}
	for _, o := range opts {
		o(options)
	}

	files := generateSdkPackage(schema)
	files = append(files, generateTestingPackage(schema)...)
	files = append(files, generateTestingSetup()...)

	if options.developmentServer {
		files = append(files, generateDevelopmentServer(schema)...)
	}

	return files, nil
}

func generateSdkPackage(schema *proto.Schema) codegen.GeneratedFiles {
	sdk := &codegen.Writer{}
	sdk.Writeln(`const { sql } = require("kysely")`)
	sdk.Writeln(`const runtime = require("@teamkeel/functions-runtime")`)
	sdk.Writeln("")

	sdkTypes := &codegen.Writer{}
	sdkTypes.Writeln(`import { Kysely, Generated } from "kysely"`)
	sdkTypes.Writeln(`import * as runtime from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln(`import { Headers } from 'node-fetch'`)
	sdkTypes.Writeln("")

	// deepFreeze is used to make the inputs object to function hooks immutable
	sdk.Writeln(`
const deepFreeze = o => {
	if (o===null || typeof o !== 'object') return o
	return new Proxy(o, {
		get(obj, prop) {
			return deepFreeze(obj[prop])
		},
		set(obj, prop) {
			throw new Error("Input " + JSON.stringify(obj) + " cannot be modified. Did you mean to modify values instead?")
		}
	})
}
	`)

	writePermissions(sdk, schema)

	writeMessages(sdkTypes, schema, false)

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
		writeFindManyParamsInterface(sdkTypes, model, false)
		writeUniqueConditionsInterface(sdkTypes, model)
		writeModelAPIDeclaration(sdkTypes, model)
		writeModelQueryBuilderDeclaration(sdkTypes, model)

		for _, action := range model.Actions {
			// We only care about custom functions for the SDK
			if action.Implementation != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}

			// writes new types to the index.d.ts to annotate the underlying vanilla javascript
			// implementation of a function with nice types
			writeFunctionWrapperType(sdkTypes, model, action)

			// if the action type is read or write, then the signature of the exported method just takes the function
			// defined by the user
			if proto.ActionIsArbitraryFunction(action) {
				sdk.Writef("module.exports.%s = (fn) => fn;", casing.ToCamel(action.Name))
			} else {
				// writes the default implementation of a function. the user can specify hooks which can
				// override the behaviour of the default implementation
				writeFunctionImplementation(sdk, schema, action)

				sdk.Writef("module.exports.%s = %s;", casing.ToCamel(action.Name), casing.ToCamel(action.Name))
			}

			sdk.Writeln("")
		}
	}

	for _, job := range schema.Jobs {
		writeJobFunctionWrapperType(sdkTypes, job)
		sdk.Writef("module.exports.%s = (fn) => fn;", job.Name)
		sdk.Writeln("")
	}

	for _, subscriber := range schema.Subscribers {
		writeSubscriberFunctionWrapperType(sdkTypes, subscriber)
		sdk.Writef("module.exports.%s = (fn) => fn;", casing.ToCamel(subscriber.Name))
		sdk.Writeln("")
	}

	writeTableConfig(sdk, schema.Models)

	writeAPIFactory(sdk, schema)

	writeDatabaseInterface(sdkTypes, schema)
	writeAPIDeclarations(sdkTypes, schema)

	sdk.Writeln("module.exports.useDatabase = runtime.useDatabase;")

	return []*codegen.GeneratedFile{
		{
			Path:     "node_modules/@teamkeel/sdk/index.js",
			Contents: sdk.String(),
		},
		{
			Path:     "node_modules/@teamkeel/sdk/index.d.ts",
			Contents: sdkTypes.String(),
		},
		{
			Path:     "node_modules/@teamkeel/sdk/package.json",
			Contents: `{"name": "@teamkeel/sdk"}`,
		},
	}
}

func writeTableInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export interface %sTable {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			continue
		}
		w.Write(casing.ToLowerCamel(field.Name))
		w.Write(": ")
		t := toTypeScriptType(field.Type, false)
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

func writeModelInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export interface %s {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			continue
		}
		w.Write(field.Name)
		w.Write(": ")
		t := toTypeScriptType(field.Type, false)
		w.Write(t)
		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeCreateValuesInterface(w *codegen.Writer, model *proto.Model) {
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
		t := toTypeScriptType(field.Type, false)
		w.Write(t)
		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeFindManyParamsInterface(w *codegen.Writer, model *proto.Model, isTestingPackage bool) {
	w.Writeln(`export type SortDirection = "asc" | "desc" | "ASC" | "DESC"`)
	w.Writef("export type %sOrderBy = {\n", model.Name)
	w.Indent()

	relevantFields := lo.Filter(model.Fields, func(f *proto.Field, _ int) bool {
		switch f.Type.Type {
		// scalar types are only permitted to sort by
		case proto.Type_TYPE_BOOL, proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_INT, proto.Type_TYPE_STRING, proto.Type_TYPE_ENUM, proto.Type_TYPE_TIMESTAMP, proto.Type_TYPE_ID:
			return true
		default:
			// includes types such as password, secret, model etc
			return false
		}
	})

	for i, f := range relevantFields {
		w.Writef("%s?: SortDirection", f.Name)

		if i < len(relevantFields)-1 {
			w.Write(",")
		}

		w.Write("\n")
	}
	w.Dedent()
	w.Write("}")

	w.Writeln("\n")
	w.Writef("export interface %sFindManyParams {\n", model.Name)
	w.Indent()
	w.Writef("where?: %sWhereConditions;\n", model.Name)
	w.Writef("limit?: number;\n")
	w.Writef("offset?: number;\n")
	w.Writef("orderBy?: %sOrderBy;\n", model.Name)
	w.Dedent()
	w.Writeln("}")
}

func writeWhereConditionsInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export interface %sWhereConditions {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		w.Write(field.Name)
		w.Write("?")
		w.Write(": ")
		if field.Type.Type == proto.Type_TYPE_MODEL {
			// Embed related models where conditions
			w.Writef("%sWhereConditions | null;", field.Type.ModelName.Value)
		} else {
			w.Write(toTypeScriptType(field.Type, false))
			w.Write(" | ")
			w.Write(toWhereConditionType(field))
			w.Write(" | null;")
		}

		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeMessages(w *codegen.Writer, schema *proto.Schema, isTestingPackage bool) {
	for _, msg := range schema.Messages {
		if msg.Name == parser.MessageFieldTypeAny {
			continue
		}
		writeMessage(w, schema, msg, isTestingPackage)
	}
}

func writeMessage(w *codegen.Writer, schema *proto.Schema, message *proto.Message, isTestingPackage bool) {
	if message.Type != nil {
		w.Writef("export type %s = ", message.Name)
		w.Write(toTypeScriptType(message.Type, isTestingPackage))
		w.Writeln(";")
		return
	}

	w.Writef("export interface %s {\n", message.Name)
	w.Indent()

	for _, field := range message.Fields {
		w.Write(field.Name)

		if field.Optional {
			w.Write("?")
		}

		w.Write(": ")

		w.Write(toTypeScriptType(field.Type, isTestingPackage))

		if field.Type.Repeated {
			w.Write("[]")
		}

		if field.Nullable {
			w.Write(" | null")
		}

		w.Writeln(";")
	}

	w.Dedent()
	w.Writeln("}")
}

func writeUniqueConditionsInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sUniqueConditions = ", model.Name)
	w.Indent()
	for _, f := range model.Fields {
		var tsType string

		switch {
		case f.Unique || f.PrimaryKey || len(f.UniqueWith) > 0:
			tsType = toTypeScriptType(f.Type, false)
		case proto.IsHasMany(f):
			// If a model "has one" of another model then you can
			// do a lookup on any of that models unique fields
			tsType = fmt.Sprintf("%sUniqueConditions", f.Type.ModelName.Value)
		default:
			// TODO: support f.UniqueWith for compound unique constraints
			continue
		}

		w.Writeln("")
		w.Writef("| {%s: %s}", f.Name, tsType)
	}
	w.Writeln(";")
	w.Dedent()
}

func writeModelAPIDeclaration(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sAPI = {\n", model.Name)
	w.Indent()

	nonOptionalFields := lo.Filter(model.Fields, func(f *proto.Field, _ int) bool {
		return !f.Optional && f.DefaultValue == nil
	})

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Create a %s record\n", model.Name)
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const record = await models.%s.create({\n", casing.ToLowerCamel(model.Name))
		w.Indent()

		for i, f := range nonOptionalFields {
			w.Writef("%s: ", casing.ToLowerCamel(f.Name))

			switch f.Type.Type {
			case proto.Type_TYPE_STRING:
				w.Write("''")
			case proto.Type_TYPE_BOOL:
				w.Write("false")
			case proto.Type_TYPE_INT:
				w.Write("0")
			case proto.Type_TYPE_DATETIME, proto.Type_TYPE_DATE, proto.Type_TYPE_TIMESTAMP:
				w.Write("new Date()")
			default:
				w.Write("undefined")
			}

			if i < len(nonOptionalFields)-1 {
				w.Write(",")
			}

			w.Write("\n")
		}
		w.Dedent()
		w.Writeln("});")
		w.Writeln("```")
	})
	w.Writef("create(values: %sCreateValues): Promise<%s>;\n", model.Name, model.Name)

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Update a %s record\n", model.Name)
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const %s = await models.%s.update(", casing.ToLowerCamel(model.Name), casing.ToLowerCamel(model.Name))
		w.Writef("{ id: \"abc\" },")
		if len(nonOptionalFields) > 0 {
			w.Writef(" { %s: XXX }", casing.ToLowerCamel(nonOptionalFields[0].Name))
		} else {
			w.Write("  {}")
		}
		w.Writeln("});")
		w.Writeln("```")
	})
	w.Writef("update(where: %sUniqueConditions, values: Partial<%s>): Promise<%s>;\n", model.Name, model.Name, model.Name)

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Deletes a %s record\n", model.Name)
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const deletedId = await models.%s.delete({ id: 'xxx' });\n", casing.ToLowerCamel(model.Name))
		w.Writeln("```")
	})
	w.Writef("delete(where: %sUniqueConditions): Promise<string>;\n", model.Name)

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Finds a single %s record\n", model.Name)
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const %s = await models.%s.findOne({ id: 'xxx' });\n", casing.ToLowerCamel(model.Name), casing.ToLowerCamel(model.Name))
		w.Writeln("```")
	})
	w.Writef("findOne(where: %sUniqueConditions): Promise<%s | null>;\n", model.Name, model.Name)
	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Finds multiple %s records\n", model.Name)
		w.Writeln("* @example")
		w.Writeln("```typescript")

		// cant seem to get markdown in vscode method signature popover to render indentation
		// so we have to get it all on one line for the meantime
		w.Writef(`const %ss = await models.%s.findMany({ where: { createdAt: { after: new Date(2022, 1, 1) } }, orderBy: { id: 'asc' }, limit: 1000, offset: 50 });`, casing.ToLowerCamel(model.Name), casing.ToLowerCamel(model.Name))
		w.Writeln("")
		w.Writeln("```")
	})
	w.Writef("findMany(params?: %sFindManyParams | undefined): Promise<%s[]>;\n", model.Name, model.Name)

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writeln("* Creates a new query builder with the given conditions applied")
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const records = await models.%s.where({ createdAt: { after: new Date(2020, 1, 1) } }).orWhere({ updatedAt: { after: new Date(2020, 1, 1) } }).findMany();\n", casing.ToLowerCamel(model.Name))
		w.Writeln("```")
	})
	w.Writef("where(where: %sWhereConditions): %sQueryBuilder;\n", model.Name, model.Name)
	w.Dedent()
	w.Writeln("}")
}

func writeModelQueryBuilderDeclaration(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sQueryBuilderParams = {\n", model.Name)
	w.Indent()
	w.Writef("limit?: number;\n")
	w.Writef("offset?: number;\n")
	w.Writef("orderBy?: %sOrderBy;\n", model.Name)
	w.Dedent()
	w.Writeln("}")

	// the following types are for the chained version of the model api
	// e.g await models.foo.where({ bar: 'bazz' }).update({ bar: 'boo' })
	w.Writef("export type %sQueryBuilder = {\n", model.Name)
	w.Indent()
	w.Writef("where(where: %sWhereConditions): %sQueryBuilder;\n", model.Name, model.Name)
	w.Writef("orWhere(where: %sWhereConditions): %sQueryBuilder;\n", model.Name, model.Name)
	w.Writef("findMany(params?: %sQueryBuilderParams): Promise<%s[]>;\n", model.Name, model.Name)
	w.Writef("findOne(params?: %sQueryBuilderParams): Promise<%s>;\n", model.Name, model.Name)

	// todo: support these.
	// w.Writef("limit(limit: number) : %sQueryBuilder;\n", model.Name)
	// w.Writef("offset(offset: number) : %sQueryBuilder;\n", model.Name)
	// w.Writef("orderBy(conditions: %sOrderBy) : %sQueryBuilder;\n", model.Name, model.Name)

	w.Writef("delete() : Promise<string>;\n")
	w.Writef("update(values: Partial<%s>) : Promise<%s>;\n", model.Name, model.Name)
	w.Dedent()
	w.Writeln("}")
}

func writeEnumObject(w *codegen.Writer, enum *proto.Enum) {
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

func writeEnum(w *codegen.Writer, enum *proto.Enum) {
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

func writeEnumWhereCondition(w *codegen.Writer, enum *proto.Enum) {
	w.Writef("export interface %sWhereCondition {\n", enum.Name)
	w.Indent()
	w.Write("equals?: ")
	w.Write(enum.Name)
	w.Writeln(" | null;")
	w.Write("oneOf?: ")
	w.Write(enum.Name)
	w.Write("[]")
	w.Writeln(" | null;")
	w.Dedent()
	w.Writeln("}")
}

func writeDatabaseInterface(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln("interface database {")
	w.Indent()
	for _, model := range schema.Models {
		w.Writef("%s: %sTable;", casing.ToSnake(model.Name), model.Name)
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
	w.Writeln("export declare function useDatabase(): Kysely<database>;")
}

func writeAPIDeclarations(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln("export type ModelsAPI = {")
	w.Indent()
	for _, model := range schema.Models {
		w.Write(casing.ToLowerCamel(model.Name))
		w.Write(": ")
		w.Writef(`%sAPI`, model.Name)
		w.Writeln(";")
	}
	w.Dedent()
	w.Writeln("}")
	w.Writeln("export declare const models: ModelsAPI;")
	w.Writeln("export declare const permissions: runtime.Permissions;")

	w.Writeln("type Environment = {")

	w.Indent()

	for _, variable := range schema.EnvironmentVariables {
		w.Writef("%s: string;\n", variable.Name)
	}

	w.Dedent()
	w.Writeln("}")
	w.Writeln("type Secrets = {")

	w.Indent()

	for _, secret := range schema.Secrets {
		w.Writef("%s: string;\n", secret.Name)
	}

	w.Dedent()
	w.Writeln("}")
	w.Writeln("")

	w.Writeln("export interface ContextAPI extends runtime.ContextAPI {")
	w.Indent()
	w.Writeln("secrets: Secrets;")
	w.Writeln("env: Environment;")
	w.Writeln("identity?: Identity;")
	w.Writeln("now(): Date;")
	w.Dedent()
	w.Writeln("}")

	w.Writeln("export interface JobContextAPI {")
	w.Indent()
	w.Writeln("secrets: Secrets;")
	w.Writeln("env: Environment;")
	w.Writeln("identity?: Identity;")
	w.Writeln("now(): Date;")
	w.Dedent()
	w.Writeln("}")

	w.Writeln("export interface SubscriberContextAPI {")
	w.Indent()
	w.Writeln("secrets: Secrets;")
	w.Writeln("env: Environment;")
	w.Writeln("now(): Date;")
	w.Dedent()
	w.Writeln("}")
}

func writeAPIFactory(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln("function createContextAPI({ responseHeaders, meta }) {")
	w.Indent()
	w.Writeln("const headers = new runtime.RequestHeaders(meta.headers);")
	w.Writeln("const response = { headers: responseHeaders }")
	w.Writeln("const now = () => { return new Date(); };")
	w.Writeln("const { identity } = meta;")
	w.Writeln("const isAuthenticated = identity != null;")
	w.Writeln("const env = {")
	w.Indent()

	for _, variable := range schema.EnvironmentVariables {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.Name, variable.Name)
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.Secrets {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.Name, secret.Name)
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("return { headers, response, identity, env, now, secrets, isAuthenticated };")
	w.Dedent()
	w.Writeln("};")

	w.Writeln("function createJobContextAPI({ meta }) {")
	w.Indent()
	w.Writeln("const now = () => { return new Date(); };")
	w.Writeln("const { identity } = meta;")
	w.Writeln("const isAuthenticated = identity != null;")
	w.Writeln("const env = {")
	w.Indent()

	for _, variable := range schema.EnvironmentVariables {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.Name, variable.Name)
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.Secrets {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.Name, secret.Name)
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("return { identity, env, now, secrets, isAuthenticated };")
	w.Dedent()
	w.Writeln("};")

	w.Writeln("function createSubscriberContextAPI({ meta }) {")
	w.Indent()
	w.Writeln("const now = () => { return new Date(); };")
	w.Writeln("const env = {")
	w.Indent()

	for _, variable := range schema.EnvironmentVariables {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.Name, variable.Name)
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.Secrets {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.Name, secret.Name)
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("return { env, now, secrets };")
	w.Dedent()
	w.Writeln("};")

	w.Writeln("function createModelAPI() {")
	w.Indent()
	w.Writeln("return {")
	w.Indent()
	for _, model := range schema.Models {
		w.Write(casing.ToLowerCamel(model.Name))
		w.Write(": ")

		// The second positional argument to the model API used to be a default values function but
		// default values are now set in the database so this is no longer needed.
		// Passing a no-op function here for backwards compatibility with older versions of the
		// functions-runtime package.
		w.Writef(`new runtime.ModelAPI("%s", () => ({}), tableConfigMap)`, casing.ToSnake(model.Name))

		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("};")
	w.Dedent()
	w.Writeln("};")

	w.Writeln("function createPermissionApi() {")
	w.Indent()
	w.Writeln("return new runtime.Permissions();")
	w.Dedent()
	w.Writeln("};")

	w.Writeln(`module.exports.models = createModelAPI();`)
	w.Writeln(`module.exports.permissions = createPermissionApi();`)
	w.Writeln("module.exports.createContextAPI = createContextAPI;")
	w.Writeln("module.exports.createJobContextAPI = createJobContextAPI;")
	w.Writeln("module.exports.createSubscriberContextAPI = createSubscriberContextAPI;")
}

func writeTableConfig(w *codegen.Writer, models []*proto.Model) {
	w.Write("const tableConfigMap = ")

	// In case the words map and string over and over aren't clear enough
	// for you see the packages/functions-runtime/src/ModelAPI.js file for
	// docs on how this object is expected to be structured
	tableConfigMap := map[string]map[string]map[string]string{}

	for _, model := range models {
		for _, field := range model.Fields {
			if field.Type.Type != proto.Type_TYPE_MODEL {
				continue
			}

			relationshipConfig := map[string]string{
				"referencesTable": casing.ToSnake(field.Type.ModelName.Value),
				"foreignKey":      casing.ToSnake(proto.GetForignKeyFieldName(models, field)),
			}

			switch {
			case proto.IsHasOne(field):
				relationshipConfig["relationshipType"] = "hasOne"
			case proto.IsHasMany(field):
				relationshipConfig["relationshipType"] = "hasMany"
			case proto.IsBelongsTo(field):
				relationshipConfig["relationshipType"] = "belongsTo"
			}

			tableConfig, ok := tableConfigMap[casing.ToSnake(model.Name)]
			if !ok {
				tableConfig = map[string]map[string]string{}
				tableConfigMap[casing.ToSnake(model.Name)] = tableConfig
			}

			tableConfig[field.Name] = relationshipConfig
		}
	}

	b, _ := json.MarshalIndent(tableConfigMap, "", "    ")
	w.Write(string(b))
	w.Writeln(";")
}

func writeBeforeQueryHook(w *codegen.Writer, action *proto.Action) {
	w.Writeln("")

	w.Writeln("// call beforeQuery hook (if defined)")
	w.Writeln("if (hooks.beforeQuery) {")
	w.Indent()
	w.Writef("let builder = models.%s.where(wheres);\n", casing.ToLowerCamel(action.ModelName))
	w.Writeln("")

	resolvedReturnType := ""
	switch action.Type {
	case proto.ActionType_ACTION_TYPE_GET:
		resolvedReturnType = action.ModelName
	case proto.ActionType_ACTION_TYPE_LIST:
		resolvedReturnType = fmt.Sprintf("%s[]", action.ModelName)
	case proto.ActionType_ACTION_TYPE_DELETE:
		resolvedReturnType = "string"
	}

	w.Writef("// we don't know if its an instance of %sQueryBuilder or Promise<%s> so we wrap in Promise.resolve to get the eventual value.\n", action.ModelName, resolvedReturnType)

	w.Writeln("let resolvedValue;")

	wrapWithSpan(w, fmt.Sprintf("%s.beforeQuery", action.Name), func(w *codegen.Writer) {
		w.Writef("resolvedValue = await hooks.beforeQuery(ctx, deepFreeze(inputs), builder);\n")
	})
	w.Writeln("")

	// we want to check if the resolved value is an instance of the runtime.QueryBuilder.
	// instanceof has some gotchas particularly between esmodules, so we revert to checking the constuctor of the resolvedValue

	w.Writeln("const constructor = resolvedValue?.constructor?.name")

	w.Writeln("if (constructor === 'QueryBuilder') {")
	w.Indent()

	w.Writeln("span.addEvent('using QueryBuilder')")
	w.Writeln("builder = resolvedValue;")

	w.Writeln("// in order to populate data, we take the QueryBuilder instance and call the relevant 'terminating' method on it to execute the query")

	w.Writeln("span.addEvent(builder.sql())")
	switch action.Type {
	case proto.ActionType_ACTION_TYPE_LIST:
		w.Writeln("data = await builder.findMany();")
	case proto.ActionType_ACTION_TYPE_GET:
		w.Writeln("data = await builder.findOne();")
	case proto.ActionType_ACTION_TYPE_DELETE:
		w.Writeln("data = await builder.delete();")
	}
	w.Dedent()
	w.Writeln("} else {")
	w.Indent()

	w.Writeln("// in this case, the data is just the resolved value of the promise")
	w.Writeln("span.addEvent('using Model API')")

	w.Writeln("data = resolvedValue;")
	w.Dedent()
	w.Writeln("}")
	w.Dedent()
	w.Write("}")

}

func writeAfterQueryHook(w *codegen.Writer, action *proto.Action) {
	w.Writeln("// call afterQuery hook (if defined)")
	w.Writeln("if (hooks.afterQuery) {")
	w.Indent()
	wrapWithSpan(w, fmt.Sprintf("%s.afterQuery", action.Name), func(w *codegen.Writer) {
		w.Writeln("data = await hooks.afterQuery(ctx, deepFreeze(inputs), data);")
	})
	w.Dedent()
	w.Writeln("}")

	w.Writeln("")
}

func writeBeforeWriteHook(w *codegen.Writer, action *proto.Action) {
	w.Writeln("")

	w.Writeln("// call beforeWrite hook (if defined)")
	w.Writeln("if (hooks.beforeWrite) {")
	w.Indent()

	wrapWithSpan(w, fmt.Sprintf("%s.beforeWrite", action.Name), func(w *codegen.Writer) {
		w.Writeln("values = await hooks.beforeWrite(ctx, deepFreeze(inputs), values);")
	})
	w.Dedent()
	w.Writeln("}")

	w.Writeln("")
}

func writeAfterWriteHook(w *codegen.Writer, action *proto.Action) {
	w.Writeln("")
	w.Writeln("")

	w.Writeln("// call afterWrite hook (if defined)")
	w.Writeln("if (hooks.afterWrite) {")
	w.Indent()

	wrapWithSpan(w, fmt.Sprintf("%s.afterWrite", action.Name), func(w *codegen.Writer) {
		w.Writeln("await hooks.afterWrite(ctx, deepFreeze(inputs), data);")
	})

	w.Dedent()
	w.Writeln("}")

	w.Writeln("")
}

func wrapWithSpan(w *codegen.Writer, name string, fn func(w *codegen.Writer)) {
	w.Writef("await runtime.tracing.withSpan('%s', async (span) => {\n", name)
	w.Indent()

	fn(w)

	w.Dedent()

	w.Writeln("});")
}

func writeFunctionImplementation(w *codegen.Writer, schema *proto.Schema, action *proto.Action) {
	w.Writef("const %s = (hooks = {}) => {\n", casing.ToCamel(action.Name))
	w.Indent()
	w.Writeln("return async function(ctx, inputs) {")
	w.Indent()

	w.Write("return ")

	wrapWithSpan(w, fmt.Sprintf("%s.DefaultImplementation", action.Name), func(w *codegen.Writer) {
		w.Writeln("const models = createModelAPI();")

		switch {
		// update actions are a special case because they have both write and query hooks, e.g:
		// - beforeQuery / afterQuery
		// - beforeWrite / afterWrite
		case action.Type == proto.ActionType_ACTION_TYPE_UPDATE:
			w.Writeln("let values = Object.assign({}, inputs.values);")
			w.Writeln("let wheres = Object.assign({}, inputs.where);")

			writeBeforeWriteHook(w, action)

			w.Writeln("let data;")

			w.Writeln("if (hooks.beforeQuery) {")
			w.Indent()

			wrapWithSpan(w, fmt.Sprintf("%s.beforeQuery", action.Name), func(w *codegen.Writer) {
				w.Writeln("data = await hooks.beforeQuery(ctx, deepFreeze(inputs), values);")
			})

			w.Dedent()
			// the else covers cases were no beforeQuery hook was defined at all,
			// so therefore we want to build up the base query without any additional help
			// from the user-defined beforeQuery
			w.Writef("} else {\n")
			w.Indent()
			w.Writeln("// when no beforeQuery hook is defined, use the default implementation")
			w.Writef("data = await models.%s.update(wheres, values);\n", casing.ToLowerCamel(action.ModelName))
			w.Dedent()
			w.Writeln("}")
			w.Writeln("")

			writeAfterQueryHook(w, action)
			writeAfterWriteHook(w, action)

			w.Writeln("return data;")
		case proto.IsReadAction(action) || action.Type == proto.ActionType_ACTION_TYPE_DELETE:
			w.Writeln("let wheres = {")
			w.Indent()
			w.Writeln("...inputs.where,")
			w.Dedent()
			w.Writeln("};")
			w.Writeln("")

			if action.Type == proto.ActionType_ACTION_TYPE_DELETE {
				w.Writeln("wheres = inputs;")
			}

			w.Writeln("let data;")

			writeBeforeQueryHook(w, action)

			// the else covers cases were no beforeQuery hook was defined at all,
			// so therefore we want to build up the base query without any additional help
			// from the user-defined beforeQuery
			w.Writeln(" else {")
			w.Indent()
			w.Writeln("// when no beforeQuery hook is defined, use the default implementation")

			switch action.Type {
			case proto.ActionType_ACTION_TYPE_LIST:
				w.Writef("data = await models.%s.findMany(inputs);\n", casing.ToLowerCamel(action.ModelName))
			case proto.ActionType_ACTION_TYPE_GET:
				w.Writef("data = await models.%s.findOne(wheres);\n", casing.ToLowerCamel(action.ModelName))
			case proto.ActionType_ACTION_TYPE_DELETE:
				w.Writef("data = await models.%s.delete(wheres);\n", casing.ToLowerCamel(action.ModelName))
			}

			w.Dedent()

			w.Writeln("}")

			writeAfterQueryHook(w, action)

			w.Writeln("return data;")
		case proto.IsWriteAction(action):
			w.Writeln("let values = {")
			w.Indent()
			w.Writeln("...inputs,")
			w.Dedent()
			w.Writeln("};")

			writeBeforeWriteHook(w, action)

			w.Writeln("// values is the mutated version of inputs.values")
			switch action.Type {
			case proto.ActionType_ACTION_TYPE_CREATE:
				w.Writef("const data = await models.%s.create(values);", casing.ToLowerCamel(action.ModelName))
			case proto.ActionType_ACTION_TYPE_UPDATE:
				w.Writef("const data = await models.%s.update(inputs.where, values);", casing.ToLowerCamel(action.ModelName))
			}

			writeAfterWriteHook(w, action)

			w.Writeln("return data;")
		}
	})

	w.Dedent()
	w.Writeln("};")

	w.Dedent()
	w.Writeln("};")
}

func writeFunctionWrapperType(w *codegen.Writer, model *proto.Model, action *proto.Action) {
	// we use the 'declare' keyword to indicate to the typescript compiler that the function
	// has already been declared in the underlying vanilla javascript and therefore we are just
	// decorating existing js code with types.
	w.Writef("export declare function %s", casing.ToCamel(action.Name))

	switch {
	case proto.ActionIsArbitraryFunction(action):
		inputType := action.InputMessageName
		if inputType == parser.MessageFieldTypeAny {
			inputType = "any"
		}

		w.Writef("(fn: (ctx: ContextAPI, inputs: %s) => ", inputType)
		w.Write(toCustomFunctionReturnType(model, action, false))
		w.Write("): ")
		w.Write(toCustomFunctionReturnType(model, action, false))
		w.Writeln(";")
	default:
		w.Writef("(hooks?: %s) : void\n", fmt.Sprintf("%sHooks", casing.ToCamel(action.Name)))

		w.Writef("export type %sHooks = {\n", casing.ToCamel(action.Name))
		w.Indent()

		inputMessage := action.InputMessageName

		switch {
		// update actions support both query and write hooks, so it's a special case
		case action.Type == proto.ActionType_ACTION_TYPE_UPDATE:
			// due to complications with our query builder needing additional support for chained .update()
			// and the typescript types becoming really complicated, we only allow a Promise of T to be returned
			// for beforeQuery hooks for update actions at the moment.
			returnType := fmt.Sprintf("Promise<%s>", action.ModelName)

			// type will be ActionNameValues
			valuesType := fmt.Sprintf("%sValues", casing.ToCamel(action.Name))

			w.Writef(`
	/**
	* beforeQuery can be used to modify the existing query, or replace it entirely.
	* If the function is marked with the async keyword, then the expected return type is a %s.
	* If the function is non-async, then the expected return type is an instance of QueryBuilder.
	*/
`, returnType)
			// the signature for beforeQuery for update is slightly different
			// as we want to pass both the original inputs, and the version of values
			// mutated by the beforeWrite hook to the function
			w.Writef("beforeQuery?: (ctx: ContextAPI, inputs: %s, values: %s) => %s\n", inputMessage, valuesType, returnType)

			w.Write(`
	/**
	* afterQuery is useful for modifying the response data purely for the purposes of presentation, performing custom permission checks, or performing other side effects. 
	*/
`)
			// afterQuery returns a Promise<T> or T
			w.Writef("afterQuery?: (ctx: ContextAPI, inputs: %s, %s: %s) => Promise<%s>\n", inputMessage, casing.ToLowerCamel(action.ModelName), action.ModelName, action.ModelName)

			w.Write(`
	/**
	* The beforeWrite hook allows you to modify the values that will be written to the database.
	*/
	`)
			w.Writef("beforeWrite?: (ctx: ContextAPI, inputs: %s, values: %s) => Promise<%s>\n", inputMessage, valuesType, valuesType)

			w.Write(`
	/**
	* The afterWrite hook allows you to perform side effects after the record has been written to the database. Common use cases include creating other models, and performing custom permission checks.
	*/
	`)
			w.Writef("afterWrite?: (ctx: ContextAPI, inputs: %s, data: %s) => Promise<void>\n", inputMessage, action.ModelName)
		case action.Type == proto.ActionType_ACTION_TYPE_LIST || action.Type == proto.ActionType_ACTION_TYPE_GET || action.Type == proto.ActionType_ACTION_TYPE_DELETE:
			resolvedReturnType := "unknown"
			queryBuilderType := fmt.Sprintf("%sQueryBuilder", action.ModelName)

			switch action.Type {
			case proto.ActionType_ACTION_TYPE_GET:
				resolvedReturnType = action.ModelName
			case proto.ActionType_ACTION_TYPE_LIST:
				resolvedReturnType = fmt.Sprintf("%s[]", action.ModelName)
			case proto.ActionType_ACTION_TYPE_DELETE:
				// todo: we could support passing the whole deleted record
				// but we'd need to read that first from the database.

				resolvedReturnType = "string" // the id of the deleted record
			}

			// the return type for beforeQuery can either be a {Model}QueryBuilder
			// or a Promise<{Model}{[]}>
			returnType := fmt.Sprintf("%sQueryBuilder | Promise<%s>", action.ModelName, resolvedReturnType)
			w.Writef(`
	/**
	* beforeQuery can be used to modify the existing query, or replace it entirely.
	* If the function is marked with the async keyword, then the expected return type is a Promise<%s>.
	* If the function is non-async, then the expected return type is an instance of QueryBuilder.
	*/
`, resolvedReturnType)
			w.Writef("beforeQuery?: (ctx: ContextAPI, inputs: %s, query: %s) => %s\n", inputMessage, queryBuilderType, returnType)

			dataVariableName := ""
			switch action.Type {
			case proto.ActionType_ACTION_TYPE_GET:
				dataVariableName = "record"
			case proto.ActionType_ACTION_TYPE_LIST:
				dataVariableName = "records"
			case proto.ActionType_ACTION_TYPE_DELETE:
				dataVariableName = "deletedId"
			}

			w.Write(`
	/**
	* afterQuery is useful for modifying the response data purely for the purposes of presentation, performing custom permission checks, or performing other side effects. 
	*/
`)
			// afterQuery returns a Promise<T> or T
			w.Writef("afterQuery?: (ctx: ContextAPI, inputs: %s, %s: %s) => Promise<%s> | %s\n", inputMessage, dataVariableName, resolvedReturnType, resolvedReturnType, resolvedReturnType)

		case action.Type == proto.ActionType_ACTION_TYPE_CREATE || action.Type == proto.ActionType_ACTION_TYPE_UPDATE:
			valuesType := ""

			switch action.Type {
			case proto.ActionType_ACTION_TYPE_CREATE:
				valuesType = fmt.Sprintf("%sCreateValues", action.ModelName)
			case proto.ActionType_ACTION_TYPE_UPDATE:
				valuesType = fmt.Sprintf("Partial<%s>", action.ModelName)
			}

			w.Write(`
	/**
	* The beforeWrite hook allows you to modify the values that will be written to the database.
	*/
`)
			w.Writef("beforeWrite?: (ctx: ContextAPI, inputs: %s, values: %s) => Promise<%s>\n", inputMessage, valuesType, valuesType)

			w.Write(`
	/**
	* The afterWrite hook allows you to perform side effects after the record has been written to the database. Common use cases include creating other models, and performing custom permission checks.
	*/
	`)
			w.Writef("afterWrite?: (ctx: ContextAPI, inputs: %s, data: %s) => Promise<void>\n", inputMessage, action.ModelName)
		}

		w.Dedent()
		w.Writeln("}")
	}
}

func toCustomFunctionReturnType(model *proto.Model, op *proto.Action, isTestingPackage bool) string {
	returnType := "Promise<"
	sdkPrefix := ""
	if isTestingPackage {
		sdkPrefix = "sdk."
	}
	switch op.Type {
	case proto.ActionType_ACTION_TYPE_CREATE:
		returnType += sdkPrefix + model.Name
	case proto.ActionType_ACTION_TYPE_UPDATE:
		returnType += sdkPrefix + model.Name
	case proto.ActionType_ACTION_TYPE_GET:
		returnType += sdkPrefix + model.Name + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		returnType += sdkPrefix + model.Name + "[]"
	case proto.ActionType_ACTION_TYPE_DELETE:
		returnType += "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		isAny := op.ResponseMessageName == parser.MessageFieldTypeAny

		if isAny {
			returnType += "any"
		} else {
			returnType += op.ResponseMessageName
		}
	}
	returnType += ">"
	return returnType
}

func writeJobFunctionWrapperType(w *codegen.Writer, job *proto.Job) {
	w.Writef("export declare function %s", casing.ToCamel(job.Name))

	inputType := job.InputMessageName

	if inputType == "" {
		w.Write("(fn: (ctx: JobContextAPI) => Promise<void>): Promise<void>")
	} else {
		w.Writef("(fn: (ctx: JobContextAPI, inputs: %s) => Promise<void>): Promise<void>", inputType)
	}

	w.Writeln(";")
}

func writeSubscriberFunctionWrapperType(w *codegen.Writer, subscriber *proto.Subscriber) {
	w.Writef("export declare function %s", casing.ToCamel(subscriber.Name))
	w.Writef("(fn: (ctx: SubscriberContextAPI, event: %s) => Promise<void>): Promise<void>", subscriber.InputMessageName)
	w.Writeln(";")
}

func toActionReturnType(model *proto.Model, op *proto.Action) string {
	returnType := "Promise<"
	sdkPrefix := "sdk."

	switch op.Type {
	case proto.ActionType_ACTION_TYPE_CREATE:
		returnType += sdkPrefix + model.Name
	case proto.ActionType_ACTION_TYPE_UPDATE:
		returnType += sdkPrefix + model.Name
	case proto.ActionType_ACTION_TYPE_GET:
		returnType += sdkPrefix + model.Name + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		returnType += "{results: " + sdkPrefix + model.Name + "[], pageInfo: runtime.PageInfo}"
	case proto.ActionType_ACTION_TYPE_DELETE:
		// todo: create ID type
		returnType += "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		returnType += op.ResponseMessageName
	}

	returnType += ">"
	return returnType
}

func generateDevelopmentServer(schema *proto.Schema) codegen.GeneratedFiles {
	w := &codegen.Writer{}
	w.Writeln(`import { handleRequest, handleJob, handleSubscriber, tracing } from '@teamkeel/functions-runtime';`)
	w.Writeln(`import { createContextAPI, createJobContextAPI, createSubscriberContextAPI, permissionFns } from '@teamkeel/sdk';`)
	w.Writeln(`import { createServer } from "http";`)

	functions := []*proto.Action{}
	for _, model := range schema.Models {
		for _, action := range model.Actions {
			if action.Implementation != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}
			functions = append(functions, action)
			// namespace import to avoid naming clashes
			w.Writef(`import function_%s from "../functions/%s.ts"`, action.Name, action.Name)
			w.Writeln(";")
		}
	}

	for _, job := range schema.Jobs {
		name := strcase.ToLowerCamel(job.Name)
		// namespace import to avoid naming clashes
		w.Writef(`import job_%s from "../jobs/%s.ts"`, name, name)
		w.Writeln(";")
	}

	for _, subscriber := range schema.Subscribers {
		name := subscriber.Name
		// namespace import to avoid naming clashes
		w.Writef(`import subscriber_%s from "../subscribers/%s.ts"`, name, name)
		w.Writeln(";")
	}

	w.Writeln("const functions = {")
	w.Indent()
	for _, fn := range functions {
		w.Writef("%s: function_%s,", fn.Name, fn.Name)
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")

	w.Writeln("const jobs = {")
	w.Indent()
	for _, job := range schema.Jobs {
		name := strcase.ToLowerCamel(job.Name)
		w.Writef("%s: job_%s,", name, name)
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")

	w.Writeln("const subscribers = {")
	w.Indent()
	for _, subscriber := range schema.Subscribers {
		name := subscriber.Name
		w.Writef("%s: subscriber_%s,", name, name)
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")

	w.Writeln("const actionTypes = {")
	w.Indent()

	for _, fn := range functions {
		w.Writef("%s: \"%s\",\n", fn.Name, fn.Type.String())
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

		let rpcResponse = null;
		switch (json.type) {
		case "action":
			rpcResponse = await handleRequest(json, {
				functions,
				createContextAPI,
				actionTypes,
				permissionFns,
			});
			break;
		case "job":
			rpcResponse = await handleJob(json, {
				jobs,
				createJobContextAPI,
			});
			break;
		case "subscriber":
			rpcResponse = await handleSubscriber(json, {
				subscribers,
				createSubscriberContextAPI,
			});
			break;
		default:
			res.statusCode = 400;
			res.end();
		}
		
		res.statusCode = 200;
		res.setHeader('Content-Type', 'application/json');
		res.write(JSON.stringify(rpcResponse));
		res.end();
		return;
	}

	res.statusCode = 400;
	res.end();
};

tracing.init();

const server = createServer(listener);
const port = (process.env.PORT && parseInt(process.env.PORT, 10)) || 3001;
server.listen(port);`)

	return []*codegen.GeneratedFile{
		{
			Path:     ".build/server.js",
			Contents: w.String(),
		},
	}
}

func generateTestingPackage(schema *proto.Schema) codegen.GeneratedFiles {
	js := &codegen.Writer{}
	types := &codegen.Writer{}

	// The testing package uses ES modules as it only used in the context of running tests
	// with Vitest
	js.Writeln(`import sdk from "@teamkeel/sdk"`)
	js.Writeln("const { useDatabase, models } = sdk;")
	js.Writeln(`import { ActionExecutor, JobExecutor, SubscriberExecutor, sql } from "@teamkeel/testing-runtime";`)
	js.Writeln("")
	js.Writeln("export { models };")
	js.Writeln("export const actions = new ActionExecutor({});")
	js.Writeln("export const jobs = new JobExecutor({});")
	js.Writeln("export const subscribers = new SubscriberExecutor({});")
	js.Writeln("export async function resetDatabase() {")
	js.Indent()
	js.Writeln("const db = useDatabase();")
	js.Write("await sql`TRUNCATE TABLE ")
	tableNames := []string{}
	for _, model := range schema.Models {
		tableNames = append(tableNames, fmt.Sprintf("\"%s\"", casing.ToSnake(model.Name)))
	}
	js.Writef("%s CASCADE", strings.Join(tableNames, ","))
	js.Writeln("`.execute(db);")
	js.Dedent()
	js.Writeln("}")

	writeTestingTypes(types, schema)

	return codegen.GeneratedFiles{
		{
			Path:     "node_modules/@teamkeel/testing/index.mjs",
			Contents: js.String(),
		},
		{
			Path:     "node_modules/@teamkeel/testing/index.d.ts",
			Contents: types.String(),
		},
		{
			Path:     "node_modules/@teamkeel/testing/package.json",
			Contents: `{"name": "@teamkeel/testing", "type": "module", "exports": "./index.mjs"}`,
		},
	}
}

func generateTestingSetup() codegen.GeneratedFiles {
	return codegen.GeneratedFiles{
		{
			Path: ".build/vitest.config.mjs",
			Contents: `
import { defineConfig } from "vitest/config";

export default defineConfig({
	test: {
		setupFiles: [__dirname + "/vitest.setup"],
		testTimeout: 100000,
	},
});
			`,
		},
		{
			Path: ".build/vitest.setup.mjs",
			Contents: `
import { expect } from "vitest";
import { toHaveError, toHaveAuthorizationError, toHaveAuthenticationError } from "@teamkeel/testing-runtime";

expect.extend({
	toHaveError,
	toHaveAuthorizationError,
	toHaveAuthenticationError,
});
			`,
		},
	}
}

func writeTestingTypes(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln(`import * as sdk from "@teamkeel/sdk";`)
	w.Writeln(`import * as runtime from "@teamkeel/functions-runtime";`)

	// We need to import the testing-runtime package to get
	// the types for the extended vitest matchers e.g. expect(v).toHaveAuthorizationError()
	w.Writeln(`import "@teamkeel/testing-runtime";`)
	w.Writeln("")

	// For the testing package we need input and response types for all actions
	writeMessages(w, schema, true)

	w.Writeln("declare class ActionExecutor {")
	w.Indent()
	w.Writeln("withIdentity(identity: sdk.Identity): ActionExecutor;")
	w.Writeln("withAuthToken(token: string): ActionExecutor;")
	for _, model := range schema.Models {
		for _, action := range model.Actions {
			msg := proto.FindMessage(schema.Messages, action.InputMessageName)

			w.Writef("%s(i", action.Name)

			// Check that all of the top level fields in the matching message are optional
			// If so, then we can make it so you don't even need to specify the key
			// example, this allows for:
			// await actions.listActivePublishersWithActivePosts();
			// instead of:
			// const { results: publishers } =
			// await actions.listActivePublishersWithActivePosts({ where: {} });
			if lo.EveryBy(msg.Fields, func(f *proto.MessageField) bool {
				return f.Optional
			}) {
				w.Write("?")
			}

			w.Writef(`: %s): %s`, action.InputMessageName, toActionReturnType(model, action))
			w.Writeln(";")
		}
	}
	w.Dedent()
	w.Writeln("}")
	if len(schema.Jobs) > 0 {
		w.Writeln("type JobOptions = { scheduled?: boolean } | null")
		w.Writeln("declare class JobExecutor {")
		w.Indent()
		w.Writeln("withIdentity(identity: sdk.Identity): JobExecutor;")
		w.Writeln("withAuthToken(token: string): JobExecutor;")
		for _, job := range schema.Jobs {
			msg := proto.FindMessage(schema.Messages, job.InputMessageName)

			// Jobs can be without inputs
			if msg != nil {
				w.Writef("%s(i", strcase.ToLowerCamel(job.Name))

				if lo.EveryBy(msg.Fields, func(f *proto.MessageField) bool {
					return f.Optional
				}) {
					w.Write("?")
				}

				w.Writef(`: %s, o?: JobOptions): %s`, job.InputMessageName, "Promise<void>")
				w.Writeln(";")
			} else {
				w.Writef("%s(o?: JobOptions): Promise<void>", strcase.ToLowerCamel(job.Name))
				w.Writeln(";")
			}

		}
		w.Dedent()
		w.Writeln("}")
		w.Writeln("export declare const jobs: JobExecutor;")
	}

	if len(schema.Subscribers) > 0 {
		w.Writeln("declare class SubscriberExecutor {")
		w.Indent()
		for _, subscriber := range schema.Subscribers {
			msg := proto.FindMessage(schema.Messages, subscriber.InputMessageName)

			w.Writef("%s(i", subscriber.Name)

			if lo.EveryBy(msg.Fields, func(f *proto.MessageField) bool {
				return f.Optional
			}) {
				w.Write("?")
			}

			w.Writef(`: %s): %s`, subscriber.InputMessageName, "Promise<void>")
			w.Writeln(";")
		}
		w.Dedent()
		w.Writeln("}")
		w.Writeln("export declare const subscribers: SubscriberExecutor;")
	}

	w.Writeln("export declare const actions: ActionExecutor;")
	w.Writeln("export declare const models: sdk.ModelsAPI;")
	w.Writeln("export declare function resetDatabase(): Promise<void>;")
}

func toTypeScriptType(t *proto.TypeInfo, isTestingPackage bool) (ret string) {
	switch t.Type {
	case proto.Type_TYPE_ID:
		ret = "string"
	case proto.Type_TYPE_STRING:
		ret = "string"
	case proto.Type_TYPE_BOOL:
		ret = "boolean"
	case proto.Type_TYPE_INT:
		ret = "number"
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		ret = "Date"
	case proto.Type_TYPE_ENUM:
		ret = t.EnumName.Value
	case proto.Type_TYPE_MESSAGE:
		ret = t.MessageName.Value
	case proto.Type_TYPE_MODEL:
		// models are imported from the sdk
		if isTestingPackage {
			ret = fmt.Sprintf("sdk.%s", t.ModelName.Value)
		} else {
			ret = t.ModelName.Value
		}
	case proto.Type_TYPE_SORT_DIRECTION:
		if isTestingPackage {
			ret = "sdk.SortDirection"
		} else {
			ret = "SortDirection"
		}
	case proto.Type_TYPE_UNION:
		// Retrieve all the types that can satisfy this union field.
		messageNames := lo.Map(t.UnionNames, func(s *wrapperspb.StringValue, _ int) string {
			return s.Value
		})
		ret = fmt.Sprintf("(%s)", strings.Join(messageNames, " | "))
	default:
		ret = "any"
	}

	return ret
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

func tsDocComment(w *codegen.Writer, f func(w *codegen.Writer)) {
	w.Writeln("/**")
	f(w)
	w.Writeln("*/")
}

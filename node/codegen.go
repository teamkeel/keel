package node

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Generate generates and returns a list of objects that represent files to be written
// to a project. Calling .Write() on the result will cause those files be written to disk.
// This function should not interact with the file system so it can be used in a backend
// context.
func Generate(ctx context.Context, schema *proto.Schema, cfg *config.ProjectConfig) (codegen.GeneratedFiles, error) {
	files := generateSdkPackage(schema, cfg)
	files = append(files, generateTestingPackage(schema)...)
	files = append(files, generateTestingSetup()...)

	return files, nil
}

func generateSdkPackage(schema *proto.Schema, cfg *config.ProjectConfig) codegen.GeneratedFiles {
	sdk := &codegen.Writer{}
	sdk.Writeln(`import { sql, NoResultError } from "kysely"`)
	sdk.Writeln(`import * as runtime from "@teamkeel/functions-runtime"`)
	sdk.Writeln("")

	sdkTypes := &codegen.Writer{}
	sdkTypes.Writeln(`import { Kysely, Generated } from "kysely"`)
	sdkTypes.Writeln(`import * as runtime from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln(`import { Headers } from 'node-fetch'`)
	sdkTypes.Writeln(`export * from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln("")

	writePermissions(sdk, schema)
	writeMessages(sdkTypes, schema, false, false)

	for _, enum := range schema.GetEnums() {
		writeEnum(sdkTypes, enum)
		writeEnumWhereCondition(sdkTypes, enum)
		writeEnumObject(sdk, enum)
	}

	writeFunctionHookHelpers(sdk)
	writeFunctionHookTypes(sdkTypes)
	writeRouteFunctionTypes(sdkTypes)

	writeTableConfig(schema, sdk, schema.GetModels())
	writeAPIFactory(sdk, schema)

	sdk.Writeln("export * from '@teamkeel/functions-runtime';")
	sdk.Writeln("export { ErrorPresets as errors } from '@teamkeel/functions-runtime';")

	for _, model := range schema.GetModels() {
		writeTableInterface(sdkTypes, model)
		writeModelInterface(sdkTypes, model, false)
		writeCreateValuesType(sdkTypes, schema, model)
		writeUpdateValuesType(sdkTypes, model)
		writeWhereConditionsInterface(sdkTypes, model)
		writeFindManyParamsInterface(sdkTypes, model)
		writeUniqueConditionsInterface(sdkTypes, model)
		writeModelAPIDeclaration(sdkTypes, model)
		writeModelQueryBuilderDeclaration(sdkTypes, model)

		for _, action := range model.GetActions() {
			// if we have an auto action with embedded data, we need to write the custom response type
			if action.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_AUTO && len(action.GetResponseEmbeds()) > 0 {
				writeEmbeddedModelInterface(sdkTypes, schema, model, toResponseType(action.GetName()), action.GetResponseEmbeds())
				continue
			}

			// We now only care about custom functions for the SDK
			if action.GetImplementation() != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}

			// writes new types to the index.d.ts to annotate the underlying vanilla javascript
			// implementation of a function with nice types
			writeFunctionWrapperType(sdkTypes, schema, model, action)

			// if the action type is read or write, then the signature of the exported method just takes the function
			// defined by the user
			if action.IsArbitraryFunction() {
				sdk.Writef("export const %s = (fn) => fn;", casing.ToCamel(action.GetName()))
				sdk.Writeln("")
			} else {
				// writes the default implementation of a function. the user can specify hooks which can
				// override the behaviour of the default implementation
				writeFunctionImplementation(sdk, schema, action)
			}
		}
	}

	sdkTypes.Writeln("export declare function AfterAuthentication(fn: (ctx: ContextAPI) => Promise<void>): Promise<void>;")
	sdkTypes.Writeln("export declare function AfterIdentityCreated(fn: (ctx: ContextAPI) => Promise<void>): Promise<void>;")

	for _, job := range schema.GetJobs() {
		writeJobFunctionWrapperType(sdkTypes, job)
		sdk.Writef("export const %s = (fn) => fn;", job.GetName())
		sdk.Writeln("")
	}

	for _, subscriber := range schema.GetSubscribers() {
		writeSubscriberFunctionWrapperType(sdkTypes, subscriber)
		sdk.Writef("export const %s = (fn) => fn;", strcase.ToCamel(subscriber.GetName()))
		sdk.Writeln("")
	}

	for _, flow := range schema.GetAllFlows() {
		writeFlowFunctionWrapperType(sdkTypes, flow)
		sdk.Writef("export const %s = (config, fn) => { return { config, fn }; };", strcase.ToCamel(flow.GetName()))
		sdk.Writeln("")
	}

	if cfg != nil {
		for _, h := range cfg.Auth.EnabledHooks() {
			sdk.Writef("export const %s = (fn) => fn;", strcase.ToCamel(string(h)))
			sdk.Writeln("")
		}
	}

	writeDatabaseInterface(sdkTypes, schema)
	writeAPIDeclarations(sdkTypes, schema)

	return []*codegen.GeneratedFile{
		{
			Path:     ".build/sdk/index.js",
			Contents: sdk.String(),
		},
		{
			Path:     ".build/sdk/index.d.ts",
			Contents: sdkTypes.String(),
		},
		{
			Path:     ".build/sdk/package.json",
			Contents: `{"name": "@teamkeel/sdk"}`,
		},
	}
}

func writeRouteFunctionTypes(w *codegen.Writer) {
	w.Writeln(`export type RouteFunction = (req: RouteFunctionRequest, ctx: Omit<ContextAPI, "headers" | "response">) => Promise<RouteFunctionResponse>;`)

	w.Writeln("export type RouteFunctionRequest = {")
	w.Indent()
	w.Writeln("body: string;")
	w.Writeln("method: string;")
	w.Writeln("path: string;")
	w.Writeln("params: {[key: string]: string | undefined};")
	w.Writeln("query: string;")
	w.Writeln("headers: Headers;")
	w.Dedent()
	w.Writeln("}")

	w.Writeln("export type RouteFunctionResponse = {")
	w.Indent()
	w.Writeln("body: string;")
	w.Writeln("statusCode?: number;")
	w.Writeln("headers?: {[key: string]: string};")
	w.Dedent()
	w.Writeln("}")
}

func writeResultInfoInterface(w *codegen.Writer, schema *proto.Schema, action *proto.Action, isClientPackage bool) {
	facetFields := proto.FacetFields(schema, action)
	if len(facetFields) == 0 {
		return
	}

	w.Writef("export interface %sResultInfo {\n", strcase.ToCamel(action.GetName()))
	w.Indent()

	for _, field := range facetFields {
		switch field.GetType().GetType() {
		case proto.Type_TYPE_DECIMAL, proto.Type_TYPE_INT:
			w.Writef("%s: { min: number, max: number, avg: number };\n", field.GetName())
		case proto.Type_TYPE_ID, proto.Type_TYPE_ENUM, proto.Type_TYPE_STRING:
			w.Writef("%s: [ { value: string, count: number } ];\n", field.GetName())
		case proto.Type_TYPE_TIMESTAMP, proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME:
			w.Writef("%s: { min: Date, max: Date };\n", field.GetName())
		case proto.Type_TYPE_DURATION:
			if isClientPackage {
				w.Writef("%s: { min: DurationString, max: DurationString, avg: DurationString };\n", field.GetName())
			} else {
				w.Writef("%s: { min: runtime.Duration, max: runtime.Duration, avg: runtime.Duration };\n", field.GetName())
			}
		}
	}

	w.Dedent()
	w.Write("}")
	w.Writeln("")
}

func writeTableInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export interface %sTable {\n", model.GetName())
	w.Indent()
	for _, field := range model.GetFields() {
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			continue
		}

		w.Write(casing.ToLowerCamel(field.GetName()))
		w.Write(": ")
		t := toDbTableType(field.GetType(), false)

		if field.GetType().GetRepeated() {
			t = fmt.Sprintf("%s[]", t)
		}

		if field.GetDefaultValue() != nil || field.GetSequence() != nil {
			t = fmt.Sprintf("Generated<%s>", t)
		}

		w.Write(t)

		if field.GetOptional() {
			w.Write(" | null")
		}
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeModelInterface(w *codegen.Writer, model *proto.Model, isClientPackage bool) {
	w.Writef("export interface %s {\n", model.GetName())
	w.Indent()
	for _, field := range model.GetFields() {
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			continue
		}

		w.Write(field.GetName())
		w.Write(": ")
		t := toTypeScriptType(field.GetType(), false, false, isClientPackage)

		if field.GetType().GetRepeated() {
			t = fmt.Sprintf("%s[]", t)
		}

		w.Write(t)

		if field.GetOptional() {
			w.Write(" | null")
		}

		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeUpdateValuesType(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sUpdateValues = {\n", model.GetName())
	w.Indent()
	for _, field := range model.GetFields() {
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			continue
		}

		if field.GetComputedExpression() != nil || field.GetSequence() != nil {
			continue
		}

		w.Write(field.GetName())
		w.Write(": ")
		t := toTypeScriptType(field.GetType(), true, false, false)

		if field.GetType().GetRepeated() {
			t = fmt.Sprintf("%s[]", t)
		}

		w.Write(t)

		if field.GetOptional() {
			w.Write(" | null")
		}

		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeEmbeddedModelInterface(w *codegen.Writer, schema *proto.Schema, model *proto.Model, name string, embeddings []string) {
	w.Writef("export interface %s ", name)
	writeEmbeddedModelFields(w, schema, model, embeddings)
	w.Writeln("")
}

func writeEmbeddedModelFields(w *codegen.Writer, schema *proto.Schema, model *proto.Model, embeddings []string) {
	w.Write("{\n")
	w.Indent()
	for _, field := range model.GetFields() {
		// if the field is of ID type, and the related model is embedded, we do not want to include it in the schema
		if field.GetType().GetType() == proto.Type_TYPE_ID && field.GetForeignKeyInfo() != nil {
			relatedModel := strings.TrimSuffix(field.GetName(), "Id")
			skip := false
			for _, embed := range embeddings {
				frags := strings.Split(embed, ".")
				if frags[0] == relatedModel {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		fieldEmbeddings := []string{}
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			found := false

			for _, embed := range embeddings {
				frags := strings.Split(embed, ".")
				if frags[0] == field.GetName() {
					found = true
					// if we have to embed a child model for this field, we need to pass them through the field schema
					// with the first segment removed
					if len(frags) > 1 {
						fieldEmbeddings = append(fieldEmbeddings, strings.Join(frags[1:], "."))
					}
				}
			}
			if !found {
				continue
			}
		}

		w.Write(field.GetName())
		w.Write(": ")

		if len(fieldEmbeddings) == 0 {
			w.Write(toTypeScriptType(field.GetType(), false, false, false))
		} else {
			fieldModel := schema.FindModel(field.GetType().GetEntityName().GetValue())
			writeEmbeddedModelFields(w, schema, fieldModel, fieldEmbeddings)
		}

		if field.GetType().GetRepeated() {
			w.Write("[]")
		}
		if field.GetOptional() {
			w.Write(" | null")
		}

		w.Writeln("")
	}
	w.Dedent()
	w.Write("}")
}

func writeCreateValuesType(w *codegen.Writer, schema *proto.Schema, model *proto.Model) {
	w.Writef("export type %sCreateValues = {\n", model.GetName())
	w.Indent()

	for _, field := range model.GetFields() {
		// For required relationship fields we don't include them in the main type but instead
		// add them after using a union.
		if (field.GetForeignKeyFieldName() != nil || field.GetForeignKeyInfo() != nil) && !field.GetOptional() {
			continue
		}

		if field.GetComputedExpression() != nil || field.GetSequence() != nil {
			continue
		}

		if field.GetForeignKeyFieldName() != nil {
			w.Writef("// if providing a value for this field do not also set %s\n", field.GetForeignKeyFieldName().GetValue())
		}
		if field.GetForeignKeyInfo() != nil {
			w.Writef("// if providing a value for this field do not also set %s\n", strings.TrimSuffix(field.GetName(), "Id"))
		}

		w.Write(field.GetName())
		if field.GetOptional() || field.GetDefaultValue() != nil || field.IsHasMany() || field.GetComputedExpression() != nil {
			w.Write("?")
		}

		w.Write(": ")

		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			if field.IsHasMany() {
				w.Write("Array<")
			}

			relation := schema.FindEntity(field.GetType().GetEntityName().GetValue())

			// For a has-many we need to omit the fields that relate to _this_ model.
			// For example if we're making the create values type for author, and this
			// field is "books" then we don't want the create values type for each book
			// to expect you to provide "author" or "authorId" - as that field will be filled
			// in when the author record is created
			if field.IsHasMany() {
				inverseField := relation.FindField(field.GetInverseFieldName().GetValue())
				w.Writef("Omit<%sCreateValues, '%s' | '%s'>", relation.GetName(), inverseField.GetName(), inverseField.GetForeignKeyFieldName().GetValue())
			} else {
				w.Writef("%sCreateValues", relation.GetName())
			}

			// ...or just an id. This API might not be ideal because by allowing just
			// "id" we make the types less strict.
			w.Writef(" | {%s: string}", relation.PrimaryKeyFieldName())

			if field.IsHasMany() {
				w.Write(">")
			}
		} else {
			t := toTypeScriptType(field.GetType(), true, false, false)
			if field.GetType().GetRepeated() {
				t = fmt.Sprintf("%s[]", t)
			}

			w.Write(t)
		}

		if field.GetOptional() {
			w.Write(" | null")
		}
		w.Writeln("")
	}

	w.Dedent()
	w.Write("}")

	// For each required belongs-to relationship add a union that lets you either set
	// the generated foreign key field or the actual model field, but not both.
	for _, field := range model.GetFields() {
		if field.GetForeignKeyFieldName() == nil || field.GetOptional() {
			continue
		}

		if field.GetComputedExpression() != nil {
			continue
		}

		w.Writeln(" & (")
		w.Indent()

		fkName := field.GetForeignKeyFieldName().GetValue()

		relation := schema.FindModel(field.GetType().GetEntityName().GetValue())
		relationPk := relation.PrimaryKeyFieldName()

		w.Writef("// Either %s or %s can be provided but not both\n", field.GetName(), fkName)
		w.Writef("| {%s: %sCreateValues | {%s: string}, %s?: undefined}\n", field.GetName(), field.GetType().GetEntityName().GetValue(), relationPk, fkName)
		w.Writef("| {%s: string, %s?: undefined}\n", fkName, field.GetName())

		w.Dedent()
		w.Write(")")
	}

	w.Writeln("")
	w.Writeln("")
}

func writeFindManyParamsInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sOrderBy = {\n", model.GetName())
	w.Indent()

	relevantFields := lo.Filter(model.GetFields(), func(f *proto.Field, _ int) bool {
		if f.GetType().GetRepeated() {
			return false
		}

		switch f.GetType().GetType() {
		// scalar types are only permitted to sort by
		case proto.Type_TYPE_BOOL, proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_INT, proto.Type_TYPE_STRING, proto.Type_TYPE_ENUM, proto.Type_TYPE_TIMESTAMP, proto.Type_TYPE_ID, proto.Type_TYPE_DECIMAL:
			return true
		default:
			// includes types such as password, secret, model etc
			return false
		}
	})

	for i, f := range relevantFields {
		w.Writef("%s?: runtime.SortDirection", f.GetName())

		if i < len(relevantFields)-1 {
			w.Write(",")
		}

		w.Write("\n")
	}
	w.Dedent()
	w.Write("}")

	w.Writeln("\n")
	w.Writef("export interface %sFindManyParams {\n", model.GetName())
	w.Indent()
	w.Writef("where?: %sWhereConditions;\n", model.GetName())
	w.Writef("limit?: number;\n")
	w.Writef("offset?: number;\n")
	w.Writef("orderBy?: %sOrderBy;\n", model.GetName())
	w.Dedent()
	w.Writeln("}")
}

func writeWhereConditionsInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export interface %sWhereConditions {\n", model.GetName())
	w.Indent()
	for _, field := range model.GetFields() {
		if field.GetType().GetType() == proto.Type_TYPE_FILE {
			continue
		}

		w.Write(field.GetName())
		w.Write("?")
		w.Write(": ")
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			// Embed related models where conditions
			w.Writef("%sWhereConditions", field.GetType().GetEntityName().GetValue())
		} else {
			w.Write(toTypeScriptType(field.GetType(), false, false, false))

			if field.GetType().GetRepeated() {
				w.Write("[]")
			}

			w.Write(" | ")
			w.Write(toWhereConditionType(field))
		}

		if field.GetOptional() {
			w.Write(" | null")
		}
		w.Write(";")

		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeMessages(w *codegen.Writer, schema *proto.Schema, isTestingPackage bool, isClientPackage bool) {
	for _, msg := range schema.GetMessages() {
		if msg.GetName() == parser.MessageFieldTypeAny {
			continue
		}

		if schema.IsActionResponseMessage(msg.GetName()) {
			writeResponseMessage(w, msg, isTestingPackage, isClientPackage)
		} else {
			writeInputMessage(w, msg, isTestingPackage, isClientPackage)
		}
	}
}

func writeInputMessage(w *codegen.Writer, message *proto.Message, isTestingPackage bool, isClientPackage bool) {
	if message.GetType() != nil {
		w.Writef("export type %s = ", message.GetName())
		w.Write(toInputTypescriptType(message.GetType(), isTestingPackage, isClientPackage))
		w.Writeln(";")
		return
	}

	w.Writef("export interface %s {\n", message.GetName())
	w.Indent()

	for _, field := range message.GetFields() {
		w.Write(field.GetName())

		if field.GetOptional() {
			w.Write("?")
		}

		w.Write(": ")

		w.Write(toInputTypescriptType(field.GetType(), isTestingPackage, isClientPackage))

		if field.GetType().GetRepeated() {
			w.Write("[]")
		}

		if field.GetNullable() {
			w.Write(" | null")
		}

		w.Writeln(";")
	}

	w.Dedent()
	w.Writeln("}")
}

func writeResponseMessage(w *codegen.Writer, message *proto.Message, isTestingPackage bool, isClientPackage bool) {
	if message.GetType() != nil {
		w.Writef("export type %s = ", message.GetName())
		w.Write(toResponseTypescriptType(message.GetType(), isTestingPackage, isClientPackage))
		w.Writeln(";")
		return
	}

	w.Writef("export interface %s {\n", message.GetName())
	w.Indent()

	for _, field := range message.GetFields() {
		w.Write(field.GetName())

		if field.GetOptional() {
			w.Write("?")
		}

		w.Write(": ")

		w.Write(toResponseTypescriptType(field.GetType(), isTestingPackage, isClientPackage))

		if field.GetType().GetRepeated() {
			w.Write("[]")
		}

		if field.GetNullable() {
			w.Write(" | null")
		}

		w.Writeln(";")
	}

	w.Dedent()
	w.Writeln("}")
}

func writeUniqueConditionsInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sUniqueConditions = ", model.GetName())
	w.Indent()

	type F struct {
		key   string
		value string
	}

	seenCompountUnique := map[string]bool{}

	for _, f := range model.GetFields() {
		entries := []*F{}

		switch {
		case f.GetUnique() || f.GetPrimaryKey() || len(f.GetUniqueWith()) > 0:
			// Collect unique fields
			fields := []*proto.Field{f}
			fieldNames := []string{f.GetName()}
			for _, v := range f.GetUniqueWith() {
				u, _ := lo.Find(model.GetFields(), func(f *proto.Field) bool {
					return f.GetName() == v
				})
				fields = append(fields, u)
				fieldNames = append(fieldNames, u.GetName())
			}

			// De-dupe compound unqique constrains
			sort.Strings(fieldNames)
			k := strings.Join(fieldNames, ":")
			if _, ok := seenCompountUnique[k]; ok {
				continue
			}
			seenCompountUnique[k] = true

			for _, f := range fields {
				if f.GetType().GetType() == proto.Type_TYPE_ENTITY {
					if f.GetForeignKeyFieldName() == nil {
						// I'm not sure this can happen, but rather than have a cryptic nil-pointer error we'll
						// panic with a hopefully more helpful error
						panic(fmt.Sprintf(
							"%s.%s is a relation field and part of a unique constraint but does not have a foreign key - this is unsupported",
							model.GetName(), f.GetName(),
						))
					}

					entries = append(entries, &F{
						key:   f.GetForeignKeyFieldName().GetValue(),
						value: "string",
					})
				} else {
					entries = append(entries, &F{
						key:   f.GetName(),
						value: toTypeScriptType(f.GetType(), false, false, false),
					})
				}
			}
		case f.IsHasMany():
			// If a field is has-many then the other side is has-one, meaning
			// you can use that fields unique conditions to look up _this_ model.
			// Example: an author has many books, but a book has one author, which
			// means given a book id you can find a single author
			entries = append(entries, &F{
				key:   f.GetName(),
				value: fmt.Sprintf("%sUniqueConditions", f.GetType().GetEntityName().GetValue()),
			})
		}

		if len(entries) == 0 {
			continue
		}

		w.Writeln("")
		w.Write("| {")
		for i, f := range entries {
			if i > 0 {
				w.Write(", ")
			}
			w.Writef("%s: %s", f.key, f.value)
		}
		w.Writef("}")
	}

	w.Writeln(";")
	w.Dedent()
}

func writeModelAPIDeclaration(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sAPI = {\n", model.GetName())
	w.Indent()

	nonOptionalFields := lo.Filter(model.GetFields(), func(f *proto.Field, _ int) bool {
		return !f.GetOptional() && f.GetDefaultValue() == nil && f.GetComputedExpression() == nil
	})

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Create a %s record\n", model.GetName())
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const record = await models.%s.create({\n", casing.ToLowerCamel(model.GetName()))
		w.Indent()

		for i, f := range nonOptionalFields {
			w.Writef("%s: ", casing.ToLowerCamel(f.GetName()))

			if f.GetType().GetRepeated() {
				w.Write("[")
			}

			switch f.GetType().GetType() {
			case proto.Type_TYPE_STRING, proto.Type_TYPE_MARKDOWN:
				w.Write("''")
			case proto.Type_TYPE_BOOL:
				w.Write("false")
			case proto.Type_TYPE_INT, proto.Type_TYPE_DECIMAL:
				w.Write("0")
			case proto.Type_TYPE_DATETIME, proto.Type_TYPE_DATE, proto.Type_TYPE_TIMESTAMP:
				w.Write("new Date()")
			case proto.Type_TYPE_FILE:
				w.Write("inputs.profilePhoto")
			default:
				w.Write("undefined")
			}

			if f.GetType().GetRepeated() {
				w.Write("]")
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
	w.Writef("create(values: %sCreateValues): Promise<%s>;\n", model.GetName(), model.GetName())

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Update a %s record\n", model.GetName())
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const %s = await models.%s.update(", casing.ToLowerCamel(model.GetName()), casing.ToLowerCamel(model.GetName()))
		w.Writef("{ id: \"abc\" },")
		if len(nonOptionalFields) > 0 {
			w.Writef(" { %s: XXX }", casing.ToLowerCamel(nonOptionalFields[0].GetName()))
		} else {
			w.Write("  {}")
		}
		w.Writeln("});")
		w.Writeln("```")
	})
	w.Writef("update(where: %sUniqueConditions, values: Partial<%sUpdateValues>): Promise<%s>;\n", model.GetName(), model.GetName(), model.GetName())

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Deletes a %s record\n", model.GetName())
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const deletedId = await models.%s.delete({ id: 'xxx' });\n", casing.ToLowerCamel(model.GetName()))
		w.Writeln("```")
	})
	w.Writef("delete(where: %sUniqueConditions): Promise<string>;\n", model.GetName())

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Finds a single %s record\n", model.GetName())
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const %s = await models.%s.findOne({ id: 'xxx' });\n", casing.ToLowerCamel(model.GetName()), casing.ToLowerCamel(model.GetName()))
		w.Writeln("```")
	})
	w.Writef("findOne(where: %sUniqueConditions): Promise<%s | null>;\n", model.GetName(), model.GetName())
	tsDocComment(w, func(w *codegen.Writer) {
		w.Writef("* Finds multiple %s records\n", model.GetName())
		w.Writeln("* @example")
		w.Writeln("```typescript")

		// cant seem to get markdown in vscode method signature popover to render indentation
		// so we have to get it all on one line for the meantime
		w.Writef(`const %ss = await models.%s.findMany({ where: { createdAt: { after: new Date(2022, 1, 1) } }, orderBy: { id: 'asc' }, limit: 1000, offset: 50 });`, casing.ToLowerCamel(model.GetName()), casing.ToLowerCamel(model.GetName()))
		w.Writeln("")
		w.Writeln("```")
	})
	w.Writef("findMany(params?: %sFindManyParams | undefined): Promise<%s[]>;\n", model.GetName(), model.GetName())

	tsDocComment(w, func(w *codegen.Writer) {
		w.Writeln("* Creates a new query builder with the given conditions applied")
		w.Writeln("* @example")
		w.Writeln("```typescript")
		w.Writef("const records = await models.%s.where({ createdAt: { after: new Date(2020, 1, 1) } }).findMany();\n", casing.ToLowerCamel(model.GetName()))
		w.Writeln("```")
	})
	w.Writef("where(where: %sWhereConditions): %sQueryBuilder;\n", model.GetName(), model.GetName())
	w.Dedent()
	w.Writeln("}")
}

func writeModelQueryBuilderDeclaration(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sQueryBuilderParams = {\n", model.GetName())
	w.Indent()
	w.Writef("limit?: number;\n")
	w.Writef("offset?: number;\n")
	w.Writef("orderBy?: %sOrderBy;\n", model.GetName())
	w.Dedent()
	w.Writeln("}")

	// the following types are for the chained version of the model api
	// e.g await models.foo.where({ bar: 'bazz' }).update({ bar: 'boo' })
	w.Writef("export type %sQueryBuilder = {\n", model.GetName())
	w.Indent()
	w.Writef("where(where: %sWhereConditions): %sQueryBuilder;\n", model.GetName(), model.GetName())
	w.Writef("findMany(params?: %sQueryBuilderParams): Promise<%s[]>;\n", model.GetName(), model.GetName())
	w.Writef("findOne(params?: %sQueryBuilderParams): Promise<%s>;\n", model.GetName(), model.GetName())

	// todo: support these.
	// w.Writef("limit(limit: number) : %sQueryBuilder;\n", model.Name)
	// w.Writef("offset(offset: number) : %sQueryBuilder;\n", model.Name)
	// w.Writef("orderBy(conditions: %sOrderBy) : %sQueryBuilder;\n", model.Name, model.Name)

	w.Writef("delete() : Promise<string>;\n")
	w.Writef("update(values: Partial<%s>) : Promise<%s>;\n", model.GetName(), model.GetName())
	w.Dedent()
	w.Writeln("}")
}

func writeEnumObject(w *codegen.Writer, enum *proto.Enum) {
	w.Writef("export const %s = {\n", enum.GetName())
	w.Indent()
	for _, v := range enum.GetValues() {
		w.Write(v.GetName())
		w.Write(": ")
		w.Writef(`"%s"`, v.GetName())
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("};")
}

func writeEnum(w *codegen.Writer, enum *proto.Enum) {
	w.Writef("export enum %s {\n", enum.GetName())
	w.Indent()
	for _, v := range enum.GetValues() {
		w.Write(v.GetName())
		w.Write(" = ")
		w.Writef(`"%s"`, v.GetName())
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeEnumWhereCondition(w *codegen.Writer, enum *proto.Enum) {
	w.Writef("export interface %sWhereCondition {\n", enum.GetName())
	w.Indent()
	w.Write("equals?: ")
	w.Write(enum.GetName())
	w.Writeln(" | null;")
	w.Write("oneOf?: ")
	w.Write(enum.GetName())
	w.Write("[]")
	w.Writeln(" | null;")
	w.Dedent()
	w.Writeln("}")

	w.Writef("export interface %sArrayWhereCondition {\n", enum.GetName())
	w.Indent()
	w.Write("equals?: ")
	w.Write(enum.GetName())
	w.Writeln("[] | null;")
	w.Write("notEquals?: ")
	w.Write(enum.GetName())
	w.Writeln("[] | null;")
	w.Write("any?: ")
	w.Write(enum.GetName())
	w.Write("ArrayQueryWhereCondition")
	w.Writeln(" | null;")
	w.Write("all?: ")
	w.Write(enum.GetName())
	w.Write("ArrayQueryWhereCondition")
	w.Writeln(" | null;")
	w.Dedent()
	w.Writeln("}")

	w.Writef("export interface %sArrayQueryWhereCondition {\n", enum.GetName())
	w.Indent()
	w.Write("equals?: ")
	w.Write(enum.GetName())
	w.Writeln(" | null;")
	w.Write("notEquals?: ")
	w.Write(enum.GetName())
	w.Writeln(" | null;")
	w.Dedent()
	w.Writeln("}")
}

func writeDatabaseInterface(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln("interface database {")
	w.Indent()
	for _, model := range schema.GetModels() {
		w.Writef("%s: %sTable;", casing.ToSnake(model.GetName()), model.GetName())
		w.Writeln("")
	}
	for _, task := range schema.GetTasks() {
		w.Writef("%s: %sTable;", casing.ToSnake(task.GetName()), task.GetName())
		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
	w.Writeln("export declare function useDatabase(): Kysely<database>;")
}

func writeAPIDeclarations(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln("export type ModelsAPI = {")
	w.Indent()
	for _, model := range schema.GetModels() {
		w.Write(casing.ToLowerCamel(model.GetName()))
		w.Write(": ")
		w.Writef(`%sAPI`, model.GetName())
		w.Writeln(";")
	}
	w.Dedent()
	w.Writeln("}")
	w.Writeln("export declare const models: ModelsAPI;")

	w.Writeln("export declare const permissions: runtime.Permissions;")
	w.Writeln("export declare const errors: runtime.Errors;")

	w.Writeln("type Environment = {")

	w.Indent()

	for _, variable := range schema.GetEnvironmentVariables() {
		w.Writef("%s: string;\n", variable.GetName())
	}

	w.Dedent()
	w.Writeln("}")
	w.Writeln("type Secrets = {")

	w.Indent()

	for _, secret := range schema.GetSecrets() {
		w.Writef("%s: string;\n", secret.GetName())
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
	w.Writeln("const headers = new Headers(meta.headers);")
	w.Writeln("const response = { headers: responseHeaders }")
	w.Writeln("const now = () => { return new Date(); };")
	w.Writeln("const { identity } = meta;")
	w.Writeln("const isAuthenticated = identity != null;")
	w.Writeln("const env = {")
	w.Indent()

	for _, variable := range schema.GetEnvironmentVariables() {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.GetName(), variable.GetName())
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.GetSecrets() {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.GetName(), secret.GetName())
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

	for _, variable := range schema.GetEnvironmentVariables() {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.GetName(), variable.GetName())
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.GetSecrets() {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.GetName(), secret.GetName())
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("return { identity, env, now, secrets, isAuthenticated };")
	w.Dedent()
	w.Writeln("};")

	w.Writeln("function createFlowContextAPI({ meta }) {")
	w.Indent()
	w.Writeln("const now = () => { return new Date(); };")
	w.Writeln("const { identity } = meta;")
	w.Writeln("const env = {")
	w.Indent()

	for _, variable := range schema.GetEnvironmentVariables() {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.GetName(), variable.GetName())
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.GetSecrets() {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.GetName(), secret.GetName())
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("return { env, now, secrets, identity };")
	w.Dedent()
	w.Writeln("};")

	w.Writeln("function createSubscriberContextAPI({ meta }) {")
	w.Indent()
	w.Writeln("const now = () => { return new Date(); };")
	w.Writeln("const env = {")
	w.Indent()

	for _, variable := range schema.GetEnvironmentVariables() {
		// fetch the value of the env var from the process.env (will pull the value based on the current environment)
		// outputs "key: process.env["key"] || []"
		w.Writef("%s: process.env[\"%s\"] || \"\",\n", variable.GetName(), variable.GetName())
	}

	w.Dedent()
	w.Writeln("};")
	w.Writeln("const secrets = {")
	w.Indent()

	for _, secret := range schema.GetSecrets() {
		w.Writef("%s: meta.secrets.%s || \"\",\n", secret.GetName(), secret.GetName())
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
	for _, model := range schema.GetModels() {
		w.Write(casing.ToLowerCamel(model.GetName()))
		w.Write(": ")

		// The second positional argument to the model API used to be a default values function but
		// default values are now set in the database so this is no longer needed.
		// Passing a no-op function here for backwards compatibility with older versions of the
		// functions-runtime package.
		w.Writef(`new runtime.ModelAPI("%s", () => ({}), tableConfigMap)`, casing.ToSnake(model.GetName()))

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

	w.Writeln("export const models = createModelAPI();")
	w.Writeln("export const permissions = createPermissionApi();")
	w.Writeln("export { createContextAPI, createJobContextAPI, createSubscriberContextAPI, createFlowContextAPI };")
}

func writeTableConfig(schema *proto.Schema, w *codegen.Writer, models []*proto.Model) {
	w.Write("const tableConfigMap = ")

	// In case the words map and string over and over aren't clear enough
	// for you see the packages/functions-runtime/src/ModelAPI.js file for
	// docs on how this object is expected to be structured
	tableConfigMap := map[string]map[string]map[string]string{}

	for _, model := range models {
		for _, field := range model.GetFields() {
			if field.GetType().GetType() != proto.Type_TYPE_ENTITY {
				continue
			}

			relationshipConfig := map[string]string{
				"referencesTable": casing.ToSnake(field.GetType().GetEntityName().GetValue()),
				"foreignKey":      casing.ToSnake(schema.GetForeignKeyFieldName(field)),
			}

			switch {
			case field.IsHasOne():
				relationshipConfig["relationshipType"] = "hasOne"
			case field.IsHasMany():
				relationshipConfig["relationshipType"] = "hasMany"
			case field.IsBelongsTo():
				relationshipConfig["relationshipType"] = "belongsTo"
			}

			tableConfig, ok := tableConfigMap[casing.ToSnake(model.GetName())]
			if !ok {
				tableConfig = map[string]map[string]string{}
				tableConfigMap[casing.ToSnake(model.GetName())] = tableConfig
			}

			tableConfig[casing.ToSnake(field.GetName())] = relationshipConfig
		}
	}

	b, _ := json.MarshalIndent(tableConfigMap, "", "    ")
	w.Write(string(b))
	w.Writeln(";")
}

var (
	//go:embed templates/**/*
	templates embed.FS
)

func writeFunctionHookHelpers(w *codegen.Writer) {
	dir := "templates/functions"
	entries, _ := fs.ReadDir(templates, dir)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".js" {
			b, _ := fs.ReadFile(templates, filepath.Join(dir, entry.Name()))
			w.Writeln(string(b))
		}
	}
}

func writeFunctionHookTypes(w *codegen.Writer) {
	b, _ := fs.ReadFile(templates, "templates/functions/types.d.ts")
	w.Writeln(string(b))
}

func writeFunctionImplementation(w *codegen.Writer, schema *proto.Schema, action *proto.Action) {
	var whereMsg *proto.Message
	var valuesMsg *proto.Message

	wheres := []string{}
	values := []string{}

	if action.GetInputMessageName() != "" {
		msg := schema.FindMessage(action.GetInputMessageName())

		switch action.GetType() {
		case proto.ActionType_ACTION_TYPE_UPDATE, proto.ActionType_ACTION_TYPE_LIST:
			for _, f := range msg.GetFields() {
				if f.GetName() == "where" {
					whereMsg = schema.FindMessage(f.GetType().GetMessageName().GetValue())
				}
				if f.GetName() == "values" {
					valuesMsg = schema.FindMessage(f.GetType().GetMessageName().GetValue())
				}
			}
		case proto.ActionType_ACTION_TYPE_CREATE:
			whereMsg = nil
			valuesMsg = msg
		default:
			whereMsg = msg
		}
	}

	// Using getter method for Fields here as it is safe even if the variables are nil
	for _, f := range whereMsg.GetFields() {
		if isModelInput(schema, f) {
			wheres = append(wheres, fmt.Sprintf(`"%s"`, f.GetName()))
		}
	}
	for _, f := range valuesMsg.GetFields() {
		if isModelInput(schema, f) {
			values = append(values, fmt.Sprintf(`"%s"`, f.GetName()))
		}
	}

	functionName := map[proto.ActionType]string{
		proto.ActionType_ACTION_TYPE_GET:    "getFunction",
		proto.ActionType_ACTION_TYPE_UPDATE: "updateFunction",
		proto.ActionType_ACTION_TYPE_CREATE: "createFunction",
		proto.ActionType_ACTION_TYPE_LIST:   "listFunction",
		proto.ActionType_ACTION_TYPE_DELETE: "deleteFunction",
	}[action.GetType()]

	w.Writeln(fmt.Sprintf(
		"export const %s = %s({model: models.%s, whereInputs: [%s], valueInputs: [%s]})",
		casing.ToCamel(action.GetName()),
		functionName,
		casing.ToLowerCamel(action.GetModelName()),
		strings.Join(wheres, ", "),
		strings.Join(values, ", "),
	))
}

// isModelInput returs true if `field` targets a model field, either directly
// or via a child field).
func isModelInput(schema *proto.Schema, field *proto.MessageField) bool {
	if len(field.GetTarget()) > 0 {
		return true
	}
	if field.GetType().GetMessageName() == nil {
		return false
	}
	msg := schema.FindMessage(field.GetType().GetMessageName().GetValue())
	for _, f := range msg.GetFields() {
		if isModelInput(schema, f) {
			return true
		}
	}
	return false
}

func writeFunctionWrapperType(w *codegen.Writer, schema *proto.Schema, model *proto.Model, action *proto.Action) {
	// we use the 'declare' keyword to indicate to the typescript compiler that the function
	// has already been declared in the underlying vanilla javascript and therefore we are just
	// decorating existing js code with types.
	w.Writef("export declare const %s: runtime.FuncWithConfig<{", casing.ToCamel(action.GetName()))

	if action.IsArbitraryFunction() {
		switch action.GetInputMessageName() {
		case parser.MessageFieldTypeAny:
			w.Write("(fn: (ctx: ContextAPI, inputs: any) => ")
		case "":
			w.Write("(fn: (ctx: ContextAPI, inputs: never) => ")
		default:
			w.Writef("(fn: (ctx: ContextAPI, inputs: %s) => ", action.GetInputMessageName())
		}

		w.Write(toCustomFunctionReturnType(model, action, false))
		w.Write("): ")
		w.Write(toCustomFunctionReturnType(model, action, false))
		w.Writeln("}>;")
		return
	}

	hooksType := fmt.Sprintf("%sHooks", casing.ToCamel(action.GetName()))

	// TODO: void return type here is wrong. It should be the type of the function e.g. (ctx, inputs) => ReturnType
	w.Writef("(hooks?: %s): void}>\n", hooksType)

	w.Writef("export type %s = ", hooksType)

	modelName := action.GetModelName()
	queryBuilder := modelName + "QueryBuilder"

	var inputsType string
	if action.GetInputMessageName() == "" {
		inputsType = "never"
	} else {
		inputsType = schema.FindMessage(action.GetInputMessageName()).GetName()
	}

	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_GET:
		w.Writef("GetFunctionHooks<%s, %s, %s>", modelName, queryBuilder, inputsType)
	case proto.ActionType_ACTION_TYPE_LIST:
		w.Writef("ListFunctionHooks<%s, %s, %s>", modelName, queryBuilder, inputsType)
	case proto.ActionType_ACTION_TYPE_CREATE:
		var beforeWriteValues string
		if action.GetInputMessageName() != "" {
			msg := schema.FindMessage(action.GetInputMessageName())
			pickKeys := lo.FilterMap(msg.GetFields(), func(f *proto.MessageField, _ int) (string, bool) {
				return fmt.Sprintf("'%s'", f.GetName()), isModelInput(schema, f)
			})

			switch len(pickKeys) {
			case len(msg.GetFields()):
				// All inputs target model fields, this means the beforeWriteValues are exactly the same as the inputs
				beforeWriteValues = action.GetInputMessageName()
			case 0:
				// No inputs target model fields - need the "empty object" type
				// https://www.totaltypescript.com/the-empty-object-type-in-typescript
				beforeWriteValues = "Record<string, never>"
			default:
				// Some inputs target model fields - so create a new type by picking from inputs
				beforeWriteValues = fmt.Sprintf("Pick<%s, %s>", action.GetInputMessageName(), strings.Join(pickKeys, " | "))
			}
		} else {
			beforeWriteValues = "Record<string, never>"
		}

		w.Writef("CreateFunctionHooks<%s, %s, %s, %s, %sCreateValues>", modelName, queryBuilder, inputsType, beforeWriteValues, modelName)
	case proto.ActionType_ACTION_TYPE_UPDATE:
		w.Writef("UpdateFunctionHooks<%s, %s, %s, %sValues>", modelName, queryBuilder, inputsType, casing.ToCamel(action.GetName()))
	case proto.ActionType_ACTION_TYPE_DELETE:
		w.Writef("DeleteFunctionHooks<%s, %s, %s>", modelName, queryBuilder, inputsType)
	}

	w.Writeln(";")
	w.Writeln("")
}

func toCustomFunctionReturnType(model *proto.Model, op *proto.Action, isTestingPackage bool) string {
	returnType := "Promise<"
	sdkPrefix := ""
	if isTestingPackage {
		sdkPrefix = "sdk."
	}
	switch op.GetType() {
	case proto.ActionType_ACTION_TYPE_CREATE:
		returnType += sdkPrefix + model.GetName()
	case proto.ActionType_ACTION_TYPE_UPDATE:
		returnType += sdkPrefix + model.GetName()
	case proto.ActionType_ACTION_TYPE_GET:
		returnType += sdkPrefix + model.GetName() + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		returnType += sdkPrefix + model.GetName() + "[]"
	case proto.ActionType_ACTION_TYPE_DELETE:
		returnType += "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		isAny := op.GetResponseMessageName() == parser.MessageFieldTypeAny

		if isAny {
			returnType += "any"
		} else {
			returnType += op.GetResponseMessageName()
		}
	}
	returnType += "| Error>"
	return returnType
}

func writeJobFunctionWrapperType(w *codegen.Writer, job *proto.Job) {
	w.Writef("export declare const %s: runtime.FuncWithConfig<{", casing.ToCamel(job.GetName()))

	inputType := job.GetInputMessageName()

	if inputType == "" {
		w.Write("(fn: (ctx: JobContextAPI) => Promise<void>): Promise<void>")
	} else {
		w.Writef("(fn: (ctx: JobContextAPI, inputs: %s) => Promise<void>): Promise<void>", inputType)
	}

	w.Writeln("}>;")
}

func writeSubscriberFunctionWrapperType(w *codegen.Writer, subscriber *proto.Subscriber) {
	w.Writef("export declare const %s: runtime.FuncWithConfig<{", casing.ToCamel(subscriber.GetName()))
	w.Writef("(fn: (ctx: SubscriberContextAPI, event: %s) => Promise<void>): Promise<void>", subscriber.GetInputMessageName())
	w.Writeln("}>;")
}

func writeFlowFunctionWrapperType(w *codegen.Writer, flow *proto.Flow) {
	var inputsType string
	if flow.GetInputMessageName() == "" {
		inputsType = "never"
	} else {
		inputsType = flow.GetInputMessageName()
	}

	w.Writef("export declare const %s: { <const C extends runtime.FlowConfig>(config: C, fn: runtime.FlowFunction<C, Environment, Secrets, Identity, %s>) };", flow.GetName(), inputsType)

	w.Writeln("")
}

func toActionReturnType(model *proto.Model, action *proto.Action) string {
	returnType := "Promise<"
	sdkPrefix := "sdk."

	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_CREATE:
		returnType += sdkPrefix + model.GetName()
	case proto.ActionType_ACTION_TYPE_UPDATE:
		returnType += sdkPrefix + model.GetName()
	case proto.ActionType_ACTION_TYPE_GET:
		className := model.GetName()
		if len(action.GetResponseEmbeds()) > 0 {
			className = toResponseType(action.GetName())
		}
		returnType += sdkPrefix + className + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		className := model.GetName()
		if len(action.GetResponseEmbeds()) > 0 {
			className = toResponseType(action.GetName())
		}

		if len(action.GetFacets()) > 0 {
			returnType += "{results: " + sdkPrefix + className + "[], resultInfo: " + strcase.ToCamel(action.GetName()) + "ResultInfo, pageInfo: runtime.PageInfo}"
		} else {
			returnType += "{results: " + sdkPrefix + className + "[], pageInfo: runtime.PageInfo}"
		}
	case proto.ActionType_ACTION_TYPE_DELETE:
		// todo: create ID type
		returnType += "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		if action.GetResponseMessageName() == parser.MessageFieldTypeAny {
			returnType += "any"
		} else {
			returnType += action.GetResponseMessageName()
		}
	}

	returnType += ">"
	return returnType
}

func generateTestingPackage(schema *proto.Schema) codegen.GeneratedFiles {
	js := &codegen.Writer{}
	types := &codegen.Writer{}

	// The testing package uses ES modules as it only used in the context of running tests
	// with Vitest
	js.Writeln(`import { useDatabase, models } from "@teamkeel/sdk"`)
	js.Writeln(`import { ActionExecutor, JobExecutor, SubscriberExecutor, Flows, FlowExecutor, sql } from "@teamkeel/testing-runtime";`)
	js.Writeln("")
	js.Writeln("export { models };")
	js.Writeln("export const actions = new ActionExecutor({});")
	js.Writeln("export const jobs = new JobExecutor({});")
	js.Writeln("export const subscribers = new SubscriberExecutor({});")
	js.Writeln("export const flows = {")
	js.Indent()
	for _, flow := range schema.GetAllFlows() {
		js.Writef("%s: new FlowExecutor({ name: \"%s\" }),", casing.ToLowerCamel(flow.GetName()), flow.GetName())
		js.Writeln("")
	}
	js.Writeln("")
	js.Dedent()
	js.Writeln("};")
	js.Writeln("export async function resetDatabase() {")
	js.Indent()
	js.Writeln("const db = useDatabase();")
	js.Write("await sql`TRUNCATE TABLE ")
	tableNames := []string{"keel_audit", "keel_storage", `"keel"."flow_run"`}
	for _, model := range schema.GetModels() {
		tableNames = append(tableNames, fmt.Sprintf("\"%s\"", casing.ToSnake(model.GetName())))
	}
	js.Writef("%s CASCADE", strings.Join(tableNames, ","))
	js.Writeln("`.execute(db);")
	js.Dedent()
	js.Writeln("}")

	writeTestingTypes(types, schema)

	return codegen.GeneratedFiles{
		{
			Path:     ".build/testing/index.mjs",
			Contents: js.String(),
		},
		{
			Path:     ".build/testing/index.d.ts",
			Contents: types.String(),
		},
		{
			Path:     ".build/testing/package.json",
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
import tsconfigPaths from 'vite-tsconfig-paths'
import * as path from 'path';

export default defineConfig({
	plugins: [
        tsconfigPaths(),
    ],
	test: {
		setupFiles: [__dirname + "/vitest.setup"],
		testTimeout: 100000,
	},
	resolve: {
		// on top of the "paths" entry in the project's tsconfig file that aliases the @teamkeel/sdk and @teamkeel/testing
		// imports so that they actually exist in the .build directory underneath the hood, for vitest we need to also add
		// the below alias section which enables vitest to pickup the same paths configuration for the code generated
		// npm modules. This is necessary because vitest isn't aware of the 'paths' configuration in typescript world at all
		alias: {
			// the __dirname below is relative to the .build directory which contains the sdk and testing directories containing
			// the codegenned sdk and testing packages.
			'@teamkeel/testing': path.resolve(__dirname, './testing'),
			'@teamkeel/sdk': path.resolve(__dirname, './sdk')
		}
	}
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
	w.Writeln(`/// <reference path="@teamkeel/functions-runtime/index.d.ts" /`)

	w.Writeln(`import * as sdk from "@teamkeel/sdk";`)
	w.Writeln(`import * as runtime from "@teamkeel/functions-runtime";`)

	// We need to import the testing-runtime package to get
	// the types for the extended vitest matchers e.g. expect(v).toHaveAuthorizationError()
	w.Writeln(`import "@teamkeel/testing-runtime";`)
	w.Writeln(`import { FlowRun, FlowExecutor } from "@teamkeel/testing-runtime";`)
	w.Writeln("")

	// For the testing package we need input and response types for all actions
	writeMessages(w, schema, true, false)

	w.Writeln("declare class ActionExecutor {")
	w.Indent()
	w.Writeln("withIdentity(identity: sdk.Identity): ActionExecutor;")
	w.Writeln("withAuthToken(token: string): ActionExecutor;")
	w.Writeln("withTimezone(timezone: string): this;")
	for _, model := range schema.GetModels() {
		for _, action := range model.GetActions() {
			args := ""
			if action.GetInputMessageName() != "" {
				msg := schema.FindMessage(action.GetInputMessageName())

				args = "i"

				// Check that all of the top level fields in the matching message are optional
				// If so, then we can make it so you don't even need to specify the key
				// example, this allows for:
				// await actions.listActivePublishersWithActivePosts();
				// instead of:
				// const { results: publishers } =
				// await actions.listActivePublishersWithActivePosts({ where: {} });
				if lo.EveryBy(msg.GetFields(), func(f *proto.MessageField) bool {
					return f.GetOptional()
				}) {
					args += "?"
				}

				argType := action.GetInputMessageName()
				if argType == parser.MessageFieldTypeAny {
					argType = "any"
				}

				args += fmt.Sprintf(": %s", argType)
			}

			w.Writef("%s(%s): %s", action.GetName(), args, toActionReturnType(model, action))
			w.Writeln(";")
		}
	}

	w.Dedent()
	w.Writeln("}")

	if len(schema.GetJobs()) > 0 {
		w.Writeln("type JobOptions = { scheduled?: boolean } | null")
		w.Writeln("declare class JobExecutor {")
		w.Indent()
		w.Writeln("withIdentity(identity: sdk.Identity): JobExecutor;")
		w.Writeln("withAuthToken(token: string): JobExecutor;")
		for _, job := range schema.GetJobs() {
			msg := schema.FindMessage(job.GetInputMessageName())

			// Jobs can be without inputs
			if msg != nil {
				w.Writef("%s(i", strcase.ToLowerCamel(job.GetName()))

				if lo.EveryBy(msg.GetFields(), func(f *proto.MessageField) bool {
					return f.GetOptional()
				}) {
					w.Write("?")
				}

				w.Writef(`: %s, o?: JobOptions): %s`, job.GetInputMessageName(), "Promise<void>")
				w.Writeln(";")
			} else {
				w.Writef("%s(o?: JobOptions): Promise<void>", strcase.ToLowerCamel(job.GetName()))
				w.Writeln(";")
			}
		}
		w.Dedent()
		w.Writeln("}")
		w.Writeln("export declare const jobs: JobExecutor;")
	}

	if len(schema.GetSubscribers()) > 0 {
		w.Writeln("declare class SubscriberExecutor {")
		w.Indent()
		for _, subscriber := range schema.GetSubscribers() {
			msg := schema.FindMessage(subscriber.GetInputMessageName())

			w.Writef("%s(e", subscriber.GetName())

			if msg.GetType().GetType() != proto.Type_TYPE_UNION && lo.EveryBy(msg.GetFields(), func(f *proto.MessageField) bool {
				return f.GetOptional()
			}) {
				w.Write("?")
			}

			w.Writef(`: %s): %s`, subscriber.GetInputMessageName(), "Promise<void>")
			w.Writeln(";")
		}
		w.Dedent()
		w.Writeln("}")
		w.Writeln("export declare const subscribers: SubscriberExecutor;")
	}

	w.Writeln("export type Flows = {")
	w.Indent()
	for _, flow := range schema.GetAllFlows() {
		input := flow.GetInputMessageName()
		if input == "" {
			w.Writef("%s: FlowExecutor<{}>;", casing.ToLowerCamel(flow.GetName()))
		} else {
			w.Writef("%s: FlowExecutor<%s>;", casing.ToLowerCamel(flow.GetName()), input)
		}

		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")

	for _, model := range schema.GetModels() {
		for _, action := range model.GetActions() {
			if action.GetType() == proto.ActionType_ACTION_TYPE_LIST {
				writeResultInfoInterface(w, schema, action, false)
			}
		}
	}

	w.Writeln("export declare const actions: ActionExecutor;")
	w.Writeln("export declare const flows: Flows;")
	w.Writeln("export declare const models: sdk.ModelsAPI;")
	w.Writeln("export declare function resetDatabase(): Promise<void>;")
}

func toDbTableType(t *proto.TypeInfo, isTestingPackage bool) (ret string) {
	switch t.GetType() {
	case proto.Type_TYPE_FILE:
		return "runtime.FileDbRecord"
	default:
		return toTypeScriptType(t, false, isTestingPackage, false)
	}
}

func toInputTypescriptType(t *proto.TypeInfo, isTestingPackage bool, isClientPackage bool) (ret string) {
	switch t.GetType() {
	case proto.Type_TYPE_DURATION:
		if isClientPackage {
			return "DurationString"
		} else {
			return "runtime.Duration"
		}
	case proto.Type_TYPE_RELATIVE_PERIOD:
		return "RelativeDateString"
	case proto.Type_TYPE_FILE:
		if isClientPackage {
			return "string"
		} else {
			if isTestingPackage {
				return "runtime.FileWriteTypes"
			} else {
				return "runtime.File"
			}
		}
	default:
		return toTypeScriptType(t, false, isTestingPackage, isClientPackage)
	}
}

func toResponseTypescriptType(t *proto.TypeInfo, isTestingPackage bool, isClientPackage bool) (ret string) {
	switch t.GetType() {
	case proto.Type_TYPE_RELATIVE_PERIOD:
		return "RelativeDateString"
	case proto.Type_TYPE_FILE:
		if isClientPackage {
			return "FileResponseObject"
		} else {
			return "runtime.File"
		}
	default:
		return toTypeScriptType(t, false, isTestingPackage, isClientPackage)
	}
}

func toTypeScriptType(t *proto.TypeInfo, includeCompatibleTypes bool, isTestingPackage bool, isClientPackage bool) (ret string) {
	switch t.GetType() {
	case proto.Type_TYPE_ID:
		ret = "string"
	case proto.Type_TYPE_STRING, proto.Type_TYPE_MARKDOWN:
		ret = "string"
	case proto.Type_TYPE_BOOL:
		ret = "boolean"
	case proto.Type_TYPE_INT, proto.Type_TYPE_DECIMAL:
		ret = "number"
	case proto.Type_TYPE_VECTOR:
		ret = "number[]"
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		ret = "Date"
	case proto.Type_TYPE_DURATION:
		if isClientPackage {
			ret = "DurationString"
		} else {
			ret = "runtime.Duration"
		}
	case proto.Type_TYPE_ENUM:
		if isTestingPackage {
			ret = "sdk." + t.GetEnumName().GetValue()
		} else {
			ret = t.GetEnumName().GetValue()
		}
	case proto.Type_TYPE_MESSAGE:
		ret = t.GetMessageName().GetValue()
	case proto.Type_TYPE_ENTITY:
		// models are imported from the sdk
		if isTestingPackage {
			ret = fmt.Sprintf("sdk.%s", t.GetEntityName().GetValue())
		} else {
			ret = t.GetEntityName().GetValue()
		}
	case proto.Type_TYPE_SORT_DIRECTION:
		if isClientPackage {
			ret = "SortDirection"
		} else {
			ret = "runtime.SortDirection"
		}
	case proto.Type_TYPE_UNION:
		// Retrieve all the types that can satisfy this union field.
		messageNames := lo.Map(t.GetUnionNames(), func(s *wrapperspb.StringValue, _ int) string {
			return s.GetValue()
		})
		ret = fmt.Sprintf("(%s)", strings.Join(messageNames, " | "))
	case proto.Type_TYPE_STRING_LITERAL:
		// Use string literal type for discriminating.
		ret = fmt.Sprintf(`"%s"`, t.GetStringLiteralValue().GetValue())

	case proto.Type_TYPE_FILE:
		if isClientPackage {
			ret = "FileResponseObject"
		} else {
			if includeCompatibleTypes {
				ret = "runtime.FileWriteTypes"
			} else {
				ret = "runtime.File"
			}
		}
	default:
		ret = "any"
	}

	return ret
}

func toWhereConditionType(f *proto.Field) string {
	if f.GetType().GetRepeated() {
		switch f.GetType().GetType() {
		case proto.Type_TYPE_ID, proto.Type_TYPE_STRING, proto.Type_TYPE_MARKDOWN:
			return "runtime.StringArrayWhereCondition"
		case proto.Type_TYPE_BOOL:
			return "runtime.BooleanArrayWhereCondition"
		case proto.Type_TYPE_INT:
			return "runtime.NumberArrayWhereCondition"
		case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			return "runtime.DateArrayWhereCondition"
		case proto.Type_TYPE_ENUM:
			return fmt.Sprintf("%sArrayWhereCondition", f.GetType().GetEnumName().GetValue())
		}
	}

	switch f.GetType().GetType() {
	case proto.Type_TYPE_ID:
		return "runtime.IDWhereCondition"
	case proto.Type_TYPE_STRING, proto.Type_TYPE_MARKDOWN:
		return "runtime.StringWhereCondition"
	case proto.Type_TYPE_BOOL:
		return "runtime.BooleanWhereCondition"
	case proto.Type_TYPE_INT, proto.Type_TYPE_DECIMAL:
		return "runtime.NumberWhereCondition"
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		return "runtime.DateWhereCondition"
	case proto.Type_TYPE_DURATION:
		return "runtime.DurationWhereCondition"
	case proto.Type_TYPE_ENUM:
		return fmt.Sprintf("%sWhereCondition", f.GetType().GetEnumName().GetValue())
	default:
		return "any"
	}
}

func tsDocComment(w *codegen.Writer, f func(w *codegen.Writer)) {
	w.Writeln("/**")
	f(w)
	w.Writeln("*/")
}

// toResponseType generates a response type name for the given action name. This is to be used for actions that contain
// embedded data.
func toResponseType(actionName string) string {
	return casing.ToCamel(actionName) + "Response"
}

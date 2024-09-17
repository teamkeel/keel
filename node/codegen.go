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
func Generate(ctx context.Context, schema *proto.Schema, cfg *config.ProjectConfig, opts ...func(o *generateOptions)) (codegen.GeneratedFiles, error) {
	options := &generateOptions{}
	for _, o := range opts {
		o(options)
	}

	files := generateSdkPackage(schema, cfg)
	files = append(files, generateTestingPackage(schema)...)
	files = append(files, generateTestingSetup()...)

	if options.developmentServer {
		files = append(files, generateDevelopmentServer(schema, cfg)...)
	}

	return files, nil
}

func generateSdkPackage(schema *proto.Schema, cfg *config.ProjectConfig) codegen.GeneratedFiles {
	sdk := &codegen.Writer{}
	sdk.Writeln(`const { sql, NoResultError } = require("kysely")`)
	sdk.Writeln(`const runtime = require("@teamkeel/functions-runtime")`)
	sdk.Writeln("")

	sdkTypes := &codegen.Writer{}
	sdkTypes.Writeln(`import { Kysely, Generated } from "kysely"`)
	sdkTypes.Writeln(`import * as runtime from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln(`import { Headers } from 'node-fetch'`)
	sdkTypes.Writeln(`export { InlineFile, File } from "@teamkeel/functions-runtime"`)
	sdkTypes.Writeln("")

	writePermissions(sdk, schema)
	writeMessages(sdkTypes, schema, false, false)

	for _, enum := range schema.Enums {
		writeEnum(sdkTypes, enum)
		writeEnumWhereCondition(sdkTypes, enum)
		writeEnumObject(sdk, enum)
	}

	writeFunctionHookHelpers(sdk)
	writeFunctionHookTypes(sdkTypes)

	writeTableConfig(sdk, schema.Models)
	writeAPIFactory(sdk, schema)

	sdk.Writeln("module.exports.useDatabase = runtime.useDatabase;")
	sdk.Writeln("module.exports.errors = runtime.ErrorPresets;")

	for _, model := range schema.Models {
		writeTableInterface(sdkTypes, model)
		writeModelInterface(sdkTypes, model, false)
		writeCreateValuesType(sdkTypes, schema, model)
		writeUpdateValuesType(sdkTypes, model)
		writeWhereConditionsInterface(sdkTypes, model)
		writeFindManyParamsInterface(sdkTypes, model)
		writeUniqueConditionsInterface(sdkTypes, model)
		writeModelAPIDeclaration(sdkTypes, model)
		writeModelQueryBuilderDeclaration(sdkTypes, model)

		for _, action := range model.Actions {
			// if we have an auto action with embedded data, we need to write the custom response type
			if action.Implementation == proto.ActionImplementation_ACTION_IMPLEMENTATION_AUTO && len(action.GetResponseEmbeds()) > 0 {
				writeEmbeddedModelInterface(sdkTypes, schema, model, toResponseType(action.Name), action.GetResponseEmbeds())
				continue
			}

			// We now only care about custom functions for the SDK
			if action.Implementation != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}

			// writes new types to the index.d.ts to annotate the underlying vanilla javascript
			// implementation of a function with nice types
			writeFunctionWrapperType(sdkTypes, schema, model, action)

			// if the action type is read or write, then the signature of the exported method just takes the function
			// defined by the user
			if action.IsArbitraryFunction() {
				sdk.Writef("module.exports.%s = (fn) => fn;", casing.ToCamel(action.Name))
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
	sdk.Writeln("")

	if cfg != nil {
		for _, h := range cfg.Auth.EnabledHooks() {
			sdk.Writef("module.exports.%s = (fn) => fn;", strcase.ToCamel(string(h)))
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

func writeTableInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export interface %sTable {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			continue
		}

		w.Write(casing.ToLowerCamel(field.Name))
		w.Write(": ")
		t := toDbTableType(field.Type, false)

		if field.Type.Repeated {
			t = fmt.Sprintf("%s[]", t)
		}

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

func writeModelInterface(w *codegen.Writer, model *proto.Model, isClientPackage bool) {
	w.Writef("export interface %s {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			continue
		}

		w.Write(field.Name)
		w.Write(": ")
		t := toTypeScriptType(field.Type, false, false, isClientPackage)

		if field.Type.Repeated {
			t = fmt.Sprintf("%s[]", t)
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

func writeUpdateValuesType(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sUpdateValues = {\n", model.Name)
	w.Indent()
	for _, field := range model.Fields {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			continue
		}

		w.Write(field.Name)
		w.Write(": ")
		t := toTypeScriptType(field.Type, true, false, false)

		if field.Type.Repeated {
			t = fmt.Sprintf("%s[]", t)
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

func writeEmbeddedModelInterface(w *codegen.Writer, schema *proto.Schema, model *proto.Model, name string, embeddings []string) {
	w.Writef("export interface %s ", name)
	writeEmbeddedModelFields(w, schema, model, embeddings)
	w.Writeln("")
}

func writeEmbeddedModelFields(w *codegen.Writer, schema *proto.Schema, model *proto.Model, embeddings []string) {
	w.Write("{\n")
	w.Indent()
	for _, field := range model.Fields {
		// if the field is of ID type, and the related model is embedded, we do not want to include it in the schema
		if field.Type.Type == proto.Type_TYPE_ID && field.ForeignKeyInfo != nil {
			relatedModel := strings.TrimSuffix(field.Name, "Id")
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
		if field.Type.Type == proto.Type_TYPE_MODEL {
			found := false

			for _, embed := range embeddings {
				frags := strings.Split(embed, ".")
				if frags[0] == field.Name {
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

		w.Write(field.Name)
		w.Write(": ")

		if len(fieldEmbeddings) == 0 {
			w.Write(toTypeScriptType(field.Type, false, false, false))
		} else {
			fieldModel := schema.FindModel(field.Type.ModelName.Value)
			writeEmbeddedModelFields(w, schema, fieldModel, fieldEmbeddings)
		}

		if field.Type.Repeated {
			w.Write("[]")
		}
		if field.Optional {
			w.Write(" | null")
		}

		w.Writeln("")
	}
	w.Dedent()
	w.Write("}")
}

func writeCreateValuesType(w *codegen.Writer, schema *proto.Schema, model *proto.Model) {
	w.Writef("export type %sCreateValues = {\n", model.Name)
	w.Indent()

	for _, field := range model.Fields {
		// For required relationship fields we don't include them in the main type but instead
		// add them after using a union.
		if (field.ForeignKeyFieldName != nil || field.ForeignKeyInfo != nil) && !field.Optional {
			continue
		}

		if field.ForeignKeyFieldName != nil {
			w.Writef("// if providing a value for this field do not also set %s\n", field.ForeignKeyFieldName.Value)
		}
		if field.ForeignKeyInfo != nil {
			w.Writef("// if providing a value for this field do not also set %s\n", strings.TrimSuffix(field.Name, "Id"))
		}

		w.Write(field.Name)
		if field.Optional || field.DefaultValue != nil || field.IsHasMany() {
			w.Write("?")
		}

		w.Write(": ")

		if field.Type.Type == proto.Type_TYPE_MODEL {
			if field.IsHasMany() {
				w.Write("Array<")
			}

			relation := schema.FindModel(field.Type.ModelName.Value)

			// For a has-many we need to omit the fields that relate to _this_ model.
			// For example if we're making the create values type for author, and this
			// field is "books" then we don't want the create values type for each book
			// to expect you to provide "author" or "authorId" - as that field will be filled
			// in when the author record is created
			if field.IsHasMany() {
				inverseField := proto.FindField(schema.Models, relation.Name, field.InverseFieldName.Value)
				w.Writef("Omit<%sCreateValues, '%s' | '%s'>", relation.Name, inverseField.Name, inverseField.ForeignKeyFieldName.Value)
			} else {
				w.Writef("%sCreateValues", relation.Name)
			}

			// ...or just an id. This API might not be ideal because by allowing just
			// "id" we make the types less strict.
			w.Writef(" | {%s: string}", relation.PrimaryKeyFieldName())

			if field.IsHasMany() {
				w.Write(">")
			}
		} else {
			t := toTypeScriptType(field.Type, true, false, false)

			if field.Type.Repeated {
				t = fmt.Sprintf("%s[]", t)
			}

			w.Write(t)
		}

		if field.Optional {
			w.Write(" | null")
		}
		w.Writeln("")
	}

	w.Dedent()
	w.Write("}")

	// For each required belongs-to relationship add a union that lets you either set
	// the generated foreign key field or the actual model field, but not both.
	for _, field := range model.Fields {
		if field.ForeignKeyFieldName == nil || field.Optional {
			continue
		}

		w.Writeln(" & (")
		w.Indent()

		fkName := field.ForeignKeyFieldName.Value

		relation := schema.FindModel(field.Type.ModelName.Value)
		relationPk := relation.PrimaryKeyFieldName()

		w.Writef("// Either %s or %s can be provided but not both\n", field.Name, fkName)
		w.Writef("| {%s: %sCreateValues | {%s: string}, %s?: undefined}\n", field.Name, field.Type.ModelName.Value, relationPk, fkName)
		w.Writef("| {%s: string, %s?: undefined}\n", fkName, field.Name)

		w.Dedent()
		w.Write(")")
	}

	w.Writeln("")
	w.Writeln("")
}

func writeFindManyParamsInterface(w *codegen.Writer, model *proto.Model) {
	w.Writef("export type %sOrderBy = {\n", model.Name)
	w.Indent()

	relevantFields := lo.Filter(model.Fields, func(f *proto.Field, _ int) bool {
		if f.Type.Repeated {
			return false
		}

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
		w.Writef("%s?: runtime.SortDirection", f.Name)

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
		if field.Type.Type == proto.Type_TYPE_FILE {
			continue
		}

		w.Write(field.Name)
		w.Write("?")
		w.Write(": ")
		if field.Type.Type == proto.Type_TYPE_MODEL {
			// Embed related models where conditions
			w.Writef("%sWhereConditions", field.Type.ModelName.Value)
		} else {
			w.Write(toTypeScriptType(field.Type, false, false, false))

			if field.Type.Repeated {
				w.Write("[]")
			}

			w.Write(" | ")
			w.Write(toWhereConditionType(field))
		}

		if field.Optional {
			w.Write(" | null")
		}
		w.Write(";")

		w.Writeln("")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeMessages(w *codegen.Writer, schema *proto.Schema, isTestingPackage bool, isClientPackage bool) {
	for _, msg := range schema.Messages {
		if msg.Name == parser.MessageFieldTypeAny {
			continue
		}

		if schema.IsActionResponseMessage(msg.Name) {
			writeResponseMessage(w, msg, isTestingPackage, isClientPackage)
		} else {
			writeInputMessage(w, msg, isTestingPackage, isClientPackage)
		}
	}
}

func writeInputMessage(w *codegen.Writer, message *proto.Message, isTestingPackage bool, isClientPackage bool) {
	if message.Type != nil {
		w.Writef("export type %s = ", message.Name)
		w.Write(toInputTypescriptType(message.Type, isTestingPackage, isClientPackage))
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

		w.Write(toInputTypescriptType(field.Type, isTestingPackage, isClientPackage))

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

func writeResponseMessage(w *codegen.Writer, message *proto.Message, isTestingPackage bool, isClientPackage bool) {
	if message.Type != nil {
		w.Writef("export type %s = ", message.Name)
		w.Write(toResponseTypescriptType(message.Type, isTestingPackage, isClientPackage))
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

		w.Write(toResponseTypescriptType(field.Type, isTestingPackage, isClientPackage))

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

	type F struct {
		key   string
		value string
	}

	seenCompountUnique := map[string]bool{}

	for _, f := range model.Fields {
		entries := []*F{}

		switch {
		case f.Unique || f.PrimaryKey || len(f.UniqueWith) > 0:
			// Collect unique fields
			fields := []*proto.Field{f}
			fieldNames := []string{f.Name}
			for _, v := range f.UniqueWith {
				u, _ := lo.Find(model.Fields, func(f *proto.Field) bool {
					return f.Name == v
				})
				fields = append(fields, u)
				fieldNames = append(fieldNames, u.Name)
			}

			// De-dupe compound unqique constrains
			sort.Strings(fieldNames)
			k := strings.Join(fieldNames, ":")
			if _, ok := seenCompountUnique[k]; ok {
				continue
			}
			seenCompountUnique[k] = true

			for _, f := range fields {
				if f.Type.Type == proto.Type_TYPE_MODEL {
					if f.ForeignKeyFieldName == nil {
						// I'm not sure this can happen, but rather than have a cryptic nil-pointer error we'll
						// panic with a hopefully more helpful error
						panic(fmt.Sprintf(
							"%s.%s is a relation field and part of a unique constraint but does not have a foreign key - this is unsupported",
							model.Name, f.Name,
						))
					}

					entries = append(entries, &F{
						key:   f.ForeignKeyFieldName.Value,
						value: "string",
					})
				} else {
					entries = append(entries, &F{
						key:   f.Name,
						value: toTypeScriptType(f.Type, false, false, false),
					})
				}
			}
		case f.IsHasMany():
			// If a field is has-many then the other side is has-one, meaning
			// you can use that fields unique conditions to look up _this_ model.
			// Example: an author has many books, but a book has one author, which
			// means given a book id you can find a single author
			entries = append(entries, &F{
				key:   f.Name,
				value: fmt.Sprintf("%sUniqueConditions", f.Type.ModelName.Value),
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

			if f.Type.Repeated {
				w.Write("[")
			}

			switch f.Type.Type {
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

			if f.Type.Repeated {
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
	w.Writef("update(where: %sUniqueConditions, values: Partial<%sUpdateValues>): Promise<%s>;\n", model.Name, model.Name, model.Name)

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

	w.Writef("export interface %sArrayWhereCondition {\n", enum.Name)
	w.Indent()
	w.Write("equals?: ")
	w.Write(enum.Name)
	w.Writeln("[] | null;")
	w.Write("notEquals?: ")
	w.Write(enum.Name)
	w.Writeln("[] | null;")
	w.Write("any?: ")
	w.Write(enum.Name)
	w.Write("ArrayQueryWhereCondition")
	w.Writeln(" | null;")
	w.Write("all?: ")
	w.Write(enum.Name)
	w.Write("ArrayQueryWhereCondition")
	w.Writeln(" | null;")
	w.Dedent()
	w.Writeln("}")

	w.Writef("export interface %sArrayQueryWhereCondition {\n", enum.Name)
	w.Indent()
	w.Write("equals?: ")
	w.Write(enum.Name)
	w.Writeln(" | null;")
	w.Write("notEquals?: ")
	w.Write(enum.Name)
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
	w.Writeln("export declare const errors: runtime.Errors;")

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
	w.Writeln("const headers = new Headers(meta.headers);")
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

	w.Writeln(`const models = createModelAPI();`)
	w.Writeln(`module.exports.InlineFile = runtime.InlineFile;`)
	w.Writeln(`module.exports.File = runtime.File;`)
	w.Writeln(`module.exports.models = models;`)
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
				"foreignKey":      casing.ToSnake(proto.GetForeignKeyFieldName(models, field)),
			}

			switch {
			case field.IsHasOne():
				relationshipConfig["relationshipType"] = "hasOne"
			case field.IsHasMany():
				relationshipConfig["relationshipType"] = "hasMany"
			case field.IsBelongsTo():
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
	msg := schema.FindMessage(action.InputMessageName)

	var whereMsg *proto.Message
	var valuesMsg *proto.Message

	wheres := []string{}
	values := []string{}

	switch action.Type {
	case proto.ActionType_ACTION_TYPE_UPDATE, proto.ActionType_ACTION_TYPE_LIST:
		for _, f := range msg.Fields {
			if f.Name == "where" {
				whereMsg = schema.FindMessage(f.Type.MessageName.Value)
			}
			if f.Name == "values" {
				valuesMsg = schema.FindMessage(f.Type.MessageName.Value)
			}
		}
	case proto.ActionType_ACTION_TYPE_CREATE:
		whereMsg = nil
		valuesMsg = msg
	default:
		whereMsg = msg
	}

	// Using getter method for Fields here as it is safe even if the variables are nil
	for _, f := range whereMsg.GetFields() {
		if isModelInput(schema, f) {
			wheres = append(wheres, fmt.Sprintf(`"%s"`, f.Name))
		}
	}
	for _, f := range valuesMsg.GetFields() {
		if isModelInput(schema, f) {
			values = append(values, fmt.Sprintf(`"%s"`, f.Name))
		}
	}

	functionName := map[proto.ActionType]string{
		proto.ActionType_ACTION_TYPE_GET:    "getFunction",
		proto.ActionType_ACTION_TYPE_UPDATE: "updateFunction",
		proto.ActionType_ACTION_TYPE_CREATE: "createFunction",
		proto.ActionType_ACTION_TYPE_LIST:   "listFunction",
		proto.ActionType_ACTION_TYPE_DELETE: "deleteFunction",
	}[action.Type]

	w.Writeln(fmt.Sprintf(
		"module.exports.%s = %s({model: models.%s, whereInputs: [%s], valueInputs: [%s]})",
		casing.ToCamel(action.Name),
		functionName,
		casing.ToLowerCamel(action.ModelName),
		strings.Join(wheres, ", "),
		strings.Join(values, ", "),
	))
}

// isModelInput returs true if `field` targets a model field, either directly
// or via a child field)
func isModelInput(schema *proto.Schema, field *proto.MessageField) bool {
	if len(field.Target) > 0 {
		return true
	}
	if field.Type.MessageName == nil {
		return false
	}
	msg := schema.FindMessage(field.Type.MessageName.Value)
	for _, f := range msg.Fields {
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
	w.Writef("export declare function %s", casing.ToCamel(action.Name))

	if action.IsArbitraryFunction() {
		inputType := action.InputMessageName
		if inputType == parser.MessageFieldTypeAny {
			inputType = "any"
		}

		w.Writef("(fn: (ctx: ContextAPI, inputs: %s) => ", inputType)
		w.Write(toCustomFunctionReturnType(model, action, false))
		w.Write("): ")
		w.Write(toCustomFunctionReturnType(model, action, false))
		w.Writeln(";")
		return
	}

	hooksType := fmt.Sprintf("%sHooks", casing.ToCamel(action.Name))

	// TODO: void return type here is wrong. It should be the type of the function e.g. (ctx, inputs) => ReturnType
	w.Writef("(hooks?: %s): void\n", hooksType)

	w.Writef("export type %s = ", hooksType)

	modelName := action.ModelName
	queryBuilder := modelName + "QueryBuilder"
	inputs := action.InputMessageName

	switch action.Type {
	case proto.ActionType_ACTION_TYPE_GET:
		w.Writef("GetFunctionHooks<%s, %s, %s>", modelName, queryBuilder, inputs)
	case proto.ActionType_ACTION_TYPE_LIST:
		w.Writef("ListFunctionHooks<%s, %s, %s>", modelName, queryBuilder, inputs)
	case proto.ActionType_ACTION_TYPE_CREATE:
		msg := schema.FindMessage(action.InputMessageName)
		pickKeys := lo.FilterMap(msg.Fields, func(f *proto.MessageField, _ int) (string, bool) {
			return fmt.Sprintf("'%s'", f.Name), isModelInput(schema, f)
		})

		var beforeWriteValues string
		switch len(pickKeys) {
		case len(msg.Fields):
			// All inputs target model fields, this means the beforeWriteValues are exactly the same as the inputs
			beforeWriteValues = inputs
		case 0:
			// No inputs target model fields - need the "empty object" type
			// https://www.totaltypescript.com/the-empty-object-type-in-typescript
			beforeWriteValues = "Record<string, never>"
		default:
			// Some inputs target model fields - so create a new type by picking from inputs
			beforeWriteValues = fmt.Sprintf("Pick<%s, %s>", inputs, strings.Join(pickKeys, " | "))
		}

		w.Writef("CreateFunctionHooks<%s, %s, %s, %s, %sCreateValues>", modelName, queryBuilder, inputs, beforeWriteValues, modelName)
	case proto.ActionType_ACTION_TYPE_UPDATE:
		w.Writef("UpdateFunctionHooks<%s, %s, %s, %sValues>", modelName, queryBuilder, inputs, casing.ToCamel(action.Name))
	case proto.ActionType_ACTION_TYPE_DELETE:
		w.Writef("DeleteFunctionHooks<%s, %s, %s>", modelName, queryBuilder, inputs)
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
	returnType += "| Error>"
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
		className := model.Name
		if len(op.GetResponseEmbeds()) > 0 {
			className = toResponseType(op.Name)
		}
		returnType += sdkPrefix + className + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		className := model.Name
		if len(op.GetResponseEmbeds()) > 0 {
			className = toResponseType(op.Name)
		}
		returnType += "{results: " + sdkPrefix + className + "[], pageInfo: runtime.PageInfo}"
	case proto.ActionType_ACTION_TYPE_DELETE:
		// todo: create ID type
		returnType += "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		returnType += op.ResponseMessageName
	}

	returnType += ">"
	return returnType
}

func generateDevelopmentServer(schema *proto.Schema, cfg *config.ProjectConfig) codegen.GeneratedFiles {
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

	for _, v := range cfg.Auth.EnabledHooks() {
		w.Writef(`import function_%s from "../functions/auth/%s.ts"`, v, v)
		w.Writeln(";")
	}

	w.Writeln("const functions = {")
	w.Indent()

	for _, fn := range functions {
		w.Writef("%s: function_%s,", fn.Name, fn.Name)
		w.Writeln("")
	}

	for _, v := range cfg.Auth.EnabledHooks() {
		w.Writef("%s: function_%s", v, v)
		w.Writeln(",")
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

	try {
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
	} catch (e) {
		console.error("Unexpected Handler Error", e)
	} finally {
		if (tracing.forceFlush) {
			await tracing.forceFlush();
		}
	}

	res.statusCode = 400;
	res.end();
};

tracing.init();

const process = require('node:process');
process.on('unhandledRejection', (reason, promise) => {
	console.error('Unhandled Promise Rejection', promise, 'Reason:', reason);
});

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
	tableNames := []string{"keel_audit", "keel_storage"}
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
import * as path from 'path';

export default defineConfig({
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
	w.Writeln("")

	// For the testing package we need input and response types for all actions
	writeMessages(w, schema, true, false)

	w.Writeln("declare class ActionExecutor {")
	w.Indent()
	w.Writeln("withIdentity(identity: sdk.Identity): ActionExecutor;")
	w.Writeln("withAuthToken(token: string): ActionExecutor;")
	for _, model := range schema.Models {
		for _, action := range model.Actions {
			msg := schema.FindMessage(action.InputMessageName)

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
			msg := schema.FindMessage(job.InputMessageName)

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
			msg := schema.FindMessage(subscriber.InputMessageName)

			w.Writef("%s(e", subscriber.Name)

			if msg.Type.Type != proto.Type_TYPE_UNION && lo.EveryBy(msg.Fields, func(f *proto.MessageField) bool {
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

func toDbTableType(t *proto.TypeInfo, isTestingPackage bool) (ret string) {
	switch t.Type {
	case proto.Type_TYPE_FILE:
		return "FileDbRecord"
	default:
		return toTypeScriptType(t, false, isTestingPackage, false)
	}
}

func toInputTypescriptType(t *proto.TypeInfo, isTestingPackage bool, isClientPackage bool) (ret string) {
	switch t.Type {
	case proto.Type_TYPE_FILE:
		if isClientPackage {
			return "string"
		} else {
			return "runtime.InlineFile"
		}
	default:
		return toTypeScriptType(t, false, isTestingPackage, isClientPackage)
	}
}

func toResponseTypescriptType(t *proto.TypeInfo, isTestingPackage bool, isClientPackage bool) (ret string) {
	switch t.Type {
	case proto.Type_TYPE_FILE:
		if isClientPackage {
			return "FileResponseObject"
		} else {
			return "runtime.File | runtime.InlineFile"
		}
	default:
		return toTypeScriptType(t, false, isTestingPackage, isClientPackage)
	}
}

func toTypeScriptType(t *proto.TypeInfo, includeCompatibleTypes bool, isTestingPackage bool, isClientPackage bool) (ret string) {
	switch t.Type {
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
		if isClientPackage {
			ret = "SortDirection"
		} else {
			ret = "runtime.SortDirection"
		}
	case proto.Type_TYPE_UNION:
		// Retrieve all the types that can satisfy this union field.
		messageNames := lo.Map(t.UnionNames, func(s *wrapperspb.StringValue, _ int) string {
			return s.Value
		})
		ret = fmt.Sprintf("(%s)", strings.Join(messageNames, " | "))
	case proto.Type_TYPE_STRING_LITERAL:
		// Use string literal type for discriminating.
		ret = fmt.Sprintf(`"%s"`, t.StringLiteralValue.Value)
	case proto.Type_TYPE_FILE:
		if isClientPackage {
			ret = "FileResponseObject"
		} else {
			if includeCompatibleTypes {
				ret = "runtime.InlineFile | runtime.File"
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
	if f.Type.Repeated {
		switch f.Type.Type {
		case proto.Type_TYPE_ID, proto.Type_TYPE_STRING, proto.Type_TYPE_MARKDOWN:
			return "runtime.StringArrayWhereCondition"
		case proto.Type_TYPE_BOOL:
			return "runtime.BooleanArrayWhereCondition"
		case proto.Type_TYPE_INT:
			return "runtime.NumberArrayWhereCondition"
		case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			return "runtime.DateArrayWhereCondition"
		case proto.Type_TYPE_ENUM:
			return fmt.Sprintf("%sArrayWhereCondition", f.Type.EnumName.Value)
		}
	}

	switch f.Type.Type {
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

// toResponseType generates a response type name for the given action name. This is to be used for actions that contain
// embedded data
func toResponseType(actionName string) string {
	return casing.ToCamel(actionName) + "Response"
}

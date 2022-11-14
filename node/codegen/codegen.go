package codegenerator

// Files to be generated:
// 1. Generate SDK
// -> dist/index.js
// -> dist/index.d.ts
// -> package.json (peer dep)
// 2. Generate Testing
// -> dist/index.js
// -> dist/index.d.ts
// -> package.json (peer dep)

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type Generator struct {
	schema *proto.Schema
	dir    string
}

func NewGenerator(schema *proto.Schema, dir string) *Generator {
	return &Generator{
		schema: schema,
		dir:    dir,
	}
}

type SourceCode struct {
	Contents string
	Path     string
}

const (
	SdkPackageName     = "sdk"
	TestingPackageName = "testing"
)

// GenerateSDK will generate a fresh @teamkeel/sdk package into the node_modules
// directory of the target directory
func (g *Generator) GenerateSDK() error {
	codez := []*SourceCode{}

	codez = append(codez, &SourceCode{
		Path:     "dist/index.js",
		Contents: g.sdkSrcCode(),
	})

	codez = append(codez, &SourceCode{
		Path:     "dist/index.d.ts",
		Contents: g.sdkTypeDefinitions(),
	})

	err := g.makeNpmPackage(SdkPackageName, codez)

	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) GenerateTesting() error {
	codez := []*SourceCode{}

	codez = append(codez, &SourceCode{
		Path:     "dist/index.js",
		Contents: g.testingSrcCode(),
	})

	codez = append(codez, &SourceCode{
		Path:     "dist/index.d.ts",
		Contents: g.testingTypeDefinitions(),
	})

	err := g.makeNpmPackage(TestingPackageName, codez)

	if err != nil {
		return err
	}

	return nil
}

// private methods

func (g *Generator) testingSrcCode() (r string) {
	renderApis := func(models []*proto.Model) (acc string) {
		for _, m := range models {
			acc += renderTemplate(TemplateTestingModelApi, map[string]interface{}{
				"Name":      m.Name,
				"TableName": strcase.ToSnake(m.Name),
			})
		}
		return acc
	}

	renderActions := func(schema *proto.Schema, withIdentity bool) (r string) {
		actions := lo.FlatMap(schema.Models, func(m *proto.Model, _ int) []*proto.Operation {
			return m.Operations
		})

		for _, action := range actions {
			returnType := ""

			switch action.Type {
			case proto.OperationType_OPERATION_TYPE_CREATE:
				returnType = fmt.Sprintf("ReturnTypes.FunctionCreateResponse<Client.%s>", action.ModelName)
			case proto.OperationType_OPERATION_TYPE_DELETE:
				returnType = fmt.Sprintf("ReturnTypes.FunctionDeleteResponse<Client.%s>", action.ModelName)
			case proto.OperationType_OPERATION_TYPE_LIST:
				returnType = fmt.Sprintf("ReturnTypes.FunctionListResponse<Client.%s>", action.ModelName)
			case proto.OperationType_OPERATION_TYPE_UPDATE:
				returnType = fmt.Sprintf("ReturnTypes.FunctionUpdateResponse<Client.%s>", action.ModelName)
			case proto.OperationType_OPERATION_TYPE_GET:
				returnType = fmt.Sprintf("ReturnTypes.FunctionGetResponse<Client.%s>", action.ModelName)
			default:
				continue
			}

			if withIdentity {
				r += fmt.Sprintf("%s\n\t", renderTemplate(TemplateInstanceProperty, map[string]interface{}{
					"Name": action.Name,
					"Type": fmt.Sprintf("async (payload: any) => await actionExecutor.execute<%s>({ actionName: '%s', payload, identity: this.identity })", returnType, action.Name),
				}))
			} else {
				r += fmt.Sprintf("%s\n\t", renderTemplate(TemplateInstanceProperty, map[string]interface{}{
					"Name": action.Name,
					"Type": fmt.Sprintf("async (payload: any) => await actionExecutor.execute<%s>({ actionName: '%s', payload })", returnType, action.Name),
				}))
			}

		}

		return r
	}

	renderAuthenticationActions := func(schema *proto.Schema) (r string) {
		actions := lo.FlatMap(schema.Models, func(m *proto.Model, _ int) []*proto.Operation {
			return m.Operations
		})

		for _, action := range actions {
			returnType := ""

			switch action.Type {
			case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
				returnType = "ReturnTypes.FunctionAuthenticateResponse"
			default:
				continue
			}

			r += fmt.Sprintf("%s\n\t", renderTemplate(TemplateInstanceProperty, map[string]interface{}{
				"Name": action.Name,
				"Type": fmt.Sprintf("async (payload: any) => await actionExecutor.execute<%s>({ actionName: '%s', payload })", returnType, action.Name),
			}))
		}

		return r
	}

	r += renderTemplate(TemplateTestingBase, map[string]interface{}{
		"TestingModelApis":    renderApis(g.schema.Models),
		"Actions":             renderAuthenticationActions(g.schema) + renderActions(g.schema, false),
		"ActionsWithIdentity": renderActions(g.schema, true),
	})

	return r
}

func (g *Generator) testingTypeDefinitions() (r string) {
	return ""
}

func (g *Generator) sdkSrcCode() (r string) {
	r += g.generateBaseTypes()
	r += g.generateBaseImports()
	r += g.generateModels()
	r += g.generateEnums(false)
	r += g.generateInputs(false)
	r += g.generateWrappers(false)
	r += g.generateAPIs(false)

	return r
}

func (g *Generator) sdkTypeDefinitions() (r string) {
	r += g.generateBaseTypes()
	r += g.generateBaseImports()
	r += g.generateModels()
	r += g.generateEnums(true)
	r += g.generateInputs(true)
	r += g.generateWrappers(true)
	r += g.generateAPIs(true)

	return r
}

//go:embed package.json.tmpl
var templatePackageJson string

func (g *Generator) makeNpmPackage(name string, srcCodes []*SourceCode) error {
	basePath := filepath.Join(g.dir, "node_modules", "@teamkeel", name)
	err := os.MkdirAll(filepath.Join(basePath, "dist"), os.ModePerm)

	if err != nil {
		return err
	}

	template, err := template.New("").Parse(templatePackageJson)

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	templateVars := map[string]any{
		"Name": name,
	}

	err = template.Execute(&buf, templateVars)

	if err != nil {
		return err
	}

	packageJsonPath := filepath.Join(basePath, "package.json")
	f, err := os.Create(packageJsonPath)

	if err != nil {
		return err
	}

	_, err = f.WriteString(buf.String())

	if err != nil {
		return err
	}

	for _, code := range srcCodes {
		f, err := os.Create(filepath.Join(basePath, code.Path))

		if err != nil {
			return err
		}

		_, err = f.WriteString(code.Contents)

		if err != nil {
			return err
		}
	}

	return nil
}

//go:embed templates/*.tmpl
var templates embed.FS

// Static template where imports to various deps are declared
// these imports will be used in many other places in the codegen
func (gen *Generator) generateBaseImports() (r string) {
	r += renderTemplate(TemplateBaseImports, map[string]interface{}{})

	return r
}

// To contain shared types and low level implementation types.
func (gen *Generator) generateBaseTypes() (r string) {
	data := map[string]interface{}{
		"Name":         "Timestamp",
		"ResolvedType": "string",
	}

	r += renderTemplate(TemplateTypeAlias, data)

	return r
}

// func (gen *Generator) generateFunction(operationName string) string {
// 	return renderTemplate(
// 		TemplateCustomFunction,
// 		map[string]interface{}{
// 			"Name": strcase.ToCamel(operationName),
// 		},
// 	)
// }

// 'Wrappers' is the collective term to describe utility functions to
// create/delete/update/get/list entities of a particular model
// e.g CreatePost(async(inputs, api)) => Promise<Post> is a wrapper func
// that encapsulates the typings for inputs/api, and also enforces the return type
// from the function.
// These sorts of utility functions save the user from typing their custom functions
// themselves
func (gen *Generator) generateWrappers(typings bool) (str string) {
	fns := proto.FilterOperations(gen.schema, func(op *proto.Operation) bool {
		return op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
	})

	for _, fn := range fns {
		switch fn.Type {
		case proto.OperationType_OPERATION_TYPE_CREATE:
			if typings {
				str += renderTemplate(
					TemplateFuncWrapperCreateTypings,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			} else {
				str += renderTemplate(
					TemplateFuncWrapperCreate,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			}
		case proto.OperationType_OPERATION_TYPE_DELETE:
			if typings {
				str += renderTemplate(
					TemplateFuncWrapperDeleteTypings,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			} else {
				str += renderTemplate(
					TemplateFuncWrapperDelete,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			}
		case proto.OperationType_OPERATION_TYPE_LIST:
			if typings {
				str += renderTemplate(
					TemplateFuncWrapperListTypings,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			} else {
				str += renderTemplate(
					TemplateFuncWrapperList,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			}
		case proto.OperationType_OPERATION_TYPE_UPDATE:
			if typings {
				str += renderTemplate(
					TemplateFuncWrapperUpdateTypings,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			} else {
				str += renderTemplate(
					TemplateFuncWrapperUpdate,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			}
		case proto.OperationType_OPERATION_TYPE_GET:
			if typings {
				str += renderTemplate(
					TemplateFuncWrapperGetTypings,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			} else {
				str += renderTemplate(
					TemplateFuncWrapperGet,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			}
		case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
			if typings {
				str += renderTemplate(
					TemplateFuncWrapperAuthenticateTypings,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			} else {
				str += renderTemplate(
					TemplateFuncWrapperAuthenticate,
					map[string]interface{}{
						"Name":  strcase.ToCamel(fn.Name),
						"Model": fn.ModelName,
					},
				)
			}
		}
	}

	return str
}

func (gen *Generator) generateEnums(typings bool) (r string) {
	for _, enum := range gen.schema.Enums {

		renderValues := func(values []*proto.EnumValue) (v string) {
			for i, value := range values {
				lastItem := i == len(values)-1

				templateResult := ""

				if typings {
					templateResult = renderTemplate(TemplateEnumValueTyping, map[string]interface{}{
						"Key":   value.Name,
						"Value": value.Name,
						"Comma": !lastItem,
					})
				} else {
					templateResult = renderTemplate(TemplateEnumValue, map[string]interface{}{
						"Key":   value.Name,
						"Value": value.Name,
						"Comma": !lastItem,
					})
				}

				if lastItem {
					v += templateResult
				} else {
					v += fmt.Sprintf("%s\n", templateResult)
				}
			}

			return v
		}

		if typings {
			r += renderTemplate(TemplateEnumTyping, map[string]interface{}{
				"Name":   enum.Name,
				"Values": renderValues(enum.Values),
			})
		} else {
			r += renderTemplate(TemplateEnum, map[string]interface{}{
				"Name":   enum.Name,
				"Values": renderValues(enum.Values),
			})
		}
	}

	return r
}

func (gen *Generator) generateModels() (r string) {
	renderFields := func(fields []*proto.Field) (acc string) {
		for i, field := range fields {
			if i == 0 {
				acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     protoTypeToTypeScriptType(field.Type),
					"Optional": field.Optional,
				}))
			} else if i < len(fields)-1 {
				acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     protoTypeToTypeScriptType(field.Type),
					"Optional": field.Optional,
				}))
			} else {
				acc += fmt.Sprintf("  %s", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     protoTypeToTypeScriptType(field.Type),
					"Optional": field.Optional,
				}))
			}
		}

		return acc
	}

	for _, model := range gen.schema.Models {
		r += renderTemplate(TemplateInterface, map[string]interface{}{
			"Name":       model.Name,
			"Properties": renderFields(model.Fields),
		})
	}

	return r
}

var APIName = "API"

func (gen *Generator) generateAPIs(typings bool) (r string) {
	renderModelApiDefs := func(models []*proto.Model) (acc string) {
		for i, model := range models {
			if i == 0 {
				acc += fmt.Sprintf("%s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name": strcase.ToLowerCamel(model.Name),
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			} else if i < len(models)-1 {
				acc += fmt.Sprintf("    %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name": strcase.ToLowerCamel(model.Name),
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			} else {
				acc += fmt.Sprintf("    %s", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name": strcase.ToLowerCamel(model.Name),
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			}
		}

		return acc
	}

	getFieldConstraintType := func(model *proto.Model, field *proto.Field) string {
		if field.Type.Type == proto.Type_TYPE_ENUM {
			return "QueryConstraints.EnumConstraint"
		}

		return fmt.Sprintf("QueryConstraints.%sConstraint", strcase.ToCamel(protoTypeToTypeScriptType(field.Type)))
	}

	buildQueryConstraints := func(model *proto.Model) (r string) {
		for i, field := range model.Fields {
			if i == 0 {
				r += fmt.Sprintf("%s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				}))
			} else if i < len(model.Fields)-1 {
				r += fmt.Sprintf("  %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				}))
			} else {
				r += fmt.Sprintf("  %s", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				}))
			}
		}

		return r
	}

	buildUniqueFields := func(model *proto.Model) (r string) {
		uniqueFields := lo.Filter(model.Fields, func(f *proto.Field, i int) bool {
			return f.Unique || f.PrimaryKey
		})

		for i, field := range uniqueFields {
			if i == len(uniqueFields)-1 {
				r += renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				})
			} else if i == 0 {
				r += fmt.Sprintf("%s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				}))
			} else if i < len(model.Fields)-1 {
				r += fmt.Sprintf("  %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				}))
			} else {
				r += fmt.Sprintf("  %s", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     field.Name,
					"Type":     getFieldConstraintType(model, field),
					"Optional": true,
				}))
			}
		}

		return r
	}

	for _, model := range gen.schema.Models {
		if typings {
			r += renderTemplate(TemplateApiTyping, map[string]interface{}{
				"Name":             model.Name,
				"QueryConstraints": buildQueryConstraints(model),
				"UniqueFields":     buildUniqueFields(model),
			})
		} else {
			r += renderTemplate(TemplateApi, map[string]interface{}{
				"Name":             model.Name,
				"TableName":        strcase.ToSnake(model.Name),
				"QueryConstraints": buildQueryConstraints(model),
				"UniqueFields":     buildUniqueFields(model),
			})
		}
	}

	if typings {
		r += fmt.Sprintf("\n%s\n", renderTemplate(TemplateKeelApiTypings, map[string]interface{}{
			"Name":      APIName,
			"ModelApis": renderModelApiDefs(gen.schema.Models),
		}))
	} else {
		r += fmt.Sprintf("\n%s\n", renderTemplate(TemplateKeelApi, map[string]interface{}{
			"Name":      APIName,
			"ModelApis": renderModelApiDefs(gen.schema.Models),
		}))
	}

	return r
}

func renderInputFields(inputs []*proto.OperationInput, filter func(input *proto.OperationInput) bool) (acc string) {
	filtered := []*proto.OperationInput{}

	for _, input := range inputs {
		if !filter(input) {
			continue
		}

		filtered = append(filtered, input)
	}

	for i, input := range filtered {

		if input.Type.Type == proto.Type_TYPE_OBJECT {
			acc += renderInputFields(input.Inputs, filter)
			continue
		}

		tsType := protoTypeToTypeScriptType(input.Type)

		if i < len(filtered)-1 {
			acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
				"Name":     input.Name,
				"Type":     tsType,
				"Optional": input.Optional,
			}))
		} else {

			acc += fmt.Sprintf("  %s", renderTemplate(TemplateProperty, map[string]interface{}{
				"Name":     input.Name,
				"Type":     tsType,
				"Optional": input.Optional,
			}))
		}
	}

	return acc
}

func (gen *Generator) generateInputs(typings bool) (r string) {
	for _, model := range gen.schema.Models {
		for _, op := range model.Operations {
			inputs := op.GetInputs()

			switch op.Type {
			case proto.OperationType_OPERATION_TYPE_CREATE:
				if typings {
					r += renderTemplate(TemplateCreateInputTypings, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_WRITE
						}),
					})
				} else {
					r += renderTemplate(TemplateCreateInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_WRITE
						}),
					})
				}
			case proto.OperationType_OPERATION_TYPE_UPDATE:
				if typings {
					r += renderTemplate(TemplateUpdateInputTypings, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Filters": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
						"Values": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_WRITE
						}),
					})
				} else {
					r += renderTemplate(TemplateUpdateInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Filters": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
						"Values": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_WRITE
						}),
					})
				}
			case proto.OperationType_OPERATION_TYPE_GET:
				if typings {
					r += renderTemplate(TemplateGetInputTypings, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
					})
				} else {
					r += renderTemplate(TemplateGetInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
					})
				}
			case proto.OperationType_OPERATION_TYPE_LIST:
				if typings {
					r += renderTemplate(TemplateListInputTypings, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
					})
				} else {
					r += renderTemplate(TemplateListInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
					})
				}
			case proto.OperationType_OPERATION_TYPE_DELETE:
				if typings {
					r += renderTemplate(TemplateDeleteInputTypings, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
					})
				} else {
					r += renderTemplate(TemplateDeleteInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
					})
				}
			case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
				if typings {
					r += renderTemplate(TemplateAuthenticateInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return true
						}),
					})
				} else {
					r += renderTemplate(TemplateAuthenticateInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Properties": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return true
						}),
					})
				}
			}

		}
	}

	return r
}

// Generates code for the @teamkeel/testing package
// The testing package mostly reuses generated code
// from the @teamkeel/client code generation above, but the invocation is
// slightly different. Database pools for code-generated APIs that talk to
// the database are constructed from within the testing package itself

var (
	TSTypeUnknown   = "unknown"
	TSTypeString    = "string"
	TSTypeBoolean   = "boolean"
	TSTypeNumber    = "number"
	TSTypeDate      = "Date"
	TSTypeTimestamp = "Timestamp"
	TSTypeIdentity  = "Identity"
)

type Typer interface {
	GetType() proto.Type
	IsRepeated() bool
}

// Maps the internal Keel field type to a corresponding valid typescript type
func protoTypeToTypeScriptType(t *proto.TypeInfo) string {
	switch t.GetType() {
	case proto.Type_TYPE_UNKNOWN:
		return TSTypeUnknown
	case proto.Type_TYPE_STRING:
		return TSTypeString
	case proto.Type_TYPE_BOOL:
		return TSTypeBoolean
	case proto.Type_TYPE_INT:
		return TSTypeNumber
	case proto.Type_TYPE_TIMESTAMP:
		return TSTypeTimestamp
	case proto.Type_TYPE_DATE:
		return TSTypeDate
	case proto.Type_TYPE_ID:
		return TSTypeString
	case proto.Type_TYPE_MODEL:
		if t.Repeated {
			return fmt.Sprintf("%s[]", t.ModelName.Value)
		}
		return t.ModelName.Value
	// case proto.Type_TYPE_CURRENCY:
	// 	return "Currency"
	case proto.Type_TYPE_DATETIME:
		return TSTypeDate
	case proto.Type_TYPE_ENUM:
		return t.EnumName.Value
	// case proto.Type_TYPE_IMAGE:
	// 	return "Image"
	case proto.Type_TYPE_PASSWORD: // todo: remove this and hide password fields going forward?
		return TSTypeString
	default:
		return TSTypeUnknown
	}
}

var (
	TemplateKeelApi                 = "keel_api"
	TemplateApi                     = "api"
	TemplateEnum                    = "enum"
	TemplateEnumValue               = "enum_value"
	TemplateProperty                = "property"
	TemplateInstanceProperty        = "instance_property"
	TemplateInterface               = "interface"
	TemplateTypeAlias               = "type_alias"
	TemplateHandler                 = "handler"
	TemplateImport                  = "import"
	TemplateObject                  = "object"
	TemplateCustomFunction          = "custom_function"
	TemplateUpdateInput             = "update_input"
	TemplateCreateInput             = "create_input"
	TemplateGetInput                = "get_input"
	TemplateListInput               = "list_input"
	TemplateDeleteInput             = "delete_input"
	TemplateAuthenticateInput       = "authenticate_input"
	TemplateFuncWrapperCreate       = "func_wrapper_create"
	TemplateFuncWrapperDelete       = "func_wrapper_delete"
	TemplateFuncWrapperUpdate       = "func_wrapper_update"
	TemplateFuncWrapperList         = "func_wrapper_list"
	TemplateFuncWrapperGet          = "func_wrapper_get"
	TemplateFuncWrapperAuthenticate = "func_wrapper_authenticate"
	TemplateBaseImports             = "base_imports"

	// Typing templates - used to generate index.d.ts file
	TemplateApiTyping                      = "api_typings"
	TemplateCreateInputTypings             = "create_input_typings"
	TemplateGetInputTypings                = "get_input_typings"
	TemplateListInputTypings               = "list_input_typings"
	TemplateDeleteInputTypings             = "delete_input_typings"
	TemplateUpdateInputTypings             = "update_input_typings"
	TemplateEnumTyping                     = "enum_typing"
	TemplateEnumValueTyping                = "enum_value_typing"
	TemplateFuncWrapperCreateTypings       = "func_wrapper_create_typings"
	TemplateFuncWrapperDeleteTypings       = "func_wrapper_delete_typings"
	TemplateFuncWrapperUpdateTypings       = "func_wrapper_update_typings"
	TemplateFuncWrapperListTypings         = "func_wrapper_list_typings"
	TemplateFuncWrapperGetTypings          = "func_wrapper_get_typings"
	TemplateFuncWrapperAuthenticateTypings = "func_wrapper_authenticate_typings"
	TemplateKeelApiTypings                 = "keel_api_typings"
	TemplateHandlerApi                     = "handler_api"

	// Templates for the @teamkeel/testing package
	TemplateTestingBase     = "testing_base"      // includes base imports used by testing package
	TemplateTestingModelApi = "testing_model_api" // template for augmented version of the @teamkeel/client model apis
)

func renderTemplate(name string, data map[string]interface{}) string {
	template, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.tmpl", name))
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer

	err = template.Execute(&tpl, data)

	if err != nil {
		panic(err)
	}

	return tpl.String()
}

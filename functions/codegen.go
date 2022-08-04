package functions

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type CodeGenerator struct {
	schema *proto.Schema
}

//go:embed codegen_templates/*.tmpl
var templates embed.FS

func NewCodeGenerator(schema *proto.Schema) *CodeGenerator {
	return &CodeGenerator{
		schema: schema,
	}
}

func (gen *CodeGenerator) GenerateClientCode() (r string) {
	r += gen.GenerateBaseTypes()
	r += gen.GenerateModels()
	r += gen.GenerateEnums(false)
	r += gen.GenerateInputs(false)
	r += gen.GenerateWrappers(false)
	r += gen.GenerateAPIs(false)

	return r
}

func (gen *CodeGenerator) GenerateClientTypings() (r string) {
	r += gen.GenerateBaseTypes()
	r += gen.GenerateModels()
	r += gen.GenerateEnums(true)
	r += gen.GenerateInputs(true)
	r += gen.GenerateWrappers(true)
	r += gen.GenerateAPIs(true)

	return r
}

func (gen *CodeGenerator) GenerateBaseTypes() (r string) {
	data := map[string]interface{}{
		"Name":         "Timestamp",
		"ResolvedType": "string",
	}

	r += renderTemplate(TemplateTypeAlias, data)

	return r
}

func (gen *CodeGenerator) GenerateFunction(operationName string) string {
	return renderTemplate(
		TemplateCustomFunction,
		map[string]interface{}{
			"Name": strcase.ToCamel(operationName),
		},
	)
}

func (gen *CodeGenerator) GenerateWrappers(typings bool) (str string) {
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
		}
	}

	return str
}

func (gen *CodeGenerator) GenerateEnums(typings bool) (r string) {
	for _, enum := range gen.schema.Enums {

		renderValues := func(values []*proto.EnumValue) (v string) {
			for i, value := range values {
				lastItem := i == len(values)-1

				templateResult := ""

				if typings {
					templateResult = renderTemplate(TemplateEnumValueTyping, map[string]interface{}{
						"Value": value.Name,
						"Index": i,
						"Comma": !lastItem,
					})
				} else {
					templateResult = renderTemplate(TemplateEnumValue, map[string]interface{}{
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

func (gen *CodeGenerator) GenerateModels() (r string) {
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

func (gen *CodeGenerator) GenerateAPIs(typings bool) (r string) {
	renderModelApiDefs := func(models []*proto.Model) (acc string) {
		modelsToUse := lo.Filter(models, func(model *proto.Model, _ int) bool {
			return model.Name != TSTypeIdentity
		})

		for i, model := range modelsToUse {
			// we do not want to expose the API for interacting with the identities table
			if model.Name == TSTypeIdentity {
				continue
			}

			if i == 0 {
				acc += fmt.Sprintf("%s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name": model.Name,
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			} else if i < len(modelsToUse)-1 {
				acc += fmt.Sprintf("    %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name": model.Name,
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			} else {
				acc += fmt.Sprintf("    %s", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name": model.Name,
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			}
		}

		return acc
	}

	if typings {
		r += renderTemplate(TemplateKeelApiTypings, map[string]interface{}{
			"Name":      APIName,
			"ModelApis": renderModelApiDefs(gen.schema.Models),
		})
	} else {
		r += renderTemplate(TemplateKeelApi, map[string]interface{}{
			"Name":      APIName,
			"ModelApis": renderModelApiDefs(gen.schema.Models),
		})
	}

	for _, model := range gen.schema.Models {
		if model.Name == TSTypeIdentity {
			continue
		}

		if typings {
			r += renderTemplate(TemplateApiTyping, map[string]interface{}{
				"Name": model.Name,
			})
		} else {
			r += renderTemplate(TemplateApi, map[string]interface{}{
				"Name": model.Name,
			})
		}
	}

	return r
}

func (gen *CodeGenerator) GenerateInputs(typings bool) (r string) {
	renderInputFields := func(inputs []*proto.OperationInput, filter func(input *proto.OperationInput) bool) (acc string) {
		filtered := []*proto.OperationInput{}

		for _, input := range inputs {
			if !filter(input) {
				continue
			}

			filtered = append(filtered, input)
		}

		for i, input := range filtered {
			if i < len(filtered)-1 {
				acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     input.Name,
					"Type":     protoTypeToTypeScriptType(input.Type),
					"Optional": input.Optional,
				}))
			} else {
				acc += fmt.Sprintf("  %s", renderTemplate(TemplateProperty, map[string]interface{}{
					"Name":     input.Name,
					"Type":     protoTypeToTypeScriptType(input.Type),
					"Optional": input.Optional,
				}))
			}
		}

		return acc
	}

	for _, model := range gen.schema.Models {
		if model.Name == TSTypeIdentity {
			continue
		}

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
						"Filters": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
						"Values": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_WRITE
						}),
					})
				} else {
					r += renderTemplate(TemplateListInput, map[string]interface{}{
						"Name": strcase.ToCamel(op.Name),
						"Filters": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_READ
						}),
						"Values": renderInputFields(inputs, func(input *proto.OperationInput) bool {
							return input.GetMode() == proto.InputMode_INPUT_MODE_WRITE
						}),
					})
				}
			}
		}
	}

	return r
}

func (gen *CodeGenerator) GenerateEntryPoint() (r string) {
	renderFunctions := func(sch *proto.Schema) (acc string) {
		for _, model := range sch.Models {
			functions := lo.Filter(model.Operations, func(o *proto.Operation, _ int) bool {
				return o.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
			})

			renderEntries := func(entries map[string]interface{}) (acc string) {
				keys := make([]string, 0)
				for k := range entries {
					keys = append(keys, k)
				}

				sort.Strings(keys)

				var i int = 0

				for _, key := range keys {
					entry := entries[key]
					contents := renderTemplate(TemplateProperty, map[string]interface{}{
						"Name": key,
						"Type": entry,
					})

					if i == 0 {
						acc += fmt.Sprintf("%s,", contents)
					} else if i < len(functions)-1 {
						acc += fmt.Sprintf(" %s,", contents)
					} else {
						acc += fmt.Sprintf(" %s", contents)
					}

					i++
				}

				return acc
			}

			for i, op := range functions {
				tmp := renderTemplate(TemplateObject, map[string]interface{}{
					"Name": op.Name,
					"Entries": renderEntries(map[string]interface{}{
						"contextModel": fmt.Sprintf("'%s'", op.ModelName),
						"call":         op.Name,
					}),
				})

				if i == 0 {
					acc += fmt.Sprintf("%s,", tmp)
				} else if i < len(functions)-1 {
					acc += fmt.Sprintf(" %s,", tmp)
				} else {
					acc += fmt.Sprintf(" %s", tmp)
				}
			}
		}

		return acc
	}

	renderImports := func(sch *proto.Schema) (acc string) {
		acc += fmt.Sprintf("\n%s\n", renderTemplate(TemplateImport, map[string]interface{}{
			"Name": "startRuntimeServer",
			"Path": "@teamkeel/runtime",
		}))

		for _, model := range sch.Models {
			functions := lo.Filter(model.Operations, func(o *proto.Operation, _ int) bool {
				return o.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
			})

			for _, op := range functions {
				acc += fmt.Sprintf("%s\n", renderTemplate(TemplateImport, map[string]interface{}{
					"Name": op.Name,
					// We need to refer to the users functions directory,
					// which will be a few levels above the @teamkeel/client/dist directory.
					// The hierarchy is as follows:
					// project/
					//   functions/
					//   node_modules/
					//     @teamkeel/
					//       client/
					//         dist/
					//           handler.js
					"Path": fmt.Sprintf("../../../../functions/%s", op.Name),
				}))
			}
		}

		return acc
	}

	r += renderTemplate(TemplateHandler, map[string]interface{}{
		"Functions": renderFunctions(gen.schema),
		"Imports":   renderImports(gen.schema),
	})

	return r
}

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
	default:
		return TSTypeUnknown
	}
}

var (
	TemplateKeelApi           = "keel_api"
	TemplateApi               = "api"
	TemplateEnum              = "enum"
	TemplateEnumValue         = "enum_value"
	TemplateProperty          = "property"
	TemplateInterface         = "interface"
	TemplateTypeAlias         = "type_alias"
	TemplateHandler           = "handler"
	TemplateImport            = "import"
	TemplateObject            = "object"
	TemplateCustomFunction    = "custom_function"
	TemplateUpdateInput       = "update_input"
	TemplateCreateInput       = "create_input"
	TemplateGetInput          = "get_input"
	TemplateListInput         = "list_input"
	TemplateFuncWrapperCreate = "func_wrapper_create"
	TemplateFuncWrapperDelete = "func_wrapper_delete"
	TemplateFuncWrapperUpdate = "func_wrapper_update"
	TemplateFuncWrapperList   = "func_wrapper_list"
	TemplateFuncWrapperGet    = "func_wrapper_get"

	// Typing templates - used to generate index.d.ts file
	TemplateApiTyping                = "api_typings"
	TemplateCreateInputTypings       = "create_input_typings"
	TemplateGetInputTypings          = "get_input_typings"
	TemplateListInputTypings         = "list_input_typings"
	TemplateUpdateInputTypings       = "update_input_typings"
	TemplateEnumTyping               = "enum_typing"
	TemplateEnumValueTyping          = "enum_value_typing"
	TemplateFuncWrapperCreateTypings = "func_wrapper_create_typings"
	TemplateFuncWrapperDeleteTypings = "func_wrapper_delete_typings"
	TemplateFuncWrapperUpdateTypings = "func_wrapper_update_typings"
	TemplateFuncWrapperListTypings   = "func_wrapper_list_typings"
	TemplateFuncWrapperGetTypings    = "func_wrapper_get_typings"
	TemplateKeelApiTypings           = "keel_api_typings"
)

func renderTemplate(name string, data map[string]interface{}) string {
	template, err := template.ParseFS(templates, fmt.Sprintf("codegen_templates/%s.tmpl", name))
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

package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type CodeGenerator struct {
	schema *proto.Schema
}

//go:embed templates/*.tmpl
var templates embed.FS

func NewCodeGenerator(schema *proto.Schema) *CodeGenerator {
	return &CodeGenerator{
		schema: schema,
	}
}

func (gen *CodeGenerator) GenerateClientCode() (r string) {
	r += gen.GenerateBaseTypes()
	r += gen.GenerateModels()
	r += gen.GenerateEnums()
	r += gen.GenerateInputs()
	r += gen.GenerateAPIs()
	r += gen.GenerateServer()
	r += gen.GenerateEntryPoint()

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

func (gen *CodeGenerator) GenerateServer() (r string) {
	return renderTemplate(TemplateServer, map[string]interface{}{})
}

func (gen *CodeGenerator) GenerateEnums() (r string) {
	for _, enum := range gen.schema.Enums {

		renderValues := func(values []*proto.EnumValue) (v string) {
			for i, value := range values {
				lastItem := i == len(values)-1

				if lastItem {
					v += renderTemplate(TemplateEnumValue, map[string]interface{}{
						"Value": value.Name,
						"Comma": false,
					})
				} else {
					v += fmt.Sprintf("%s\n", renderTemplate(TemplateEnumValue, map[string]interface{}{
						"Value": value.Name,
						"Comma": true,
					}))
				}
			}

			return v
		}

		r += renderTemplate(TemplateEnum, map[string]interface{}{
			"Name":   enum.Name,
			"Values": renderValues(enum.Values),
		})
	}

	return r
}

func (gen *CodeGenerator) GenerateModels() (r string) {
	renderFields := func(fields []*proto.Field) (acc string) {
		for i, field := range fields {
			if i == 0 {
				acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": field.Name,
					"Type": protoTypeToTypeScriptType(field),
				}))
			} else if i < len(fields)-1 {
				acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": field.Name,
					"Type": protoTypeToTypeScriptType(field),
				}))
			} else {
				acc += fmt.Sprintf("  %s", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": field.Name,
					"Type": protoTypeToTypeScriptType(field),
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

var ExcludedInputs = []string{"id", "updatedAt", "createdAt"}

func (gen *CodeGenerator) GenerateAPIs() (r string) {
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
				acc += fmt.Sprintf("%s\n", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": model.Name,
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			} else if i < len(modelsToUse)-1 {
				acc += fmt.Sprintf("    %s\n", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": model.Name,
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			} else {
				acc += fmt.Sprintf("    %s", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": model.Name,
					"Type": fmt.Sprintf("%sApi", model.Name),
				}))
			}
		}

		return acc
	}

	r += renderTemplate(TemplateKeelApi, map[string]interface{}{
		"Name":      APIName,
		"ModelApis": renderModelApiDefs(gen.schema.Models),
	})

	for _, model := range gen.schema.Models {
		if model.Name == TSTypeIdentity {
			continue
		}

		r += renderTemplate(TemplateApi, map[string]interface{}{
			"Name": model.Name,
		})
	}

	return r
}

func (gen *CodeGenerator) GenerateInputs() (r string) {
	// render input type interfaces
	renderInputFields := func(fields []*proto.Field) (acc string) {
		fieldsToUse := lo.Filter(fields, func(field *proto.Field, _ int) bool {
			return !lo.Contains(ExcludedInputs, field.Name)
		})

		for i, field := range fieldsToUse {
			if i < len(fieldsToUse)-1 {
				acc += fmt.Sprintf("  %s\n", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": field.Name,
					"Type": protoTypeToTypeScriptType(field),
				}))
			} else {
				acc += fmt.Sprintf("  %s", renderTemplate(TemplateInterfaceProperty, map[string]interface{}{
					"Name": field.Name,
					"Type": protoTypeToTypeScriptType(field),
				}))
			}
		}

		return acc
	}

	for _, model := range gen.schema.Models {
		if model.Name == TSTypeIdentity {
			continue
		}

		r += renderTemplate(TemplateInterface, map[string]interface{}{
			"Name":       fmt.Sprintf("%sInputs", model.Name),
			"Properties": renderInputFields(model.Fields),
		})
	}

	return r
}

func (gen *CodeGenerator) GenerateEntryPoint() (r string) {
	renderFunctions := func(sch *proto.Schema) (acc string) {
		for _, model := range sch.Models {
			functions := lo.Filter(model.Operations, func(o *proto.Operation, _ int) bool {
				return o.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
			})

			for i, op := range functions {
				if i == 0 {
					acc += fmt.Sprintf("%s,\n", op.Name)
				} else if i < len(functions)-1 {
					acc += fmt.Sprintf("    %s,\n", op.Name)
				} else {
					acc += fmt.Sprintf("    %s", op.Name)
				}
			}
		}

		return acc
	}

	renderImports := func(sch *proto.Schema) (acc string) {
		for _, model := range sch.Models {
			functions := lo.Filter(model.Operations, func(o *proto.Operation, _ int) bool {
				return o.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
			})

			for _, op := range functions {
				acc += fmt.Sprintf("%s\n", renderTemplate(TemplateImport, map[string]interface{}{
					"Name": op.Name,
					"Path": fmt.Sprintf("../functions/%s", op.Name),
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

func protoTypeToTypeScriptType(f *proto.Field) string {
	switch f.Type.Type {
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
		if f.Type.Repeated {
			return fmt.Sprintf("%s[]", f.Type.ModelName.Value)
		}
		return f.Type.ModelName.Value
	// case proto.Type_TYPE_CURRENCY:
	// 	return "Currency"
	case proto.Type_TYPE_DATETIME:
		return TSTypeDate
	case proto.Type_TYPE_ENUM:
		return f.Type.EnumName.Value
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
	TemplateInterfaceProperty = "interface_property"
	TemplateInterface         = "interface"
	TemplateTypeAlias         = "type_alias"
	TemplateHandler           = "handler"
	TemplateServer            = "server"
	TemplateImport            = "import"
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

package functions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type CodeGenerator struct {
	schema *proto.Schema
}

func NewCodeGenerator(schema *proto.Schema) *CodeGenerator {
	return &CodeGenerator{
		schema: schema,
	}
}

func (gen *CodeGenerator) Generate() (r string) {
	// Generates common data types such as Timestamp
	r += gen.GenerateBaseTypes()
	// Generates typedefs for all models defined in a schema
	r += gen.GenerateModels()
	// Generates all enums defined in a schema
	r += gen.GenerateEnums()
	// Generates Models API definitions for all models
	r += gen.GenerateAPIs()

	return r
}

func (gen *CodeGenerator) GenerateBaseTypes() (r string) {
	r += "type Timestamp = string\n\n"
	return r
}

func (gen *CodeGenerator) GenerateEnums() (r string) {
	for _, enum := range gen.schema.Enums {
		r += fmt.Sprintf("export enum %s {\n", enum.Name)

		for i, v := range enum.Values {
			if i == len(enum.Values)-1 {
				r += fmt.Sprintf("  %s\n", v.Name)
			} else {
				r += fmt.Sprintf("  %s,\n", v.Name)
			}
		}

		r += "}\n\n"
	}

	return r
}

func (gen *CodeGenerator) GenerateModels() (r string) {
	for _, model := range gen.schema.Models {

		r += newTSInterface(model.Name, func(acc string) string {
			for _, field := range model.Fields {
				acc += newTSInterfaceProperty(field.Name, protoTypeToTypeScriptType(field))
			}

			return acc
		})
	}

	return r
}

var APIName = "API"

var (
	ActionCreate   = "create"
	ActionDelete   = "delete"
	ActionFind     = "find"
	ActionFindMany = "findMany"
	ActionUpdate   = "update"
)

var ExcludedInputs = []string{"id", "updatedAt", "createdAt"}

func (gen *CodeGenerator) GenerateAPIs() (r string) {
	r += newTSInterface(APIName, func(acc string) string {
		acc += "  models: {\n"

		for _, model := range gen.schema.Models {
			if model.Name == TSTypeIdentity {
				continue
			}
			acc += "  " + newTSInterfaceProperty(model.Name, fmt.Sprintf("%sApi", model.Name))
		}

		acc += "  }\n"

		return acc
	})

	for _, model := range gen.schema.Models {
		if model.Name == TSTypeIdentity {
			continue
		}

		// Inputs for creating, excludes ID, createdAt, updatedAt
		r += newTSInterface(fmt.Sprintf("%sInputs", model.Name), func(acc string) string {

			for _, field := range model.Fields {
				if lo.Contains(ExcludedInputs, field.Name) {
					continue
				}

				acc += newTSInterfaceProperty(field.Name, protoTypeToTypeScriptType(field))
			}
			return acc
		})

		r += newTSInterface(fmt.Sprintf("%sApi", model.Name), func(acc string) string {
			acc += newTSInterfaceProperty(ActionCreate, fmt.Sprintf("(inputs: %sInputs) => Promise<%s>", model.Name, model.Name))
			acc += newTSInterfaceProperty(ActionDelete, "(id: string) => Promise<boolean>")
			acc += newTSInterfaceProperty(ActionFind, fmt.Sprintf("(p: Partial<%s>) => Promise<%s>", model.Name, model.Name))
			acc += newTSInterfaceProperty(ActionUpdate, fmt.Sprintf("(id: string, inputs: %sInputs) => Promise<%s>", model.Name, model.Name))
			acc += newTSInterfaceProperty(ActionFindMany, fmt.Sprintf("(p: Partial<%s>) => Promise<%s[]>", model.Name, model.Name))

			return acc
		})
	}

	return r
}

func newTSInterface(name string, bodyFunc func(acc string) string) (r string) {
	r += fmt.Sprintf("export interface %s {\n", name)

	r += bodyFunc("")

	r += "}\n\n"

	return r
}

func newTSInterfaceProperty(name string, t string) string {
	return fmt.Sprintf("  %s: %s\n", name, t)
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

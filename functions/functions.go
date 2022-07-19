package functions

import (
	"fmt"

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
		r += fmt.Sprintf("export interface %s {\n", model.Name)

		for _, field := range model.Fields {

			r += fmt.Sprintf("  %s: %s\n", field.Name, ProtoTypeToTypeScriptType(field))
		}

		r += "}\n\n"
	}

	return r
}

func (gen *CodeGenerator) GenerateAPIs() string {
	return ""
}

var (
	TSTypeUnknown   = "unknown"
	TSTypeString    = "string"
	TSTypeBoolean   = "boolean"
	TSTypeNumber    = "number"
	TSTypeDate      = "Date"
	TSTypeTimestamp = "Timestamp"
)

func ProtoTypeToTypeScriptType(f *proto.Field) string {
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
	case proto.Type_TYPE_IDENTITY:
		return "Identity"
	// case proto.Type_TYPE_IMAGE:
	// 	return "Image"
	default:
		return TSTypeUnknown
	}
}

package validation

import (
	"fmt"

	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
)

type SchemaValidator struct {
	declarations *parser.Schema // Input schema
	protoSchema *proto.Schema // Constructed incrementally during validation.
}

func NewSchemaValidator(declarationsAST *parser.Schema) *SchemaValidator {
	return &SchemaValidator{
		declarations: declarationsAST,
		protoSchema: &proto.Schema{
			Models: []*proto.Model{},
		},
	}
}

func (sv *SchemaValidator) Validate() (*proto.Schema, error) {
	for _, decl := range sv.declarations.Declarations {
		model := decl.Model
		protoModel, err := sv.makeProtoModel(model)
		if err != nil {
			return nil, fmt.Errorf("makeProtoModel() failed with: %v", err)
		}
		sv.protoSchema.Models = append(sv.protoSchema.Models, protoModel)
	}
	return sv.protoSchema, nil
}

func (sv *SchemaValidator) makeProtoModel(parserModel *parser.Model) (*proto.Model, error) {
	model := &proto.Model{
		Fields: []*proto.Field{},
		Operations: []*proto.Operation{},
		Attributes: []*proto.Attribute{},
	}
	err := sv.validateModelName(parserModel.Name)
	if err != nil {
		return nil, fmt.Errorf("validateModelName() failed with: %v", err)
	}
	model.Name = parserModel.Name

	for _, section := range parserModel.Sections {
		model.Fields, err = sv.makeFields(section.Fields, model.Name)
		if err != nil {
			return nil, fmt.Errorf("makeFields() failed with: %v", err)
		}

		model.Operations, err = sv.makeOperations(section.Operation)
		if err != nil {
			return nil, fmt.Errorf("makeOperations() failed with: %v", err)
		}

		model.Attributes, err = sv.makeAtributes(section.Attributes)
		if err != nil {
			return nil, fmt.Errorf("makeAttributes() failed with: %v", err)
		}
	}

	return model, nil
}

func (sv *SchemaValidator) validateModelName(incomingName string) error {
	// Model names must be PascalCase
	if !ModelsUpperCamel(incomingName) {
		return fmt.Errorf("Invalid model name (must be Pascal case): ", incomingName)
	}
	// Model names must be unique
	for _, mdl := range sv.protoSchema.Models {
		if incomingName == mdl.Name {
			return fmt.Errorf("Invalid model name (already used): ", incomingName)
		}
	}
	return nil
}

func (sv *SchemaValidator) makeFields(parserFields []*parser.ModelField, modelName string) ([]*proto.Field, error) {
	protoFields := []*proto.Field{}
	for _, parserField := range parserFields {
		protoField, err := sv.makeField(parserField, modelName)
		if err != nil {
			return nil, fmt.Errorf("makeField() failed with: %v", err)
		}
		protoFields = append(protoFields, protoField)
	}
	return protoFields, nil
}


func (sv *SchemaValidator) makeField(parserField *parser.ModelField, modelName string) (*proto.Field, error) {
	protoField := &proto.Field{
		ModelName: modelName,
	}

	err := sv.validateFieldName(parserField.Name)
	if err != nil {
		return nil, fmt.Errorf("validateFieldName() failed with: %v", err)
	}
	protoField.Name = parserField.Name

	protoField.Type, err = sv.mapFieldType(parserField.Type)
	if err != nil {
		return nil, fmt.Errorf("mapFieldType() failed with: %v", err)
	}

	protoField.Attributes, err = sv.makeAttributes(parserField.Attributes)
	if err != nil {
		return nil, fmt.Errorf("makeAttributes() failed with: %v", err)
	}

	return protoField, nil
}
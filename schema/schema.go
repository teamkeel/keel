package schema

import (
	"fmt"

	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/validation"
)

// A Schema knows how to produce a (validated) proto.Schema,
// from a given Keel Schema. Construct one, then call the Make method.
type Schema struct {
	keelSchema string
}

func NewSchema(keelSchema string) *Schema {
	return &Schema{
		keelSchema: keelSchema,
	}
}

// Make constructs a proto.Schema from the Keel Schema provided
// at construction time.
func (scm *Schema) Make() (*proto.Schema, error) {
	// Four mains steps:
	// 1. Parse to AST
	// 2. Insert built-in fields
	// 3. Validate
	// 4. Convert to proto model
	declarations, err := parser.Parse(scm.keelSchema)
	if err != nil {
		return nil, fmt.Errorf("parser.Parse() failed with: %v", err)
	}

	scm.insertBuiltInFields(declarations)

	v := validation.NewValidator(declarations)
	err = v.RunAllValidators()
	if err != nil {
		return nil, fmt.Errorf("RunAllValidators() failed with: %v", err)
	}

	protoModels := scm.makeProtoModels(declarations)
	return protoModels, nil
}

// insertBuiltInFields injects new fields into the parser schema, to represent
// our implicit (or built-in) fields. For example every Model has an <id> field.
func (scm *Schema) insertBuiltInFields(declarations *parser.Schema) {
	for _, decl := range declarations.Declarations {
		if decl.Model == nil {
			continue
		}
		field := &parser.ModelField{
			BuiltIn: true,
			Name:    "id", // todo - replace magic string with a more widely shared const.
			Type:    parser.FieldTypeID,
			Attributes: []*parser.Attribute{
				{
					Name: "primaryKey",
				},
			},
		}
		section := &parser.ModelSection{
			Fields: []*parser.ModelField{field},
		}
		model := decl.Model
		model.Sections = append(model.Sections, section)
	}
}

// makeProtoModels derives and returns a proto.Schema from the given (known to be valid) parsed AST.
func (scm *Schema) makeProtoModels(parserSchema *parser.Schema) *proto.Schema {
	protoSchema := &proto.Schema{}

	for _, decl := range parserSchema.Declarations {
		if decl.Model == nil {
			continue
		}
		protoModel := scm.makeModel(decl)
		protoSchema.Models = append(protoSchema.Models, protoModel)
	}
	return protoSchema
}

func (scm *Schema) makeModel(decl *parser.Declaration) *proto.Model {
	parserModel := decl.Model
	protoModel := &proto.Model{
		Name: parserModel.Name,
	}
	for _, section := range parserModel.Sections {
		switch {

		case section.Fields != nil:
			protoModel.Fields = scm.makeFields(section.Fields, protoModel.Name)

		case section.Functions != nil:
			protoModel.Operations = scm.makeOperations(section.Functions, protoModel.Name, proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM)

		case section.Operations != nil:
			protoModel.Operations = scm.makeOperations(section.Operations, protoModel.Name, proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO)

		case section.Attribute != nil:
			// TODO: implement support for attributes on model
		default:
			panic("unrecognized case")
		}
	}

	return protoModel
}

func (scm *Schema) makeFields(parserFields []*parser.ModelField, modelName string) []*proto.Field {
	protoFields := []*proto.Field{}
	for _, parserField := range parserFields {
		protoField := scm.makeField(parserField, modelName)
		protoFields = append(protoFields, protoField)
	}
	return protoFields
}

func (scm *Schema) makeField(parserField *parser.ModelField, modelName string) *proto.Field {
	protoField := &proto.Field{
		ModelName: modelName,
		Name:      parserField.Name,
		Type:      proto.FieldType_FIELD_TYPE_BOOL, // todo need to map parserField.Type,
	}
	// todo protoField.Attributes = nil // todo
	return protoField
}

func (scm *Schema) makeOperations(parserFunctions []*parser.ModelAction, modelName string, impl proto.OperationImplementation) []*proto.Operation {
	protoOps := []*proto.Operation{}
	for _, parserFunc := range parserFunctions {
		protoOp := scm.makeOp(parserFunc, modelName, impl)
		protoOps = append(protoOps, protoOp)
	}
	return protoOps
}

func (scm *Schema) makeOp(parserFunction *parser.ModelAction, modelName string, impl proto.OperationImplementation) *proto.Operation {

	protoOp := &proto.Operation{
		ModelName:      modelName,
		Name:           parserFunction.Name,
		Implementation: impl,
	}

	// Todo the proto type also supports other operation types - like "delete", but don't know how to choose them
	protoOp.Type = proto.OperationType_OPERATION_TYPE_GET
	if parserFunction.Type == parser.ActionTypeCreate {
		protoOp.Type = proto.OperationType_OPERATION_TYPE_CREATE
	}

	protoOp.Inputs = scm.makeOpInputs(parserFunction)

	// todo protoOp.Attributes = nil // todo

	return protoOp
}

func (scm *Schema) makeOpInputs(parserFunction *parser.ModelAction) []*proto.OperationInput {
	// todo - a bit lost here
	return nil
}

package schema2model

import (
	"fmt"

	"github.com/teamkeel/keel/internal/validation"
	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
)

// A Schema2Model knows how to produce a (validated) proto.Schema,
// from a given Keel Schema.
type Schema2Model struct {
	keelSchema string
}

func NewSchema2Model(keelSchema string) *Schema2Model {
	return &Schema2Model{
		keelSchema: keelSchema,
	}
}

// Make constructs a proto.Schema from the Keel Schema provided
// at construction time.
func (scm *Schema2Model) Make() (*proto.Schema, error) {
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
	return protoModels, err
}

// insertBuiltInFields injects new fields into the parser schema, to represent
// our implicit (or built-in) fields. For example every Model has an <id> field.
func (scm *Schema2Model) insertBuiltInFields(declarations *parser.Schema) {
	for _, decl := range declarations.Declarations {
		field := &parser.ModelField{
			Name: "ID",            // todo - replace magic string with a more widely shared const.
			Type: "field_type_id", // todo what is the proper type?
		}
		section := &parser.ModelSection{
			Fields: []*parser.ModelField{field},
		}
		model := decl.Model
		model.Sections = append(model.Sections, section)
	}
}

// makeProtoModels derives and returns a proto.Schema from the given (known to be valid) parsed AST.
func (scm *Schema2Model) makeProtoModels(parserSchema *parser.Schema) *proto.Schema {
	protoSchema := &proto.Schema{}

	for _, decl := range parserSchema.Declarations {
		protoModel := scm.makeModel(decl)
		protoSchema.Models = append(protoSchema.Models, protoModel)
	}
	return protoSchema
}

func (scm *Schema2Model) makeModel(decl *parser.Declaration) *proto.Model {
	parserModel := decl.Model
	protoModel := &proto.Model{
		Name: parserModel.Name,
	}
	for _, section := range parserModel.Sections {
		switch {

		case section.Fields != nil:
			protoModel.Fields = scm.makeFields(section.Fields, protoModel.Name)

		case section.Functions != nil:
			protoModel.Operations = scm.makeOperations(section.Functions)

		case section.Attribute != nil:
			panic("not implemented yet")
		default:
			panic("unrecognized case")
		}
	}

	return protoModel
}

func (scm *Schema2Model) makeFields(parserFields []*parser.ModelField, modelName string) []*proto.Field {
	protoFields := []*proto.Field{}
	for _, parserField := range parserFields {
		protoField := scm.makeField(parserField, modelName)
		protoFields = append(protoFields, protoField)
	}
	return protoFields
}

func (scm *Schema2Model) makeField(parserField *parser.ModelField, modelName string) *proto.Field {
	protoField := &proto.Field{
		ModelName: modelName,
		Name:      parserField.Name,
		Type:      proto.FieldType_FIELD_TYPE_BOOL, // todo need to map parserField.Type,
	}
	protoField.Attributes = nil // todo
	return protoField
}

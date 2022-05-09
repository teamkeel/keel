package schema

import (
	"fmt"
	"io/ioutil"

	"github.com/teamkeel/keel/inputs"
	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/validation"
)

// A Schema knows how to produce a (validated) proto.Schema,
// from a given Keel Schema. Construct one, then call the Make method.
type Schema struct {
}

// MakeFromDirectory constructs a proto.Schema from the .keel files present in the given
// directory.
func (scm *Schema) MakeFromDirectory(directory string) (*proto.Schema, error) {
	allInputFiles, err := inputs.Assemble(directory)
	if err != nil {
		return nil, fmt.Errorf("Error assembling input files: %v", err)
	}
	return scm.makeFromInputs(allInputFiles)
}


// MakeFromFile constructs a proto.Schema from the given .keel file.
func (scm *Schema) MakeFromFile(filename string) (*proto.Schema, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %v", err)
	}
	schemaFile := inputs.SchemaFile{
		FileName: filename,
		Contents: string(fileBytes),
	}
	allInputFiles := &inputs.Inputs{
		Directory: "Unspecified",
		SchemaFiles: []inputs.SchemaFile{schemaFile},
	}
	return scm.makeFromInputs(allInputFiles)
}

func (scm *Schema) makeFromInputs(allInputFiles *inputs.Inputs) (*proto.Schema, error) {
	// - For each of the .keel (schema) files specified...
	// 		- Parse to AST
	// 		- Add built-in fields
	// - With the parsed (AST) schemas as a set:
	// 		- Validate them (as a set)
	// 		- Convert the set to a single / aggregate proto model
	validationInputs := []validation.Input{}
	for _, oneInputSchemaFile := range allInputFiles.SchemaFiles {
		declarations, err := parser.Parse(oneInputSchemaFile.Contents)
		if err != nil {
			return nil, fmt.Errorf("parser.Parse() failed on file: %s, with error %v", oneInputSchemaFile.FileName, err)
		}
		scm.insertBuiltInFields(declarations)
		validationInputs = append(validationInputs, validation.Input{
			FileName: oneInputSchemaFile.FileName,
			ParsedSchema: declarations,
		})
	}

	v := validation.NewValidator(validationInputs)
	err := v.RunAllValidators()
	if err != nil {
		return nil, fmt.Errorf("RunAllValidators() failed with: %v", err)
	}

	validatedSchemas := []*parser.Schema{}
	for _, vs := range validationInputs {
		validatedSchemas = append(validatedSchemas, vs.ParsedSchema)
	}
	protoModels := scm.makeProtoModels(validatedSchemas)
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
func (scm *Schema) makeProtoModels(parserSchemas []*parser.Schema) *proto.Schema {
	protoSchema := &proto.Schema{}

	for _, parserSchema := range parserSchemas {
		for _, decl := range parserSchema.Declarations {
			if decl.Model == nil {
				continue
			}
			protoModel := scm.makeModel(decl)
			protoSchema.Models = append(protoSchema.Models, protoModel)
		}
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

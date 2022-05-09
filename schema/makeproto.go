package schema

import (

	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
)

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
			scm.applyModelAttributes(parserModel, protoModel, section.Attribute)
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
	// todo:
	// set optional if attr
	// set unique if attr
	// 

	// Todo the proto type also supports other operation types - like "delete", but don't know how to choose them
	protoOp.Type = proto.OperationType_OPERATION_TYPE_GET
	if parserFunction.Type == parser.ActionTypeCreate {
		protoOp.Type = proto.OperationType_OPERATION_TYPE_CREATE
	}

	protoOp.Inputs = scm.makeArguments(parserFunction)
	scm.applyFunctionAttributes(parserFunction, protoOp)

	return protoOp
}

func (scm *Schema) makeArguments(parserFunction *parser.ModelAction) []*proto.OperationInput {
	// todo - for each, then
	// LHS, RHS and Operation
	return nil
}
 
func (scm *Schema) applyModelAttributes(parserModel *parser.Model, protoModel *proto.Model, attribute *parser.Attribute) {
	// todo - think we need to upgrade the protobuf model structure to support this
}

func (scm *Schema) applyFunctionAttributes(parserFunction *parser.ModelAction, protoOperation *proto.Operation) {
	// todo
}



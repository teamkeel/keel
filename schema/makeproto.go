package schema

import (
	"fmt"

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// makeProtoModels derives and returns a proto.Schema from the given (known to be valid) parsed AST.
func (scm *Schema) makeProtoModels(parserSchemas []*parser.Schema) *proto.Schema {
	protoSchema := &proto.Schema{}

	for _, parserSchema := range parserSchemas {
		for _, decl := range parserSchema.Declarations {
			switch {
			case decl.Model != nil:
				protoModel := scm.makeModel(decl)
				protoSchema.Models = append(protoSchema.Models, protoModel)
			case decl.API != nil:
				// todo API not yet supported in proto
			default:
				panic("Case not recognized")
			}
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
			scm.applyModelAttribute(parserModel, protoModel, section.Attribute)
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
	}

	// We establish the field type when possible using the 1:1 mapping between parser enums
	// and proto enums. However, when the parsed field type is not one of the built in types, we
	// infer that it must refer to one of the Models defined in the schema, and is therefore of type
	// relationship.
	switch parserField.Type {
	case parser.FieldTypeBoolean:
		protoField.Type = proto.FieldType_FIELD_TYPE_BOOL
	case parser.FieldTypeText:
		protoField.Type = proto.FieldType_FIELD_TYPE_STRING
	case parser.FieldTypeCurrency:
		protoField.Type = proto.FieldType_FIELD_TYPE_CURRENCY
	case parser.FieldTypeDate:
		protoField.Type = proto.FieldType_FIELD_TYPE_DATE
	case parser.FieldTypeDatetime:
		protoField.Type = proto.FieldType_FIELD_TYPE_DATETIME
	case parser.FieldTypeEnum:
		protoField.Type = proto.FieldType_FIELD_TYPE_ENUM
	case parser.FieldTypeID:
		protoField.Type = proto.FieldType_FIELD_TYPE_ID
	case parser.FieldTypeImage:
		protoField.Type = proto.FieldType_FIELD_TYPE_IMAGE
	case parser.FieldTypeNumber:
		protoField.Type = proto.FieldType_FIELD_TYPE_INT
	case parser.FieldTypeIdentity:
		protoField.Type = proto.FieldType_FIELD_TYPE_IDENTITY
	default:
		protoField.Type = proto.FieldType_FIELD_TYPE_RELATIONSHIP
	}
	scm.applyFieldAttributes(parserField, protoField)
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

	switch parserFunction.Type {
	case parser.ActionTypeCreate:
		protoOp.Type = proto.OperationType_OPERATION_TYPE_CREATE
	case parser.ActionTypeUpdate:
		protoOp.Type = proto.OperationType_OPERATION_TYPE_UPDATE
	case parser.ActionTypeGet:
		protoOp.Type = proto.OperationType_OPERATION_TYPE_GET
	case parser.ActionTypeList:
		protoOp.Type = proto.OperationType_OPERATION_TYPE_LIST
	default:
		panic("Action type not recognized")
	}

	protoOp.Inputs = scm.makeArguments(parserFunction, modelName)
	scm.applyFunctionAttributes(parserFunction, protoOp)

	return protoOp
}

func (scm *Schema) makeArguments(parserFunction *parser.ModelAction, modelName string) []*proto.OperationInput {
	// Currently, we only support arguments of the form <modelname>.
	operationInputs := []*proto.OperationInput{}
	for _, parserArg := range parserFunction.Arguments {
		operationInput := proto.OperationInput{
			Name:      parserArg.Name,
			Type:      proto.OperationInputType_OPERATION_INPUT_TYPE_FIELD,
			ModelName: wrapperspb.String(parserArg.Name),
			FieldName: wrapperspb.String(parserArg.Name),
		}
		operationInputs = append(operationInputs, &operationInput)
	}

	return operationInputs
}

func (scm *Schema) applyModelAttribute(parserModel *parser.Model, protoModel *proto.Model, attribute *parser.Attribute) {
	switch attribute.Name {
	case parser.AttributePermission:
		scm.applyModelPermission(attribute, protoModel)
	}
}

func (scm *Schema) applyFunctionAttributes(parserFunction *parser.ModelAction, protoOperation *proto.Operation) {
	for _, attribute := range parserFunction.Attributes {
		scm.applyFunctionAttribute(attribute, protoOperation)
	}
}

func (scm *Schema) applyFunctionAttribute(attribute *parser.Attribute, protoOperation *proto.Operation) {
	// permission, where, or set
	switch attribute.Name {
	case parser.AttributePermission:
		// todo await attr/expr support in parser
	case parser.AttributeWhere:
		// todo await attr/exp support in parser
	case parser.AttributeSet:
		// todo await attr/exp support in parser
	}
}

func (scm *Schema) applyFieldAttributes(parserField *parser.ModelField, protoField *proto.Field) {
	for _, fieldAttribute := range parserField.Attributes {
		switch fieldAttribute.Name {
		case parser.AttributeUnique:
			protoField.Unique = true
		case parser.AttributeOptional:
			protoField.Optional = true
		}
	}
}

func (scm *Schema) applyModelPermission(permissionAttribute *parser.Attribute, protoModel *proto.Model) {
	args := permissionAttribute.Arguments
	switch {
	// The first form we support is Permission(conditional, actions)
	case len(args) == 2 && args[0].Expression != nil:
		// todo - see if we can remove error from ToString return values
		conditional, _ := expressions.ToString(args[0].Expression)
		actions, _ := expressions.ToString(args[1].Expression)

		permissionRule := &proto.PermissionRule{
			ModelName: protoModel.Name,
			OperationName: nil,
			RoleName: nil,
			Expression: conditional,
			OperationsTypes: actions,
		}

		protoModel.Permissions = []*proto.PermissionRule{permissionRule}

	// todo - extend cases to support Model permissions of the form @Permission(role, actions)
	default:
		panic("Permission attribute malformed")
	}
}

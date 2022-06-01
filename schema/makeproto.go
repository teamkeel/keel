package schema

import (
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// makeProtoModels derives and returns a proto.Schema from the given (known to be valid) set of parsed AST.
func (scm *Schema) makeProtoModels(parserSchemas []*parser.Schema) *proto.Schema {
	protoSchema := &proto.Schema{}

	for _, parserSchema := range parserSchemas {
		for _, decl := range parserSchema.Declarations {
			switch {
			case decl.Model != nil:
				protoModel := scm.makeModel(decl)
				protoSchema.Models = append(protoSchema.Models, protoModel)
			case decl.Role != nil:
				protoRole := scm.makeRole(decl)
				protoSchema.Roles = append(protoSchema.Roles, protoRole)
			case decl.API != nil:
				protoAPI := scm.makeAPI(decl)
				protoSchema.Apis = append(protoSchema.Apis, protoAPI)
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
			fields := scm.makeFields(section.Fields, protoModel.Name)
			protoModel.Fields = append(protoModel.Fields, fields...)

		case section.Functions != nil:
			ops := scm.makeOperations(section.Functions, protoModel.Name, proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM)
			protoModel.Operations = append(protoModel.Operations, ops...)

		case section.Operations != nil:
			ops := scm.makeOperations(section.Operations, protoModel.Name, proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO)
			protoModel.Operations = append(protoModel.Operations, ops...)

		case section.Attribute != nil:
			scm.applyModelAttribute(parserModel, protoModel, section.Attribute)
		default:
			panic("unrecognized case")
		}
	}

	return protoModel
}

func (scm *Schema) makeRole(decl *parser.Declaration) *proto.Role {
	parserRole := decl.Role
	protoRole := &proto.Role{
		Name: parserRole.Name,
	}
	for _, section := range parserRole.Sections {
		for _, parserDomain := range section.Domains {
			protoRole.Domains = append(protoRole.Domains, stripQuotes(parserDomain.Domain))
		}
		for _, parserEmail := range section.Emails {
			protoRole.Emails = append(protoRole.Emails, stripQuotes(parserEmail.Email))
		}
	}
	return protoRole
}

func (scm *Schema) makeAPI(decl *parser.Declaration) *proto.Api {
	parserAPI := decl.API
	protoAPI := &proto.Api{
		Name:      parserAPI.Name,
		ApiModels: []*proto.ApiModel{},
	}
	for _, section := range parserAPI.Sections {
		switch {
		case section.Attribute != nil:
			protoAPI.Type = scm.mapToAPIType(section.Attribute.Name)
		case len(section.Models) > 0:
			for _, parserApiModel := range section.Models {
				protoModel := &proto.ApiModel{
					ModelName: parserApiModel.ModelName,
				}
				protoAPI.ApiModels = append(protoAPI.ApiModels, protoModel)
			}
		}
	}
	return protoAPI
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
		Type:           scm.mapToOperationType(parserFunction.Type),
	}
	protoOp.Inputs = scm.makeArguments(parserFunction, modelName)
	scm.applyFunctionAttributes(parserFunction, protoOp, modelName)

	return protoOp
}

func (scm *Schema) makeArguments(parserFunction *parser.ModelAction, modelName string) []*proto.OperationInput {
	// Currently, we only support arguments of the form <modelname>.
	operationInputs := []*proto.OperationInput{}
	for _, parserArg := range parserFunction.Arguments {
		operationInput := proto.OperationInput{
			Name:      parserArg.Name,
			Type:      proto.OperationInputType_OPERATION_INPUT_TYPE_FIELD,
			ModelName: wrapperspb.String(modelName),
			FieldName: wrapperspb.String(parserArg.Name),
		}
		operationInputs = append(operationInputs, &operationInput)
	}
	return operationInputs
}

func (scm *Schema) applyModelAttribute(parserModel *parser.Model, protoModel *proto.Model, attribute *parser.Attribute) {
	switch attribute.Name {
	case parser.AttributePermission:
		perm := scm.permissionAttributeToProtoPermission(attribute)
		perm.ModelName = protoModel.Name
		protoModel.Permissions = append(protoModel.Permissions, perm)
	}
}

func (scm *Schema) applyFunctionAttributes(parserFunction *parser.ModelAction, protoOperation *proto.Operation, modelName string) {
	for _, attribute := range parserFunction.Attributes {
		switch attribute.Name {
		case parser.AttributePermission:
			perm := scm.permissionAttributeToProtoPermission(attribute)
			perm.ModelName = modelName
			perm.OperationName = wrapperspb.String(protoOperation.Name)
			protoOperation.Permissions = append(protoOperation.Permissions, perm)
		case parser.AttributeWhere:
			// todo hope to remove error return from ToString
			expr, _ := expressions.ToString(attribute.Arguments[0].Expression)
			where := &proto.Expression{Source: expr}
			protoOperation.WhereExpressions = append(protoOperation.WhereExpressions, where)
		case parser.AttributeSet:
			// todo hope to remove error return from ToString
			expr, _ := expressions.ToString(attribute.Arguments[0].Expression)
			set := &proto.Expression{Source: expr}
			protoOperation.SetExpressions = append(protoOperation.SetExpressions, set)
		}
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

func (scm *Schema) permissionAttributeToProtoPermission(attr *parser.Attribute) *proto.PermissionRule {
	pr := &proto.PermissionRule{}
	for _, arg := range attr.Arguments {
		switch arg.Name {
		// todo use parser constants for "expression" etc below
		case "expression":
			expr, _ := expressions.ToString(arg.Expression)
			pr.Expression = &proto.Expression{Source: expr}
		case "role":
			value, _ := expressions.ToValue(arg.Expression)
			pr.RoleName = value.Ident[0]
		case "actions":
			value, _ := expressions.ToValue(arg.Expression)
			for _, v := range value.Array.Values {
				pr.OperationsTypes = append(pr.OperationsTypes, scm.mapToOperationType(v.Ident[0]))
			}
		}
	}
	return pr
}

func (scm *Schema) mapToOperationType(parsedOperation string) proto.OperationType {
	switch parsedOperation {
	case parser.ActionTypeCreate:
		return proto.OperationType_OPERATION_TYPE_CREATE
	case parser.ActionTypeUpdate:
		return proto.OperationType_OPERATION_TYPE_UPDATE
	case parser.ActionTypeGet:
		return proto.OperationType_OPERATION_TYPE_GET
	case parser.ActionTypeList:
		return proto.OperationType_OPERATION_TYPE_LIST
	default:
		return proto.OperationType_OPERATION_TYPE_UNKNOWN
	}
}

// stripQuotes removes all double quotes from the given string, regardless of where they are.
func stripQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}

func (scm *Schema) mapToAPIType(parserAPIType string) proto.ApiType {
	switch parserAPIType {
	case parser.APITypeGraphQL:
		return proto.ApiType_API_TYPE_GRAPHQL
	case parser.APITypeRPC:
		return proto.ApiType_API_TYPE_RPC
	default:
		return proto.ApiType_API_TYPE_UNKNOWN
	}
}

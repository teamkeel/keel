package schema

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// makeProtoModels derives and returns a proto.Schema from the given (known to be valid) set of parsed AST.
func (scm *Builder) makeProtoModels() *proto.Schema {
	protoSchema := &proto.Schema{}

	for _, parserSchema := range scm.asts {
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
			case decl.Enum != nil:
				protoEnum := scm.makeEnum(decl)
				protoSchema.Enums = append(protoSchema.Enums, protoEnum)
			default:
				panic("Case not recognized")
			}
		}
	}
	// Now that we've built all the models from all the input schema files, we have a global set,
	// and we can finish up populating some of the data concerned with relationship fields.

	// Alongside every HasOne relationship field of type Model defined in the schema,
	// we auto-generate another field in the same model to carry the foreign key values
	// that can hold the "pointer" to  the target related object. In order to name these new fields,
	// we need to access the related model and find out which of its fields is its primary key.
	// (Canonically "id"). E.g. if the HasOne field is named "author", and the related model's primary key
	// is "id", then we name the new FK field "authorId".
	affectedFields := calculateForeignKeyFieldNames(protoSchema)

	// Now we go ahead and actually auto-insert the new FK fields.
	createRelationshipIdFields(protoSchema, affectedFields)

	return protoSchema
}

func (scm *Builder) makeModel(decl *parser.DeclarationNode) *proto.Model {
	parserModel := decl.Model
	protoModel := &proto.Model{
		Name: parserModel.Name.Value,
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
			// this is possible if the user defines an empty block in the schema e.g. "fields {}"
			// this isn't really an error so we can just ignore these sections
		}
	}

	if decl.Model.Name.Value == parser.ImplicitIdentityModelName {
		protoOp := proto.Operation{
			ModelName:      parser.ImplicitIdentityModelName,
			Name:           parser.ImplicitAuthenticateOperationName,
			Implementation: proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO,
			Type:           proto.OperationType_OPERATION_TYPE_AUTHENTICATE,
			Inputs: []*proto.OperationInput{
				{
					ModelName:     parser.ImplicitIdentityModelName,
					OperationName: parser.ImplicitAuthenticateOperationName,
					Name:          "createIfNotExists",
					Type:          &proto.TypeInfo{Type: proto.Type_TYPE_BOOL},
					Optional:      true,
				},
				{
					ModelName:     parser.ImplicitIdentityModelName,
					OperationName: parser.ImplicitAuthenticateOperationName,
					Name:          "emailPassword",
					Type:          &proto.TypeInfo{Type: proto.Type_TYPE_OBJECT},
					Optional:      false,
					Inputs: []*proto.OperationInput{
						{
							ModelName:     parser.ImplicitIdentityModelName,
							OperationName: parser.ImplicitAuthenticateOperationName,
							Name:          "email",
							Type:          &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
							Optional:      false,
						},
						{
							ModelName:     parser.ImplicitIdentityModelName,
							OperationName: parser.ImplicitAuthenticateOperationName,
							Name:          "password",
							Type:          &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
							Optional:      false,
						},
					},
				},
			},
			Outputs: []*proto.OperationOutput{
				{
					ModelName:     parser.ImplicitIdentityModelName,
					OperationName: parser.ImplicitAuthenticateOperationName,
					Name:          "identityCreated",
					Type:          &proto.TypeInfo{Type: proto.Type_TYPE_BOOL},
				},
				{
					ModelName:     parser.ImplicitIdentityModelName,
					OperationName: parser.ImplicitAuthenticateOperationName,
					Name:          "token",
					Type:          &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				},
			},
		}

		protoModel.Operations = append(protoModel.Operations, &protoOp)
	}

	return protoModel
}

func (scm *Builder) makeRole(decl *parser.DeclarationNode) *proto.Role {
	parserRole := decl.Role
	protoRole := &proto.Role{
		Name: parserRole.Name.Value,
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

func (scm *Builder) makeAPI(decl *parser.DeclarationNode) *proto.Api {
	parserAPI := decl.API
	protoAPI := &proto.Api{
		Name:      parserAPI.Name.Value,
		ApiModels: []*proto.ApiModel{},
	}
	for _, section := range parserAPI.Sections {
		switch {
		case section.Attribute != nil:
			protoAPI.Type = scm.mapToAPIType(section.Attribute.Name.Value)
		case len(section.Models) > 0:
			for _, parserApiModel := range section.Models {
				protoModel := &proto.ApiModel{
					ModelName: parserApiModel.Name.Value,
				}
				protoAPI.ApiModels = append(protoAPI.ApiModels, protoModel)
			}
		}
	}
	return protoAPI
}

func (scm *Builder) makeEnum(decl *parser.DeclarationNode) *proto.Enum {
	parserEnum := decl.Enum
	enum := &proto.Enum{
		Name:   parserEnum.Name.Value,
		Values: []*proto.EnumValue{},
	}
	for _, value := range parserEnum.Values {
		enum.Values = append(enum.Values, &proto.EnumValue{
			Name: value.Name.Value,
		})
	}
	return enum
}

func (scm *Builder) makeFields(parserFields []*parser.FieldNode, modelName string) []*proto.Field {
	protoFields := []*proto.Field{}
	for _, parserField := range parserFields {
		protoField := scm.makeField(parserField, modelName)
		protoFields = append(protoFields, protoField)
	}
	return protoFields
}

func (scm *Builder) makeField(parserField *parser.FieldNode, modelName string) *proto.Field {
	typeInfo := scm.parserFieldToProtoTypeInfo(parserField)

	protoField := &proto.Field{
		ModelName: modelName,
		Name:      parserField.Name.Value,
		Type:      typeInfo,
		Optional:  parserField.Optional,
	}

	// Handle @unique attribute at model level which expresses
	// unique constrains across multiple fields
	model := query.Model(scm.asts, modelName)
	for _, attr := range query.ModelAttributes(model) {
		if attr.Name.Value != parser.AttributeUnique {
			continue
		}

		value, _ := attr.Arguments[0].Expression.ToValue()
		fieldNames := lo.Map(value.Array.Values, func(v *parser.Operand, i int) string {
			return v.Ident.ToString()
		})

		if !lo.Contains(fieldNames, parserField.Name.Value) {
			continue
		}

		protoField.UniqueWith = lo.Filter(fieldNames, func(v string, i int) bool {
			return v != parserField.Name.Value
		})
	}

	scm.applyFieldAttributes(parserField, protoField)
	return protoField
}

func (scm *Builder) makeOperations(parserFunctions []*parser.ActionNode, modelName string, impl proto.OperationImplementation) []*proto.Operation {
	protoOps := []*proto.Operation{}
	for _, parserFunc := range parserFunctions {
		protoOp := scm.makeOp(parserFunc, modelName, impl)
		protoOps = append(protoOps, protoOp)
	}
	return protoOps
}

func (scm *Builder) makeOp(parserFunction *parser.ActionNode, modelName string, impl proto.OperationImplementation) *proto.Operation {
	protoOp := &proto.Operation{
		ModelName:      modelName,
		Name:           parserFunction.Name.Value,
		Implementation: impl,
		Type:           scm.mapToOperationType(parserFunction.Type.Value),
	}

	model := query.Model(scm.asts, modelName)

	for _, input := range parserFunction.Inputs {
		protoInput := scm.makeOperationInput(model, parserFunction, input, proto.InputMode_INPUT_MODE_READ)
		protoOp.Inputs = append(protoOp.Inputs, protoInput)
	}

	for _, input := range parserFunction.With {
		protoInput := scm.makeOperationInput(model, parserFunction, input, proto.InputMode_INPUT_MODE_WRITE)
		protoOp.Inputs = append(protoOp.Inputs, protoInput)
	}

	scm.applyActionAttributes(parserFunction, protoOp, modelName)

	return protoOp
}

func (scm *Builder) makeOperationInput(
	model *parser.ModelNode,
	op *parser.ActionNode,
	input *parser.ActionInputNode,
	mode proto.InputMode,
) (inputs *proto.OperationInput) {

	idents := input.Type.Fragments
	protoType := scm.parserTypeToProtoType(idents[0].Fragment)

	behaviour := proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT
	target := []string{}

	var modelName *wrapperspb.StringValue
	var enumName *wrapperspb.StringValue

	if protoType == proto.Type_TYPE_UNKNOWN {
		behaviour = proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT

		var field *parser.FieldNode
		currModel := model

		for _, ident := range idents {
			target = append(target, ident.Fragment)
			field = query.ModelField(currModel, ident.Fragment)
			currModel = query.Model(scm.asts, field.Type)
		}

		protoType = scm.parserFieldToProtoTypeInfo(field).Type

		if protoType == proto.Type_TYPE_MODEL {
			modelName = &wrapperspb.StringValue{
				Value: field.Type,
			}
		}

		if protoType == proto.Type_TYPE_ENUM {
			enumName = &wrapperspb.StringValue{
				Value: field.Type,
			}
		}
	}

	return &proto.OperationInput{
		ModelName:     model.Name.Value,
		OperationName: op.Name.Value,
		Name:          input.Name(),
		Type: &proto.TypeInfo{
			Type:      protoType,
			Repeated:  input.Repeated,
			ModelName: modelName,
			EnumName:  enumName,
		},
		Optional:  input.Optional,
		Mode:      mode,
		Behaviour: behaviour,
		Target:    target,
	}
}

// parserType could be a built-in type or a user-defined model or enum
func (scm *Builder) parserTypeToProtoType(parserType string) proto.Type {
	switch {
	case parserType == parser.FieldTypeText:
		return proto.Type_TYPE_STRING
	case parserType == parser.FieldTypeID:
		return proto.Type_TYPE_ID
	case parserType == parser.FieldTypeBoolean:
		return proto.Type_TYPE_BOOL
	case parserType == parser.FieldTypeNumber:
		return proto.Type_TYPE_INT
	case parserType == parser.FieldTypeDate:
		return proto.Type_TYPE_DATE
	case parserType == parser.FieldTypeDatetime:
		return proto.Type_TYPE_DATETIME
	case parserType == parser.FieldTypeSecret:
		return proto.Type_TYPE_SECRET
	case parserType == parser.FieldTypePassword:
		return proto.Type_TYPE_PASSWORD
	case query.IsIdentityModel(scm.asts, parserType):
		return proto.Type_TYPE_IDENTITY
	case query.IsModel(scm.asts, parserType):
		return proto.Type_TYPE_MODEL
	case query.IsEnum(scm.asts, parserType):
		return proto.Type_TYPE_ENUM
	default:
		return proto.Type_TYPE_UNKNOWN
	}
}

func (scm *Builder) parserFieldToProtoTypeInfo(field *parser.FieldNode) *proto.TypeInfo {

	protoType := scm.parserTypeToProtoType(field.Type)
	var modelName *wrapperspb.StringValue
	var enumName *wrapperspb.StringValue

	switch protoType {

	case proto.Type_TYPE_MODEL, proto.Type_TYPE_IDENTITY:
		modelName = &wrapperspb.StringValue{
			Value: query.Model(scm.asts, field.Type).Name.Value,
		}
	case proto.Type_TYPE_ENUM:
		enumName = &wrapperspb.StringValue{
			Value: query.Enum(scm.asts, field.Type).Name.Value,
		}
	}

	return &proto.TypeInfo{
		Type:      protoType,
		Repeated:  field.Repeated,
		ModelName: modelName,
		EnumName:  enumName,
	}
}

func (scm *Builder) applyModelAttribute(parserModel *parser.ModelNode, protoModel *proto.Model, attribute *parser.AttributeNode) {
	switch attribute.Name.Value {
	case parser.AttributePermission:
		perm := scm.permissionAttributeToProtoPermission(attribute)
		perm.ModelName = protoModel.Name
		protoModel.Permissions = append(protoModel.Permissions, perm)
	}
}

func (scm *Builder) applyActionAttributes(action *parser.ActionNode, protoOperation *proto.Operation, modelName string) {
	for _, attribute := range action.Attributes {
		switch attribute.Name.Value {
		case parser.AttributePermission:
			perm := scm.permissionAttributeToProtoPermission(attribute)
			perm.ModelName = modelName
			perm.OperationName = wrapperspb.String(protoOperation.Name)
			protoOperation.Permissions = append(protoOperation.Permissions, perm)
		case parser.AttributeWhere:
			expr, _ := attribute.Arguments[0].Expression.ToString()
			where := &proto.Expression{Source: expr}
			protoOperation.WhereExpressions = append(protoOperation.WhereExpressions, where)
		case parser.AttributeSet:
			expr, _ := attribute.Arguments[0].Expression.ToString()
			set := &proto.Expression{Source: expr}
			protoOperation.SetExpressions = append(protoOperation.SetExpressions, set)
		case parser.AttributeValidate:
			expr, _ := attribute.Arguments[0].Expression.ToString()
			set := &proto.Expression{Source: expr}
			protoOperation.ValidationExpressions = append(protoOperation.ValidationExpressions, set)
		}
	}
}

func (scm *Builder) applyFieldAttributes(parserField *parser.FieldNode, protoField *proto.Field) {
	for _, fieldAttribute := range parserField.Attributes {
		switch fieldAttribute.Name.Value {
		case parser.AttributeUnique:
			protoField.Unique = true
		case parser.AttributePrimaryKey:
			protoField.PrimaryKey = true
		case parser.AttributeDefault:
			defaultValue := &proto.DefaultValue{}
			if len(fieldAttribute.Arguments) == 1 {
				expr := fieldAttribute.Arguments[0].Expression
				source, _ := expr.ToString()
				defaultValue.Expression = &proto.Expression{
					Source: source,
				}
			} else {
				defaultValue.UseZeroValue = true
			}
			protoField.DefaultValue = defaultValue
		}
	}
}

func (scm *Builder) permissionAttributeToProtoPermission(attr *parser.AttributeNode) *proto.PermissionRule {
	pr := &proto.PermissionRule{}
	for _, arg := range attr.Arguments {
		switch arg.Label.Value {
		case "expression":
			expr, _ := arg.Expression.ToString()
			pr.Expression = &proto.Expression{Source: expr}
		case "roles":
			value, _ := arg.Expression.ToValue()
			for _, item := range value.Array.Values {
				pr.RoleNames = append(pr.RoleNames, item.Ident.Fragments[0].Fragment)
			}
		case "actions":
			value, _ := arg.Expression.ToValue()
			for _, v := range value.Array.Values {
				pr.OperationsTypes = append(pr.OperationsTypes, scm.mapToOperationType(v.Ident.Fragments[0].Fragment))
			}
		}
	}
	return pr
}

func (scm *Builder) mapToOperationType(parsedOperation string) proto.OperationType {
	switch parsedOperation {
	case parser.ActionTypeCreate:
		return proto.OperationType_OPERATION_TYPE_CREATE
	case parser.ActionTypeUpdate:
		return proto.OperationType_OPERATION_TYPE_UPDATE
	case parser.ActionTypeGet:
		return proto.OperationType_OPERATION_TYPE_GET
	case parser.ActionTypeList:
		return proto.OperationType_OPERATION_TYPE_LIST
	case parser.ActionTypeDelete:
		return proto.OperationType_OPERATION_TYPE_DELETE
	default:
		return proto.OperationType_OPERATION_TYPE_UNKNOWN
	}
}

// stripQuotes removes all double quotes from the given string, regardless of where they are.
func stripQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}

func (scm *Builder) mapToAPIType(parserAPIType string) proto.ApiType {
	switch parserAPIType {
	case parser.AttributeGraphQL:
		return proto.ApiType_API_TYPE_GRAPHQL
	case parser.AttributeRPC:
		return proto.ApiType_API_TYPE_RPC
	default:
		return proto.ApiType_API_TYPE_UNKNOWN
	}
}

// calculateForeignKeyFieldNames locates all the model fields of type Model that represent
// a HasOne relation, calculates, and populates that field's ForeignKeyFieldName value.
// To do so it has to locate the referred-to model to find out which of its fields is
// its primary key. (Because the foreign key name incorporates that).
func calculateForeignKeyFieldNames(schema *proto.Schema) (affectedFields []*proto.Field) {
	hasOneModelFields := allHasOneModelFields(schema)
	for _, thisField := range hasOneModelFields {
		if !proto.IfIsTypeModelAndHasOne(thisField) {
			continue
		}

		relatedModel := proto.FindModel(schema.Models, thisField.Type.ModelName.Value)
		relatedPrimaryKeyFieldName := proto.PrimaryKeyFieldName(relatedModel)
		foreignKeyName := strcase.ToLowerCamel(thisField.Name + "_" + relatedPrimaryKeyFieldName)
		thisField.ForeignKeyFieldName = &wrapperspb.StringValue{
			Value: foreignKeyName,
		}
	}
	return hasOneModelFields
}

// allHasOneModelFields provides a list of all the fields in the schema
// that represent HasOne relationships.
func allHasOneModelFields(schema *proto.Schema) []*proto.Field {
	allFields := proto.AllFields(schema)
	hasOneFields := lo.Filter(allFields, func(field *proto.Field, _ int) bool {
		return proto.IfIsTypeModelAndHasOne(field)
	})
	return hasOneFields
}

// createRelationshipIdFields visits all the given hasOneModelFields (which
// are defined in the Keel Schema), and auto generates a sibling field of type Id in the
// same model to carry the foreign key that references the related object. Note this sibling field does not show up
// in the keel schema - they are implied, similarly to the way each model has an implied CreatedAt
// field.
func createRelationshipIdFields(schema *proto.Schema, hasOneModelFields []*proto.Field) {
	for _, hasOneModelField := range hasOneModelFields {
		relnIdField := &proto.Field{
			// The new field belongs to the same model as the hasOneModelField.
			ModelName: hasOneModelField.ModelName,

			// The hasOneModelField tells us explicitly what to call this new Id field.
			Name: hasOneModelField.ForeignKeyFieldName.Value,

			Type: &proto.TypeInfo{
				// The new Id field is of type ID.
				Type: proto.Type_TYPE_ID,

				// The hasOneModelField can tell us which related model this Id value refers to.
				ModelName: hasOneModelField.Type.ModelName,

				// The new Id field has a hard-coded name - i.e. ID.
				FieldName: &wrapperspb.StringValue{Value: parser.ImplicitFieldNameId},
			},
		}
		thisModel := proto.FindModel(schema.Models, hasOneModelField.ModelName)
		thisModel.Fields = append(thisModel.Fields, relnIdField)
	}
}

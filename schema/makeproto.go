package schema

import (
	"fmt"
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
	scm.proto = &proto.Schema{}

	for _, parserSchema := range scm.asts {
		for _, decl := range parserSchema.Declarations {
			switch {
			case decl.Model != nil:
				scm.makeModel(decl)
			case decl.Role != nil:
				scm.makeRole(decl)
			case decl.API != nil:
				scm.makeAPI(decl)
			case decl.Enum != nil:
				scm.makeEnum(decl)
			default:
				panic("Case not recognized")
			}
		}

		for _, envVar := range parserSchema.EnvironmentVariables {
			scm.proto.EnvironmentVariables = append(scm.proto.EnvironmentVariables, &proto.EnvironmentVariable{
				Name: envVar,
			})
		}
	}

	return scm.proto
}

// Adds a set of proto.Messages to top level Messages registry for all inputs of an Action
func (scm *Builder) makeActionInputMessages(model *parser.ModelNode, action *parser.ActionNode, impl proto.OperationImplementation) {
	switch action.Type.Value {
	case parser.ActionTypeCreate:
		values := []*proto.MessageField{}

		for _, value := range action.With {
			typeInfo, target := scm.inferParserInputType(model, action, value, impl)

			values = append(values, &proto.MessageField{
				Name:     value.Name(),
				Type:     typeInfo,
				Target:   target,
				Optional: value.Optional,
			})
		}

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name:   fmt.Sprintf("%sInput", strcase.ToCamel(action.Name.Value)),
			Fields: values,
		})
	case parser.ActionTypeGet, parser.ActionTypeDelete:
		fields := []*proto.MessageField{}

		for _, input := range action.Inputs {
			typeInfo, target := scm.inferParserInputType(model, action, input, impl)

			fields = append(fields, &proto.MessageField{
				Name:     input.Name(),
				Type:     typeInfo,
				Target:   target,
				Optional: input.Optional,
			})
		}

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name:   fmt.Sprintf("%sInput", strcase.ToCamel(action.Name.Value)),
			Fields: fields,
		})
	case parser.ActionTypeUpdate:
		wheres := []*proto.MessageField{}

		for _, where := range action.Inputs {
			typeInfo, target := scm.inferParserInputType(model, action, where, impl)

			wheres = append(wheres, &proto.MessageField{
				Name:     where.Name(),
				Type:     typeInfo,
				Target:   target,
				Optional: where.Optional,
			})
		}

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name:   fmt.Sprintf("%sWhereInput", strcase.ToCamel(action.Name.Value)),
			Fields: wheres,
		})

		values := []*proto.MessageField{}

		for _, value := range action.With {
			typeInfo, target := scm.inferParserInputType(model, action, value, impl)

			values = append(values, &proto.MessageField{
				Name:     value.Name(),
				Type:     typeInfo,
				Target:   target,
				Optional: value.Optional,
			})
		}

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name:   fmt.Sprintf("%sValuesInput", strcase.ToCamel(action.Name.Value)),
			Fields: values,
		})
		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name: fmt.Sprintf("%sInput", strcase.ToCamel(action.Name.Value)),
			Fields: []*proto.MessageField{
				{
					Name: "where",
					Type: &proto.TypeInfo{
						MessageName: wrapperspb.String(fmt.Sprintf("%sWhereInput", strcase.ToCamel(action.Name.Value))),
					},
				},
				{
					Name: "values",
					Type: &proto.TypeInfo{
						MessageName: wrapperspb.String(fmt.Sprintf("%sValuesInput", strcase.ToCamel(action.Name.Value))),
					},
				},
			},
		})
	case parser.ActionTypeList:
		wheres := []*proto.MessageField{}

		for _, where := range action.Inputs {
			typeInfo, target := scm.inferParserInputType(model, action, where, impl)

			wheres = append(wheres, &proto.MessageField{
				Name:     where.Name(),
				Type:     typeInfo,
				Target:   target,
				Optional: where.Optional,
			})
		}

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name:   fmt.Sprintf("%sWhereInput", strcase.ToCamel(action.Name.Value)),
			Fields: wheres,
		})

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name: fmt.Sprintf("%sInput", strcase.ToCamel(action.Name.Value)),
			Fields: []*proto.MessageField{
				{
					Name: "where",
					Type: &proto.TypeInfo{
						MessageName: wrapperspb.String(fmt.Sprintf("%sWhereInput", strcase.ToCamel(action.Name.Value))),
					},
				},
			},
		})
	}
}

func (scm *Builder) makeModel(decl *parser.DeclarationNode) {
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
			ModelName:           parser.ImplicitIdentityModelName,
			Name:                parser.ImplicitAuthenticateOperationName,
			Implementation:      proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO,
			Type:                proto.OperationType_OPERATION_TYPE_AUTHENTICATE,
			InputMessageName:    "AuthenticateInput",
			ResponseMessageName: "AuthenticateResponse",
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

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name: "EmailPasswordInput",
			Fields: []*proto.MessageField{
				{
					Name:     "email",
					Type:     &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
					Optional: false,
				},
				{
					Name:     "password",
					Type:     &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
					Optional: false,
				},
			},
		})

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name: "AuthenticateInput",
			Fields: []*proto.MessageField{
				{
					Name:     "createIfNotExists",
					Type:     &proto.TypeInfo{Type: proto.Type_TYPE_BOOL},
					Optional: true,
				},
				{
					Name:     "emailPassword",
					Type:     &proto.TypeInfo{Type: proto.Type_TYPE_MESSAGE, MessageName: wrapperspb.String("EmailPasswordInput")},
					Optional: false,
				},
			},
		})

		scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
			Name: "AuthenticateResponse",
			Fields: []*proto.MessageField{
				{
					Name:     "identityCreated",
					Type:     &proto.TypeInfo{Type: proto.Type_TYPE_BOOL},
					Optional: true,
				},
				{
					Name:     "token",
					Type:     &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
					Optional: false,
				},
			},
		})

		protoModel.Operations = append(protoModel.Operations, &protoOp)
	}

	scm.proto.Models = append(scm.proto.Models, protoModel)
}

func (scm *Builder) makeRole(decl *parser.DeclarationNode) {
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
	scm.proto.Roles = append(scm.proto.Roles, protoRole)
}

func (scm *Builder) makeAPI(decl *parser.DeclarationNode) {
	parserAPI := decl.API
	protoAPI := &proto.Api{
		Name:      parserAPI.Name.Value,
		ApiModels: []*proto.ApiModel{},
	}
	for _, section := range parserAPI.Sections {
		switch {
		case len(section.Models) > 0:
			for _, parserApiModel := range section.Models {
				protoModel := &proto.ApiModel{
					ModelName: parserApiModel.Name.Value,
				}
				protoAPI.ApiModels = append(protoAPI.ApiModels, protoModel)
			}
		}
	}
	scm.proto.Apis = append(scm.proto.Apis, protoAPI)
}

func (scm *Builder) makeEnum(decl *parser.DeclarationNode) {
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
	scm.proto.Enums = append(scm.proto.Enums, enum)
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

	// Apply data about relationship fields captured during the parsing phase.

	if fki := parserField.FkInfo; fki != nil {
		switch {
		// This is a foreign key field
		case protoField.Name == fki.ForeignKeyField.Name.Value:
			protoField.ForeignKeyInfo = &proto.ForeignKeyInfo{
				RelatedModelName:  fki.ReferredToModel.Name.Value,
				RelatedModelField: fki.ReferredToModelPrimaryKey.Name.Value,
			}

			// This is model field that "owns" a FK field.
		case protoField.Name == fki.OwningField.Name.Value:
			protoField.ForeignKeyFieldName = wrapperspb.String(fki.ForeignKeyField.Name.Value)
		}
	}

	return protoField
}

func (scm *Builder) makeOperations(parserFunctions []*parser.ActionNode, modelName string, impl proto.OperationImplementation) []*proto.Operation {
	protoOps := []*proto.Operation{}
	for _, parserFunc := range parserFunctions {
		protoOp := scm.makeOperation(parserFunc, modelName, impl)
		protoOps = append(protoOps, protoOp)
	}
	return protoOps
}

func (scm *Builder) makeOperation(parserFunction *parser.ActionNode, modelName string, impl proto.OperationImplementation) *proto.Operation {
	protoOp := &proto.Operation{
		ModelName:      modelName,
		Name:           parserFunction.Name.Value,
		Implementation: impl,
		Type:           scm.mapToOperationType(parserFunction.Type.Value),
	}

	model := query.Model(scm.asts, modelName)

	scm.makeActionInputMessages(model, parserFunction, impl)

	for _, input := range parserFunction.Inputs {
		protoInput := scm.makeOperationInput(model, parserFunction, input, proto.InputMode_INPUT_MODE_READ, impl)
		protoOp.Inputs = append(protoOp.Inputs, protoInput)
	}

	for _, input := range parserFunction.With {
		protoInput := scm.makeOperationInput(model, parserFunction, input, proto.InputMode_INPUT_MODE_WRITE, impl)
		protoOp.Inputs = append(protoOp.Inputs, protoInput)
	}

	scm.applyActionAttributes(parserFunction, protoOp, modelName)

	return protoOp
}

func (scm *Builder) inferParserInputType(
	model *parser.ModelNode,
	op *parser.ActionNode,
	input *parser.ActionInputNode,
	impl proto.OperationImplementation,
) (t *proto.TypeInfo, target []string) {
	idents := input.Type.Fragments
	protoType := scm.parserTypeToProtoType(idents[0].Fragment)

	var modelName *wrapperspb.StringValue
	var fieldName *wrapperspb.StringValue
	var enumName *wrapperspb.StringValue

	if protoType == proto.Type_TYPE_ENUM {
		enumName = &wrapperspb.StringValue{
			Value: idents[0].Fragment,
		}
	}

	if protoType == proto.Type_TYPE_UNKNOWN {
		// If we haven't been able to resolve the type of the input it
		// must be a model field, so we need to resolve it

		var field *parser.FieldNode
		currModel := model

		for _, ident := range idents {
			// For operations, inputs that refer to model fields are handled automatically
			// by the runtime. For this to work we need to store the path to the field
			// that the input refers to, as it may be in a nested model.
			if impl == proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
				target = append(target, ident.Fragment)
			}

			field = query.ModelField(currModel, ident.Fragment)
			m := query.Model(scm.asts, field.Type)
			if m != nil {
				currModel = m
			}
		}

		protoType = scm.parserFieldToProtoTypeInfo(field).Type

		modelName = &wrapperspb.StringValue{
			Value: currModel.Name.Value,
		}
		fieldName = &wrapperspb.StringValue{
			Value: field.Name.Value,
		}

		if protoType == proto.Type_TYPE_ENUM {
			enumName = &wrapperspb.StringValue{
				Value: field.Type,
			}
		}
	}

	return &proto.TypeInfo{
		Type:      protoType,
		Repeated:  input.Repeated,
		ModelName: modelName,
		FieldName: fieldName,
		EnumName:  enumName,
	}, target
}

func (scm *Builder) makeOperationInput(
	model *parser.ModelNode,
	op *parser.ActionNode,
	input *parser.ActionInputNode,
	mode proto.InputMode,
	impl proto.OperationImplementation,
) (inputs *proto.OperationInput) {

	idents := input.Type.Fragments
	protoType := scm.parserTypeToProtoType(idents[0].Fragment)

	target := []string{}

	var modelName *wrapperspb.StringValue
	var fieldName *wrapperspb.StringValue
	var enumName *wrapperspb.StringValue

	if protoType == proto.Type_TYPE_ENUM {
		enumName = &wrapperspb.StringValue{
			Value: idents[0].Fragment,
		}
	}

	if protoType == proto.Type_TYPE_UNKNOWN {
		// If we haven't been able to resolve the type of the input it
		// must be a model field, so we need to resolve it

		var field *parser.FieldNode
		currModel := model

		for _, ident := range idents {
			// For operations, inputs that refer to model fields are handled automatically
			// by the runtime. For this to work we need to store the path to the field
			// that the input refers to, as it may be in a nested model.
			if impl == proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
				target = append(target, ident.Fragment)
			}

			field = query.ModelField(currModel, ident.Fragment)
			m := query.Model(scm.asts, field.Type)
			if m != nil {
				currModel = m
			}
		}

		protoType = scm.parserFieldToProtoTypeInfo(field).Type

		modelName = &wrapperspb.StringValue{
			Value: currModel.Name.Value,
		}
		fieldName = &wrapperspb.StringValue{
			Value: field.Name.Value,
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
			FieldName: fieldName,
			EnumName:  enumName,
		},
		Optional: input.Optional,
		Mode:     mode,
		Target:   target,
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

	case proto.Type_TYPE_MODEL:
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

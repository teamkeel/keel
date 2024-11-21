package expressions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// OperandResolverCel hides some of the complexity of expression parsing so that the runtime action code
// can reason about and execute expression logic without stepping through the AST.
type OperandResolverCel struct {
	Context context.Context
	Schema  *proto.Schema
	Model   *proto.Model
	Action  *proto.Action
	Operand *expr.Expr
}

func NewOperandResolverCel(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, operand *expr.Expr) *OperandResolverCel {
	return &OperandResolverCel{
		Context: ctx,
		Schema:  schema,
		Model:   model,
		Action:  action,
		Operand: operand,
	}
}

func (resolver *OperandResolverCel) Fragments() ([]string, error) {
	expre := []string{}
	e := resolver.Operand

	for {
		switch {
		}
		if s, ok := e.ExprKind.(*expr.Expr_SelectExpr); ok {
			expre = append([]string{e.GetSelectExpr().GetField()}, expre...)
			e = s.SelectExpr.Operand
		} else if _, ok := e.ExprKind.(*expr.Expr_IdentExpr); ok {
			expre = append([]string{e.GetIdentExpr().Name}, expre...)
			break
		} else {
			return nil, errors.New("unhandled expression type")
		}
	}

	//TODO: normalise?

	return expre, nil
}

// NormalisedFragments will return the expression fragments "in full" so that they can be processed for query building
// For example, note the two expressions in the condition @where(account in ctx.identity.primaryAccount.following.followee)
// NormalisedFragments will transform each of these operands as follows:
//
//	account.id
//	ctx.identity.primaryAccount.following.followeeId
func (resolver *OperandResolverCel) NormalisedFragments() ([]string, error) {
	fragments, err := resolver.Fragments()
	if err != nil {
		return nil, err
	}

	operandType, _, err := resolver.GetOperandType()
	if err != nil {
		return nil, err
	}

	if operandType == proto.Type_TYPE_MODEL && len(fragments) == 1 {
		// One fragment is only possible if the expression is only referencing the model.
		// For example, @where(account in ...)
		// Add a new fragment 'id'
		fragments = append(fragments, parser.FieldNameId)
	} else if operandType == proto.Type_TYPE_MODEL {
		i := 0
		if fragments[0] == "ctx" {
			i++
		}

		modelTarget := resolver.Schema.FindModel(casing.ToCamel(fragments[i]))
		if modelTarget == nil {
			return nil, fmt.Errorf("model '%s' does not exist in schema", casing.ToCamel(fragments[i]))
		}

		var fieldTarget *proto.Field
		for i := i + 1; i < len(fragments); i++ {
			fieldTarget = proto.FindField(resolver.Schema.Models, modelTarget.Name, fragments[i])
			if fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
				modelTarget = resolver.Schema.FindModel(fieldTarget.Type.ModelName.Value)
				if modelTarget == nil {
					return nil, fmt.Errorf("model '%s' does not exist in schema", fieldTarget.Type.ModelName.Value)
				}
			}
		}

		if fieldTarget.IsHasOne() || fieldTarget.IsHasMany() {
			// Add a new fragment 'id'
			fragments = append(fragments, parser.FieldNameId)
		} else {
			// Replace the last fragment with the foreign key field
			fragments[len(fragments)-1] = fmt.Sprintf("%sId", fragments[len(fragments)-1])
		}
	}

	return fragments, nil
}

// IsLiteral returns true if the expression operand is a literal type.
// For example, a number or string literal written straight into the Keel schema,
// such as the right-hand side operand in @where(person.age > 21).
func (resolver *OperandResolverCel) IsLiteral() bool {
	// // Check if literal or array of literals, such as a "keel" or ["keel", "weave"]
	// isLiteral, _ := resolver.operand.IsLiteralType()
	// if isLiteral {
	// 	return true
	// }

	// // Check if an enum, such as Sport.Cricket
	// isEnumLiteral := resolver.operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.operand.Ident.Fragments[0].Fragment)
	// if isEnumLiteral {
	// 	return true
	// }

	// if resolver.operand.Ident == nil && resolver.operand.Array != nil {
	// 	// Check if an empty array, such as []
	// 	isEmptyArray := resolver.operand.Ident == nil && resolver.operand.Array != nil && len(resolver.operand.Array.Values) == 0
	// 	if isEmptyArray {
	// 		return true
	// 	}

	// 	// Check if an array of enums, such as [Sport.Cricket, Sport.Rugby]
	// 	isEnumLiteralArray := true
	// 	for _, item := range resolver.operand.Array.Values {
	// 		if !proto.EnumExists(resolver.Schema.Enums, item.Ident.Fragments[0].Fragment) {
	// 			isEnumLiteralArray = false
	// 		}
	// 	}
	// 	if isEnumLiteralArray {
	// 		return true
	// 	}
	// }

	return false
}

// IsImplicitInput returns true if the expression operand refers to an implicit input on an action.
// For example, an input value provided in a create action might require validation,
// such as: create createThing() with (name) @validation(name != "")
func (resolver *OperandResolverCel) IsImplicitInput() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		panic(err)
		//return nil, err
	}

	if len(fragments) > 1 {
		return false
	}

	foundImplicitWhereInput := false
	foundImplicitValueInput := false

	whereInputs := proto.FindWhereInputMessage(resolver.Schema, resolver.Action.Name)
	if whereInputs != nil {
		_, foundImplicitWhereInput = lo.Find(whereInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == fragments[0] && in.IsModelField()
		})
	}

	valuesInputs := proto.FindValuesInputMessage(resolver.Schema, resolver.Action.Name)
	if valuesInputs != nil {
		_, foundImplicitValueInput = lo.Find(valuesInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == fragments[0] && in.IsModelField()
		})
	}

	return foundImplicitWhereInput || foundImplicitValueInput
}

// IsExplicitInput returns true if the expression operand refers to an explicit input on an action.
// For example, a where condition might use an explicit input,
// such as: list listThings(isActive: Boolean) @where(thing.isActive == isActive)
func (resolver *OperandResolverCel) IsExplicitInput() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		panic(err)
		//return nil, err
	}

	if len(fragments) > 1 {
		return false
	}

	foundExplicitWhereInput := false
	foundExplicitValueInput := false

	whereInputs := proto.FindWhereInputMessage(resolver.Schema, resolver.Action.Name)
	if whereInputs != nil {
		_, foundExplicitWhereInput = lo.Find(whereInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == fragments[0] && !in.IsModelField()
		})
	}

	valuesInputs := proto.FindValuesInputMessage(resolver.Schema, resolver.Action.Name)
	if valuesInputs != nil {
		_, foundExplicitValueInput = lo.Find(valuesInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == fragments[0] && !in.IsModelField()
		})
	}

	return foundExplicitWhereInput || foundExplicitValueInput
}

// IsModelDbColumn returns true if the expression operand refers to a field value residing in the database.
// For example, a where condition might filter on reading data,
// such as: @where(post.author.isActive)
func (resolver *OperandResolverCel) IsModelDbColumn() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	return fragments[0] == strcase.ToLowerCamel(resolver.Model.Name)
}

// IsContextDbColumn returns true if the expression refers to a value on the context
// which will require database access (such as with identity backlinks),
// such as: @permission(expression: ctx.identity.user.isActive)
func (resolver *OperandResolverCel) IsContextDbColumn() bool {
	return resolver.IsContextIdentity() && !resolver.IsContextIdentityId()

}

func (resolver *OperandResolverCel) IsContextIdentity() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if !resolver.IsContext() {
		return false
	}

	if len(fragments) > 1 && fragments[1] == "identity" {
		return true
	}

	return false
}

func (resolver *OperandResolverCel) IsContextIdentityId() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if !resolver.IsContextIdentity() {
		return false
	}

	if len(fragments) == 2 {
		return true
	}

	if len(fragments) == 3 && fragments[2] == "id" {
		return true
	}

	return false
}

func (resolver *OperandResolverCel) IsContextIsAuthenticatedField() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if resolver.IsContext() && len(fragments) == 2 {
		return fragments[1] == "isAuthenticated"
	}

	return false
}

// IsContextField returns true if the expression operand refers to a value on the context
// which does not require to be read from the database.
// For example, a permission condition may check against the current identity,
// such as: @permission(thing.identity == ctx.identity)
//
// However if the expression traverses onwards from identity (using an Identity-backlink)
// like this:
// "ctx.identity.user"
// then it returns false, because that can no longer be resolved solely from the
// in memory context data.
func (resolver *OperandResolverCel) IsContextField() bool {
	return resolver.IsContext() && !resolver.IsContextDbColumn()
}

func (resolver *OperandResolverCel) IsContext() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	return fragments[0] == "ctx"
}

func (resolver *OperandResolverCel) IsContextNowField() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if resolver.IsContext() && len(fragments) == 2 {
		return fragments[1] == "now"
	}
	return false
}

func (resolver *OperandResolverCel) IsContextHeadersField() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if resolver.IsContext() && len(fragments) == 3 {
		return fragments[1] == "headers"
	}
	return false
}

func (resolver *OperandResolverCel) IsContextEnvField() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if resolver.IsContext() && len(fragments) == 3 {
		return fragments[1] == "env"
	}
	return false
}

func (resolver *OperandResolverCel) IsContextSecretField() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
	}

	if resolver.IsContext() && len(fragments) == 3 {
		return fragments[1] == "secrets"
	}
	return false
}

// GetOperandType returns the equivalent protobuf type for the expression operand and whether it is an array or not
func (resolver *OperandResolverCel) GetOperandType() (proto.Type, bool, error) {
	action := resolver.Action
	schema := resolver.Schema

	fragments, err := resolver.Fragments()
	if err != nil {
		return proto.Type_TYPE_UNKNOWN, false, err
	}

	switch {
	// case resolver.IsLiteral():
	// 	if operand.Ident == nil {
	// 		switch {
	// 		case operand.String != nil:
	// 			return proto.Type_TYPE_STRING, false, nil
	// 		case operand.Number != nil:
	// 			return proto.Type_TYPE_INT, false, nil
	// 		case operand.Decimal != nil:
	// 			return proto.Type_TYPE_DECIMAL, false, nil
	// 		case operand.True || operand.False:
	// 			return proto.Type_TYPE_BOOL, false, nil
	// 		case operand.Array != nil:
	// 			return proto.Type_TYPE_UNKNOWN, true, nil
	// 		case operand.Null:
	// 			return proto.Type_TYPE_UNKNOWN, false, nil
	// 		default:
	// 			return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("cannot handle operand type")
	// 		}
	// 	} else if resolver.operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.operand.Ident.Fragments[0].Fragment) {
	// 		return proto.Type_TYPE_ENUM, false, nil
	// 	} else {
	// 		return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("unknown literal type")
	// 	}

	case resolver.IsModelDbColumn(), resolver.IsContextDbColumn():
		if resolver.IsContextDbColumn() {
			// If this is a context backlink, then remove the first "ctx" fragment.
			fragments = fragments[1:]
		}

		// The first fragment will always be the root model name, e.g. "author" in author.posts.title
		modelTarget := schema.FindModel(casing.ToCamel(fragments[0]))
		if modelTarget == nil {
			return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("model '%s' does not exist in schema", casing.ToCamel(fragments[0]))
		}

		var fieldTarget *proto.Field
		for i := 1; i < len(fragments); i++ {
			fieldTarget = proto.FindField(schema.Models, modelTarget.Name, fragments[i])
			if fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
				modelTarget = schema.FindModel(fieldTarget.Type.ModelName.Value)
				if modelTarget == nil {
					return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("model '%s' does not exist in schema", fieldTarget.Type.ModelName.Value)
				}
			}
		}

		// If no field is provided, for example: @where(account in ...)
		// Or if the target field is a MODEL, for example:
		if fieldTarget == nil || fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
			return proto.Type_TYPE_MODEL, false, nil
		}

		return fieldTarget.Type.Type, fieldTarget.Type.Repeated, nil
	case resolver.IsImplicitInput():
		modelTarget := casing.ToCamel(action.ModelName)
		inputName := fragments[0]
		field := proto.FindField(schema.Models, modelTarget, inputName)
		return field.Type.Type, field.Type.Repeated, nil
	case resolver.IsExplicitInput():
		inputName := fragments[0]
		var field *proto.MessageField
		switch action.Type {
		case proto.ActionType_ACTION_TYPE_CREATE:
			message := proto.FindValuesInputMessage(schema, action.Name)
			field = message.FindField(inputName)
		case proto.ActionType_ACTION_TYPE_GET, proto.ActionType_ACTION_TYPE_LIST, proto.ActionType_ACTION_TYPE_DELETE:
			message := proto.FindWhereInputMessage(schema, action.Name)
			field = message.FindField(inputName)
		case proto.ActionType_ACTION_TYPE_UPDATE:
			message := proto.FindValuesInputMessage(schema, action.Name)
			field = message.FindField(inputName)
			if field == nil {
				message := proto.FindWhereInputMessage(schema, action.Name)
				field = message.FindField(inputName)
			}
		default:
			return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("unhandled action type %s for explicit input", action.Type)
		}
		if field == nil {
			return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("could not find explicit input %s on action %s", inputName, action.Name)
		}
		return field.Type.Type, field.Type.Repeated, nil
	case resolver.IsContextNowField():
		return proto.Type_TYPE_TIMESTAMP, false, nil
	case resolver.IsContextEnvField():
		return proto.Type_TYPE_STRING, false, nil
	case resolver.IsContextSecretField():
		return proto.Type_TYPE_STRING, false, nil
	case resolver.IsContextHeadersField():
		return proto.Type_TYPE_STRING, false, nil
	case resolver.IsContext():
		fieldName := fragments[1]
		return runtimectx.ContextFieldTypes[fieldName], false, nil
	default:
		return proto.Type_TYPE_UNKNOWN, false, fmt.Errorf("cannot handle operand target %s", fragments[0])
	}
}

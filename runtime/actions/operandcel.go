package actions

import (
	"context"
	"errors"
	"fmt"
	"net/textproto"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
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
	Inputs  map[string]any
}

func NewOperandResolverCel(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, operand *expr.Expr, inputs map[string]any) *OperandResolverCel {
	return &OperandResolverCel{
		Context: ctx,
		Schema:  schema,
		Model:   model,
		Action:  action,
		Operand: operand,
		Inputs:  inputs,
	}
}

func (resolver *OperandResolverCel) Fragments() ([]string, error) {
	expre := []string{}
	e := resolver.Operand

	for {
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

	switch resolver.Operand.ExprKind.(type) {
	case *expr.Expr_ListExpr:
		return true // TODO: check elements
	case *expr.Expr_ConstExpr:
		return true
	}

	return false
}

// // IsImplicitInput returns true if the expression operand refers to an implicit input on an action.
// // For example, an input value provided in a create action might require validation,
// // such as: create createThing() with (name) @validation(name != "")
func (resolver *OperandResolverCel) IsImplicitInput() bool {
	fragments, err := resolver.Fragments()
	if err != nil {
		return false
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
		return false
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
	//action := resolver.Action
	schema := resolver.Schema

	fragments, err := resolver.Fragments()
	if err != nil {
		return proto.Type_TYPE_UNKNOWN, false, err
	}

	if fragments[0] == "ctx" {
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

}

// Generates a QueryOperand which
func (resolver *OperandResolverCel) QueryOperand() (*QueryOperand, error) {
	var queryOperand *QueryOperand

	switch {
	case resolver.IsLiteral():
		switch resolver.Operand.ExprKind.(type) {
		case *expr.Expr_ListExpr:
			l := resolver.Operand.GetListExpr()
			elems := l.GetElements()

			arr := make([]any, len(elems))

			for i, elem := range elems {
				switch elem.ExprKind.(type) {
				case *expr.Expr_SelectExpr:
					// Enum values
					s := elem.GetSelectExpr()
					op := s.GetField()
					arr[i] = op
				case *expr.Expr_ConstExpr:
					// Literal values
					c := elem.GetConstExpr()
					v, err := toNative(c)
					if err != nil {
						return nil, err
					}
					arr[i] = v
				}
			}

			return Value(arr), nil
		case *expr.Expr_ConstExpr:
			c := resolver.Operand.GetConstExpr()

			v, err := toNative(c)
			if err != nil {
				return nil, err
			}

			return Value(v), nil
		default:
			return nil, errors.New("unknown literal")
		}

	case resolver.IsContextDbColumn():
		// If this is a value from ctx that requires a database read (such as with identity backlinks),
		// then construct an inline query for this operand.  This is necessary because we can't retrieve this value
		// from the current query builder.

		fragments, err := resolver.NormalisedFragments()
		if err != nil {
			return nil, err
		}

		// Remove the ctx fragment
		fragments = fragments[1:]

		identityModel := resolver.Schema.FindModel(strcase.ToCamel(fragments[0]))
		query := NewQuery(identityModel)

		identityId := ""
		if auth.IsAuthenticated(resolver.Context) {
			identity, err := auth.GetIdentity(resolver.Context)
			if err != nil {
				return nil, err
			}
			identityId = identity[parser.FieldNameId].(string)
		}

		err = query.AddJoinFromFragments(resolver.Schema, fragments)
		if err != nil {
			return nil, err
		}

		err = query.Where(IdField(), Equals, Value(identityId))
		if err != nil {
			return nil, err
		}

		selectField := ExpressionField(fragments[:len(fragments)-1], fragments[len(fragments)-1], false)

		// If there are no matches in the subquery then null will be returned, but null
		// will cause IN and NOT IN filtering of this subquery result to always evaluate as false.
		// Therefore we need to filter out null.
		query.And()
		err = query.Where(selectField, NotEquals, Null())
		if err != nil {
			return nil, err
		}

		currModel := identityModel
		for i := 1; i < len(fragments)-1; i++ {
			name := proto.FindField(resolver.Schema.Models, currModel.Name, fragments[i]).Type.ModelName.Value
			currModel = resolver.Schema.FindModel(name)
		}
		currField := proto.FindField(resolver.Schema.Models, currModel.Name, fragments[len(fragments)-1])

		if currField.Type.Repeated {
			query.SelectUnnested(selectField)
		} else {
			query.Select(selectField)
		}

		queryOperand = InlineQuery(query, selectField)

	case resolver.IsImplicitInput(), resolver.IsExplicitInput():
		fragments, err := resolver.Fragments()
		if err != nil {
			return nil, err
		}

		inputName := fragments[0]
		value, ok := resolver.Inputs[inputName]
		if !ok {
			return nil, fmt.Errorf("implicit or explicit input '%s' does not exist in arguments", inputName)
		}
		return Value(value), nil

	case resolver.IsModelDbColumn():
		// If this is a model field then generate the appropriate column operand for the database query.

		fragments, err := resolver.NormalisedFragments()
		if err != nil {
			return nil, err
		}

		// Generate QueryOperand from the fragments that make up the expression operand
		queryOperand, err = operandFromFragments(resolver.Schema, fragments)
		if err != nil {
			return nil, err
		}
	case resolver.IsContextIdentityId():
		isAuthenticated := auth.IsAuthenticated(resolver.Context)
		if !isAuthenticated {
			return nil, nil
		}

		identity, err := auth.GetIdentity(resolver.Context)
		if err != nil {
			return nil, err
		}
		return Value(identity[parser.FieldNameId].(string)), nil
	case resolver.IsContextIdentityId():
		isAuthenticated := auth.IsAuthenticated(resolver.Context)
		if !isAuthenticated {
			return nil, nil
		}

		identity, err := auth.GetIdentity(resolver.Context)
		if err != nil {
			return nil, err
		}
		return Value(identity[parser.FieldNameId].(string)), nil
	case resolver.IsContextIsAuthenticatedField():
		isAuthenticated := auth.IsAuthenticated(resolver.Context)
		return Value(isAuthenticated), nil
	case resolver.IsContextNowField():
		return Value(runtimectx.GetNow()), nil
	case resolver.IsContextEnvField():
		fragments, err := resolver.Fragments()
		if err != nil {
			return nil, err
		}

		envVarName := fragments[2]
		return Value(os.Getenv(envVarName)), nil
	case resolver.IsContextSecretField():
		fragments, err := resolver.Fragments()
		if err != nil {
			return nil, err
		}

		secret, err := runtimectx.GetSecret(resolver.Context, fragments[2])
		if err != nil {
			return nil, err
		}

		return Value(secret), nil
	case resolver.IsContextHeadersField():
		fragments, err := resolver.Fragments()
		if err != nil {
			return nil, err
		}

		headerName := fragments[2]

		// First we parse the header name to kebab. MyCustomHeader will become my-custom-header.
		kebab := strcase.ToKebab(headerName)

		// Then get canonical name. my-custom-header will become My-Custom-Header.
		// https://pkg.go.dev/net/http#Header.Get
		canonicalName := textproto.CanonicalMIMEHeaderKey(kebab)

		headers, err := runtimectx.GetRequestHeaders(resolver.Context)
		if err != nil {
			return nil, err
		}
		if value, ok := headers[canonicalName]; ok {
			return Value(strings.Join(value, ", ")), nil
		}
		return Value(""), nil

	}

	return queryOperand, nil
}

func toNative(c *expr.Constant) (any, error) {
	switch c.ConstantKind.(type) {
	case *expr.Constant_BoolValue:
		return c.GetBoolValue(), nil
	case *expr.Constant_DoubleValue:
		return c.GetDoubleValue(), nil
	case *expr.Constant_Int64Value:
		return c.GetInt64Value(), nil
	case *expr.Constant_NullValue:
		return nil, nil
	case *expr.Constant_StringValue:
		return c.GetStringValue(), nil
	case *expr.Constant_Uint64Value:
		return c.GetUint64Value(), nil
	default:
		return nil, fmt.Errorf("not implemented : %v", c)
	}
}

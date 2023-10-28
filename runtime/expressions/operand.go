package expressions

import (
	"context"
	"errors"
	"fmt"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// OperandResolver hides some of the complexity of expression parsing so that the runtime action code
// can reason about and execute expression logic without stepping through the AST.
type OperandResolver struct {
	Context context.Context
	Schema  *proto.Schema
	Model   *proto.Model
	Action  *proto.Action
	operand *parser.Operand
}

func NewOperandResolver(ctx context.Context, schema *proto.Schema, model *proto.Model, action *proto.Action, operand *parser.Operand) *OperandResolver {
	return &OperandResolver{
		Context: ctx,
		Schema:  schema,
		Model:   model,
		Action:  action,
		operand: operand,
	}
}

// NormalisedFragments will return the expression fragments "in full" so that they can be processed for query building
// For example, note the two expressions in the condition @where(account in ctx.identity.primaryAccount.following.followee)
// NormalisedFragments will transform each of these operands as follows:
//
//	account.id
//	ctx.identity.primaryAccount.following.followeeId
func (resolver *OperandResolver) NormalisedFragments() ([]string, error) {
	fragments := lo.Map(resolver.operand.Ident.Fragments, func(fragment *parser.IdentFragment, _ int) string { return fragment.Fragment })

	operandType, err := resolver.GetOperandType()
	if err != nil {
		return nil, err
	}

	if operandType == proto.Type_TYPE_MODEL && len(fragments) == 1 {
		// One fragment is only possible if the expression is only referencing the model.
		// For example, @where(account in ...)
		// Add a new fragment 'id'
		fragments = append(fragments, parser.ImplicitFieldNameId)
	} else if operandType == proto.Type_TYPE_MODEL {
		i := 0
		if fragments[0] == "ctx" {
			i++
		}

		modelTarget := proto.FindModel(resolver.Schema.Models, casing.ToCamel(fragments[i]))
		if modelTarget == nil {
			return nil, fmt.Errorf("model '%s' does not exist in schema", casing.ToCamel(fragments[i]))
		}

		var fieldTarget *proto.Field
		for i := i + 1; i < len(fragments); i++ {
			fieldTarget = proto.FindField(resolver.Schema.Models, modelTarget.Name, fragments[i])
			if fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
				modelTarget = proto.FindModel(resolver.Schema.Models, fieldTarget.Type.ModelName.Value)
				if modelTarget == nil {
					return nil, fmt.Errorf("model '%s' does not exist in schema", fieldTarget.Type.ModelName.Value)
				}
			}
		}

		if proto.IsHasOne(fieldTarget) || proto.IsHasMany(fieldTarget) {
			// Add a new fragment 'id'
			fragments = append(fragments, parser.ImplicitFieldNameId)
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
func (resolver *OperandResolver) IsLiteral() bool {
	isLiteral, _ := resolver.operand.IsLiteralType()
	isEnumLiteral := resolver.operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.operand.Ident.Fragments[0].Fragment)
	return isLiteral || isEnumLiteral
}

// IsImplicitInput returns true if the expression operand refers to an implicit input on an action.
// For example, an input value provided in a create action might require validation,
// such as: create createThing() with (name) @validation(name != "")
func (resolver *OperandResolver) IsImplicitInput() bool {
	isSingleFragment := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) == 1

	if !isSingleFragment {
		return false
	}

	foundImplicitWhereInput := false
	foundImplicitValueInput := false

	whereInputs := proto.FindWhereInputMessage(resolver.Schema, resolver.Action.Name)
	if whereInputs != nil {
		_, foundImplicitWhereInput = lo.Find(whereInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == resolver.operand.Ident.Fragments[0].Fragment && in.IsModelField()
		})
	}

	valuesInputs := proto.FindValuesInputMessage(resolver.Schema, resolver.Action.Name)
	if valuesInputs != nil {
		_, foundImplicitValueInput = lo.Find(valuesInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == resolver.operand.Ident.Fragments[0].Fragment && in.IsModelField()
		})
	}

	return foundImplicitWhereInput || foundImplicitValueInput
}

// IsExplicitInput returns true if the expression operand refers to an explicit input on an action.
// For example, a where condition might use an explicit input,
// such as: list listThings(isActive: Boolean) @where(thing.isActive == isActive)
func (resolver *OperandResolver) IsExplicitInput() bool {
	isSingleFragmentIdent := resolver.operand.Ident != nil && len(resolver.operand.Ident.Fragments) == 1

	if !isSingleFragmentIdent {
		return false
	}

	foundExplicitWhereInput := false
	foundExplicitValueInput := false

	whereInputs := proto.FindWhereInputMessage(resolver.Schema, resolver.Action.Name)
	if whereInputs != nil {
		_, foundExplicitWhereInput = lo.Find(whereInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == resolver.operand.Ident.Fragments[0].Fragment && !in.IsModelField()
		})
	}

	valuesInputs := proto.FindValuesInputMessage(resolver.Schema, resolver.Action.Name)
	if valuesInputs != nil {
		_, foundExplicitValueInput = lo.Find(valuesInputs.Fields, func(in *proto.MessageField) bool {
			return in.Name == resolver.operand.Ident.Fragments[0].Fragment && !in.IsModelField()
		})
	}

	return foundExplicitWhereInput || foundExplicitValueInput
}

// IsModelDbColumn returns true if the expression operand refers to a field value residing in the database.
// For example, a where condition might filter on reading data,
// such as: @where(post.author.isActive)
func (resolver *OperandResolver) IsModelDbColumn() bool {
	return !resolver.IsLiteral() &&
		!resolver.IsContext() &&
		!resolver.IsExplicitInput() &&
		!resolver.IsImplicitInput()
}

// IsContextDbColumn returns true if the expression refers to a value on the context
// which will require database access (such as with identity backlinks),
// such as: @permission(expression: ctx.identity.user.isActive)
func (resolver *OperandResolver) IsContextDbColumn() bool {
	return resolver.operand.Ident.IsContextIdentity() && !resolver.operand.Ident.IsContextIdentityId()
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
func (resolver *OperandResolver) IsContextField() bool {
	return resolver.operand.Ident.IsContext() && !resolver.IsContextDbColumn()
}

func (resolver *OperandResolver) IsContext() bool {
	return resolver.operand.Ident.IsContext()
}

// GetOperandType returns the equivalent protobuf type for the expression operand.
func (resolver *OperandResolver) GetOperandType() (proto.Type, error) {
	operand := resolver.operand
	action := resolver.Action
	schema := resolver.Schema

	switch {
	case resolver.IsLiteral():
		if operand.Ident == nil {
			switch {
			case operand.String != nil:
				return proto.Type_TYPE_STRING, nil
			case operand.Number != nil:
				return proto.Type_TYPE_INT, nil
			case operand.True || operand.False:
				return proto.Type_TYPE_BOOL, nil
			case operand.Array != nil:
				return proto.Type_TYPE_UNKNOWN, nil
			case operand.Null:
				return proto.Type_TYPE_UNKNOWN, nil
			default:
				return proto.Type_TYPE_UNKNOWN, fmt.Errorf("cannot handle operand type")
			}
		} else if resolver.operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.operand.Ident.Fragments[0].Fragment) {
			return proto.Type_TYPE_ENUM, nil
		} else {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("unknown literal type")
		}

	case resolver.IsModelDbColumn(), resolver.IsContextDbColumn():
		fragments := operand.Ident.Fragments

		if resolver.IsContextDbColumn() {
			// If this is a context backlink, then remove the first "ctx" fragment.
			fragments = fragments[1:]
		}

		// The first fragment will always be the root model name, e.g. "author" in author.posts.title
		modelTarget := proto.FindModel(schema.Models, casing.ToCamel(fragments[0].Fragment))
		if modelTarget == nil {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("model '%s' does not exist in schema", casing.ToCamel(fragments[0].Fragment))
		}

		var fieldTarget *proto.Field
		for i := 1; i < len(fragments); i++ {
			fieldTarget = proto.FindField(schema.Models, modelTarget.Name, fragments[i].Fragment)
			if fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
				modelTarget = proto.FindModel(schema.Models, fieldTarget.Type.ModelName.Value)
				if modelTarget == nil {
					return proto.Type_TYPE_UNKNOWN, fmt.Errorf("model '%s' does not exist in schema", fieldTarget.Type.ModelName.Value)
				}
			}
		}

		// If no field is provided, for example: @where(account in ...)
		// Or if the target field is a MODEL, for example:
		if fieldTarget == nil || fieldTarget.Type.Type == proto.Type_TYPE_MODEL {
			return proto.Type_TYPE_MODEL, nil
		}

		return fieldTarget.Type.Type, nil
	case resolver.IsImplicitInput():
		modelTarget := casing.ToCamel(action.ModelName)
		inputName := operand.Ident.Fragments[0].Fragment
		operandType := proto.FindField(schema.Models, modelTarget, inputName).Type.Type
		return operandType, nil
	case resolver.IsExplicitInput():
		inputName := operand.Ident.Fragments[0].Fragment
		var field *proto.MessageField
		switch action.Type {
		case proto.ActionType_ACTION_TYPE_CREATE:
			message := proto.FindValuesInputMessage(schema, action.Name)
			field = proto.FindMessageField(message, inputName)
		case proto.ActionType_ACTION_TYPE_GET, proto.ActionType_ACTION_TYPE_LIST, proto.ActionType_ACTION_TYPE_DELETE:
			message := proto.FindWhereInputMessage(schema, action.Name)
			field = proto.FindMessageField(message, inputName)
		case proto.ActionType_ACTION_TYPE_UPDATE:
			message := proto.FindValuesInputMessage(schema, action.Name)
			field = proto.FindMessageField(message, inputName)
			if field == nil {
				message := proto.FindWhereInputMessage(schema, action.Name)
				field = proto.FindMessageField(message, inputName)
			}
		default:
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("unhandled action type %s for explicit input", action.Type)
		}
		if field == nil {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("could not find explicit input %s on action %s", inputName, action.Name)
		}
		return field.Type.Type, nil
	case operand.Ident.IsContext():
		fragmentCount := len(operand.Ident.Fragments)
		if fragmentCount > 2 && operand.Ident.IsContextEnvField() {
			return proto.Type_TYPE_STRING, nil
		}

		fieldName := operand.Ident.Fragments[1].Fragment
		return runtimectx.ContextFieldTypes[fieldName], nil
	default:
		return proto.Type_TYPE_UNKNOWN, fmt.Errorf("cannot handle operand target %s", operand.Ident.Fragments[0].Fragment)
	}
}

// ResolveValue returns the actual value of the operand, provided a database read is not required.
func (resolver *OperandResolver) ResolveValue(args map[string]any) (any, error) {
	operandType, err := resolver.GetOperandType()
	if err != nil {
		return nil, err
	}

	switch {
	case resolver.IsLiteral():
		isLiteral, _ := resolver.operand.IsLiteralType()
		if isLiteral {
			return toNative(resolver.operand, operandType)
		} else if resolver.operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.operand.Ident.Fragments[0].Fragment) {
			return resolver.operand.Ident.Fragments[1].Fragment, nil
		} else {
			return nil, errors.New("unknown literal type")
		}
	case resolver.IsImplicitInput(), resolver.IsExplicitInput():
		inputName := resolver.operand.Ident.Fragments[0].Fragment
		value, ok := args[inputName]
		if !ok {
			return nil, fmt.Errorf("implicit or explicit input '%s' does not exist in arguments", inputName)
		}
		return value, nil
	case resolver.IsModelDbColumn(), resolver.IsContextDbColumn():
		// todo: https://linear.app/keel/issue/RUN-153/set-attribute-to-support-targeting-database-fields
		panic("cannot resolve operand value from the database")
	case resolver.operand.Ident.IsContextIdentityId():
		isAuthenticated := auth.IsAuthenticated(resolver.Context)
		if !isAuthenticated {
			return nil, nil
		}

		identity, err := auth.GetIdentity(resolver.Context)
		if err != nil {
			return nil, err
		}
		return identity.Id, nil
	case resolver.operand.Ident.IsContextIsAuthenticatedField():
		isAuthenticated := auth.IsAuthenticated(resolver.Context)
		return isAuthenticated, nil
	case resolver.operand.Ident.IsContextNowField():
		return runtimectx.GetNow(), nil
	case resolver.operand.Ident.IsContextEnvField():
		envVarName := resolver.operand.Ident.Fragments[2].Fragment
		return os.Getenv(envVarName), nil
	case resolver.operand.Ident.IsContextSecretField():
		secret := resolver.operand.Ident.Fragments[2].Fragment
		return runtimectx.GetSecret(resolver.Context, secret)
	case resolver.operand.Ident.IsContextHeadersField():
		headerName := resolver.operand.Ident.Fragments[2].Fragment
		// Get canonical name, as this is what header keys are transformed into
		// https://pkg.go.dev/net/http#Header.Get
		canonicalName := textproto.CanonicalMIMEHeaderKey(headerName)
		headers, err := runtimectx.GetRequestHeaders(resolver.Context)
		if err != nil {
			return nil, err
		}
		if value, ok := headers[canonicalName]; ok {
			return strings.Join(value, ", "), nil
		}
		return "", nil
	case resolver.operand.Type() == parser.TypeArray:
		return nil, fmt.Errorf("cannot yet handle operand of type non-literal array")
	default:
		return nil, fmt.Errorf("cannot handle operand of unknown type")
	}
}

func toNative(v *parser.Operand, fieldType proto.Type) (any, error) {
	if v.Array != nil {
		values := []any{}
		for _, v := range v.Array.Values {
			value, err := toNative(v, fieldType)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	}

	switch {
	case v.False:
		return false, nil
	case v.True:
		return true, nil
	case v.Number != nil:
		return *v.Number, nil
	case v.String != nil:
		v := *v.String
		v = strings.TrimPrefix(v, `"`)
		v = strings.TrimSuffix(v, `"`)
		switch fieldType {
		case proto.Type_TYPE_DATE:
			return toDate(v), nil
		case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			return toTime(v), nil
		}
		return v, nil
	case v.Null:
		return nil, nil
	case fieldType == proto.Type_TYPE_ENUM:
		return v.Ident.Fragments[1].Fragment, nil
	default:
		return nil, fmt.Errorf("toNative() does yet support this expression operand: %+v", v)
	}
}

func toDate(s string) time.Time {
	segments := strings.Split(s, `/`)
	day, _ := strconv.Atoi(segments[0])
	month, _ := strconv.Atoi(segments[1])
	year, _ := strconv.Atoi(segments[2])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func toTime(s string) time.Time {
	tm, _ := time.Parse(time.RFC3339, s)
	return tm
}

// // isContextIdentityDbColumn works out if this operand traverses an Identity field and requires the database.
// func (resolver *OperandResolver) isContextIdentityDbColumn() bool {
// 	if resolver.operand.Ident == nil {
// 		return false
// 	}
// 	fragments := lo.Map(resolver.operand.Ident.Fragments, func(frag *parser.IdentFragment, _ int) string {
// 		return frag.Fragment
// 	})

// 	if len(fragments) < 3 {
// 		return false
// 	}
// 	if fragments[0] != "ctx" {
// 		return false
// 	}
// 	if fragments[1] != "identity" {
// 		return false
// 	}
// 	if resolver.operand.is

// 	return true

// 	// field := proto.FindField(resolver.Schema.Models, "Identity", fragments[2])
// 	// if field == nil {
// 	// 	return false
// 	// }

// 	// return field.Type.Type == proto.Type_TYPE_MODEL
// }

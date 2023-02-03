package expressions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// OperandResolver hides some of the complexity of expression parsing so that the runtime action code
// can reason about and execute expression logic without stepping through the AST.
type OperandResolver struct {
	Context   context.Context
	Schema    *proto.Schema
	Operation *proto.Operation
	Operand   *parser.Operand
}

func NewOperandResolver(ctx context.Context, schema *proto.Schema, operation *proto.Operation, operand *parser.Operand) *OperandResolver {
	return &OperandResolver{
		Context:   ctx,
		Schema:    schema,
		Operation: operation,
		Operand:   operand,
	}
}

// IsLiteral returns true if the expression operand is a literal type.
// For example, a number or string literal written straight into the Keel schema,
// such as the right-hand side operand in @where(person.age > 21).
func (resolver *OperandResolver) IsLiteral() bool {
	isLiteral, _ := resolver.Operand.IsLiteralType()
	isEnumLiteral := resolver.Operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.Operand.Ident.Fragments[0].Fragment)
	return isLiteral || isEnumLiteral
}

// IsImplicitInput returns true if the expression operand refers to an implicit input on an operation.
// For example, an input value provided in a create operation might require validation,
// such as: create createThing() with (name) @validation(name != "")
func (resolver *OperandResolver) IsImplicitInput() bool {
	isSingleFragment := resolver.Operand.Ident != nil && len(resolver.Operand.Ident.Fragments) == 1

	if !isSingleFragment {
		return false
	}

	input, found := lo.Find(resolver.Operation.Inputs, func(in *proto.OperationInput) bool {
		return in.Name == resolver.Operand.Ident.Fragments[0].Fragment
	})

	return found && input.IsModelField()
}

// IsExplicitInput returns true if the expression operand refers to an explicit input on an operation.
// For example, a where condition might use an explicit input,
// such as: list listThings(isActive: Boolean) @where(thing.isActive == isActive)
func (resolver *OperandResolver) IsExplicitInput() bool {
	isSingleFragmentIdent := resolver.Operand.Ident != nil && len(resolver.Operand.Ident.Fragments) == 1

	if !isSingleFragmentIdent {
		return false
	}

	input, found := lo.Find(resolver.Operation.Inputs, func(in *proto.OperationInput) bool {
		return in.Name == resolver.Operand.Ident.Fragments[0].Fragment
	})

	return found && !input.IsModelField()
}

// IsDatabaseColumn returns true if the expression operand refers to a field value residing in the database.
// For example, a where condition might filter on reading data,
// such as: @where(post.author.isActive)
func (resolver *OperandResolver) IsDatabaseColumn() bool {
	isMultiFragmentIdent := resolver.Operand.Ident != nil && len(resolver.Operand.Ident.Fragments) > 1

	if !isMultiFragmentIdent {
		return false
	}

	modelTarget := resolver.Operand.Ident.Fragments[0].Fragment

	return modelTarget == strcase.ToLowerCamel(resolver.Operation.ModelName)
}

// IsContextField returns true if the expression operand refers to a value on the context.
// For example, a permission condition may check against the current identity,
// such as: @permission(thing.identity == ctx.identity)
func (resolver *OperandResolver) IsContextField() bool {
	return resolver.Operand.Ident.IsContext()
}

// GetOperandType returns the equivalent protobuf type for the expression operand.
func (resolver *OperandResolver) GetOperandType() (proto.Type, error) {
	operand := resolver.Operand
	operation := resolver.Operation
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
		} else if resolver.Operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.Operand.Ident.Fragments[0].Fragment) {
			return proto.Type_TYPE_ENUM, nil
		} else {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("unknown literal type")
		}
	case resolver.IsDatabaseColumn():
		fragmentCount := len(operand.Ident.Fragments)
		modelTarget := strcase.ToCamel(operand.Ident.Fragments[0].Fragment)

		if fragmentCount > 2 {
			for i := 1; i < fragmentCount-1; i++ {
				field := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), operand.Ident.Fragments[i].Fragment)
				modelTarget = field.Type.ModelName.Value
			}
		}

		fieldName := operand.Ident.Fragments[fragmentCount-1].Fragment
		if !proto.ModelHasField(schema, strcase.ToCamel(modelTarget), fieldName) {
			return proto.Type_TYPE_UNKNOWN, fmt.Errorf("this model: %s, does not have a field of name: %s", modelTarget, fieldName)
		}

		operandType := proto.FindField(schema.Models, strcase.ToCamel(modelTarget), fieldName).Type.Type
		return operandType, nil
	case resolver.IsImplicitInput():
		modelTarget := strcase.ToCamel(operation.ModelName)
		inputName := operand.Ident.Fragments[0].Fragment
		operandType := proto.FindField(schema.Models, modelTarget, inputName).Type.Type
		return operandType, nil
	case resolver.IsExplicitInput():
		inputName := operand.Ident.Fragments[0].Fragment
		input := proto.FindInput(operation, inputName)
		return input.Type.Type, nil
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
		isLiteral, _ := resolver.Operand.IsLiteralType()
		if isLiteral {
			return ToNative(resolver.Operand, operandType)
		} else if resolver.Operand.Ident != nil && proto.EnumExists(resolver.Schema.Enums, resolver.Operand.Ident.Fragments[0].Fragment) {
			return resolver.Operand.Ident.Fragments[1].Fragment, nil
		} else {
			return nil, errors.New("unknown literal type")
		}
	case resolver.IsImplicitInput(), resolver.IsExplicitInput():
		inputName := resolver.Operand.Ident.Fragments[0].Fragment
		value, ok := args[inputName]
		if !ok {
			return nil, fmt.Errorf("implicit or explicit input '%s' does not exist in arguments", inputName)
		}
		return value, nil
	case resolver.IsDatabaseColumn():
		// todo: https://linear.app/keel/issue/RUN-153/set-attribute-to-support-targeting-database-fields
		panic("cannot resolve operand value when IsDatabaseColumn() is true")
	case resolver.Operand.Ident.IsContextIdentityField():
		isAuthenticated := runtimectx.IsAuthenticated(resolver.Context)
		if !isAuthenticated {
			return nil, nil
		}

		identity, err := runtimectx.GetIdentity(resolver.Context)
		if err != nil {
			return nil, err
		}
		return identity.Id, nil
	case resolver.Operand.Ident.IsContextIsAuthenticatedField():
		isAuthenticated := runtimectx.IsAuthenticated(resolver.Context)
		return isAuthenticated, nil
	case resolver.Operand.Ident.IsContextNowField():
		return runtimectx.GetNow(), nil
	case resolver.Operand.Ident.IsContextEnvField():
		envVarName := resolver.Operand.Ident.Fragments[2].Fragment
		return os.Getenv(envVarName), nil
	case resolver.Operand.Type() == parser.TypeArray:
		return nil, fmt.Errorf("cannot yet handle operand of type non-literal array")
	default:
		return nil, fmt.Errorf("cannot handle operand of unknown type")
	}
}

func ToNative(v *parser.Operand, fieldType proto.Type) (any, error) {
	if v.Array != nil {
		values := []any{}
		for _, v := range v.Array.Values {
			value, err := ToNative(v, fieldType)
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

package actions

import (
	"context"
	"errors"
	"fmt"
	"net/textproto"
	"os"
	"strings"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
)

// GenerateCtxQuery visits the expression and adds filter conditions to the provided query builder.
func GenerateCtxQuery(ctx context.Context, query *QueryBuilder, schema *proto.Schema) resolve.Visitor[*QueryBuilder] {
	return &ctxQueryGen{
		ctx:       ctx,
		query:     query,
		schema:    schema,
		operators: arraystack.New(),
		operands:  arraystack.New(),
	}
}

var _ resolve.Visitor[*QueryBuilder] = new(ctxQueryGen)

type ctxQueryGen struct {
	ctx       context.Context
	query     *QueryBuilder
	schema    *proto.Schema
	operators *arraystack.Stack
	operands  *arraystack.Stack
}

func (v *ctxQueryGen) StartTerm(nested bool) error {
	if op, ok := v.operators.Peek(); ok && op == Not {
		_, _ = v.operators.Pop()
		v.query.Not()
	}

	// Only add parenthesis if we're in a nested condition
	if nested {
		v.query.OpenParenthesis()
	}

	return nil
}

func (v *ctxQueryGen) EndTerm(nested bool) error {
	if _, ok := v.operators.Peek(); ok && v.operands.Size() == 2 {
		operator, _ := v.operators.Pop()

		r, ok := v.operands.Pop()
		if !ok {
			return errors.New("expected rhs operand")
		}
		l, ok := v.operands.Pop()
		if !ok {
			return errors.New("expected lhs operand")
		}

		lhs := l.(*QueryOperand)
		rhs := r.(*QueryOperand)

		v.query.And()
		err := v.query.Where(lhs, operator.(ActionOperator), rhs)
		if err != nil {
			return err
		}
	} else if _, ok := v.operators.Peek(); !ok {
		l, hasOperand := v.operands.Pop()
		if hasOperand {
			lhs := l.(*QueryOperand)
			v.query.And()
			err := v.query.Where(lhs, Equals, Value(true))
			if err != nil {
				return err
			}
		}
	}

	// Only close parenthesis if we're nested
	if nested {
		v.query.CloseParenthesis()
	}

	return nil
}

func (v *ctxQueryGen) StartFunction(name string) error {
	return nil
}

func (v *ctxQueryGen) EndFunction() error {
	return nil
}

func (v *ctxQueryGen) StartArgument(num int) error {
	return nil
}

func (v *ctxQueryGen) EndArgument() error {
	return nil
}

func (v *ctxQueryGen) VisitAnd() error {
	v.query.And()
	return nil
}

func (v *ctxQueryGen) VisitOr() error {
	v.query.Or()
	return nil
}

func (v *ctxQueryGen) VisitNot() error {
	v.operators.Push(Not)
	return nil
}

func (v *ctxQueryGen) VisitOperator(op string) error {
	operator, err := toActionOperator(op)
	if err != nil {
		return err
	}

	v.operators.Push(operator)

	return nil
}

func (v *ctxQueryGen) VisitLiteral(value any) error {
	if value == nil {
		v.operands.Push(Null())
	} else {
		v.operands.Push(Value(value))
	}
	return nil
}

func (v *ctxQueryGen) VisitIdent(ident *parser.ExpressionIdent) error {
	operand, err := generateOperandForCtxQuery(v.ctx, v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	if ident.Fragments[0] == "ctx" && ident.Fragments[1] == "identity" {
		idents := ident.Fragments[1:]

		err = v.query.AddJoinFromFragments(v.schema, idents)
		if err != nil {
			return err
		}

		identityId := ""
		if auth.IsAuthenticated(v.ctx) {
			identity, err := auth.GetIdentity(v.ctx)
			if err != nil {
				return err
			}
			identityId = identity[parser.FieldNameId].(string)
		}

		err = v.query.Where(IdField(), Equals, Value(identityId))
		if err != nil {
			return err
		}
	}

	v.operands.Push(operand)

	return nil
}

func (v *ctxQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	arr := []string{}
	for _, e := range idents {
		arr = append(arr, e.Fragments[1])
	}

	v.operands.Push(Value(arr))

	return nil
}

func (v *ctxQueryGen) Result() (*QueryBuilder, error) {
	return v.query, nil
}

func generateOperandForCtxQuery(ctx context.Context, schema *proto.Schema, fragments []string) (*QueryOperand, error) {
	ident, err := NormaliseFragments(schema, fragments)
	if err != nil {
		return nil, err
	}

	switch {
	case expressions.IsContextDbColumn(ident):
		return operandFromFragments(schema, ident[1:])
	case expressions.IsContextIdentityId(ident):
		isAuthenticated := auth.IsAuthenticated(ctx)
		if !isAuthenticated {
			return Null(), nil
		} else {
			identity, err := auth.GetIdentity(ctx)
			if err != nil {
				return nil, err
			}
			return Value(identity[parser.FieldNameId].(string)), nil
		}
	case expressions.IsContextIsAuthenticatedField(ident):
		isAuthenticated := auth.IsAuthenticated(ctx)
		return Value(isAuthenticated), nil
	case expressions.IsContextNowField(ident):
		return Value(runtimectx.GetNow()), nil
	case expressions.IsContextEnvField(ident):
		envVarName := ident[2]
		return Value(os.Getenv(envVarName)), nil
	case expressions.IsContextSecretField(ident):
		secret, err := runtimectx.GetSecret(ctx, ident[2])
		if err != nil {
			return nil, err
		}
		return Value(secret), nil
	case expressions.IsContextHeadersField(ident):
		headerName := ident[2]

		// First we parse the header name to kebab. MyCustomHeader will become my-custom-header.
		kebab := strcase.ToKebab(headerName)

		// Then get canonical name. my-custom-header will become My-Custom-Header.
		// https://pkg.go.dev/net/http#Header.Get
		canonicalName := textproto.CanonicalMIMEHeaderKey(kebab)

		headers, err := runtimectx.GetRequestHeaders(ctx)
		if err != nil {
			return nil, err
		}
		if value, ok := headers[canonicalName]; ok {
			return Value(strings.Join(value, ", ")), nil
		} else {
			return Value(""), nil
		}
	}

	return nil, fmt.Errorf("cannot handle ctx fragments: %s", strings.Join(ident, "."))
}

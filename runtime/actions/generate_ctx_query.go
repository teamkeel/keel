package actions

import (
	"context"
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
	entity := proto.FindModel(schema.GetModels(), "Identity")

	return &baseQueryGen{
		ctx:       ctx,
		query:     query,
		schema:    schema,
		entity:    entity,
		action:    nil,
		inputs:    nil,
		operators: arraystack.New(),
		operands:  arraystack.New(),
		identHandler: func(ctx context.Context, query *QueryBuilder, schema *proto.Schema, ident *parser.ExpressionIdent, operands *arraystack.Stack) error {
			operand, err := generateOperandForCtxQuery(ctx, schema, ident.Fragments)
			if err != nil {
				return err
			}

			if ident.Fragments[0] == "ctx" && ident.Fragments[1] == "identity" {
				idents := ident.Fragments[1:]

				err = query.AddJoinFromFragments(schema, idents)
				if err != nil {
					return err
				}

				identityId := ""
				if auth.IsAuthenticated(ctx) {
					identity, err := auth.GetIdentity(ctx)
					if err != nil {
						return err
					}
					identityId = identity[parser.FieldNameId].(string)
				}

				err = query.Where(IdField(), Equals, Value(identityId))
				if err != nil {
					return err
				}
			}

			operands.Push(operand)

			return nil
		},
	}
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

package actions

import (
	"context"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// GenerateComputedFunction visits the expression and generates a SQL expression.
func GenerateComputedFunction(ctx context.Context, schema *proto.Schema, entity proto.Entity, field *proto.Field) resolve.Visitor[string] {
	gen := GenerateSQLExpression(SQLExpressionGeneratorConfig{
		Ctx:           ctx,
		Schema:        schema,
		Entity:        entity,
		Action:        nil,
		Inputs:        nil,
		TableAlias:    "r",
		ResultType:    field.GetType().GetType(),
		EmbedLiterals: true,
	})

	return &computedQueryGenWrapper{gen: gen}
}

// computedQueryGenWrapper adapts SQLExpression visitor to return string (for backward compatibility)
type computedQueryGenWrapper struct {
	gen resolve.Visitor[*SQLExpression]
}

func (w *computedQueryGenWrapper) StartTerm(nested bool) error {
	return w.gen.StartTerm(nested)
}

func (w *computedQueryGenWrapper) EndTerm(nested bool) error {
	return w.gen.EndTerm(nested)
}

func (w *computedQueryGenWrapper) StartFunction(name string) error {
	return w.gen.StartFunction(name)
}

func (w *computedQueryGenWrapper) EndFunction() error {
	return w.gen.EndFunction()
}

func (w *computedQueryGenWrapper) StartArgument(num int) error {
	return w.gen.StartArgument(num)
}

func (w *computedQueryGenWrapper) EndArgument() error {
	return w.gen.EndArgument()
}

func (w *computedQueryGenWrapper) VisitAnd() error {
	return w.gen.VisitAnd()
}

func (w *computedQueryGenWrapper) VisitOr() error {
	return w.gen.VisitOr()
}

func (w *computedQueryGenWrapper) VisitNot() error {
	return w.gen.VisitNot()
}

func (w *computedQueryGenWrapper) VisitOperator(op string) error {
	return w.gen.VisitOperator(op)
}

func (w *computedQueryGenWrapper) VisitLiteral(value any) error {
	return w.gen.VisitLiteral(value)
}

func (w *computedQueryGenWrapper) VisitIdent(ident *parser.ExpressionIdent) error {
	return w.gen.VisitIdent(ident)
}

func (w *computedQueryGenWrapper) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return w.gen.VisitIdentArray(idents)
}

func (w *computedQueryGenWrapper) Result() (string, error) {
	expr, err := w.gen.Result()
	if err != nil {
		return "", err
	}
	return expr.SQL, nil
}

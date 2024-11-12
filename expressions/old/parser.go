package expressions

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/iancoleman/strcase"

	"github.com/google/cel-go/checker/decls"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// Parser performs parsing, validation and query building of Keel expressions
type Parser struct {
	env *cel.Env
	ast *cel.Ast
}

func NewParser(schema *proto.Schema, model *proto.Model) (*Parser, error) {

	typeProvider := NewTypeProvider(schema)

	env, err := cel.NewCustomEnv(
		KeelLib(),
		cel.ClearMacros(),
		cel.CustomTypeProvider(typeProvider),
		cel.Declarations(
			decls.NewVar(strcase.ToLowerCamel(model.Name), decls.NewObjectType(model.Name)),
			decls.NewVar("ctx", decls.NewObjectType("Context")),
		),
		cel.EagerlyValidateDeclarations(true),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	return &Parser{
		env: env,
	}, nil
}

// Validate parses and validates the expression
func (p *Parser) Validate(expression string, expectedOutoutType *proto.TypeInfo) ([]string, error) {
	ast, issues := p.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		validationErrors := []string{}
		for _, e := range issues.Errors() {
			validationErrors = append(validationErrors, e.Message)
		}
		return validationErrors, nil
	}

	if ast.OutputType() != fromKeel(expectedOutoutType) {
		return []string{fmt.Sprintf("expression expected to resolve to type '%s'", expectedOutoutType.GetType())}, nil
	}

	p.ast = ast

	// Valid expression
	return nil, nil
}

// Evaluate will evaluate the expression in-proc if possible (early evaluation) without hitting the DB
func (p *Parser) Evaluate(schema *proto.Schema, model *proto.Model, expression string) {

}

// Build will construct a SQL statement for the expression
func (p *Parser) Build(query *actions.QueryBuilder, expression string, input map[string]any) error {
	checkedExpr, err := cel.AstToCheckedExpr(p.ast)
	if err != nil {
		return err
	}

	un := &builder{
		query: query,
	}
	if err := un.visit(checkedExpr.Expr); err != nil {
		return err
	}

	return nil
}

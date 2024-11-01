package expressions

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/iancoleman/strcase"

	"github.com/google/cel-go/checker/decls"
	"github.com/teamkeel/keel/proto"
)

func Validate(schema *proto.Schema, model *proto.Model, expression string, expectedOutoutType *proto.TypeInfo) ([]string, error) {

	typeProvider := NewTypeProvider(schema)

	env, err := cel.NewCustomEnv(
		KeelLib(),

		cel.CustomTypeProvider(typeProvider),

		cel.Declarations(
			decls.NewVar(strcase.ToLowerCamel(model.Name), decls.NewObjectType(model.Name)),
			decls.NewVar("ctx", decls.NewObjectType("Context")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	ast, issues := env.Compile(expression)
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

	// Valid expression
	return nil, nil
}

func Evaluate(schema *proto.Schema, model *proto.Model, expression string) {

}

func ToSQL(schema *proto.Schema, model *proto.Model, expression string) (string, error) {
	typeProvider := NewTypeProvider(schema)

	env, err := cel.NewCustomEnv(
		KeelLib(),
		cel.CustomTypeProvider(typeProvider),
		cel.Declarations(
			decls.NewVar(strcase.ToLowerCamel(model.Name), decls.NewObjectType(model.Name)),
			decls.NewVar("ctx", decls.NewObjectType("Context")),
		),
	)
	if err != nil {
		return "", fmt.Errorf("program setup err: %s", err)
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return "", issues.Err()
	}

	sql, err := Convert(ast)

	return sql, err

}

package resolve

import (
	"errors"

	"github.com/google/cel-go/cel"
)

// AsString expects and evaluates a string from the expression
func AsString(expression string) (string, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return "", errors.New("could not")
	}

	ast, issues := env.Parse(expression)
	if issues != nil && len(issues.Errors()) > 0 {
		return "", errors.New("could not")
	}

	prg, err := env.Program(ast)
	if err != nil {
		return "", errors.New("could not")
	}

	out, _, err := prg.Eval(map[string]any{})

	if err != nil {
		return "", err
	}

	return out.Value().(string), nil
}

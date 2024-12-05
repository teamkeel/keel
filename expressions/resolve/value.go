package resolve

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
)

// AsString expects and evaluates a string from the expression
func ToValue[T any](expression string) (T, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return *new(T), errors.New("could not")
	}

	ast, issues := env.Parse(expression)
	if issues != nil && len(issues.Errors()) > 0 {
		return *new(T), errors.New("could not")
	}

	prg, err := env.Program(ast)
	if err != nil {
		return *new(T), errors.New("could not")
	}

	out, _, err := prg.Eval(map[string]any{})

	if err != nil {
		return *new(T), err
	}

	if value, ok := out.Value().(T); ok {
		return value, nil
	} else {
		return *new(T), fmt.Errorf("value is of type '%T' and cannot assert type '%T'", out.Value(), *new(T))
	}
}

package resolve

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"github.com/teamkeel/keel/schema/parser"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToValue expects and resolves to a specific type by evaluating the expression
func ToValue[T any](expression *parser.Expression) (T, bool, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return *new(T), false, err
	}

	ast, issues := env.Parse(expression.String())
	if issues != nil && len(issues.Errors()) > 0 {
		return *new(T), false, errors.New("expression has validation errors and cannot be evaluated")
	}

	prg, err := env.Program(ast)
	if err != nil {
		return *new(T), false, err
	}

	out, _, err := prg.Eval(map[string]any{})
	if err != nil {
		return *new(T), false, err
	}

	value := out.Value()

	if _, ok := value.(structpb.NullValue); ok {
		return *new(T), true, nil
	} else if value, ok := value.(T); ok {
		return value, false, nil
	} else {
		return *new(T), false, fmt.Errorf("value is of type '%T' and cannot assert type '%T'", out.Value(), *new(T))
	}
}

// ToValueArray expects and resolves to a specific array type by evaluating the expression
func ToValueArray[T any](expression *parser.Expression) ([]T, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return nil, err
	}

	ast, issues := env.Parse(expression.String())
	if issues != nil && len(issues.Errors()) > 0 {
		return nil, errors.New("expression has validation errors and cannot be evaluated")
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	out, _, err := prg.Eval(map[string]any{})
	if err != nil {
		return nil, err
	}

	values, ok := out.Value().([]ref.Val)
	if !ok {
		return nil, errors.New("value is not an array")
	}
	arr := *new([]T)
	for _, v := range values {
		item, ok := v.Value().(T)
		if !ok {
			return nil, errors.New("element is not correct type")
		}
		arr = append(arr, item)
	}

	return arr, nil
}

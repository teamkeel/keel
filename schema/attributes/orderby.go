package attributes

import (
	"errors"

	"github.com/teamkeel/keel/expressions/attributes/expressions"
)

var ErrNotValue = errors.New("expression is not a single value")

func NewOrderByParser() (*expressions.ExpressionParser, error) {
	p, err := expressions.NewParser()
	if err != nil {
		return nil, err
	}

	return p, nil

}

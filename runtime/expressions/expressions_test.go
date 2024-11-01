package expressions_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/parser"
)

func TestThing(t *testing.T) {

	exp := "thing.id"

	expression, err := parser.ParseExpression(exp)
	require.NoError(t, err)

	fmt.Println(expression.Conditions()[0].LHS.Type())

	require.Equal(t, expression.Conditions()[0].Type(), "")
}

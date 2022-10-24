package actions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func FooTest(t *testing.T) {
	scope := makeScope(t)
	rslv := NewExpressionResolver(scope)
	qry, err := rslv.Resolve(expr, args)
	require.NoError(t, err)
}

func makeScope(t *testing.T) *Scope {

}

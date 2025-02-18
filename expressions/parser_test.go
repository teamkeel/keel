package expressions_test

import (
	"testing"

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/test-go/testify/require"
)

func TestNewOptions(t *testing.T) {
	opt1 := options.WithVariable("first", "Text", false)

	parser, err := expressions.NewParser(opt1)
	require.NoError(t, err)

	expr1 := "first"
	_, issues := parser.CelEnv.Compile(expr1)
	require.Len(t, issues.Errors(), 0)

	opt2 := options.WithVariable("second", "Text", false)

	err = opt2(parser)
	require.NoError(t, err)

	_, issues = parser.CelEnv.Compile(expr1)
	require.Len(t, issues.Errors(), 0)

	expr2 := "second"
	_, issues = parser.CelEnv.Compile(expr2)
	require.Len(t, issues.Errors(), 0)
}

func TestExtendsVariables(t *testing.T) {
	opt1 := options.WithVariable("first", "Text", false)

	parser1, err := expressions.NewParser(opt1)
	require.NoError(t, err)

	expr1 := "first"
	_, issues := parser1.CelEnv.Compile(expr1)
	require.Len(t, issues.Errors(), 0)

	opt2 := options.WithVariable("second", "Text", false)

	parser2, err := parser1.Extend(opt2)
	require.NoError(t, err)

	_, issues = parser2.CelEnv.Compile(expr1)
	require.Len(t, issues.Errors(), 0)

	expr2 := "second"
	_, issues = parser2.CelEnv.Compile(expr2)
	require.Len(t, issues.Errors(), 0)

	// "second" should not exist in parser1
	_, issues = parser1.CelEnv.Compile(expr2)
	require.Len(t, issues.Errors(), 1)
}

func TestExtendsTypeProvider(t *testing.T) {
	parser1, err := expressions.NewParser()
	require.NoError(t, err)

	expr := "ctx.identity"
	_, issues := parser1.CelEnv.Compile(expr)
	require.Len(t, issues.Errors(), 1)

	opt2 := options.WithCtx()

	parser2, err := parser1.Extend(opt2)
	require.NoError(t, err)

	_, issues = parser2.CelEnv.Compile(expr)
	require.Len(t, issues.Errors(), 0)

	// "ctx.identity" should not exist in parser1
	_, issues = parser1.CelEnv.Compile(expr)
	require.Len(t, issues.Errors(), 1)
}

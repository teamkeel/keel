package actions

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"text/template"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestExpressionResolver(t *testing.T) {

	for _, tCase := range testCases {
		_ = tCase

		t.Run(tCase.name, func(t *testing.T) {

			// Set up the fixture required for this test - passing in parameters that
			// adjust the generated schema to this test case. E.g. the model field's type,
			// or the operation's where expression.
			scope, sqlDb := makeDbAndScope(t, schemaParams{FieldType: tCase.fieldType})
			defer sqlDb.Close()

			// Fish out the Where expression to test from the schema.
			exprSource := scope.operation.WhereExpressions[0].Source
			parsedExpr, err := parser.ParseExpression(exprSource)
			require.NoError(t, err)

			// Create some input arguments to represent an incoming request.
			// Todo: these should come from the testCase.
			requestArgs := RequestArguments{
				"coolTitle": "Good Morning",
			}

			writeArgs := map[string]any{}

			// Fire the function under test.
			rslv := NewExpressionResolver(scope)
			updatedQry, err := rslv.ResolveQueryStatement(parsedExpr, requestArgs, writeArgs)
			require.NoError(t, err)

			// Fish out the now-populated gorm data structures that represent the generated SQL for inspection.
			c := findGormWhereClause(t, updatedQry)

			// Fire the assertions.
			// Todo: these should come from the test case.
			require.Equal(t, "my_model.my_field = ?", c.SQL)
			require.Equal(t, "Good Morning", c.Vars[0])
		})

	}
}

// makeDbAndScope constructs a Scope based on a keel schema,
// and with a mock database.
func makeDbAndScope(t *testing.T, schemaParams schemaParams) (*Scope, *sql.DB) {
	sqldb, _, err := sqlmock.New()
	require.NoError(t, err)

	gormdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	require.NoError(t, err)

	schema := protoSchema(t, parameterisedSchema(t, schemaParams))
	op := proto.FindOperation(schema, "myOperation")

	ctx := runtimectx.WithDatabase(context.Background(), gormdb)
	scope, err := NewScope(ctx, op, schema, nil)
	require.NoError(t, err)
	return scope, sqldb
}

// protoSchema generates a proto.Schema based on the given keel schema string.
func protoSchema(t *testing.T, keelSchema string) *proto.Schema {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromInputs(&reader.Inputs{
		SchemaFiles: []reader.SchemaFile{
			{
				Contents: keelSchema,
			},
		},
	})
	require.NoError(t, err)
	return schema
}

// findGormWhereClause extracts the clause.Expr that represents the
// Where clause from inside the given *gorm.DB query structure.
func findGormWhereClause(t *testing.T, qry *gorm.DB) clause.Expr {
	clauses, ok := qry.Statement.Clauses["WHERE"]
	require.True(t, ok)
	asWhere, ok := clauses.Expression.(clause.Where)
	require.True(t, ok)
	first := asWhere.Exprs[0]
	asExpr, ok := first.(clause.Expr)
	require.True(t, ok)
	return asExpr
}

// A schemaTemplate is a keel (text) schema, with placeholders that can
// be replaced using go's template.Template.
const schemaTemplate string = `
model MyModel {
	fields {
		myField {{.FieldType}}
	}
	operations {
		list myOperation(coolTitle: Text) {
			@where(myModel.myField == coolTitle )
		}
	}
}
`

// parameterisedSchema generates a template-based keel schema using
// the given plug-in values.
func parameterisedSchema(t *testing.T, pluginValues schemaParams) string {
	tmpl, err := template.New("test").Parse(schemaTemplate)
	require.NoError(t, err)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pluginValues)
	require.NoError(t, err)
	return buf.String()
}

type schemaParams struct {
	FieldType string
}

type testCase struct {
	name      string
	fieldType string
}

var testCases []testCase = []testCase{
	{
		name:      "getting started",
		fieldType: "Text",
	},
}

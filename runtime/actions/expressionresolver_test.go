package actions

import (
	"context"
	"database/sql"
	"testing"

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
	scope, sqlDb := makeDbAndScope(t)
	defer sqlDb.Close()

	rslv := NewExpressionResolver(scope)
	expr, err := parser.ParseExpression(`post.Title == coolTitle`)
	require.NoError(t, err)

	updatedQry, err := rslv.Resolve(
		expr,
		RequestArguments{
			"coolTitle": "Good Morning",
		})
	require.NoError(t, err)

	c := findGormWhereClause(t, updatedQry)

	require.Equal(t, "post.title IS ?", c.SQL)
	require.Equal(t, "Good Morning", c.Vars[0])
}

// makeDbAndScope constructs a Scope based on a keel schema,
// and with a mock database.
func makeDbAndScope(t *testing.T) (*Scope, *sql.DB) {
	sqldb, _, err := sqlmock.New()
	require.NoError(t, err)

	gormdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	require.NoError(t, err)

	schema := protoSchema(t, schemaString)
	op := proto.FilterOperations(schema, func(op *proto.Operation) bool {
		return op.Name == "listPosts"
	})[0]

	ctx := runtimectx.WithDatabase(context.Background(), gormdb)
	scope, err := NewScope(ctx, op, schema)
	require.NoError(t, err)
	return scope, sqldb
}

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

const schemaString string = `
model Post {
	fields {
		title Text 
	}
	operations {
		list listPosts(coolTitle: Text) {
			@where(post.title == coolTitle )
		}
	}
}
`

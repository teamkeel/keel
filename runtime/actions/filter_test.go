package actions

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TestAddFilter exercises the addFilter() function with a variety of operator and operand types,
// and checks that the generated SQL inside the generated gorm Statement is what it should be.
// It uses go-sqlmock to configure gorm instead of a live connection to a real database.
func TestAddFilter(t *testing.T) {

	for _, theCase := range buildTestCases() {
		t.Run(theCase.name, func(t *testing.T) {
			scope, sqldb, _ := initDbAndScope(t)
			defer sqldb.Close()

			// Call the function under test.
			err := addFilter(scope, "myCol", theCase.operator, theCase.operand)
			require.NoError(t, err)

			w := findGeneratedSqlWhere(t, scope)

			// Check for correctly generated SQL inside gorm.
			require.Equal(t, theCase.expectedSQL, w.SQL)       // e.g. "my_col = ?""
			require.Equal(t, theCase.expectedValue, w.Vars[0]) // e.g. "harry"
		})
	}
}

// initDbAndScope constructs a gorm.DB using a mocked sql.DB, and partially
// constructs a Scope object with that gorm.DB, so it can be used as an argument
// to the addFilter function - i.e. the function under test.
func initDbAndScope(t *testing.T) (*Scope, *sql.DB, *gorm.DB) {
	sqldb, _, err := sqlmock.New()
	require.NoError(t, err)

	gormdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	require.NoError(t, err)

	scope := &Scope{ // only the query field is needed for this testing use case.
		query: gormdb,
	}
	return scope, sqldb, gormdb
}

// findGeneratedSqlWhere navigates the gorm.DB inside the given Scope object
// to reach the generated WHERE statement clauses and returns them in the form of
// a gorm clause.Expr - which is good for making assertions on.
func findGeneratedSqlWhere(t *testing.T, scope *Scope) clause.Expr {
	// Note: for reasons I do not understand, you can't simply look at the Statement field in
	// scope.query (which is a *gorm.DB), I could only find the required data with the
	// navigation coded below.
	expr := scope.query.Statement.DB.Clauses().Statement.Clauses["WHERE"].Expression
	asWhere, ok := expr.(clause.Where)
	require.True(t, ok)
	first := asWhere.Exprs[0]
	asExpr, ok := first.(clause.Expr)
	require.True(t, ok)
	return asExpr
}

type testCase struct {
	name          string
	operator      ActionOperator
	operand       any
	expectedSQL   string
	expectedValue any
}

type typeAndValue struct {
	typeName string
	operand  any
}

func buildTestCases() []testCase {
	testCases := []testCase{}

	// Equals and NotEquals.
	for _, operator := range []ActionOperator{Equals, NotEquals} {
		symbol := lo.Ternary(operator == Equals, "=", "!=")
		for _, tv := range []typeAndValue{
			{"string", "harry"},
			{"int", 42},
			{"bool", true},
			{"time", time.Time{}},
		} {
			testCases = append(testCases, testCase{
				name:          fmt.Sprintf("%v %t", operator, tv.operand), // e.g. "Equals 42"
				operator:      operator,
				operand:       tv.operand,
				expectedSQL:   fmt.Sprintf("my_col %s ?", symbol), // e.g. "my_col != ?"
				expectedValue: tv.operand,
			})
		}
	}

	// StartsWith
	testCases = append(testCases, testCase{
		name:          "startsWith",
		operator:      StartsWith,
		operand:       "Fr",
		expectedSQL:   "my_col LIKE ?",
		expectedValue: "Fr%",
	})

	// EndsWith
	testCases = append(testCases, testCase{
		name:          "endsWith",
		operator:      EndsWith,
		operand:       "Fr",
		expectedSQL:   "my_col LIKE ?",
		expectedValue: "%Fr",
	})

	// Contains
	testCases = append(testCases, testCase{
		name:          "contains",
		operator:      Contains,
		operand:       "Fr",
		expectedSQL:   "my_col LIKE ?",
		expectedValue: "%Fr%",
	})

	// OneOf (string)
	testCases = append(testCases, testCase{
		name:          "oneOf string",
		operator:      OneOf,
		operand:       []any{"apple", "pear"},
		expectedSQL:   "my_col in ?",
		expectedValue: []interface{}{"apple", "pear"},
	})

	// OneOf (number)
	testCases = append(testCases, testCase{
		name:          "oneOf number",
		operator:      OneOf,
		operand:       []any{41, 42},
		expectedSQL:   "my_col in ?",
		expectedValue: []interface{}{41, 42},
	})

	// LessThan
	testCases = append(testCases, testCase{
		name:          "lessThan",
		operator:      LessThan,
		operand:       42,
		expectedSQL:   "my_col < ?",
		expectedValue: 42,
	})

	// GreaterThan
	testCases = append(testCases, testCase{
		name:          "greaterThan",
		operator:      GreaterThan,
		operand:       42,
		expectedSQL:   "my_col > ?",
		expectedValue: 42,
	})

	return testCases
}

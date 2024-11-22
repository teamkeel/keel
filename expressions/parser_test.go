package expressions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func TestParser_Variable(t *testing.T) {
	expression := `myVar == "Keel"`

	parser, err := NewParser(
		WithCtx(),
		WithVariable("myVar", parser.FieldTypeText),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_TextEquality(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression := `person.name == 'Keel'`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_TextInequality(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression := `person.name != 'Keel'`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_Array(t *testing.T) {
	expression := `[1,2,3]`

	parser, err := NewParser(
		WithCtx(),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeNumber, true))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ExpectedArray(t *testing.T) {
	expression := `[1,2,3]`

	parser, err := NewParser(
		WithCtx(),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeNumber, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'int' but it is 'list(int)'", issues[0])
}

func TestParser_In(t *testing.T) {
	expression := `1 in [1,2,3]`

	parser, err := NewParser(
		WithCtx(),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_InInvalid(t *testing.T) {
	expression := `"keel" in "keel"`

	parser, err := NewParser(
		WithCtx(),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '@in' applied to '(string, string)'", issues[0])
}

func TestParser_InInvalidTypes(t *testing.T) {
	expression := `"keel" in [1,2,3]`

	parser, err := NewParser(
		WithCtx(),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '@in' applied to '(string, list(int))'", issues[0])
}

func TestParser_UnknownVariable(t *testing.T) {
	expression := `person.name == 'Keel'`

	parser, err := NewParser(
		WithCtx(),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undeclared reference to 'person' (in container '')", issues[0])
}

func TestParser_UnknownField(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression := `person.n == 'Keel'`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'n'", issues[0])
}

func TestParser_UnknownOperators(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				age Number
			}
		}`})

	expression := `person.age == 1 + 1`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 2)
	require.Equal(t, "undeclared reference to '_==_' (in container '')", issues[0])
	require.Equal(t, "undeclared reference to '_+_' (in container '')", issues[1])
}

func TestParser_TypeMismatch(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression := `person.name == 123`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(string, int)'", issues[0])
}

func TestParser_ReturnAssertion(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression := `person.name`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type 'bool' but it is 'string'", issues[0])
}

func TestParser_EnumEquals(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status == Status.Married`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_EnumNotEquals(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status != Status.Married`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_EnumInvalidOperator(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status > Status.Married`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_>_' applied to '(Status, Status)'", issues[0])
}

func TestParser_EnumInvalidValue(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status == Status.NotExists`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'NotExists'", issues[0])
}

func TestParser_EnumWithoutValue(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
			}
		}
		enum Status {
			Married
			Single
		}`})

	expression := `person.status == Status`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(Status, Status_EnumDefinition)'", issues[0])
}

func TestParser_EnumTypeMismatch(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				status Status
				employment Employment
			}
		}
		enum Status {
			Married
			Single
		}
		enum Employment {
			Permanent
			Temporary
			Unemployed
		}`})

	expression := `person.status == Employment.Permanent`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(Status, Employment)'", issues[0])
}

func TestParser_ArrayString(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				names Text[]
			}
		}`})

	expression := `person.names == ["Keel","Weave"]`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ArrayInt(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				numbers Number[]
			}
		}`})

	expression := `person.numbers == [-1,2,3]`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ArrayDouble(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				numbers Decimal[]
			}
		}`})

	expression := `person.numbers == [1.2, 2.1, 3.9]`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ArrayEmpty(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				names Text[]
			}
		}`})

	expression := `person.names == []`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ArrayTypeMismatch(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				names Text[]
			}
		}`})

	expression := `person.names == 'Keel'`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '_==_' applied to '(list(string), string)'", issues[0])
}

func TestParser_ModelEquals(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				p Person?
			}
		}`})

	expression := `person == person.p`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ModelIn(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
			model Account {
				fields {
					identity Identity @unique
					friends Befriend[]
				}
			}
			model Befriend {
				fields {
					follower Account 
				}
				@unique([follower, followee])
			}
			model Identity {
				fields {
					account Account
				}
			}`})

	expression := `account in ctx.identity.account.friends.follower`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("account", "Account"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
	require.Len(t, issues, 0)
}

func TestParser_ModelInNotToMany(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
			model Account {
				fields {
					identity Identity @unique
				}
			}
			model Identity {
				fields {
					account Account
				}
			}`})

	expression := `account in ctx.identity.account`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("account", "Account"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '@in' applied to '(Account, Account)'", issues[0])
}

func TestParser_ModelInWrongType(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
			model Account {
				fields {
					identity Identity @unique
					friends Befriend[]
				}
			}
			model Befriend {
				fields {
					follower Account 
				}
				@unique([follower, followee])
			}
			model Identity {
				fields {
					account Account
				}
			}`})

	expression := `account in ctx.identity.account.friends`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("account", "Account"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "found no matching overload for '@in' applied to '(Account, Befriend[])'", issues[0])
}

func TestParser_ToOneRelationship(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				org Organisation
			}
		}
		model Organisation {
			fields {
				companyName Text
				people Person[]
			}
		}`})

	expression := `person.org.companyName == "Keel"`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("person", "Person"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ToManyRelationship(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				org Organisation
			}
		}
		model Organisation {
			fields {
				companyName Text
				people Person[]
			}
		}`})

	expression := `organisation.people.name == "Keel"`

	parser, err := NewParser(
		WithCtx(),
		WithSchema(schema),
		WithVariable("organisation", "Organisation"),
		WithComparisonOperators(),
		WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func parse(t *testing.T, s *reader.SchemaFile) []*parser.AST {
	schema, err := parser.Parse(s)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return []*parser.AST{schema}
}

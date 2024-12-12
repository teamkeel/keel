package expressions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func TestParser_Variable(t *testing.T) {
	expression, err := parser.ParseExpression(`myVar == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithVariable("myVar", parser.FieldTypeText, false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`person.name == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`person.name != "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_DateTimeEquality(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				created Timestamp
			}
		}`})

	expression, err := parser.ParseExpression(`person.created == ctx.now`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_DateTimeComparison(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
				created Timestamp
			}
		}`})

	expression, err := parser.ParseExpression(`person.created > ctx.now`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_NumberDecimalComparison(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				number Number
				decimal Decimal
			}
		}`})

	expression, err := parser.ParseExpression(`person.number > person.decimal`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_CompareNullWithRequiredField(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Post {
			fields {
				name Text?
			}
			actions {
				list listPosts() 
			}
		}`})

	expression, err := parser.ParseExpression(`post.name == null`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithSchemaTypes(schema),
		options.WithVariable("post", "Post", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_Array(t *testing.T) {
	expression, err := parser.ParseExpression(`[1,2,3]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithReturnTypeAssertion(parser.FieldTypeNumber, true))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_NotExpectedArray(t *testing.T) {
	expression, err := parser.ParseExpression(`[1,2,3]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithReturnTypeAssertion(parser.FieldTypeNumber, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Number but it is Number[]", issues[0].Message)
}

func TestParser_In(t *testing.T) {
	expression, err := parser.ParseExpression(`1 in [1,2,3]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_InInvalid(t *testing.T) {
	expression, err := parser.ParseExpression(`"keel" in "keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator 'in' with types Text and Text", issues[0].Message)
	require.Equal(t, errorhandling.LexerPos{Offset: 7, Line: 1, Column: 7}, issues[0].Pos)
	require.Equal(t, errorhandling.LexerPos{Offset: 9, Line: 1, Column: 9}, issues[0].EndPos)
}

func TestParser_InInvalidTypes(t *testing.T) {
	expression, err := parser.ParseExpression(`"keel" in [1,2,3]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator 'in' with types Text and Number[]", issues[0].Message)
}

func TestParser_UnknownVariable(t *testing.T) {
	expression, err := parser.ParseExpression(`person.name == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "unknown identifier 'person'", issues[0].Message)
}

func TestParser_UnknownField(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression, err := parser.ParseExpression(`person.n == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'n'", issues[0].Message)
}

func TestParser_UnknownOperators(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				age Number
			}
		}`})

	expression, err := parser.ParseExpression(`person.age == 1 + 1`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 2)
	require.Equal(t, "operator '==' not supported in this context", issues[0].Message)
	require.Equal(t, "operator '+' not supported in this context", issues[1].Message)
}

func TestParser_TypeMismatch(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression, err := parser.ParseExpression(`person.name == 123`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator '==' with types Text and Number", issues[0].Message)
}

func TestParser_ReturnAssertion(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				name Text
			}
		}`})

	expression, err := parser.ParseExpression(`person.name`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type Boolean but it is Text", issues[0].Message)
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

	expression, err := parser.ParseExpression(`person.status == Status.Married`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
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

	expression, err := parser.ParseExpression(`person.status != Status.Married`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
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

	expression, err := parser.ParseExpression(`person.status > Status.Married`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator '>' with types Status and Status", issues[0].Message)
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

	expression, err := parser.ParseExpression(`person.status == Status.NotExists`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "undefined field 'NotExists'", issues[0].Message)
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

	expression, err := parser.ParseExpression(`person.status == Status`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator '==' with types Status and Status_Enum", issues[0].Message)
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

	expression, err := parser.ParseExpression(`person.status == Employment.Permanent`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	)
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator '==' with types Status and Employment", issues[0].Message)
}

func TestParser_TimestampEquality(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				date Timestamp
			}
		}`})

	expression, err := parser.ParseExpression(`person.date == ctx.now`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)
	require.Empty(t, issues)
}

func TestParser_ArrayString(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				names Text[]
			}
		}`})

	expression, err := parser.ParseExpression(`person.names == ["Keel","Weave"]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`person.numbers == [-1,2,3]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`person.numbers == [1.2, 2.1, 3.9]`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`person.names == []`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`person.names == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator '==' with types Text[] and Text", issues[0].Message)
}

func TestParser_ModelEquals(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Person {
			fields {
				p Person?
			}
		}`})

	expression, err := parser.ParseExpression(`person == person.p`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`account in ctx.identity.account.friends.follower`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("account", "Account", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`account in ctx.identity.account`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("account", "Account", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator 'in' with types Account and Account", issues[0].Message)
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

	expression, err := parser.ParseExpression(`account in ctx.identity.account.friends`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable("account", "Account", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
	require.NoError(t, err)

	issues, err := parser.Validate(expression)
	require.NoError(t, err)

	require.Len(t, issues, 1)
	require.Equal(t, "cannot use operator 'in' with types Account and Befriend[]", issues[0].Message)
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

	expression, err := parser.ParseExpression(`person.org.companyName == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithSchemaTypes(schema),
		options.WithVariable("person", "Person", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

	expression, err := parser.ParseExpression(`organisation.people.name == "Keel"`)
	require.NoError(t, err)

	parser, err := expressions.NewParser(
		options.WithSchemaTypes(schema),
		options.WithVariable("organisation", "Organisation", false),
		options.WithComparisonOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false))
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

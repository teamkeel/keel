package attributes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/reader"
)

func TestUnique_Valid(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Account {
			fields {
				code Text
				country Country
			}
			@unique([code, country])
		}
		enum Country {
			ZA
			UK
			US
		}`})

	model := query.Model(schema, "Account")
	expression := model.Sections[1].Attribute.Arguments[0].Expression

	issues, err := attributes.ValidateCompositeUnique(model, expression)
	require.NoError(t, err)
	require.Len(t, issues, 0)
}

func TestUnique_NotArray(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Account {
			fields {
				code Text
				country Country
			}
			@unique(code)
		}
		enum Country {
			ZA
			UK
			US
		}`})

	model := query.Model(schema, "Account")
	expression := model.Sections[1].Attribute.Arguments[0].Expression

	issues, err := attributes.ValidateCompositeUnique(model, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)

	require.Equal(t, "expression expected to resolve to type FieldName[] but it is FieldName", issues[0].Message)
	require.Equal(t, 7, issues[0].Pos.Line)
	require.Equal(t, 12, issues[0].Pos.Column)
	require.Equal(t, 81, issues[0].Pos.Offset)
	require.Equal(t, 7, issues[0].EndPos.Line)
	require.Equal(t, 16, issues[0].EndPos.Column)
	require.Equal(t, 85, issues[0].EndPos.Offset)
}

func TestUnique_UnknownIdentifier(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Account {
			fields {
				code Text
				country Country
			}
			@unique([unknown])
		}
		enum Country {
			ZA
			UK
			US
		}`})

	model := query.Model(schema, "Account")
	expression := model.Sections[1].Attribute.Arguments[0].Expression

	issues, err := attributes.ValidateCompositeUnique(model, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)

	require.Equal(t, "unknown identifier 'unknown'", issues[0].Message)
	require.Equal(t, 7, issues[0].Pos.Line)
	require.Equal(t, 12, issues[0].Pos.Column)
	require.Equal(t, 82, issues[0].Pos.Offset)
	require.Equal(t, 7, issues[0].EndPos.Line)
	require.Equal(t, 12, issues[0].EndPos.Column)
	require.Equal(t, 89, issues[0].EndPos.Offset)
}

func TestUnique_ArrayField(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Account {
			fields {
				code Text
				country Country
				tags Text[]
			}
			@unique([code, tags])
		}
		enum Country {
			ZA
			UK
			US
		}`})

	model := query.Model(schema, "Account")
	expression := model.Sections[1].Attribute.Arguments[0].Expression

	issues, err := attributes.ValidateCompositeUnique(model, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)

	require.Equal(t, "unknown identifier 'tags'", issues[0].Message)
	require.Equal(t, 8, issues[0].Pos.Line)
	require.Equal(t, 12, issues[0].Pos.Column)
	require.Equal(t, 104, issues[0].Pos.Offset)
	require.Equal(t, 8, issues[0].EndPos.Line)
	require.Equal(t, 12, issues[0].EndPos.Column)
	require.Equal(t, 108, issues[0].EndPos.Offset)
}

func TestUnique_IncorrectType(t *testing.T) {
	schema := parse(t, &reader.SchemaFile{FileName: "test.keel", Contents: `
		model Account {
			fields {
				code Text
				country Country
				tags Text[]
			}
			@unique([code, 1])
		}
		enum Country {
			ZA
			UK
			US
		}`})

	model := query.Model(schema, "Account")
	expression := model.Sections[1].Attribute.Arguments[0].Expression

	issues, err := attributes.ValidateCompositeUnique(model, expression)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "expression expected to resolve to type FieldName[] but it is dyn[]", issues[0].Message)
}

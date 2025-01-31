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
}

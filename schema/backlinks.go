package schema

import (
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// insertAllBackLinkFields works out which of the Models in the schema have
// relationship fields that:
// 1) point to the Identity model
// 2) are of type hasOne
// 3) are marked @unique
//
// It is for these that we auto-add corresponding
// back-link relationship fields to the Identity model.
func (scm *Builder) insertAllBackLinkFields(
	asts []*parser.AST) *errorhandling.ErrorDetails {
	identityModel := query.Model(asts, parser.IdentityModelName)

	// Traverse all fields of all models to find "forward" relationships to Identity models.
	// And for each found, delegate to insertOneBackLinkField() the creation of
	// the corresponding backlink field in the Identity model.
	for _, model := range query.Models(asts) {
		if model == identityModel {
			continue
		}

		backlinkFields := []*parser.FieldNode{}
		for _, f := range query.ModelFields(model) {
			if f.Type.Value != parser.IdentityModelName {
				continue
			}
			if f.Repeated {
				continue
			}
			if !query.FieldHasAttribute(f, parser.AttributeUnique) {
				continue
			}

			backlinkFields = append(backlinkFields, f)
		}

		for _, f := range backlinkFields {
			if errorDetails := scm.insertBackLinkField(identityModel, model, f); errorDetails != nil {
				return errorDetails
			}
		}
	}
	return nil
}

func (scm *Builder) insertBackLinkField(
	identityModel *parser.ModelNode,
	parentModel *parser.ModelNode,
	forwardRelnField *parser.FieldNode) *errorhandling.ErrorDetails {
	// The backlink field is named after the name of the model it is back
	// linking to unless @relation is defined.  If @relation(myFieldName) exists,
	// then the backlink field will be named using the value provided (i.e. myFieldName).
	backlinkName := casing.ToLowerCamel(parentModel.Name.Value)
	relation := query.FieldGetAttribute(forwardRelnField, parser.AttributeRelation)
	if relation != nil {
		relationValue, _ := relation.Arguments[0].Expression.ToValue()
		backlinkName = relationValue.ToString()
	}

	// If the field already exists don't add another one as this will just create a
	// duplicate field name error that is confusing. This is will be an error but will
	// be caught by relationship validation since it is not possible for Identity
	// to have any fields which use a user-defined model
	if query.Field(identityModel, backlinkName) != nil {
		return nil
	}

	backLinkField := &parser.FieldNode{
		Name: parser.NameNode{
			Value: backlinkName,
		},
		Type: parser.NameNode{
			Value: parentModel.Name.Value,
		},
		Repeated: false,
		Optional: true,
		BuiltIn:  false,
	}

	for _, section := range identityModel.Sections {
		if section.Fields != nil {
			section.Fields = append(section.Fields, backLinkField)
		}
	}

	return nil
}

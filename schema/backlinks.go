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

	identityModel := query.Model(asts, parser.ImplicitIdentityModelName)

	// Traverse all fields of all models to find "forward" relationships to Identity models.
	// And for each found, delegate to insertOneBackLinkField() the creation of
	// the corresponding backlink field in the Identity model.
	for _, model := range query.Models(asts) {
		if model == identityModel {
			continue
		}

		//	hasNoRelationAttribute := false
		backlinkFields := []*parser.FieldNode{}
		for _, f := range query.ModelFields(model) {
			if f.Type.Value != parser.ImplicitIdentityModelName {
				continue
			}
			if f.Repeated {
				continue
			}
			if !query.FieldHasAttribute(f, parser.AttributeUnique) {
				continue
			}
			// if !query.FieldHasAttribute(f, parser.AttributeRelation) {
			// 	hasNoRelationAttribute = true
			// }

			backlinkFields = append(backlinkFields, f)
		}

		//	if !(len(backlinkFields) > 1 && hasNoRelationAttribute) {
		for i, f := range backlinkFields {
			if errorDetails := scm.insertOneBackLinkField(identityModel, asts, model, f, i); errorDetails != nil {
				return errorDetails
			}
		}
		//}
	}
	return nil
}

func (scm *Builder) insertOneBackLinkField(
	identityModel *parser.ModelNode,
	asts []*parser.AST,
	parentModel *parser.ModelNode,
	forwardRelnField *parser.FieldNode,
	i int) *errorhandling.ErrorDetails {

	// The backlink field is named (for now) after the name of the model it is back
	// linking to. For example "user".

	backlinkName := casing.ToLowerCamel(parentModel.Name.Value)
	relation := query.FieldGetAttribute(forwardRelnField, parser.AttributeRelation)
	if relation != nil {
		relationValue, _ := relation.Arguments[0].Expression.ToValue()
		backlinkName = relationValue.ToString()
	}

	backlinkType := parentModel.Name.Value

	backLinkField := &parser.FieldNode{
		Name: parser.NameNode{
			Value: backlinkName,
		},
		Type: parser.NameNode{
			Value: backlinkType,
		},
		Repeated: false,
		Optional: true,
		BuiltIn:  false, // TODO: needs to be false
	}

	// XXXX todo DRY up this boiler place - it is repeated all over the place, but all it does is insert a
	// given field into a given model.
	// and it is replicated elsewhere.
	for _, section := range identityModel.Sections {
		if section.Fields != nil {
			section.Fields = append(section.Fields, backLinkField)
		}
	}

	return nil
}

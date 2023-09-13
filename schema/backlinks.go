package schema

import (
	"fmt"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// insertAllBackLinkFields works out which of the Models in the schema have
// relationship fields that point to Identity models, and adds corresponding
// back-link relationship fields to the Identity model.
func (scm *Builder) insertAllBackLinkFields(
	asts []*parser.AST) *errorhandling.ErrorDetails {

	identityModel := query.Model(asts, parser.ImplicitIdentityModelName)

	// Traverse all fields of all models to find "forward" relationships to Identity models.
	// And for each found, delegate to insertOneBackLinkField() the creation of
	// the corresponding backlink field.
	for _, model := range query.Models(asts) {
		fmt.Printf("XXXX consider model: %s\n", model.Name.Value)
		if model == identityModel {
			continue
		}
		for _, f := range query.ModelFields(model) {
			if f.Type.Value != parser.ImplicitIdentityModelName {
				continue
			}
			if errorDetails := scm.insertOneBackLinkField(identityModel, asts, model, f); errorDetails != nil {
				return errorDetails
			}
		}
	}
	return nil
}

func (scm *Builder) insertOneBackLinkField(
	identityModel *parser.ModelNode,
	asts []*parser.AST,
	parentModel *parser.ModelNode,
	forwardRelnField *parser.FieldNode) *errorhandling.ErrorDetails {

	backlinkName := casing.ToLowerCamel(parentModel.Name.Value)
	backlinkType := parentModel.Name.Value

	repeated := false

	backLinkField := &parser.FieldNode{
		Name: parser.NameNode{
			Value: backlinkName,
		},
		Type: parser.NameNode{
			Value: backlinkType,
		},
		Repeated: repeated,
		Optional: false,
		BuiltIn:  true,
	}

	// XXXX todo DRY up this boiler place - all it does is insert a field into a model.
	// and it is replicated elsewhere.
	for _, section := range identityModel.Sections {
		if section.Fields != nil {
			section.Fields = append(section.Fields, backLinkField)
		}
	}

	return nil
}

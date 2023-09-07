package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type InvalidManyToManyDetails struct {
	InverseModel *parser.ModelNode
	InverseField *parser.FieldNode
	ThisModel    *parser.ModelNode
	ThisField    *parser.FieldNode
}

type ManyToOneRelation struct {
	ManyModel string
	ManyField string
	OneModel  string
	OneField  string
}

// DirectManyToManyRule checks that a direct many to many relationship between
// two models hasn't been defined, and recommends creating a join model.
func DirectManyToManyRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	registry := map[string]map[string]*InvalidManyToManyDetails{}
	relationRegistry := map[string]map[string]*ManyToOneRelation{}

	var currentModel *parser.ModelNode

	// This function checks for the existence of a relation attribute and, if one exists, stores the details of that
	// relation in a map keyed by the Many side of the relationship
	checkForRelation := func(currentModel *parser.ModelNode, currentField *parser.FieldNode, otherModel *parser.ModelNode) {
		relationAttribute := query.FieldGetAttribute(currentField, parser.AttributeRelation)
		if relationAttribute != nil {
			relationField, _ := relationAttribute.Arguments[0].Expression.ToValue()
			relation := &ManyToOneRelation{
				OneModel:  currentModel.Name.Value,
				OneField:  currentField.Name.Value,
				ManyModel: otherModel.Name.Value,
				ManyField: relationField.ToString(),
			}
			if relationRegistry[relation.ManyModel] == nil {
				relationRegistry[relation.ManyModel] = map[string]*ManyToOneRelation{}
			}
			relationRegistry[relation.ManyModel][relation.ManyField] = relation
		}
	}

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			if m.BuiltIn {
				return
			}
			currentModel = m

			registry[currentModel.Name.Value] = map[string]*InvalidManyToManyDetails{}
		},
		EnterField: func(currentField *parser.FieldNode) {
			if currentModel == nil {
				return
			}
			if details, ok := registry[currentModel.Name.Value][currentField.Name.Value]; ok {
				errs.AppendError(invalidManyToManyError(details, currentField.Node))
				return
			}
			otherModel := query.Model(asts, currentField.Type.Value)
			if otherModel == nil || currentModel.Name.Value == otherModel.Name.Value {
				return
			}
			if query.IsHasManyModelField(asts, currentField) {
				for _, otherField := range query.ModelFields(otherModel) {
					if !query.IsHasManyModelField(asts, otherField) {
						// check to see if there is a relation defined on this field and store it so we
						// can check it later
						checkForRelation(otherModel, otherField, currentModel)
						continue
					}

					if otherField.Type.Value == currentModel.Name.Value {
						// check to see if there is a relation stored for the other side of
						// this m:m that means the relation is not refering to the current field
						if relations, ok := relationRegistry[otherModel.Name.Value]; ok {
							if _, ok := relations[otherField.Name.Value]; ok {
								continue
							}
						}

						invalidManyToManyDetails := InvalidManyToManyDetails{
							InverseModel: otherModel,
							InverseField: otherField,
							ThisModel:    currentModel,
							ThisField:    currentField,
						}

						errs.AppendError(invalidManyToManyError(&invalidManyToManyDetails, currentField.Node))

						if registry[otherModel.Name.Value] == nil {
							registry[otherModel.Name.Value] = map[string]*InvalidManyToManyDetails{}
						}

						registry[otherModel.Name.Value][otherField.Name.Value] = &invalidManyToManyDetails
					}
				}
			} else {
				// check whether a relation exists for this field
				checkForRelation(currentModel, currentField, otherModel)
			}
		},
	}
}

func invalidManyToManyError(invalidDetails *InvalidManyToManyDetails, node node.Node) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.RelationshipError,
		errorhandling.ErrorDetails{
			Message: fmt.Sprintf("Cannot have a direct many to many between '%s' and '%s'", invalidDetails.ThisModel.Name.Value, invalidDetails.InverseModel.Name.Value),
			Hint:    "Visit https://docs.keel.so/models#many-to-many-relationships for information on how to create a many-to-many relationship",
		},
		node,
	)
}

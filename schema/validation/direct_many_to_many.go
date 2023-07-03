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

	checkForRelation := func(currentModel *parser.ModelNode, currentField *parser.FieldNode, otherModel *parser.ModelNode) {
		relation := query.FieldGetAttribute(currentField, parser.AttributeRelation)
		if relation != nil {
			relationValue, _ := relation.Arguments[0].Expression.ToValue()
			m21Relation := &ManyToOneRelation{
				OneModel:  currentModel.Name.Value,
				OneField:  currentField.Name.Value,
				ManyModel: otherModel.Name.Value,
				ManyField: relationValue.ToString(),
			}
			if m21Relation != nil {
				if relationRegistry[m21Relation.ManyModel] == nil {
					relationRegistry[m21Relation.ManyModel] = map[string]*ManyToOneRelation{}
				}
				relationRegistry[m21Relation.ManyModel][m21Relation.ManyField] = m21Relation
			}
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
			if otherModel == nil {
				return
			}
			if query.IsHasManyModelField(asts, currentField) {
				for _, otherField := range query.ModelFields(otherModel) {
					if !query.IsHasManyModelField(asts, otherField) {
						checkForRelation(otherModel, otherField, currentModel)
						continue
					}

					if otherField.Type.Value == currentModel.Name.Value {
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
			Hint:    "Visit https://keel.notaku.site/documentation/models for information on how to create a many-to-many relationship",
		},
		node,
	)
}

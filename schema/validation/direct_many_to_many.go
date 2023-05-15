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

// DirectManyToManyRule checks that a direct many to many relationship between
// two models hasn't been defined, and recommends creating a join model.
func DirectManyToManyRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	registry := map[string]map[string]*InvalidManyToManyDetails{}
	var currentModel *parser.ModelNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			if m.BuiltIn {
				return
			}
			currentModel = m

			registry[currentModel.Name.Value] = map[string]*InvalidManyToManyDetails{}
		},
		EnterField: func(f *parser.FieldNode) {
			if currentModel == nil {
				return
			}
			if details, ok := registry[currentModel.Name.Value][f.Name.Value]; ok {
				errs.AppendError(invalidManyToManyError(details, f.Node))

				return
			}
			if query.IsHasManyModelField(asts, f) {
				otherModel := query.Model(asts, f.Type.Value)

				for _, otherField := range query.ModelFields(otherModel) {
					if !query.IsHasManyModelField(asts, otherField) {
						continue
					}

					if otherField.Type.Value == currentModel.Name.Value {
						errs.AppendError(invalidManyToManyError(&InvalidManyToManyDetails{
							InverseModel: otherModel,
							InverseField: otherField,
							ThisModel:    currentModel,
							ThisField:    f,
						}, f.Node))

						if registry[otherModel.Name.Value] == nil {
							registry[otherModel.Name.Value] = map[string]*InvalidManyToManyDetails{}
						}

						registry[otherModel.Name.Value][otherField.Name.Value] = &InvalidManyToManyDetails{
							InverseModel: currentModel,
							InverseField: f,
							ThisModel:    otherModel,
							ThisField:    otherField,
						}
					}
				}
			}
		},
	}
}

func invalidManyToManyError(invalidDetails *InvalidManyToManyDetails, node node.Node) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.RelationshipError,
		errorhandling.ErrorDetails{
			Message: fmt.Sprintf("Cannot have a direct many to many between '%s' and '%s'", invalidDetails.ThisModel.Name.Value, invalidDetails.InverseModel.Name.Value),
			Hint:    fmt.Sprintf("Remove field '%s' from the '%s' model, and create a join model called '%s' that acts as a bridge between either side", invalidDetails.InverseField.Name.Value, invalidDetails.InverseModel.Name.Value, fmt.Sprintf("%s%s", invalidDetails.ThisModel.Name.Value, invalidDetails.InverseModel.Name.Value)),
		},
		node,
	)
}

package relationships

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func InvalidOneToOneRelationshipRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	processed := map[string]bool{}
	allModelNames := query.ModelNames(asts)

	for _, model := range query.Models(asts) {
		if ok := processed[model.Name.Value]; ok {
			continue
		}

		for _, field := range query.ModelFields(model) {
			if lo.Contains(allModelNames, field.Type) {
				otherModel := query.Model(asts, field.Type)

				otherModelFields := query.ModelFields(otherModel)

				for _, otherField := range otherModelFields {
					if otherField.Type != model.Name.Value {
						continue
					}

					// If either the field on model A is repeated
					// or the corresponding field on the other side is repeated
					// then we are not interested
					if otherField.Repeated || field.Repeated {
						continue
					}

					errs.Append(
						errorhandling.ErrorInvalidOneToOneRelationship,
						map[string]string{
							"ModelA": model.Name.Value,
							"ModelB": field.Type,
						},
						field,
					)

					processed[model.Name.Value] = true
					processed[otherModel.Name.Value] = true
				}
			}
		}
	}

	return
}

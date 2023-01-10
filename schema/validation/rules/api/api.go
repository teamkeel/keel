package api

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueAPINamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	seenAPINames := map[string]bool{}
	for _, api := range query.APIs(asts) {
		if _, ok := seenAPINames[api.Name.Value]; ok {
			errs.Append(errorhandling.ErrorUniqueAPIGlobally,

				map[string]string{
					"Name": api.Name.Value,
				},

				api.Name,
			)

			continue
		}

		seenAPINames[api.Name.Value] = true
	}

	return
}

func NamesCorrespondToModels(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	modelNames := query.ModelNames(asts)
	for _, api := range query.APIs(asts) {
		for _, section := range api.Sections {
			for _, model := range section.Models {
				if !lo.Contains(modelNames, model.Name.Value) {
					errs.Append(errorhandling.ErrorModelNotFound,
						map[string]string{
							"API":   api.Name.Value,
							"Model": model.Name.Value,
						},
						api.Name,
					)
				}
			}
		}
	}

	return
}

func ModelsToHaveQueryOperations(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, api := range query.APIs(asts) {
		for _, section := range api.Sections {
			for _, model := range section.Models {
				m := query.Model(asts, model.Name.Value)
				if m == nil || m.BuiltIn {
					continue
				}

				actions := query.ModelActions(m)
				hasQuery := lo.SomeBy(actions, func(action *parser.ActionNode) bool {
					return action.IsRead()
				})

				if !hasQuery {
					errs.Append(errorhandling.ErrorModelHasNoQueryActions,
						map[string]string{
							"API":   api.Name.Value,
							"Model": model.Name.Value,
						},
						api.Name)
				}
			}
		}
	}

	return errs
}

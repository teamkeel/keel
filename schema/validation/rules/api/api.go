package api

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueAPINamesRule(asts []*parser.AST) (errors []error) {
	seenAPINames := map[string]bool{}

	for _, api := range query.APIs(asts) {
		if _, ok := seenAPINames[api.Name.Value]; ok {
			errors = append(
				errors,
				errorhandling.NewValidationError(errorhandling.ErrorUniqueAPIGlobally,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Name": api.Name.Value,
						},
					},
					api.Name,
				),
			)

			continue
		}

		seenAPINames[api.Name.Value] = true
	}

	return errors
}

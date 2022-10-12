package api

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueAPINamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	seenAPINames := map[string]bool{}

	for _, api := range parser.APIs(asts) {
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

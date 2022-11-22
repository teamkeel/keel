package enum

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueEnumsRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	seenEnumNames := map[string]bool{}

	for _, enum := range query.Enums(asts) {
		if _, ok := seenEnumNames[enum.Name.Value]; ok {
			errs.Append(errorhandling.ErrorUniqueEnumGlobally,
				map[string]string{
					"Name": enum.Name.Value,
				},
				enum.Name,
			)

			continue
		}

		seenEnumNames[enum.Name.Value] = true
	}

	return
}

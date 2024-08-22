package model

import (
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// ModelNamesMaxLengthRule will validate that model names are smaller than the maximum allowed by postgres (63 bytes).
//
// The maximum field length is given by: 63 - 11 (to accommodate for the longest trigger name suffix: _updated_at) hard limit.
// This maximum length is applied to the snake cased version of the field name.
func ModelNamesMaxLengthRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	const (
		maxBytes  = 63
		maxSuffix = "_updated_at"
	)

	for _, model := range query.Models(asts) {
		identifier := casing.ToSnake(model.Name.Value) + maxSuffix

		if len(identifier) > maxBytes {
			errs.Append(errorhandling.ErrorModelNamesMaxLength,
				map[string]string{
					"Name": model.Name.Value,
				},
				model.Name,
			)
		}
	}

	return
}

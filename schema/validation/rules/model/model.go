package model

import (
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// NamesMaxLengthRule will validate that model and task names are smaller than the maximum allowed by postgres (63 bytes).
//
// The maximum field length is given by: 63 - 11 (to accommodate for the longest trigger name suffix: _updated_at) hard limit.
// This maximum length is applied to the snake cased version of the field name.
func NamesMaxLengthRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	const (
		maxBytes  = 63
		maxSuffix = "_updated_at"
	)

	for _, entity := range query.Entities(asts) {
		identifier := casing.ToSnake(entity.GetName()) + maxSuffix

		if len(identifier) > maxBytes {
			errs.Append(errorhandling.ErrorModelNamesMaxLength,
				map[string]string{
					"Name":      entity.GetName(),
					"DefinedOn": entity.EntityType(),
				},
				entity.Node().Name,
			)
		}
	}

	return
}

package role

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueRoleNamesRule(asts []*parser.AST) (errors []error) {
	seenRoleNames := map[string]bool{}

	for _, role := range query.Roles(asts) {
		if _, ok := seenRoleNames[role.Name.Value]; ok {
			errors = append(
				errors,
				errorhandling.NewValidationError(errorhandling.ErrorUniqueRoleGlobally,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Name": role.Name.Value,
						},
					},
					role.Name,
				),
			)

			continue
		}
		seenRoleNames[role.Name.Value] = true
	}

	return errors
}

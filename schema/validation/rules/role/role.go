package role

import (
	"github.com/teamkeel/keel/schema/foreignkeys"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueRoleNamesRule(asts []*parser.AST, fkInfo []*foreignkeys.ForeignKeyInfo) (errs errorhandling.ValidationErrors) {
	seenRoleNames := map[string]bool{}

	for _, role := range query.Roles(asts) {
		if _, ok := seenRoleNames[role.Name.Value]; ok {
			errs.Append(errorhandling.ErrorUniqueRoleGlobally,
				map[string]string{
					"Name": role.Name.Value,
				},

				role.Name,
			)

			continue
		}
		seenRoleNames[role.Name.Value] = true
	}

	return
}

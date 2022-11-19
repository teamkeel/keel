package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/foreignkeys"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	reservedModelNames = []string{"query"}
)

func ModelNamingRule(asts []*parser.AST, fkInfo []*foreignkeys.ForeignKeyInfo) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		// todo - these MustCompile regex would be better at module scope, to
		// make the MustCompile panic a load-time thing rather than a runtime thing.
		reg := regexp.MustCompile("([A-Z][a-z0-9]+)+")

		if reg.FindString(model.Name.Value) != model.Name.Value {
			suggested := strcase.ToCamel(strings.ToLower(model.Name.Value))

			errs.Append(errorhandling.ErrorUpperCamel,
				map[string]string{
					"Model":     model.Name.Value,
					"Suggested": suggested,
				},
				model.Name,
			)
		}

	}

	return
}

func ReservedModelNamesRule(asts []*parser.AST, fkInfo []*foreignkeys.ForeignKeyInfo) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, name := range reservedModelNames {
			if strings.EqualFold(name, model.Name.Value) {
				errs.Append(errorhandling.ErrorReservedModelName,
					map[string]string{
						"Name":       model.Name.Value,
						"Suggestion": fmt.Sprintf("%ser", model.Name.Value),
					},
					model.Name,
				)
			}
		}
	}

	return
}

func UniqueModelNamesRule(asts []*parser.AST, fkInfo []*foreignkeys.ForeignKeyInfo) (errs errorhandling.ValidationErrors) {
	seenModelNames := map[string]bool{}

	for _, model := range query.Models(asts) {
		if _, ok := seenModelNames[model.Name.Value]; ok {
			errs.Append(errorhandling.ErrorUniqueModelsGlobally,
				map[string]string{
					"Name": model.Name.Value,
				},
				model.Name,
			)
			continue
		}
		seenModelNames[model.Name.Value] = true
	}

	return
}

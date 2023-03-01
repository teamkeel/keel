package model

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	reservedModelNames = []string{"query"}
)

func ReservedModelNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
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

func UniqueModelNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
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

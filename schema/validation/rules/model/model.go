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

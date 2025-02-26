package validation

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func RouteFunctions(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterRoutes: func(routes *parser.RoutesNode) {
			for _, route := range routes.Routes {
				allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, "ALL"}
				if !lo.Contains(allowedMethods, strings.ToUpper(route.Method.Value)) {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not a valid route type. Valid types are get, post, put, delete and all", route.Method.Value),
						},
						route.Method,
					))
				}
			}
		},
	}
}

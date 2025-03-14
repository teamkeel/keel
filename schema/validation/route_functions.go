package validation

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
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

				pattern := strings.TrimPrefix(route.Pattern.Value, `"`)
				pattern = strings.TrimSuffix(pattern, `"`)

				if !strings.HasPrefix(pattern, `/`) {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: "a route pattern must start with \"/\"",
						},
						route.Pattern,
					))
				}

				invalid, ok := validateURLPath(pattern)
				if !ok {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("route pattern contains invalid characters: \"%s\"", invalid),
						},
						route.Pattern,
					))
				}

				u, _ := url.Parse(pattern)
				if u != nil && u.RawQuery != "" {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: "route pattern cannot contain query string",
						},
						route.Pattern,
					))
				}

				if casing.ToLowerCamel(route.Handler.Value) != route.Handler.Value {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.NamingError,
						errorhandling.ErrorDetails{
							Message: "a route handler must be named using lowerCamelCase",
						},
						route.Handler,
					))
				}
			}
		},
	}
}

// Regex to match any invalid character in a URL path
var invalidURLPathCharRegex = regexp.MustCompile(`[^A-Za-z0-9\-._~:/?#\[\]@!$&'()*+,;=]`)

func validateURLPath(path string) (string, bool) {
	loc := invalidURLPathCharRegex.FindStringIndex(path)
	if loc != nil {
		// Return the first invalid character
		return string(path[loc[0]]), false
	}
	return "", true
}

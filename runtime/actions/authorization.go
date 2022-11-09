package actions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func DefaultIsAuthorised(
	scope *Scope,
	args WhereArgs,
) (authorized bool, err error) {
	permissions := []*proto.PermissionRule{}

	// Add permissions defined at the model level
	model := proto.FindModel(scope.schema.Models, scope.operation.ModelName)
	modelPermissions := lo.Filter(model.Permissions, func(modelPermission *proto.PermissionRule, _ int) bool {
		return slices.Contains(modelPermission.OperationsTypes, scope.operation.Type)
	})
	permissions = append(permissions, modelPermissions...)

	// Add permissions defined at the operation level
	if scope.operation.Permissions != nil {
		permissions = append(permissions, scope.operation.Permissions...)
	}

	// No permissions declared means no permission can be granted.
	if len(permissions) == 0 {
		return false, nil
	}

	// We do a first pass - considering only Role-based permissions.
	// Because if one of these grants permission, we short-circuit the need to compose
	// and execute database queries to evaluate permission expressions.
	granted, err := roleBasedPermissionGranted(permissions, scope)
	if err != nil {
		return false, err
	}
	if granted {
		return true, nil
	}

	// Dropping through to Expression-based permissions logic...

	constraints := scope.query.Session(&gorm.Session{NewDB: true})

	for i, permission := range permissions {
		if permission.Expression != nil {
			expression, err := parser.ParseExpression(permission.Expression.Source)
			if err != nil {
				return false, err
			}

			// New expression resolver to generate a database query statement
			resolver := NewExpressionResolver(scope)

			// First check to see if we can resolve the condition "in proc"
			if resolver.CanResolveInMemory(expression) {
				if resolver.ResolveInMemory(expression, args, scope.writeValues) {
					return true, nil
				} else if i == len(permissions)-1 {
					return false, nil
				}
			} else {
				// Resolve the database statement for this expression
				statement, err := resolver.ResolveQueryStatement(expression, args, scope.writeValues)
				if err != nil {
					return false, err
				}

				// Logical OR between each of the permission expressions
				constraints = constraints.Or(statement)
			}
		}
	}

	// Logical AND between the implicit/explicit filters and all the permission conditions
	permissionQuery := scope.query.
		Session(&gorm.Session{}).
		Where(constraints)

	// Determine the number of rows in the current query which don't satisfy the permission conditions
	results := map[string]any{}
	scope.query.Session(&gorm.Session{NewDB: true}).Raw("SELECT COUNT(id) as unauthorised FROM (? EXCEPT ?) as unauthorisedrows",
		scope.query,
		permissionQuery,
	).Scan(&results)
	unauthorisedRows, ok := results["unauthorised"].(int64)

	if !ok {
		return false, fmt.Errorf("failed to query or parse unauthorised rows from database")
	}

	return unauthorisedRows == 0, nil
}

// roleBasedPermissionGranted returns true if there is a role-based permission among the
// given list of permissions that passes.
func roleBasedPermissionGranted(permissions []*proto.PermissionRule, scope *Scope) (granted bool, err error) {

	// If the context does not have an authenticated user, then we cannot grant
	// any role-based permissions.
	if !runtimectx.IsAuthenticated(scope.context) {
		return false, nil
	}

	currentUserEmail, currentUserDomain, err := getEmailAndDomain(scope)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		for _, roleName := range perm.RoleNames {
			role := proto.FindRole(roleName, scope.schema)
			for _, email := range role.Emails {
				if email == currentUserEmail {
					return true, nil
				}
			}

			for _, domain := range role.Domains {
				if domain == currentUserDomain {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// getEmailAndDomain requires that the the given scope's context
// contains an authenticated user, extracts the id of that
// authenticated user, and does a database query to fetch their
// email and domain.
func getEmailAndDomain(scope *Scope) (string, string, error) {

	// Use the authenticated user's id to lookup their email address.
	userKSUID, err := runtimectx.GetIdentity(scope.context)
	if err != nil {
		return "", "", err
	}

	db, err := runtimectx.GetDatabase(scope.context)
	if err != nil {
		return "", "", err
	}
	rows := []map[string]any{}
	tableName := strcase.ToSnake(parser.ImplicitIdentityModelName)
	response := db.Table(tableName).Where("id = ?", userKSUID.String()).Find(&rows)
	if response.Error != nil {
		return "", "", err
	}
	if response.RowsAffected != 1 {
		return "", "", ErrNotOneRow
	}
	row := rows[0]
	email := row["email"].(string)
	segments := strings.Split(email, "@")
	domain := segments[1]
	return email, domain, nil
}

var (
	ErrNotOneRow = errors.New("should be one row")
)

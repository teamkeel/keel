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

	// Combine all the permissions defined at model level, with those defined at
	// operation level.
	model := proto.FindModel(scope.schema.Models, scope.operation.ModelName)
	modelPermissions := lo.Filter(model.Permissions, func(modelPermission *proto.PermissionRule, _ int) bool {
		return slices.Contains(modelPermission.OperationsTypes, scope.operation.Type)
	})
	permissions = append(permissions, modelPermissions...)

	if scope.operation.Permissions != nil {
		permissions = append(permissions, scope.operation.Permissions...)
	}

	// No permissions declared means no permission can be granted.
	if len(permissions) == 0 {
		return false, nil
	}

	// Does one of the role-based rules grant permission?
	//
	// This is good to check first, because it avoids the composition
	// and execution of a complex SQL query.
	if runtimectx.IsAuthenticated(scope.context) {
		roleBasedPerms := proto.PermissionsWithRole(permissions)
		granted, err := roleBasedPermissionGranted(roleBasedPerms, scope)
		if err != nil {
			return false, err
		}
		if granted {
			return true, nil
		}
	}

	// Dropping through to Expression-based permissions logic...

	exprBasedPerms := proto.PermissionsWithExpression(permissions)

	// If there are zero Expression-based permissions provided - we know we cannot grant permission.
	// This isn't just an optimisation - the code below spuriously grants access in this case.
	if len(exprBasedPerms) == 0 {
		return false, nil
	}

	constraints := scope.query.Session(&gorm.Session{NewDB: true})
	for i, permission := range exprBasedPerms {

		expression, err := parser.ParseExpression(permission.Expression.Source)
		if err != nil {
			return false, err
		}

		// New expression resolver to generate a database query statement
		resolver := NewExpressionResolver(scope) // todo: would it be better to have this outside the loop?

		// First check to see if we can resolve the condition "in proc"
		if resolver.CanResolveInMemory(expression) {
			if resolver.ResolveInMemory(expression, args, scope.writeValues) {
				return true, nil
			} else if i == len(exprBasedPerms)-1 {
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
func roleBasedPermissionGranted(roleBasedPermissions []*proto.PermissionRule, scope *Scope) (granted bool, err error) {

	// todo: nicer if this came in the Scope or Token?
	// Because it costs a database query.
	currentUserEmail, currentUserDomain, err := getEmailAndDomain(scope)
	if err != nil {
		return false, err
	}

	for _, perm := range roleBasedPermissions {
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
//
// todo: it would be nicer if the current user's email name was
// available directly in the scope object?
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

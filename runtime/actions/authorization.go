package actions

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/exp/slices"
)

func (query *QueryBuilder) isAuthorised(scope *Scope, args map[string]any) (authorized bool, err error) {
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
		granted, err := roleBasedPermissionGranted(scope, roleBasedPerms)
		if err != nil {
			return false, err
		}
		if granted {
			return true, nil
		}
	}

	// Create a copy of the current action query
	permissionQuery := query.Copy()

	// Dropping through to Expression-based permissions logic...
	exprBasedPerms := proto.PermissionsWithExpression(permissions)

	// No expression permissions left means no permission can be granted.
	if len(exprBasedPerms) == 0 {
		return false, nil
	}

	for i, permission := range exprBasedPerms {
		expression, err := parser.ParseExpression(permission.Expression.Source)
		if err != nil {
			return false, err
		}

		// First check to see if we can resolve the condition "in proc"
		if expressions.CanResolveInMemory(scope.context, scope.schema, scope.operation, expression) {
			if expressions.ResolveInMemory(scope.context, scope.schema, scope.operation, expression, args) {
				return true, nil
			} else if i == len(exprBasedPerms)-1 {
				return false, nil
			}
		} else {
			// Resolve the database statement for this expression
			err = permissionQuery.whereByExpression(scope, expression, args)
			if err != nil {
				return false, err
			}
			// Or with the next
			permissionQuery.Or()
		}
	}

	// Determine the number of rows in the current query which don't satisfy the permission conditions
	stmt := &Statement{
		template: fmt.Sprintf("SELECT COUNT(id) as unauthorised FROM (%v EXCEPT %v) as unauthorisedrows",
			query.SelectStatement().template,
			permissionQuery.SelectStatement().template),
		args: append(query.args, permissionQuery.args...)}

	results, _, _, err := stmt.ExecuteToMany(scope.context)
	if err != nil {
		return false, err
	}

	unauthorisedRows, ok := results[0]["unauthorised"].(int64)

	if !ok {
		return false, fmt.Errorf("failed to query or parse unauthorised rows from database")
	}

	return unauthorisedRows == 0, nil
}

// roleBasedPermissionGranted returns true if there is a role-based permission among the
// given list of permissions that passes.
func roleBasedPermissionGranted(scope *Scope, roleBasedPermissions []*proto.PermissionRule) (granted bool, err error) {
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

	// Use the authenticated identity's id to lookup their email address.
	identityId, err := runtimectx.GetIdentity(scope.context)
	if err != nil {
		return "", "", err
	}

	identity, err := FindIdentityById(scope.context, identityId)
	if err != nil {
		return "", "", err
	}

	if identity == nil {
		return "", "", ErrIdentityNotFound
	}

	segments := strings.Split(identity.Email, "@")
	domain := segments[1]
	return identity.Email, domain, nil
}

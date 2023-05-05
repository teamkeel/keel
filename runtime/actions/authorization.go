package actions

import (
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func AuthoriseSingle(scope *Scope, args map[string]any, rowToAuthorise map[string]any) (authorized bool, err error) {
	return Authorise(scope, args, []map[string]any{rowToAuthorise})
}

func Authorise(scope *Scope, args map[string]any, rowsToAuthorise []map[string]any) (authorized bool, err error) {
	ctx, span := tracer.Start(scope.context, "Check Permissions")
	defer span.End()

	scope = scope.WithContext(ctx)
	permissions := proto.PermissionsForAction(scope.schema, scope.operation)

	// No permissions declared means no permission can be granted.
	if len(permissions) == 0 {
		span.SetAttributes(attribute.Bool("result", false))
		span.SetAttributes(attribute.String("reason", "no permission rules"))
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
			span.SetAttributes(attribute.Bool("result", true))
			span.SetAttributes(attribute.String("reason", "role"))
			return true, nil
		}
	}

	// Create a copy of the current action query
	//permissionQuery := query.Copy()

	// Dropping through to Expression-based permissions logic...
	exprBasedPerms := proto.PermissionsWithExpression(permissions)

	// No expression permissions left means no permission can be granted.
	if len(exprBasedPerms) == 0 {
		span.SetAttributes(attribute.Bool("result", false))
		span.SetAttributes(attribute.String("reason", "no matching permission rules"))
		return false, nil
	}

	query := NewQuery(scope.model)
	query.OpenParenthesis()
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
			err = query.whereByExpression(scope, expression, args)
			if err != nil {
				return false, err
			}
			// Or with the next permission attribute
			query.Or()
		}
	}
	query.CloseParenthesis()

	ids := lo.Map(rowsToAuthorise, func(row map[string]interface{}, _ int) any {
		return row["id"]
	})

	query.And()
	err = query.Where(IdField(), OneOf, Value(ids))

	stmt := query.SelectStatement()

	results, _, err := stmt.ExecuteToMany(scope.context, nil)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	authorised := len(results) == len(ids)

	// if !authorised {
	// 	err := fmt.Errorf("failed to query or parse unauthorised rows from database")
	// 	span.RecordError(err, trace.WithStackTrace(true))
	// 	span.SetStatus(codes.Error, err.Error())
	// 	span.SetAttributes(attribute.Bool("result", false))
	// 	return false, err
	// }

	//span.SetAttributes(attribute.Bool("result", unauthorisedRows == 0))
	return authorised, nil
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
	identity, err := runtimectx.GetIdentity(scope.context)
	if err != nil {
		return "", "", err
	}

	if identity == nil {
		return "", "", ErrIdentityNotFound
	}

	if identity.Email == "" {
		return "", "", nil
	}

	segments := strings.Split(identity.Email, "@")
	domain := segments[1]
	return identity.Email, domain, nil
}

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

func AuthoriseSingle(scope *Scope, rowToAuthorise map[string]any) (authorized bool, err error) {
	return Authorise(scope, []map[string]any{rowToAuthorise})
}

func Authorise(scope *Scope, rowsToAuthorise []map[string]any) (authorized bool, err error) {
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

	// Do one of the role-based rules grant permission?
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

	span.SetAttributes(attribute.String("reason", "permission rules"))

	// If there are no expression permissions to satisfy, then access cannot be granted.
	if len(proto.PermissionsWithExpression(permissions)) == 0 {
		span.SetAttributes(attribute.Bool("result", false))
		return false, nil
	}

	// Test if any expressions can be resolved and satisfied without a database operation.
	canResolve, authorised, err := tryResolveInMemory(scope)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}
	if canResolve {
		span.SetAttributes(attribute.Bool("result", authorised))
		return authorised, nil
	}

	// Generate SQL for the permission expressions.
	stmt, err := GeneratePermissionStatement(scope, rowsToAuthorise)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	// Execute permission query against the database.
	results, _, err := stmt.ExecuteToMany(scope.context, nil)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	// TODO: compare ids matching in both slices
	authorised = len(results) == len(rowsToAuthorise)

	if !authorised {
		span.SetAttributes(attribute.Bool("result", false))
		return false, err
	}

	span.SetAttributes(attribute.Bool("result", authorised))
	return authorised, nil
}

func tryResolveInMemory(scope *Scope) (canResolve bool, authorised bool, err error) {
	permissions := proto.PermissionsForAction(scope.schema, scope.operation)
	exprBasedPerms := proto.PermissionsWithExpression(permissions)

	for i, permission := range exprBasedPerms {
		expression, err := parser.ParseExpression(permission.Expression.Source)
		if err != nil {
			return false, false, err
		}

		// First check to see if we can resolve the condition "in proc"
		if expressions.CanResolveInMemory(scope.context, scope.schema, scope.operation, expression) {
			if expressions.ResolveInMemory(scope.context, scope.schema, scope.operation, expression, map[string]any{}) {
				return true, true, nil
			} else if i == len(exprBasedPerms)-1 {
				return true, false, nil
			}
		}
	}

	return false, false, nil
}

func GeneratePermissionStatement(scope *Scope, rowsToAuthorise []map[string]any) (*Statement, error) {
	permissions := proto.PermissionsForAction(scope.schema, scope.operation)
	exprBasedPerms := proto.PermissionsWithExpression(permissions)
	query := NewQuery(scope.model)

	// Append SQL where conditions for each permission attribute.
	query.OpenParenthesis()
	for _, permission := range exprBasedPerms {
		expression, err := parser.ParseExpression(permission.Expression.Source)
		if err != nil {
			return nil, err
		}

		err = query.whereByExpression(scope, expression, map[string]any{})
		if err != nil {
			return nil, err
		}
		// Or with the next permission attribute
		query.Or()
	}
	query.CloseParenthesis()

	ids := lo.Map(rowsToAuthorise, func(row map[string]interface{}, _ int) any {
		return row["id"]
	})

	// Filter by the IDs of the rows we want to authorise.
	query.And()
	err := query.Where(IdField(), OneOf, Value(ids))
	if err != nil {
		return nil, err
	}

	// Select distinct IDs.
	query.AppendSelect(IdField())
	query.AppendDistinctOn(IdField())

	return query.SelectStatement(), nil
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
// contains an authenticated user
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

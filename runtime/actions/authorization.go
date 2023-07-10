package actions

import (
	"context"
	"errors"
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

// AuthoriseAction checks authorisation for rows using the permission and role rules applicable for an action,
// which could be defined at model- and action- levels.
func AuthoriseAction(scope *Scope, rowsToAuthorise []map[string]any) (authorised bool, err error) {
	if scope.Operation == nil {
		return false, errors.New("cannot authorise with AuthoriseAction if no operation is provided in scope")
	}

	permissions := proto.PermissionsForAction(scope.Schema, scope.Operation)
	return Authorise(scope, permissions, rowsToAuthorise)
}

// AuthoriseForActionType checks authorisation for rows using permission and role rules defined for some operation type,
// i.e. agnostic to any action.
func AuthoriseForActionType(scope *Scope, opType proto.OperationType, rowsToAuthorise []map[string]any) (authorised bool, err error) {
	permissions := proto.PermissionsForOperationType(scope.Schema, scope.Model.Name, opType)
	return Authorise(scope, permissions, rowsToAuthorise)
}

// Authorise checks authorisation for rows using the slice of permission rules provided.
func Authorise(scope *Scope, permissions []*proto.PermissionRule, rowsToAuthorise []map[string]any) (authorized bool, err error) {
	ctx, span := tracer.Start(scope.Context, "Check permissions")
	defer span.End()

	scope = scope.WithContext(ctx)

	// No permissions declared means no permission can be granted.
	if len(permissions) == 0 {
		span.SetAttributes(attribute.Bool("result", false))
		span.SetAttributes(attribute.String("reason", "no permission rules"))
		return false, nil
	}

	// Do one of the role-based rules grant permission?
	if runtimectx.IsAuthenticated(scope.Context) {
		roleBasedPerms := proto.PermissionsWithRole(permissions)
		granted, err := RoleBasedPermissionGranted(scope.Context, scope.Schema, roleBasedPerms)
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
	canResolve, authorised, err := tryResolveInMemory(scope, permissions)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}
	if canResolve && authorised {
		span.SetAttributes(attribute.Bool("result", authorised))
		return authorised, nil
	}

	// Generate SQL for the permission expressions.
	stmt, err := GeneratePermissionStatement(scope, permissions, rowsToAuthorise)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	// Execute permission query against the database.
	results, _, err := stmt.ExecuteToMany(scope.Context, nil)
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

func tryResolveInMemory(scope *Scope, permissions []*proto.PermissionRule) (canResolve bool, authorised bool, err error) {
	permissions = proto.PermissionsWithExpression(permissions)

	for i, permission := range permissions {
		expression, err := parser.ParseExpression(permission.Expression.Source)
		if err != nil {
			return false, false, err
		}

		// Check to see if we can resolve the condition "in proc"
		if expressions.CanResolveInMemory(scope.Context, scope.Schema, scope.Model, scope.Operation, expression) {
			if expressions.ResolveInMemory(scope.Context, scope.Schema, scope.Model, scope.Operation, expression, map[string]any{}) {
				return true, true, nil
			} else if i == len(permissions)-1 {
				return true, false, nil
			}
		}
	}

	return false, false, nil
}

func GeneratePermissionStatement(scope *Scope, permissions []*proto.PermissionRule, rowsToAuthorise []map[string]any) (*Statement, error) {
	permissions = proto.PermissionsWithExpression(permissions)
	query := NewQuery(scope.Model)

	// Append SQL where conditions for each permission attribute.
	query.OpenParenthesis()
	for _, permission := range permissions {
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

// RoleBasedPermissionGranted returns true if there is a role-based permission among the
// given list of permissions that passes.
func RoleBasedPermissionGranted(ctx context.Context, schema *proto.Schema, roleBasedPermissions []*proto.PermissionRule) (granted bool, err error) {
	// todo: nicer if this came in the Scope or Token?
	// Because it costs a database query.
	currentUserEmail, currentUserDomain, err := getEmailAndDomain(ctx)
	if err != nil {
		return false, err
	}

	for _, perm := range roleBasedPermissions {
		for _, roleName := range perm.RoleNames {
			role := proto.FindRole(roleName, schema)
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
func getEmailAndDomain(ctx context.Context) (string, string, error) {
	// Use the authenticated identity's id to lookup their email address.
	identity, err := runtimectx.GetIdentity(ctx)
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

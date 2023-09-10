package actions

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// AuthoriseAction checks authorisation for rows using the permission and role rules applicable for an action,
// which could be defined at model- and action- levels.
func AuthoriseAction(scope *Scope, input map[string]any, rowsToAuthorise []map[string]any) (authorised bool, err error) {
	if scope.Action == nil {
		return false, errors.New("cannot authorise with AuthoriseAction if no operation is provided in scope")
	}

	if scope.Action.Type == proto.ActionType_ACTION_TYPE_UPDATE || scope.Action.Type == proto.ActionType_ACTION_TYPE_LIST {
		var ok bool
		input, ok = input["where"].(map[string]any)
		if !ok {
			input = map[string]any{}
		}
	}

	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)
	return authorise(scope, permissions, input, rowsToAuthorise)
}

// AuthoriseForActionType checks authorisation for rows using permission and role rules defined for some action type,
// i.e. agnostic to any action.
func AuthoriseForActionType(scope *Scope, opType proto.ActionType, rowsToAuthorise []map[string]any) (authorised bool, err error) {
	permissions := proto.PermissionsForActionType(scope.Schema, scope.Model.Name, opType)
	return authorise(scope, permissions, map[string]any{}, rowsToAuthorise)
}

// authorise checks authorisation for rows using the slice of permission rules provided.
func authorise(scope *Scope, permissions []*proto.PermissionRule, input map[string]any, queryResults []map[string]any) (authorised bool, err error) {
	ctx, span := tracer.Start(scope.Context, "Check permissions")
	defer span.End()

	scope = scope.WithContext(ctx)

	// No permissions declared means no permission can be granted.
	if len(permissions) == 0 {
		span.SetAttributes(attribute.Bool("result", false))
		span.SetAttributes(attribute.String("reason", "no permission rules"))
		return false, nil
	}

	canResolve, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if canResolve {
		return authorised, nil
	}

	span.SetAttributes(attribute.String("reason", "permission rules"))

	// If there are no expression permissions to satisfy, then access cannot be granted.
	if len(proto.PermissionsWithExpression(permissions)) == 0 {
		span.SetAttributes(attribute.Bool("result", false))
		return false, nil
	}

	// Generate SQL for the permission expressions.
	stmt, err := GeneratePermissionStatement(scope, permissions, input)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	// Execute permission query against the database.
	permissionResults, _, err := stmt.ExecuteToMany(scope.Context, nil)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	permissionResultIds := lo.Map(permissionResults, func(row map[string]interface{}, _ int) string {
		return row["id"].(string)
	})

	queryResultIds := lo.Map(queryResults, func(row map[string]interface{}, _ int) string {
		return row["id"].(string)
	})

	authorised = compare(permissionResultIds, queryResultIds)

	if !authorised {
		span.SetAttributes(attribute.Bool("result", false))
		return false, err
	}

	span.SetAttributes(attribute.Bool("result", authorised))
	return authorised, nil
}

// TryResolveAuthorisationEarly will attempt to check authorisation early without row-based querying.
// This will take into account logical conditions and multiple expression and role permission attributes.
func TryResolveAuthorisationEarly(scope *Scope, permissions []*proto.PermissionRule) (canResolveAll bool, authorised bool, err error) {
	hasDatabaseCheck := false
	canResolveAll = false
	for _, permission := range permissions {
		canResolve := false
		authorised := false
		switch {
		case permission.Expression != nil:
			expression, err := parser.ParseExpression(permission.Expression.Source)
			if err != nil {
				return false, false, err
			}

			// Try resolve the permission early.
			canResolve, authorised = expressions.TryResolveExpressionEarly(scope.Context, scope.Schema, scope.Model, scope.Action, expression, map[string]any{})

			if !canResolve {
				hasDatabaseCheck = true
			}

		case permission.RoleNames != nil:
			// Roles can always be resolved early.
			canResolve = true

			// Check if this role permission is satisfied.
			authorised, err = resolveRolePermissionRule(scope.Context, scope.Schema, permission)
			if err != nil {
				return false, false, err
			}
		}

		// If this permission can be resolved now and is satisfied,
		// then we know the permission will be granted because
		// permission attributes are ORed.
		if canResolve && authorised {
			return true, true, nil
		}

		// If this permission can be resolved now and
		// there hasn't been a row/db permission, then
		// assume we can still resolve the entire action.
		canResolveAll = canResolve && !hasDatabaseCheck
	}

	return canResolveAll, false, nil
}

// resolveRolePermissionRule returns true if there is a role-based permission among the
// given list of permissions that passes.
func resolveRolePermissionRule(ctx context.Context, schema *proto.Schema, permission *proto.PermissionRule) (bool, error) {
	// If there is no authenticated user, then no role permissions can be satisfied.
	if !auth.IsAuthenticated(ctx) {
		return false, nil
	}

	identityEmail, identityDomain, err := getEmailAndDomain(ctx)
	if err != nil {
		return false, err
	}

	authorised := false
	for _, roleName := range permission.RoleNames {
		role := proto.FindRole(roleName, schema)
		for _, email := range role.Emails {
			if email == identityEmail {
				authorised = true
			}
		}

		for _, domain := range role.Domains {
			if domain == identityDomain {
				authorised = true
			}
		}
	}

	return authorised, nil
}

func GeneratePermissionStatement(scope *Scope, permissions []*proto.PermissionRule, input map[string]any) (*Statement, error) {
	permissions = proto.PermissionsWithExpression(permissions)
	query := NewQuery(scope.Context, scope.Model, WithJoinType(JoinTypeLeft))

	// Implicit and explicit filters need to be included in the permissions query,
	// otherwise we'll be testing against records which aren't part of the the result set
	if scope.Action.Type == proto.ActionType_ACTION_TYPE_LIST {
		err := query.applyImplicitFiltersForList(scope, input)
		if err != nil {
			return nil, err
		}
		query.And()
	} else {
		err := query.applyImplicitFilters(scope, input)
		if err != nil {
			return nil, err
		}
		query.And()
	}

	err := query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}
	query.And()

	if len(permissions) > 0 {
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
	}

	// Select distinct IDs.
	query.AppendSelect(IdField())
	query.AppendDistinctOn(IdField())

	return query.SelectStatement(), nil
}

// getEmailAndDomain requires that the the given scope's context
// contains an authenticated user
func getEmailAndDomain(ctx context.Context) (string, string, error) {
	// Use the authenticated identity's id to lookup their email address.
	identity, err := auth.GetIdentity(ctx)
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

func compare(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

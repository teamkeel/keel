package flows

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/schema/parser"
)

// AuthoriseFlow will check that the context's identity is authorised to access the given flow in the context of the given schema.
func AuthoriseFlow(ctx context.Context, schema *proto.Schema, flow *proto.Flow) (bool, error) {
	// if the flow doesn't have any permission rules, do not authorise
	if len(flow.GetPermissions()) == 0 {
		return false, nil
	}

	for _, permission := range flow.GetPermissions() {
		switch {
		case permission.GetExpression() != nil:
			expression, err := parser.ParseExpression(permission.GetExpression().GetSource())
			if err != nil {
				return false, err
			}

			// Try resolve the permission early.
			canAuthorise, authorised := actions.TryResolveExpressionEarly(ctx, schema, nil, nil, permission.GetExpression().GetSource(), nil)

			// If access can be concluded by role permissions alone
			if canAuthorise {
				return authorised, nil
			}

			query := actions.NewQuery(schema.FindModel(parser.IdentityModelName))
			query.SelectClause("COUNT(*) as authorised")

			_, err = resolve.RunCelVisitor(expression, actions.GenerateCtxQuery(ctx, query, schema))
			if err != nil {
				return false, err
			}

			stmt := query.SelectStatement()

			// Execute permission query against the database.
			permissionResults, err := stmt.ExecuteToSingle(ctx)
			if err != nil {
				return false, err
			}

			if len(permissionResults) != 1 {
				return false, errors.New("could not parse permission result as there are multiple rows")
			}

			authorisedValue, ok := permissionResults["authorised"].(int64)
			if !ok {
				return false, errors.New("could not parse authorised result")
			}

			if authorisedValue == 0 {
				return false, nil
			}

			return true, nil

		case permission.RoleNames != nil:
			if permission.RoleNames != nil {
				authorised, err := actions.ResolveRolePermissionRule(ctx, schema, permission)
				if err != nil {
					return false, err
				}

				// rules are OR'ed so if one resolves to true then the user is authorised.
				if authorised {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// AuthorisedFlows returns a list of flows from the given schema for which the current context user is authorised to view.
func AuthorisedFlows(ctx context.Context, schema *proto.Schema) ([]*proto.Flow, error) {
	flows := []*proto.Flow{}
	for _, f := range schema.GetAllFlows() {
		authorised, err := AuthoriseFlow(ctx, schema, f)
		if err != nil {
			return nil, err
		}
		// not authorised to view this flow, continue
		if authorised {
			flows = append(flows, f)
		}
	}

	return flows, nil
}

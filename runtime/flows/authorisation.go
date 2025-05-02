package flows

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// AuthoriseFlow will check that the context's identity is authorised to access the given flow in the context of the given schema
func AuthoriseFlow(ctx context.Context, schema *proto.Schema, flow *proto.Flow) (bool, error) {
	// if the flow doesn't have any permission rules, do not authorise
	if len(flow.Permissions) == 0 {
		return false, nil
	}

	for _, permission := range flow.Permissions {
		if permission.RoleNames != nil {
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

// AuthorisedFlows returns a list of flows from the given schema for which the current context user is authorised to view
func AuthorisedFlows(ctx context.Context, schema *proto.Schema) ([]*proto.Flow, error) {
	flows := []*proto.Flow{}
	for _, f := range schema.Flows {
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

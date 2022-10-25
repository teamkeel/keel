package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func DefaultIsAuthorised(
	scope *Scope,
	args map[string]any,
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

	// todo: remove this once we make permissions a requirement for any access
	// https://linear.app/keel/issue/RUN-135/permissions-required-for-access-at-all
	if len(permissions) == 0 {
		return true, nil
	}

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

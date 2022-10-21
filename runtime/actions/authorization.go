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

	// Generation SQL condition for each permission expression
	sqlConditions := []string{}
	for _, permission := range permissions {
		if permission.Expression != nil {
			expression, err := parser.ParseExpression(permission.Expression.Source)
			if err != nil {
				return false, err
			}

			condition, err := expressionToSqlCondition(scope.context, expression, scope.operation, scope.schema, args)
			if err != nil {
				return false, err
			}

			sqlConditions = append(sqlConditions, condition)
		}
	}

	// Logical OR between all the permission expressions
	conditions := scope.permissionQuery.Session(&gorm.Session{NewDB: true}) // todo: remove NewDB?
	for _, sqlSegment := range sqlConditions {
		conditions = conditions.Or(sqlSegment)
	}

	// Logical AND between the filters and the permission conditions
	scope.permissionQuery = scope.permissionQuery.Where(conditions)

	// Determine the number of rows which don't satisfy the permission conditions
	results := map[string]any{}
	scope.query.Session(&gorm.Session{NewDB: true}).Raw("SELECT COUNT(*) as unauthorised FROM (? EXCEPT ?) as unauthorisedrows",
		scope.query,
		scope.permissionQuery,
	).Scan(&results)
	unauthorisedRows, ok := results["unauthorised"].(int64)

	if !ok {
		return false, fmt.Errorf("failed to query or parse unauthorised rows from database")
	}

	return unauthorisedRows == 0, nil
}

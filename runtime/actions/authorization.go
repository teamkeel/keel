package actions

import (
	"fmt"

	"github.com/iancoleman/strcase"
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

	conditions := scope.permissionQuery.Session(&gorm.Session{NewDB: true}) // todo: remove NewDB?

	for _, permission := range permissions {
		if permission.Expression != nil {
			expression, err := parser.ParseExpression(permission.Expression.Source)
			if err != nil {
				return false, err
			}

			if len(expression.Conditions()) != 1 {
				//return "", nil // fmt.Errorf("cannot yet handle multiple conditions, have: %d", len(conditions))
			}
			condition := expression.Conditions()[0]

			operatorStr := condition.Operator.ToString()
			operator, _ := expressionOperatorToActionOperator(operatorStr)

			queryTemplate, queryArguments := conditionToSqlStatement(scope, condition, operator, args)
			if err != nil {
				return false, err
			}

			// Logical OR between all the permission expressions
			conditions = conditions.Or(queryTemplate, queryArguments...)
		}
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

func toSqlStatement(scope *Scope, operator ActionOperator, data map[string]any)

func conditionToSqlStatement(scope *Scope, condition *parser.Condition, operator ActionOperator, data map[string]any) (queryTemplate string, queryArguments []any) {

	if condition.Type() != parser.ValueCondition && condition.Type() != parser.LogicalCondition {
		return "", nil // fmt.Errorf("can only handle condition type of LogicalCondition, have: %s", condition.Type())
	}

	var lhsOperandType, rhsOperandType proto.Type

	lhsOperand := condition.LHS
	rhsOperand := condition.RHS

	lhsResolver := NewOperandResolver(scope.context, lhsOperand, scope.operation, scope.schema)
	rhsResolver := NewOperandResolver(scope.context, rhsOperand, scope.operation, scope.schema)
	lhsOperandType, _ = lhsResolver.GetOperandType()

	if condition.Type() == parser.ValueCondition {
		if lhsOperandType != proto.Type_TYPE_BOOL {
			//todo: err - must be a bool
		}

		// A value condition only has one operand in the expression,
		// so we must set the operator and RHS value (= true) ourselves.
		rhsOperandType = lhsOperandType
		operator = Equals
	} else {
		rhsOperandType, _ = rhsResolver.GetOperandType()
	}

	var template string
	var queryArgs []any

	lhsSqlOperand := "?"
	rhsSqlOperand := "?"

	if !lhsResolver.IsModelField() {
		lhsValue, _ := lhsResolver.ResolveValue(data, lhsOperandType)
		// todo: date and time parsing
		queryArgs = append(queryArgs, lhsValue)
	} else {
		modelTarget := strcase.ToSnake(lhsResolver.operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(lhsResolver.operand.Ident.Fragments[1].Fragment)
		lhsSqlOperand = fmt.Sprintf("%s.%s", modelTarget, fieldName)
	}

	if condition.Type() == parser.ValueCondition {
		queryArgs = append(queryArgs, true)
	} else if !rhsResolver.IsModelField() {
		rhsValue, _ := rhsResolver.ResolveValue(data, rhsOperandType)
		// todo: date and time parsing
		queryArgs = append(queryArgs, rhsValue)
	} else {
		modelTarget := strcase.ToSnake(rhsResolver.operand.Ident.Fragments[0].Fragment)
		fieldName := strcase.ToSnake(rhsResolver.operand.Ident.Fragments[1].Fragment)
		rhsSqlOperand = fmt.Sprintf("%s.%s", modelTarget, fieldName)
	}

	template = generateFilterTemplate(lhsSqlOperand, rhsSqlOperand, operator, rhsOperandType)

	return template, queryArgs
}

func generateFilterTemplate(lhsSqlOperand any, rhsSqlOperand any, operator ActionOperator, rhsOperandType proto.Type) string {
	var template string

	switch operator {
	case Equals:
		if rhsOperandType == proto.Type_TYPE_UNKNOWN {
			template = fmt.Sprintf("%s IS %s", lhsSqlOperand, rhsSqlOperand)
		} else {
			template = fmt.Sprintf("%s = %s", lhsSqlOperand, rhsSqlOperand)
		}
	case NotEquals:
		if rhsOperandType == proto.Type_TYPE_UNKNOWN {
			template = fmt.Sprintf("%s IS NOT %s", lhsSqlOperand, rhsSqlOperand)
		} else {
			template = fmt.Sprintf("%s != %s", lhsSqlOperand, rhsSqlOperand)
		}
	case StartsWith:
		template = fmt.Sprintf("%s LIKE %s%s", lhsSqlOperand, "%%", rhsSqlOperand)
	case EndsWith:
		template = fmt.Sprintf("%s LIKE %s%s", lhsSqlOperand, rhsSqlOperand, "%%")
	case Contains:
		template = fmt.Sprintf("%s LIKE %s%s%s", lhsSqlOperand, "%%", rhsSqlOperand, "%%")
	case OneOf:
		template = fmt.Sprintf("%s in %s", lhsSqlOperand, rhsSqlOperand)
	case LessThan:
		template = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case LessThanEquals:
		template = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThan:
		template = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThanEquals:
		template = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
	case Before:
		template = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case After:
		template = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrBefore:
		template = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrAfter:
		template = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
		//default:
		//	return fmt.Errorf("operator: %v is not yet supported", operator)
	}

	return template
}

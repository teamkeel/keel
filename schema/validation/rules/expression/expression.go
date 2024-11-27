package expression

// type Rule func(asts []*parser.AST, expression *parser.Expression, context expressions.ExpressionContext) []error

// func ValidateExpression(asts []*parser.AST, expression *parser.Expression, rules []Rule, context expressions.ExpressionContext) (errors []error) {
// 	for _, rule := range rules {
// 		errs := rule(asts, expression, context)
// 		errors = append(errors, errs...)
// 	}

// 	return errors
// }

// // Validates that the field type has a zero value (no expression necessary).
// // Zero values are the following:
// // * Text -> ""
// // * Number => 0
// // * Boolean -> false
// // * ID -> a ksuid
// // * Timestamp -> now
// func DefaultCanUseZeroValueRule(asts []*parser.AST, attr *parser.AttributeNode, context expressions.ExpressionContext) (errors []*errorhandling.ValidationError) {
// 	typesWithZeroValue := []string{"Text", "Number", "Boolean", "ID", "Timestamp"}

// 	if !lo.Contains(typesWithZeroValue, context.Field.Type.Value) {
// 		errors = append(errors,
// 			errorhandling.NewValidationError(
// 				errorhandling.ErrorDefaultExpressionNeeded,
// 				errorhandling.TemplateLiterals{},
// 				attr,
// 			),
// 		)
// 		return errors
// 	}

// 	return errors
// }

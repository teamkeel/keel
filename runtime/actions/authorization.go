package actions

import "errors"

func DefaultAuthorizeAction(scope *Scope, args RequestArguments, result map[string]any) error {
	authorized, err := EvaluatePermissions(scope.context, scope.operation, scope.schema, toLowerCamelMap(result))

	if err != nil {
		return err
	}

	if !authorized {
		return errors.New("not authorized to access this operation")
	}

	return nil
}

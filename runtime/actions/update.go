package actions

type UpdateAction struct {
	Action
}

func (action *UpdateAction) Execute(args RequestArguments) (*ActionResult, error) {
	err := action.query.Updates(action.Scope.writeValues).Error

	if err != nil {
		return nil, err
	}

	result := ActionResult(toLowerCamelMap(action.writeValues))

	return &result, nil
}

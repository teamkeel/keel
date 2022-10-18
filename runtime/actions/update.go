package actions

type UpdateAction struct {
	*Action[UpdateResult]
}

type UpdateResult struct {
	Object map[string]any `json:"object"`
}

func (action *UpdateAction) Initialise(scope *Scope) ActionBuilder[UpdateResult] {
	action.Action = &Action[UpdateResult]{
		Scope: scope,
	}
	return action
}

func (action *UpdateAction) Execute(args RequestArguments) (*ActionResult[UpdateResult], error) {
	err := action.query.Updates(action.Scope.writeValues).Error

	if err != nil {
		return nil, err
	}

	return &ActionResult[UpdateResult]{
		Value: UpdateResult{
			Object: toLowerCamelMap(action.writeValues),
		},
	}, nil
}

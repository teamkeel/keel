package actions

type UpdateAction struct {
	Action[UpdateResult]
}

type UpdateResult struct {
	Object map[string]any `json:"object"`
}

func (action *UpdateAction) Execute(args RequestArguments) (*ActionResult[UpdateResult], error) {
	err := action.query.Updates(action.Scope.writeValues).Error

	if err != nil {
		return nil, err
	}

	return &ActionResult[UpdateResult]{
		Value: UpdateResult{
			Object: map[string]any{
				"object": toLowerCamelMap(action.writeValues),
			},
		},
	}, nil
}

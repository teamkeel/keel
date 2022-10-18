package actions

type CreateAction struct {
	*Action[CreateResult]
}

type CreateResult struct {
	Object map[string]any `json:"object"`
}

func (action *CreateAction) Initialise(scope *Scope) ActionBuilder[CreateResult] {
	action.Action = &Action[CreateResult]{
		Scope: scope,
	}
	return action
}

func (c *CreateAction) Execute(args RequestArguments) (*ActionResult[CreateResult], error) {
	err := c.query.Create(c.Scope.writeValues).Error

	if err != nil {
		return nil, err
	}
	result := toLowerCamelMap(c.writeValues)

	return &ActionResult[CreateResult]{
		Value: CreateResult{
			Object: result,
		},
	}, nil
}

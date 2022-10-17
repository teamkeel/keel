package actions

type CreateAction struct {
	Action[CreateResult]
}

type CreateResult struct {
	Object map[string]any `json:"object"`
}

func (c *CreateAction) Execute(args RequestArguments) (*ActionResult[CreateResult], error) {
	err := c.query.Create(c.Scope.writeValues).Error

	if err != nil {
		return nil, err
	}
	result := toLowerCamelMap(c.writeValues)

	return &ActionResult[CreateResult]{
		Value: CreateResult{
			Object: map[string]any{
				"object": result,
			},
		},
	}, nil
}

// insert into posts (id, title, created_at, updated_at) values('djdjd', 'a post', 20201i, 2023020) returning *

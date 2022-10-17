package actions

type CreateAction struct {
	Action
}

func (c *CreateAction) Execute(args RequestArguments) (*ActionResult, error) {
	err := c.query.Create(c.Scope.writeValues).Error

	if err != nil {
		return nil, err
	}
	result := toLowerCamelMap(c.writeValues)
	res := ActionResult(result)

	return &res, nil
}

// insert into posts (id, title, created_at, updated_at) values('djdjd', 'a post', 20201i, 2023020) returning *

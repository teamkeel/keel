package actions

type CreateAction struct {
	Action
}

func (c *CreateAction) Execute() (*ActionResult, error) {
	return &ActionResult{}, nil
}

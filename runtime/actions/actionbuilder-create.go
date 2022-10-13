package actions

type CreateAction struct {
	Action
}

func (c *CreateAction) Execute() (*Result, error) {
	return &Result{}, nil
}

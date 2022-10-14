package actions

type UpdateAction struct {
	Action
}

func (action *UpdateAction) Execute() (*ActionResult, error) {
	return &ActionResult{}, nil
}

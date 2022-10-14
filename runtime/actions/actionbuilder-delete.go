package actions

type DeleteAction struct {
	Action
}

func (action *DeleteAction) Execute() (*ActionResult, error) {
	return &ActionResult{}, nil
}

package actions

type DeleteAction struct {
	Action
}

func (action *DeleteAction) Execute() (*Result, error) {
	return &Result{}, nil
}

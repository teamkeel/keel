package actions

type UpdateAction struct {
	Action
}

func (action *UpdateAction) Execute() (*Result, error) {
	return &Result{}, nil
}

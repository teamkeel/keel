package actions

type ListAction struct {
	Action
}

func (action *ListAction) Execute() (*Result, error) {
	return &Result{}, nil
}

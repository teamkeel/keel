package actions

type DeleteAction struct {
	Action
}

func (action *DeleteAction) Execute(args RequestArguments) (*ActionResult, error) {
	record := []map[string]any{}
	err := action.query.Delete(record).Error

	result := ActionResult(map[string]any{
		"success": err == nil,
	})

	if err != nil {
		return &result, err
	}

	return &result, nil
}

package actions

type DeleteAction struct {
	Action[DeleteResult]
}

type DeleteResult struct {
	Success bool `json:"success"`
}

func (action *DeleteAction) Execute(args RequestArguments) (*ActionResult[DeleteResult], error) {
	record := []map[string]any{}
	err := action.query.Delete(record).Error

	result := ActionResult[DeleteResult]{
		Value: DeleteResult{
			Success: err != nil,
		},
	}

	if err != nil {
		return &result, err
	}

	return &result, nil
}

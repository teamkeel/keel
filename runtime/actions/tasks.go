package actions

import (
	"errors"

	"github.com/karlseguin/typed"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

func CreateTask(scope *Scope, input map[string]any) (map[string]any, error) {
	var err error
	typedInput := typed.New(input)

	topic := typedInput.String("topic")
	taskModel := proto.FindModel(scope.Schema.Models, parser.TaskModelName)
	if taskModel == nil {
		return nil, errors.New("topic does not exist")
	}

	input = map[string]any{
		"typeType": "Type",
		"status":   "New",
	}

	query := NewQuery(taskModel)
	err = query.captureWriteValues(scope, input)
	if err != nil {
		return nil, err
	}
	query.AppendReturning(AllFields())
	statement := query.InsertStatement(scope.Context)

	newTask, err := statement.ExecuteToSingle(scope.Context)
	if err != nil {
		return nil, err
	}

	_ = proto.FindModel(scope.Schema.Models, topic) //

	return newTask, nil
}

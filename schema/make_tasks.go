package schema

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	createTaskInputMessageName = "CreateTaskInput"
	updateTaskInputMessageName = "UpdateTaskInput"
)

// makeBuiltInTasks will make all the items required for Keel Tasks: Task Model, TaskStatus & TaskType Enum
func (scm *Builder) makeBuiltInTasks() {
	statusEnum := &proto.Enum{
		Name: parser.TaskStatusEnumName,
		Values: []*proto.EnumValue{
			{Name: parser.TaskStatusOpen},
			{Name: parser.TaskStatusAssigned},
			{Name: parser.TaskStatusDeferred},
			{Name: parser.TaskStatusCompleted},
			{Name: parser.TaskStatusCancelled},
		},
	}
	typeEnum := &proto.Enum{
		Name:   parser.TaskTypeEnumName,
		Values: []*proto.EnumValue{},
	}
	scm.proto.Enums = append(scm.proto.Enums, statusEnum, typeEnum)

	// add the create task input message
	scm.proto.Messages = append(scm.proto.Messages, makeCreateTaskInputMessage())
	// add the update task input message
	scm.proto.Messages = append(scm.proto.Messages, makeUpdateTaskInputMessage())

	protoModel := &proto.Model{
		Name: parser.TaskModelName,
		Fields: []*proto.Field{
			{
				ModelName:  parser.TaskModelName,
				Name:       parser.FieldNameId,
				PrimaryKey: true,
				Unique:     true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameType,
				Optional:  false,
				Type: &proto.TypeInfo{
					Type:     proto.Type_TYPE_ENUM,
					EnumName: wrapperspb.String(parser.TaskTypeEnumName),
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameStatus,
				Optional:  false,
				Type: &proto.TypeInfo{
					Type:     proto.Type_TYPE_ENUM,
					EnumName: wrapperspb.String(parser.TaskStatusEnumName),
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameAssignedTo,
				Optional:  true,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_MODEL,
					ModelName: wrapperspb.String(parser.IdentityModelName),
				},
				ForeignKeyFieldName: wrapperspb.String(fmt.Sprintf("%sId", parser.TaskFieldNameAssignedTo)),
			},
			{
				ModelName: parser.TaskModelName,
				Name:      fmt.Sprintf("%sId", parser.TaskFieldNameAssignedTo),
				Optional:  true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
				ForeignKeyInfo: &proto.ForeignKeyInfo{
					RelatedModelName:  parser.IdentityModelName,
					RelatedModelField: parser.FieldNameId,
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameAssignedAt,
				Optional:  true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_DATETIME,
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameResolvedBy,
				Optional:  true,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_MODEL,
					ModelName: wrapperspb.String(parser.IdentityModelName),
				},
				ForeignKeyFieldName: wrapperspb.String(fmt.Sprintf("%sId", parser.TaskFieldNameResolvedBy)),
			},
			{
				ModelName: parser.TaskModelName,
				Name:      fmt.Sprintf("%sId", parser.TaskFieldNameResolvedBy),
				Optional:  true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
				ForeignKeyInfo: &proto.ForeignKeyInfo{
					RelatedModelName:  parser.IdentityModelName,
					RelatedModelField: parser.FieldNameId,
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameResolvedAt,
				Optional:  true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_DATETIME,
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameDeferredUntil,
				Optional:  true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_DATETIME,
				},
			},
			{
				ModelName: parser.TaskModelName,
				Name:      parser.TaskFieldNameVisibleFrom,
				Optional:  true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_DATETIME,
				},
			},
			{
				ModelName:    parser.TaskModelName,
				Name:         parser.FieldNameCreatedAt,
				Type:         &proto.TypeInfo{Type: proto.Type_TYPE_DATETIME},
				DefaultValue: &proto.DefaultValue{UseZeroValue: true},
			},
			{
				ModelName:    parser.TaskModelName,
				Name:         parser.FieldNameUpdatedAt,
				Type:         &proto.TypeInfo{Type: proto.Type_TYPE_DATETIME},
				DefaultValue: &proto.DefaultValue{UseZeroValue: true},
			},
		},
		Actions: []*proto.Action{
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameCreateTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_CREATE,
				InputMessageName:    createTaskInputMessageName, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameGetTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_READ,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameUpdateTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    updateTaskInputMessageName, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameCompleteTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    updateTaskInputMessageName, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameAssignTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameDeferTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameCancelTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameGetNextTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_READ,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameListTopics,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_READ,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameListTasks,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_READ,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
		},
	}

	scm.proto.Models = append(scm.proto.Models, protoModel)
}

// makeTopic will add a new task topic to the schema
//
// This includes:
// * adding new type to the TaskType enum
// * adding a new model for the topic's Fields
// * adding a new model for the topic's Inputs
func (scm *Builder) makeTopic(decl *parser.DeclarationNode) {
	// add new value to the TaskType enum
	topicNode := decl.Topic
	topicName := topicNode.Name.Value
	if typeEnum := proto.FindEnum(scm.proto.Enums, parser.TaskTypeEnumName); typeEnum != nil {
		typeEnum.Values = append(typeEnum.Values, &proto.EnumValue{Name: topicName})
	}

	// create new fields model with a relationship field to the task model
	fieldsModelName := makeTopicFieldsModelName(topicName)
	fieldsModel := &proto.Model{
		Name: fieldsModelName,
		Fields: []*proto.Field{
			{
				ModelName:  fieldsModelName,
				Name:       parser.FieldNameId,
				PrimaryKey: true,
				Unique:     true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
			},
			{
				ModelName: fieldsModelName,
				Name:      parser.TaskFieldNameTask,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_MODEL,
					ModelName: wrapperspb.String(parser.TaskModelName),
				},
				Optional:            false,
				Unique:              true,
				ForeignKeyFieldName: wrapperspb.String(fmt.Sprintf("%sId", parser.TaskFieldNameTask)),
			},
			{
				ModelName: fieldsModelName,
				Name:      fmt.Sprintf("%sId", parser.TaskFieldNameTask),
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
				Optional: false,
				Unique:   true,
				ForeignKeyInfo: &proto.ForeignKeyInfo{
					RelatedModelName:  parser.TaskModelName,
					RelatedModelField: parser.FieldNameId,
				},
			},
			{
				ModelName:    fieldsModelName,
				Name:         parser.FieldNameCreatedAt,
				Type:         &proto.TypeInfo{Type: proto.Type_TYPE_DATETIME},
				DefaultValue: &proto.DefaultValue{UseZeroValue: true},
			},
			{
				ModelName:    fieldsModelName,
				Name:         parser.FieldNameUpdatedAt,
				Type:         &proto.TypeInfo{Type: proto.Type_TYPE_DATETIME},
				DefaultValue: &proto.DefaultValue{UseZeroValue: true},
			},
		},
	}

	// create new inputs model with a relationship field to the task model
	inputsModelName := makeTopicInputsModelName(topicName)
	inputsModel := &proto.Model{
		Name: inputsModelName,
		Fields: []*proto.Field{
			{
				ModelName:  inputsModelName,
				Name:       parser.FieldNameId,
				PrimaryKey: true,
				Unique:     true,
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
			},
			{
				ModelName: inputsModelName,
				Name:      parser.TaskFieldNameTask,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_MODEL,
					ModelName: wrapperspb.String(parser.TaskModelName),
				},
				Optional:            false,
				Unique:              true,
				ForeignKeyFieldName: wrapperspb.String(fmt.Sprintf("%sId", parser.TaskFieldNameTask)),
			},
			{
				ModelName: inputsModelName,
				Name:      fmt.Sprintf("%sId", parser.TaskFieldNameTask),
				Type: &proto.TypeInfo{
					Type: proto.Type_TYPE_ID,
				},
				Optional: false,
				Unique:   true,
				ForeignKeyInfo: &proto.ForeignKeyInfo{
					RelatedModelName:  parser.TaskModelName,
					RelatedModelField: parser.FieldNameId,
				},
			},
			{
				ModelName:    inputsModelName,
				Name:         parser.FieldNameCreatedAt,
				Type:         &proto.TypeInfo{Type: proto.Type_TYPE_DATETIME},
				DefaultValue: &proto.DefaultValue{UseZeroValue: true},
			},
			{
				ModelName:    inputsModelName,
				Name:         parser.FieldNameUpdatedAt,
				Type:         &proto.TypeInfo{Type: proto.Type_TYPE_DATETIME},
				DefaultValue: &proto.DefaultValue{UseZeroValue: true},
			},
		},
	}

	// make the message used to create a task for this topic. This message will be used to populate the Fields model
	fieldsMessage := &proto.Message{
		Name: makeTopicCreateMessageName(topicName),
		Fields: []*proto.MessageField{
			{
				MessageName: makeTopicCreateMessageName(topicName),
				Name:        parser.TaskFieldNameType,
				Type: &proto.TypeInfo{
					Type:               proto.Type_TYPE_STRING_LITERAL,
					StringLiteralValue: wrapperspb.String(topicName),
				},
			},
		},
	}
	// make the message used to update a task for this topic. This message will be used to populate the Inputs model
	inputsMessage := &proto.Message{
		Name: makeTopicUpdateMessageName(topicNode.Name.Value),
	}

	for _, section := range topicNode.Sections {
		switch {
		case section.Fields != nil:
			// make the model fields
			fields := scm.makeFields(section.Fields, fieldsModel.Name)
			fieldsModel.Fields = append(fieldsModel.Fields, fields...)

			// add the fields to the input message
			for _, field := range section.Fields {
				fieldsMessage.Fields = append(fieldsMessage.Fields, &proto.MessageField{
					Name:        field.Name.Value,
					Type:        scm.parserFieldToProtoTypeInfo(field),
					Optional:    field.Optional,
					Nullable:    false, // TODO: can explicit inputs use the null value?
					MessageName: fieldsMessage.Name,
				})
			}
		case section.Inputs != nil:
			// make the model fields
			fields := scm.makeFields(section.Inputs, inputsModel.Name)
			inputsModel.Fields = append(inputsModel.Fields, fields...)

			// add the fields to the input message
			for _, field := range section.Inputs {
				inputsMessage.Fields = append(inputsMessage.Fields, &proto.MessageField{
					Name:        field.Name.Value,
					Type:        scm.parserFieldToProtoTypeInfo(field),
					Optional:    field.Optional,
					Nullable:    false, // TODO: can explicit inputs use the null value?
					MessageName: inputsMessage.Name,
				})
			}
		}
	}

	scm.proto.Models = append(scm.proto.Models, fieldsModel)
	scm.proto.Models = append(scm.proto.Models, inputsModel)

	// add the create (fieldsMessage) and update (inputsMessage) input messages
	scm.proto.Messages = append(scm.proto.Messages, fieldsMessage, inputsMessage)

	// and add the message to the union type of the createTaskInput
	cm := proto.FindMessage(scm.proto.Messages, createTaskInputMessageName)
	cm.Type.UnionNames = append(cm.Type.UnionNames, wrapperspb.String(fieldsMessage.Name))

	// and add the message to the union type of the updateTaskInput
	um := proto.FindMessage(scm.proto.Messages, updateTaskInputMessageName)
	umf := proto.FindMessageField(um, "values")
	umf.Type.UnionNames = append(umf.Type.UnionNames, wrapperspb.String(inputsMessage.Name))
}

// makeCreateTaskInputMessage creates a input message to be used as part of the createTask action
func makeCreateTaskInputMessage() *proto.Message {
	return &proto.Message{
		Name:   createTaskInputMessageName,
		Fields: []*proto.MessageField{},
		Type: &proto.TypeInfo{
			Type:       proto.Type_TYPE_UNION,
			UnionNames: []*wrapperspb.StringValue{},
		},
	}
}

// makeUpdateTaskInputMessage creates a input message to be used as part of the updateTask/completeTask action
func makeUpdateTaskInputMessage() *proto.Message {
	return &proto.Message{
		Name: updateTaskInputMessageName,
		Fields: []*proto.MessageField{
			{
				MessageName: updateTaskInputMessageName,
				Name:        "task_id",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(parser.TaskModelName),
					FieldName: wrapperspb.String(parser.FieldNameId),
				},
			},
			{
				MessageName: updateTaskInputMessageName,
				Name:        "values",
				Type: &proto.TypeInfo{
					Type:       proto.Type_TYPE_UNION,
					UnionNames: []*wrapperspb.StringValue{},
				},
			},
		},
	}
}

// makeTopicFieldsModelName returns a model name for fields model for the given topic
func makeTopicFieldsModelName(topicName string) string {
	return fmt.Sprintf("%sFields", topicName)
}

// makeTopicInputsModelName returns a model name for inputs model for the given topic
func makeTopicInputsModelName(topicName string) string {
	return fmt.Sprintf("%sInputs", topicName)
}

// makeTopicCreateMessageName returns a name for the message used to create a new task for the given topic
func makeTopicCreateMessageName(topicName string) string {
	return fmt.Sprintf("%sCreateInput", topicName)
}

// makeTopicUpdateMessageName returns a name for the message used to update a task for the given topic
func makeTopicUpdateMessageName(topicName string) string {
	return fmt.Sprintf("%sUpdateInput", topicName)
}

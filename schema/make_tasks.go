package schema

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
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
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
				ResponseMessageName: parser.MessageFieldTypeAny, // TODO: make this something else
			},
			{
				ModelName:           parser.TaskModelName,
				Name:                parser.TaskActionNameCompleteTask,
				Implementation:      proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME,
				Type:                proto.ActionType_ACTION_TYPE_WRITE,
				InputMessageName:    parser.MessageFieldTypeAny, // TODO: make this something else
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
	if typeEnum := proto.FindEnum(scm.proto.Enums, parser.TaskTypeEnumName); typeEnum != nil {
		typeEnum.Values = append(typeEnum.Values, &proto.EnumValue{Name: topicNode.Name.Value})
	}

	// create new fields model with a relationship field to the task model
	fieldsModelName := makeTopicFieldsModelName(topicNode.Name.Value)
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
	inputsModelName := makeTopicInputsModelName(topicNode.Name.Value)
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

	for _, section := range topicNode.Sections {
		switch {
		case section.Fields != nil:
			fields := scm.makeFields(section.Fields, fieldsModel.Name)
			fieldsModel.Fields = append(fieldsModel.Fields, fields...)
		case section.Inputs != nil:
			fields := scm.makeFields(section.Inputs, inputsModel.Name)
			inputsModel.Fields = append(inputsModel.Fields, fields...)
		}
	}

	scm.proto.Models = append(scm.proto.Models, fieldsModel)
	scm.proto.Models = append(scm.proto.Models, inputsModel)
}

// makeTopicFieldsModelName returns a model name for fields model for the given topic
func makeTopicFieldsModelName(topicName string) string {
	return fmt.Sprintf("%sFields", topicName)
}

// makeTopicInputsModelName returns a model name for inputs model for the given topic
func makeTopicInputsModelName(topicName string) string {
	return fmt.Sprintf("%sInputs", topicName)
}

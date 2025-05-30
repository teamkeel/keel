package proto

import (
	"sort"

	"github.com/samber/lo"
)

// FindApi locates the API of the given name.
func FindApi(s *Schema, name string) *Api {
	api, _ := lo.Find(s.Apis, func(m *Api) bool {
		return m.Name == name
	})

	return api
}

// ApiModels provides the models defined in an API.
func ApiModels(s *Schema, api *Api) []*Model {
	return lo.Map(api.ApiModels, func(a *ApiModel, _ int) *Model {
		return FindModel(s.Models, a.ModelName)
	})
}

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
//
// Deprecated: Use Schema.ModelNames() instead.
func ModelNames(p *Schema) []string {
	names := lo.Map(p.Models, func(x *Model, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
}

// FieldNames provides a (sorted) list of the fields in the model of
// the given name.
//
// Deprecated: Please use Model.FieldNames() instead.
func FieldNames(m *Model) []string {
	names := lo.Map(m.Fields, func(x *Field, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
}

// IsTypeModel returns true of the field's type is Model.
//
// Deprecated: Please use Field.IsTypeModel() instead.
func IsTypeModel(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL
}

// IsTypeRepeated returns true if the field is specified as
// being "repeated".
//
// Deprecated: Please use Field.IsRepeated() instead.
func IsRepeated(field *Field) bool {
	return field.Type.Repeated
}

// PrimaryKeyFieldName returns the name of the field in the given model,
// that is marked as being the model's primary key. (Or empty string).
//
// Deprecated: please use Model.PrimaryKeyFieldName() instead.
func PrimaryKeyFieldName(model *Model) string {
	field, _ := lo.Find(model.Fields, func(f *Field) bool {
		return f.PrimaryKey
	})
	if field != nil {
		return field.Name
	}
	return ""
}

// AllFields provides a list of all the model fields specified in the schema.
//
// Deprecated: please use Schema.AllFields() instead.
func AllFields(p *Schema) []*Field {
	fields := []*Field{}
	for _, model := range p.Models {
		fields = append(fields, model.Fields...)
	}
	return fields
}

// ForeignKeyFields returns all the fields in the given model which have their ForeignKeyInfo
// populated.
//
// Deprecated: please use Model.ForeignKeyFields() instead.
func ForeignKeyFields(model *Model) []*Field {
	return lo.Filter(model.Fields, func(f *Field, _ int) bool {
		return f.ForeignKeyInfo != nil
	})
}

// Deprecated: please use Field.IsHasMany() instead.
func IsHasMany(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL && field.ForeignKeyFieldName == nil && field.Type.Repeated
}

// Deprecated: please use Field.IsHasOne() instead.
func IsHasOne(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL && field.ForeignKeyFieldName == nil && !field.Type.Repeated
}

// Deprecated: please use Field.IsBelongsTo() instead.
func IsBelongsTo(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL && field.ForeignKeyFieldName != nil && !field.Type.Repeated
}

// GetForeignKeyFieldName returns the foreign key field name for the given field if it
// represents a relationship to another model. It returns an empty string if field's type is
// not a model.
// The foreign key returned might exists on field's parent model, or on the model field
// it is related to, so this function would normally be used in conjunction with
// IsBelongsTo or it's counterparts to determine on which side the foreign
// key lives.
func GetForeignKeyFieldName(models []*Model, field *Field) string {
	// The query is not meaningful if the field is not of type Model.
	if field.Type.Type != Type_TYPE_MODEL {
		return ""
	}

	// The answer is trivial if the field is already marked with a FK field name.
	if field.ForeignKeyFieldName != nil {
		return field.ForeignKeyFieldName.Value
	}

	// Repeated model fields will "know" their inverse field name if was defined in the input
	// schema with an @relation attribute.
	//
	// When that is the case we can go off and find
	// that field in the related model, and that related model field will in turn,
	// know the name of its sibling foreign key field name.
	if field.InverseFieldName != nil {
		relatedModelName := field.Type.ModelName.Value
		inverseField := FindField(models, relatedModelName, field.InverseFieldName.Value)
		fkName := inverseField.ForeignKeyFieldName.Value
		return fkName
	}

	// If we get this far, we must search for fields in the related thisModel to infer the answer.
	// NB. Schema validation guarantees that there will never be more than one
	// candidate in the latter case.
	thisModel := FindModel(models, field.ModelName)
	relatedModel := FindModel(models, field.Type.ModelName.Value)
	relatedField, _ := lo.Find(relatedModel.Fields, func(field *Field) bool {
		return field.Type.Type == Type_TYPE_MODEL && field.Type.ModelName.Value == thisModel.Name
	})
	return relatedField.ForeignKeyFieldName.Value
}

// ModelsExists returns true if the given schema contains a
// model with the given name.
func ModelExists(models []*Model, name string) bool {
	for _, m := range models {
		if m.Name == name {
			return true
		}
	}
	return false
}

// FindModel locates the model of the given name.
//
// Deprecated: use Schema.FindModel() instead.
func FindModel(models []*Model, name string) *Model {
	model, _ := lo.Find(models, func(m *Model) bool {
		return m.Name == name
	})
	return model
}

// FindEnum locates the enum of the given name.
func FindEnum(enums []*Enum, name string) *Enum {
	enum, _ := lo.Find(enums, func(m *Enum) bool {
		return m.Name == name
	})
	return enum
}

// Deprecated: Use Schema.FilterActions() instead.
func FilterActions(p *Schema, filter func(op *Action) bool) (ops []*Action) {
	for _, model := range p.Models {
		actions := model.Actions

		for _, o := range actions {
			if filter(o) {
				ops = append(ops, o)
			}
		}
	}

	return ops
}

// Deprecated: Use Schema.FindAction() instead.
func FindAction(schema *Schema, actionName string) *Action {
	actions := schema.FilterActions(func(op *Action) bool {
		return op.Name == actionName
	})
	if len(actions) != 1 {
		return nil
	}
	return actions[0]
}

// Deprecated: Use Action.IsFunction() instead.
func ActionIsFunction(action *Action) bool {
	return action.Implementation == ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
}

// Deprecated: Use Action.IsArbitraryFunction() instead.
func ActionIsArbitraryFunction(action *Action) bool {
	return action.IsFunction() && (action.Type == ActionType_ACTION_TYPE_READ || action.Type == ActionType_ACTION_TYPE_WRITE)
}

// Deprecated: Use Action.IsWriteAction() instead.
func IsWriteAction(action *Action) bool {
	switch action.Type {
	case ActionType_ACTION_TYPE_CREATE, ActionType_ACTION_TYPE_DELETE, ActionType_ACTION_TYPE_WRITE, ActionType_ACTION_TYPE_UPDATE:
		return true
	default:
		return false
	}
}

// Deprecated: Use Action.IsReadAction() instead.
func IsReadAction(action *Action) bool {
	switch action.Type {
	case ActionType_ACTION_TYPE_GET, ActionType_ACTION_TYPE_LIST, ActionType_ACTION_TYPE_READ:
		return true
	default:
		return false
	}
}

// FindModels locates and returns the models whose names match up with those
// specified in the given names to find.
func FindModels(allModels []*Model, namesToFind []string) (foundModels []*Model) {
	for _, candidateModel := range allModels {
		if lo.Contains(namesToFind, candidateModel.Name) {
			foundModels = append(foundModels, candidateModel)
		}
	}
	return foundModels
}

func FindField(models []*Model, modelName string, fieldName string) *Field {
	model := FindModel(models, modelName)
	for _, field := range model.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// ModelHasField returns true IF the schema contains a model of the given name AND
// that model has a field of the given name.
func ModelHasField(schema *Schema, model string, field string) bool {
	for _, m := range schema.Models {
		if m.Name != model {
			continue
		}
		for _, f := range m.Fields {
			if f.Name == field {
				return true
			}
		}
	}
	return false
}

// EnumExists returns true if the given schema contains a
// enum with the given name.
func EnumExists(enums []*Enum, name string) bool {
	for _, m := range enums {
		if m.Name == name {
			return true
		}
	}
	return false
}

// FindRole locates and returns the Role object that has the given name.
func FindRole(roleName string, schema *Schema) *Role {
	for _, role := range schema.Roles {
		if role.Name == roleName {
			return role
		}
	}
	return nil
}

// FindJob locates the job of the given name.
//
// Deprecated: please use Schema.FindJob() instead.
func FindJob(jobs []*Job, name string) *Job {
	job, _ := lo.Find(jobs, func(m *Job) bool {
		return m.Name == name
	})
	return job
}

// GetActionNamesForApi returns all the actions available on an API.
func GetActionNamesForApi(p *Schema, api *Api) []string {
	actions := []string{}
	for _, v := range api.ApiModels {
		for _, f := range v.ModelActions {
			actions = append(actions, f.ActionName)
		}
	}

	return actions
}

// PermissionsWithRole returns a list of those permission present in the given permissions
// list, which have at least one Role-based permission rule. This does not imply that the
// returned Permissions might not also have some expression-based rules.
func PermissionsWithRole(permissions []*PermissionRule) []*PermissionRule {
	withRoles := []*PermissionRule{}
	for _, perm := range permissions {
		if len(perm.RoleNames) > 0 {
			withRoles = append(withRoles, perm)
		}
	}
	return withRoles
}

type PermissionFilter = func(p *PermissionRule) bool

func PermissionsForAction(schema *Schema, action *Action, filters ...PermissionFilter) (permissions []*PermissionRule) {
	// if there are any action level permissions, then these take priority
	if len(action.Permissions) > 0 {
		return action.Permissions
	}

	// if there are no action level permissions, then we fallback to model level permissions
	// that match the type of the action
	opTypePermissions := PermissionsForActionType(schema, action.ModelName, action.Type)
	permissions = append(permissions, opTypePermissions...)

	if len(filters) == 0 {
		return permissions
	}

	filtered := []*PermissionRule{}

permissions:
	for _, permission := range permissions {
		for _, filter := range filters {
			if !filter(permission) {
				filtered = append(filtered, permission)

				continue permissions
			}
		}
	}

	return filtered
}

// PermissionsForActionType returns a list of permissions defined for an action type on a model.
func PermissionsForActionType(schema *Schema, modelName string, actionType ActionType) []*PermissionRule {
	permissions := []*PermissionRule{}

	model := FindModel(schema.Models, modelName)

	for _, perm := range model.Permissions {
		if lo.Contains(perm.ActionTypes, actionType) {
			permissions = append(permissions, perm)
		}
	}

	return permissions
}

// PermissionsWithExpression returns a list of those permission present in the given permissions
// list, which have at least one expression-based permission rule. This does not imply that the
// returned Permissions might not also have some role-based rules.
func PermissionsWithExpression(permissions []*PermissionRule) []*PermissionRule {
	withPermissions := []*PermissionRule{}
	for _, perm := range permissions {
		if perm.Expression != nil {
			withPermissions = append(withPermissions, perm)
		}
	}
	return withPermissions
}

// FindMessage will find a message type defined in a Keel schema based on the name of the message
// e.g
// FindMessage("MyMessage") will return this node:
// message MyMessage {}
//
// Deprecated: Please use Schema.FindMessage instead.
func FindMessage(messages []*Message, messageName string) *Message {
	message, _ := lo.Find(messages, func(m *Message) bool {
		return m.Name == messageName
	})
	return message
}

// Deprecated: Use Message.FindField() instead.
func FindMessageField(message *Message, fieldName string) *MessageField {
	for _, field := range message.Fields {
		if field.Name == fieldName {
			return field
		}
	}

	return nil
}

// For built-in action types, returns the "values" input message, which may be nested inside the
// root message for some action types, or returns nil if not found.
func FindValuesInputMessage(schema *Schema, actionName string) *Message {
	action := schema.FindAction(actionName)
	message := schema.FindMessage(action.InputMessageName)

	switch action.Type {
	case ActionType_ACTION_TYPE_CREATE:
		return message
	case ActionType_ACTION_TYPE_UPDATE:
		for _, v := range message.Fields {
			if v.Name == "values" && v.Type.Type == Type_TYPE_MESSAGE {
				return schema.FindMessage(v.Type.MessageName.Value)
			}
		}
	}
	return nil
}

// For built-in action types, returns the "where" input message, which may be nested inside the
// root message for some action types, or returns nil if not found.
func FindWhereInputMessage(schema *Schema, actionName string) *Message {
	action := schema.FindAction(actionName)
	message := schema.FindMessage(action.InputMessageName)

	switch action.Type {
	case ActionType_ACTION_TYPE_GET,
		ActionType_ACTION_TYPE_DELETE:
		return message
	case ActionType_ACTION_TYPE_LIST,
		ActionType_ACTION_TYPE_UPDATE:
		for _, v := range message.Fields {
			if v.Name == "where" && v.Type.Type == Type_TYPE_MESSAGE {
				return schema.FindMessage(v.Type.MessageName.Value)
			}
		}
	}
	return nil
}

func MessageUsedAsResponse(schema *Schema, msgName string) bool {
	for _, model := range schema.Models {
		for _, action := range model.Actions {
			if action.ResponseMessageName == msgName {
				return true
			}
		}
	}

	return false
}

// FindSubscriber locates the subscriber of the given name.
func FindSubscriber(subscribers []*Subscriber, name string) *Subscriber {
	subscriber, _ := lo.Find(subscribers, func(m *Subscriber) bool {
		return m.Name == name
	})
	return subscriber
}

// FindEvent locates the event of the given name.
func FindEvent(subscribers []*Event, name string) *Event {
	event, _ := lo.Find(subscribers, func(m *Event) bool {
		return m.Name == name
	})
	return event
}

// FindEventSubscriptions locates the subscriber of the given name.
//
// Deprecated: use Schema.FindEventSubscribers instead.
func FindEventSubscriptions(schema *Schema, event *Event) []*Subscriber {
	subscribers := lo.Filter(schema.Subscribers, func(m *Subscriber, _ int) bool {
		return lo.Contains(m.EventNames, event.Name)
	})
	return subscribers
}

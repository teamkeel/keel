package proto

import (
	"sort"

	"github.com/samber/lo"
)

// FindApi locates the API of the given name.
func FindApi(s *Schema, name string) *Api {
	api, _ := lo.Find(s.GetApis(), func(m *Api) bool {
		return m.GetName() == name
	})

	return api
}

// ApiModels provides the models defined in an API.
func ApiModels(s *Schema, api *Api) []*Model {
	return lo.Map(api.GetApiModels(), func(a *ApiModel, _ int) *Model {
		return FindModel(s.GetModels(), a.GetModelName())
	})
}

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
//
// Deprecated: Use Schema.ModelNames() instead.
func ModelNames(p *Schema) []string {
	names := lo.Map(p.GetModels(), func(x *Model, _ int) string {
		return x.GetName()
	})
	sort.Strings(names)
	return names
}

// FieldNames provides a (sorted) list of the fields in the model of
// the given name.
//
// Deprecated: Please use Model.FieldNames() instead.
func FieldNames(m *Model) []string {
	names := lo.Map(m.GetFields(), func(x *Field, _ int) string {
		return x.GetName()
	})
	sort.Strings(names)
	return names
}

// IsTypeModel returns true of the field's type is Model.
//
// Deprecated: Please use Field.IsTypeModel() instead.
func IsTypeModel(field *Field) bool {
	return field.GetType().GetType() == Type_TYPE_MODEL
}

// IsTypeRepeated returns true if the field is specified as
// being "repeated".
//
// Deprecated: Please use Field.IsRepeated() instead.
func IsRepeated(field *Field) bool {
	return field.GetType().GetRepeated()
}

// PrimaryKeyFieldName returns the name of the field in the given model,
// that is marked as being the model's primary key. (Or empty string).
//
// Deprecated: please use Model.PrimaryKeyFieldName() instead.
func PrimaryKeyFieldName(model *Model) string {
	field, _ := lo.Find(model.GetFields(), func(f *Field) bool {
		return f.GetPrimaryKey()
	})
	if field != nil {
		return field.GetName()
	}
	return ""
}

// AllFields provides a list of all the model fields specified in the schema.
//
// Deprecated: please use Schema.AllFields() instead.
func AllFields(p *Schema) []*Field {
	fields := []*Field{}
	for _, model := range p.GetModels() {
		fields = append(fields, model.GetFields()...)
	}
	return fields
}

// ForeignKeyFields returns all the fields in the given model which have their ForeignKeyInfo
// populated.
//
// Deprecated: please use Model.ForeignKeyFields() instead.
func ForeignKeyFields(model *Model) []*Field {
	return lo.Filter(model.GetFields(), func(f *Field, _ int) bool {
		return f.GetForeignKeyInfo() != nil
	})
}

// Deprecated: please use Field.IsHasMany() instead.
func IsHasMany(field *Field) bool {
	return field.GetType().GetType() == Type_TYPE_MODEL && field.GetForeignKeyFieldName() == nil && field.GetType().GetRepeated()
}

// Deprecated: please use Field.IsHasOne() instead.
func IsHasOne(field *Field) bool {
	return field.GetType().GetType() == Type_TYPE_MODEL && field.GetForeignKeyFieldName() == nil && !field.GetType().GetRepeated()
}

// Deprecated: please use Field.IsBelongsTo() instead.
func IsBelongsTo(field *Field) bool {
	return field.GetType().GetType() == Type_TYPE_MODEL && field.GetForeignKeyFieldName() != nil && !field.GetType().GetRepeated()
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
	if field.GetType().GetType() != Type_TYPE_MODEL {
		return ""
	}

	// The answer is trivial if the field is already marked with a FK field name.
	if field.GetForeignKeyFieldName() != nil {
		return field.GetForeignKeyFieldName().GetValue()
	}

	// Repeated model fields will "know" their inverse field name if was defined in the input
	// schema with an @relation attribute.
	//
	// When that is the case we can go off and find
	// that field in the related model, and that related model field will in turn,
	// know the name of its sibling foreign key field name.
	if field.GetInverseFieldName() != nil {
		relatedModelName := field.GetType().GetModelName().GetValue()
		inverseField := FindField(models, relatedModelName, field.GetInverseFieldName().GetValue())
		fkName := inverseField.GetForeignKeyFieldName().GetValue()
		return fkName
	}

	// If we get this far, we must search for fields in the related thisModel to infer the answer.
	// NB. Schema validation guarantees that there will never be more than one
	// candidate in the latter case.
	thisModel := FindModel(models, field.GetModelName())
	relatedModel := FindModel(models, field.GetType().GetModelName().GetValue())
	relatedField, _ := lo.Find(relatedModel.GetFields(), func(field *Field) bool {
		return field.GetType().GetType() == Type_TYPE_MODEL && field.GetType().GetModelName().GetValue() == thisModel.GetName()
	})
	return relatedField.GetForeignKeyFieldName().GetValue()
}

// ModelsExists returns true if the given schema contains a
// model with the given name.
func ModelExists(models []*Model, name string) bool {
	for _, m := range models {
		if m.GetName() == name {
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
		return m.GetName() == name
	})
	return model
}

// FindEnum locates the enum of the given name.
func FindEnum(enums []*Enum, name string) *Enum {
	enum, _ := lo.Find(enums, func(m *Enum) bool {
		return m.GetName() == name
	})
	return enum
}

// Deprecated: Use Schema.FilterActions() instead.
func FilterActions(p *Schema, filter func(op *Action) bool) (ops []*Action) {
	for _, model := range p.GetModels() {
		actions := model.GetActions()

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
		return op.GetName() == actionName
	})
	if len(actions) != 1 {
		return nil
	}
	return actions[0]
}

// Deprecated: Use Action.IsFunction() instead.
func ActionIsFunction(action *Action) bool {
	return action.GetImplementation() == ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
}

// Deprecated: Use Action.IsArbitraryFunction() instead.
func ActionIsArbitraryFunction(action *Action) bool {
	return action.IsFunction() && (action.GetType() == ActionType_ACTION_TYPE_READ || action.GetType() == ActionType_ACTION_TYPE_WRITE)
}

// Deprecated: Use Action.IsWriteAction() instead.
func IsWriteAction(action *Action) bool {
	switch action.GetType() {
	case ActionType_ACTION_TYPE_CREATE, ActionType_ACTION_TYPE_DELETE, ActionType_ACTION_TYPE_WRITE, ActionType_ACTION_TYPE_UPDATE:
		return true
	default:
		return false
	}
}

// Deprecated: Use Action.IsReadAction() instead.
func IsReadAction(action *Action) bool {
	switch action.GetType() {
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
		if lo.Contains(namesToFind, candidateModel.GetName()) {
			foundModels = append(foundModels, candidateModel)
		}
	}
	return foundModels
}

func FindField(models []*Model, modelName string, fieldName string) *Field {
	model := FindModel(models, modelName)
	for _, field := range model.GetFields() {
		if field.GetName() == fieldName {
			return field
		}
	}
	return nil
}

// ModelHasField returns true IF the schema contains a model of the given name AND
// that model has a field of the given name.
func ModelHasField(schema *Schema, model string, field string) bool {
	for _, m := range schema.GetModels() {
		if m.GetName() != model {
			continue
		}
		for _, f := range m.GetFields() {
			if f.GetName() == field {
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
		if m.GetName() == name {
			return true
		}
	}
	return false
}

// FindRole locates and returns the Role object that has the given name.
func FindRole(roleName string, schema *Schema) *Role {
	for _, role := range schema.GetRoles() {
		if role.GetName() == roleName {
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
		return m.GetName() == name
	})
	return job
}

// GetActionNamesForApi returns all the actions available on an API.
func GetActionNamesForApi(p *Schema, api *Api) []string {
	actions := []string{}
	for _, v := range api.GetApiModels() {
		for _, f := range v.GetModelActions() {
			actions = append(actions, f.GetActionName())
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
		if len(perm.GetRoleNames()) > 0 {
			withRoles = append(withRoles, perm)
		}
	}
	return withRoles
}

type PermissionFilter = func(p *PermissionRule) bool

func PermissionsForAction(schema *Schema, action *Action, filters ...PermissionFilter) (permissions []*PermissionRule) {
	// if there are any action level permissions, then these take priority
	if len(action.GetPermissions()) > 0 {
		return action.GetPermissions()
	}

	// if there are no action level permissions, then we fallback to model level permissions
	// that match the type of the action
	opTypePermissions := PermissionsForActionType(schema, action.GetModelName(), action.GetType())
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

	model := FindModel(schema.GetModels(), modelName)

	for _, perm := range model.GetPermissions() {
		if lo.Contains(perm.GetActionTypes(), actionType) {
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
		if perm.GetExpression() != nil {
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
		return m.GetName() == messageName
	})
	return message
}

// Deprecated: Use Message.FindField() instead.
func FindMessageField(message *Message, fieldName string) *MessageField {
	for _, field := range message.GetFields() {
		if field.GetName() == fieldName {
			return field
		}
	}

	return nil
}

// For built-in action types, returns the "values" input message, which may be nested inside the
// root message for some action types, or returns nil if not found.
func FindValuesInputMessage(schema *Schema, actionName string) *Message {
	action := schema.FindAction(actionName)
	message := schema.FindMessage(action.GetInputMessageName())

	switch action.GetType() {
	case ActionType_ACTION_TYPE_CREATE:
		return message
	case ActionType_ACTION_TYPE_UPDATE:
		for _, v := range message.GetFields() {
			if v.GetName() == "values" && v.GetType().GetType() == Type_TYPE_MESSAGE {
				return schema.FindMessage(v.GetType().GetMessageName().GetValue())
			}
		}
	}
	return nil
}

// For built-in action types, returns the "where" input message, which may be nested inside the
// root message for some action types, or returns nil if not found.
func FindWhereInputMessage(schema *Schema, actionName string) *Message {
	action := schema.FindAction(actionName)
	message := schema.FindMessage(action.GetInputMessageName())

	switch action.GetType() {
	case ActionType_ACTION_TYPE_GET,
		ActionType_ACTION_TYPE_DELETE:
		return message
	case ActionType_ACTION_TYPE_LIST,
		ActionType_ACTION_TYPE_UPDATE:
		for _, v := range message.GetFields() {
			if v.GetName() == "where" && v.GetType().GetType() == Type_TYPE_MESSAGE {
				return schema.FindMessage(v.GetType().GetMessageName().GetValue())
			}
		}
	}
	return nil
}

func MessageUsedAsResponse(schema *Schema, msgName string) bool {
	for _, model := range schema.GetModels() {
		for _, action := range model.GetActions() {
			if action.GetResponseMessageName() == msgName {
				return true
			}
		}
	}

	return false
}

// FindSubscriber locates the subscriber of the given name.
func FindSubscriber(subscribers []*Subscriber, name string) *Subscriber {
	subscriber, _ := lo.Find(subscribers, func(m *Subscriber) bool {
		return m.GetName() == name
	})
	return subscriber
}

// FindEvent locates the event of the given name.
func FindEvent(subscribers []*Event, name string) *Event {
	event, _ := lo.Find(subscribers, func(m *Event) bool {
		return m.GetName() == name
	})
	return event
}

// FindEventSubscriptions locates the subscriber of the given name.
//
// Deprecated: use Schema.FindEventSubscribers instead.
func FindEventSubscriptions(schema *Schema, event *Event) []*Subscriber {
	subscribers := lo.Filter(schema.GetSubscribers(), func(m *Subscriber, _ int) bool {
		return lo.Contains(m.GetEventNames(), event.GetName())
	})
	return subscribers
}

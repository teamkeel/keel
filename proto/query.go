package proto

import (
	"sort"

	"github.com/samber/lo"
)

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
func ModelNames(p *Schema) []string {
	names := lo.Map(p.Models, func(x *Model, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
}

// FieldNames provides a (sorted) list of the fields in the model of
// the given name.
func FieldNames(m *Model) []string {
	names := lo.Map(m.Fields, func(x *Field, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
}

// IsTypeModel returns true of the field's type is Model.
func IsTypeModel(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL
}

// IsTypeRepeated returns true if the field is specified as
// being "repeated".
func IsRepeated(field *Field) bool {
	return field.Type.Repeated
}

// PrimaryKeyFieldName returns the name of the field in the given model,
// that is marked as being the model's primary key. (Or empty string).
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
func AllFields(p *Schema) []*Field {
	fields := []*Field{}
	for _, model := range p.Models {
		fields = append(fields, model.Fields...)
	}
	return fields
}

// IdFields returns all the fields in the given model which have type Type_TYPE_ID.
func ForeignKeyFields(model *Model) []*Field {
	return lo.Filter(model.Fields, func(f *Field, _ int) bool {
		return f.ForeignKeyInfo != nil
	})
}

func IsHasMany(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL && field.ForeignKeyFieldName == nil && field.Type.Repeated
}

func IsHasOne(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL && field.ForeignKeyFieldName == nil && !field.Type.Repeated
}

func IsBelongsTo(field *Field) bool {
	return field.Type.Type == Type_TYPE_MODEL && field.ForeignKeyFieldName != nil && !field.Type.Repeated
}

// GetForignKeyFieldName returns the foreign key field name for the relationship that
// field has to another model, or an empty string if field's type is not a model.
// Foreign key returned might exists on field's parent model, or on the model field
// is related to, so this function would normally be used in conjunction with
// IsBelongsTo or it's counterparts to determine on which side the foreign
// key lives
func GetForignKeyFieldName(models []*Model, field *Field) string {
	if field.Type.Type != Type_TYPE_MODEL {
		return ""
	}

	if field.ForeignKeyFieldName != nil {
		return field.ForeignKeyFieldName.Value
	}

	model := FindModel(models, field.ModelName)
	relatedModel := FindModel(models, field.Type.ModelName.Value)
	relatedField, _ := lo.Find(relatedModel.Fields, func(field *Field) bool {
		return field.Type.Type == Type_TYPE_MODEL && field.Type.ModelName.Value == model.Name
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

func FilterOperations(p *Schema, filter func(op *Operation) bool) (ops []*Operation) {
	for _, model := range p.Models {
		operations := model.Operations

		for _, o := range operations {
			if filter(o) {
				ops = append(ops, o)
			}
		}
	}

	return ops
}

func FindOperation(schema *Schema, operationName string) *Operation {
	operations := FilterOperations(schema, func(op *Operation) bool {
		return op.Name == operationName
	})
	if len(operations) != 1 {
		return nil
	}
	return operations[0]
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

// ModelHasField returns true IFF the schema contains a model of the given name AND
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

// FindInput returns the input on a given operation
func FindInput(op *Operation, name string) *OperationInput {
	for _, input := range op.Inputs {
		if input.Name == name {
			return input
		}
	}
	return nil
}

// OperationHasInput returns true if the given Operation defines
// an input of the given name.
func OperationHasInput(op *Operation, name string) bool {
	for _, input := range op.Inputs {
		if input.Name == name {
			return true
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

func GetActionNamesForApi(p *Schema, api *Api) []string {
	modelNames := lo.Map(api.ApiModels, func(m *ApiModel, _ int) string {
		return m.ModelName
	})

	models := FindModels(p.Models, modelNames)

	actions := []string{}
	for _, m := range models {
		for _, op := range m.Operations {
			actions = append(actions, op.Name)
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

// IsModelField returns true if the input targets a model field
// and is handled automatically by the runtime.
// This will only be true for inputs that are part of operations,
// as functions never have this behaviour.
func (i *OperationInput) IsModelField() bool {
	return len(i.Target) > 0
}

func FindMessage(messages []*Message, messageName string) *Message {
	message, _ := lo.Find(messages, func(m *Message) bool {
		return m.Name == messageName
	})
	return message
}

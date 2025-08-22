package query

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func APIs(asts []*parser.AST) (res []*parser.APINode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.API != nil {
				res = append(res, decl.API)
			}
		}
	}
	return res
}

// APIModelNodes returns all the models included in the given API.
func APIModelNodes(api *parser.APINode) (res []*parser.APIModelNode) {
	for _, section := range api.Sections {
		res = append(res, section.Models...)
	}

	return res
}

type ModelFilter func(m *parser.ModelNode) bool

func ExcludeBuiltInModels(m *parser.ModelNode) bool {
	return !m.BuiltIn
}

func Entities(asts []*parser.AST) (res []parser.Entity) {
	for _, model := range Models(asts) {
		res = append(res, model)
	}
	for _, task := range Tasks(asts) {
		res = append(res, task)
	}
	return res
}

func Models(asts []*parser.AST, filters ...ModelFilter) (res []*parser.ModelNode) {
	for _, ast := range asts {
	models:
		for _, decl := range ast.Declarations {
			if decl.Model == nil {
				continue
			}

			for _, filter := range filters {
				if !filter(decl.Model) {
					continue models
				}
			}

			res = append(res, decl.Model)
		}
	}
	return res
}

func Tasks(asts []*parser.AST) (res []*parser.TaskNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Task != nil {
				res = append(res, decl.Task)
			}
		}
	}
	return res
}

func Model(asts []*parser.AST, name string) *parser.ModelNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil && decl.Model.Name.Value == name {
				return decl.Model
			}
		}
	}
	return nil
}

// Entity returns the model or task matching the given name.
func Entity(asts []*parser.AST, name string) parser.Entity {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil && decl.Model.Name.Value == name {
				return decl.Model
			}
			if decl.Task != nil && decl.Task.Name.Value == name {
				return decl.Task
			}
		}
	}
	return nil
}

func Action(asts []*parser.AST, name string) *parser.ActionNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, sec := range decl.Model.Sections {
					for _, action := range sec.Actions {
						if action.Name.Value == name {
							return action
						}
					}
				}
			}
		}
	}
	return nil
}

func ActionModel(asts []*parser.AST, name string) *parser.ModelNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, sec := range decl.Model.Sections {
					for _, action := range sec.Actions {
						if action.Name.Value == name {
							return decl.Model
						}
					}
				}
			}
		}
	}
	return nil
}

func IsModel(asts []*parser.AST, name string) bool {
	return Model(asts, name) != nil
}

func IsForeignKey(asts []*parser.AST, entity parser.Entity, field *parser.FieldNode) bool {
	if !field.BuiltIn {
		return false
	}
	f := entity.Field(strings.TrimSuffix(field.Name.Value, "Id"))
	return f != nil && Entity(asts, f.Type.Value) != nil
}

func IsIdentityModel(asts []*parser.AST, name string) bool {
	return name == parser.IdentityModelName
}

func ModelAttributes(model *parser.ModelNode) (res []*parser.AttributeNode) {
	for _, section := range model.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}
	return res
}

func Enums(asts []*parser.AST) (res []*parser.EnumNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Enum != nil {
				res = append(res, decl.Enum)
			}
		}
	}
	return res
}

func Enum(asts []*parser.AST, name string) *parser.EnumNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Enum != nil && decl.Enum.Name.Value == name {
				return decl.Enum
			}
		}
	}
	return nil
}

func MessageNames(asts []*parser.AST) (ret []string) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Message != nil {
				ret = append(ret, decl.Message.Name.Value)
			}
		}
	}

	return ret
}

func Messages(asts []*parser.AST) (ret []*parser.MessageNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Message != nil {
				ret = append(ret, decl.Message)
			}
		}
	}

	return ret
}

// Message finds the message in the schema of the given name. Or nil.
func Message(asts []*parser.AST, name string) *parser.MessageNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Message != nil && decl.Message.Name.Value == name {
				return decl.Message
			}
		}
	}

	return nil
}

func Jobs(asts []*parser.AST) (ret []*parser.JobNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Job != nil {
				ret = append(ret, decl.Job)
			}
		}
	}

	return ret
}

func Flows(asts []*parser.AST) (ret []*parser.FlowNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Flow != nil {
				ret = append(ret, decl.Flow)
			}
		}
	}
	return ret
}

func IsEnum(asts []*parser.AST, name string) bool {
	return Enum(asts, name) != nil
}

func IsMessage(asts []*parser.AST, name string) bool {
	return Message(asts, name) != nil
}

func Roles(asts []*parser.AST) (res []*parser.RoleNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Role != nil {
				res = append(res, decl.Role)
			}
		}
	}
	return res
}

func IsUserDefinedType(asts []*parser.AST, name string) bool {
	return Entity(asts, name) != nil || Enum(asts, name) != nil
}

func UserDefinedTypes(asts []*parser.AST) (res []string) {
	for _, entity := range Entities(asts) {
		res = append(res, entity.GetName())
	}
	for _, enum := range Enums(asts) {
		res = append(res, enum.Name.Value)
	}
	return res
}

// ModelCreateActions returns all the actions in the given model, which
// are create-type actions.
func ModelCreateActions(model *parser.ModelNode, filters ...ModelActionFilter) (res []*parser.ActionNode) {
	allFilters := []ModelActionFilter{}
	allFilters = append(allFilters, filters...)
	allFilters = append(allFilters, func(a *parser.ActionNode) bool {
		return a.Type.Value == parser.ActionTypeCreate
	})
	return ModelActions(model, allFilters...)
}

type ModelActionFilter func(a *parser.ActionNode) bool

func ModelActions(model *parser.ModelNode, filters ...ModelActionFilter) (res []*parser.ActionNode) {
	for _, section := range model.Sections {
		if len(section.Actions) > 0 {
		actions:
			for _, action := range section.Actions {
				for _, filter := range filters {
					if !filter(action) {
						continue actions
					}
				}

				res = append(res, action)
			}
		}
	}

	return res
}

type FieldFilter func(f *parser.FieldNode) bool

func ExcludeBuiltInFields(f *parser.FieldNode) bool {
	return !f.BuiltIn
}

func ModelFields(model *parser.ModelNode, filters ...FieldFilter) (res []*parser.FieldNode) {
	for _, section := range model.Sections {
		if section.Fields == nil {
			continue
		}

	fields:
		for _, field := range section.Fields {
			for _, filter := range filters {
				if !filter(field) {
					continue fields
				}
			}

			res = append(res, field)
		}
	}

	return res
}

func TaskField(task *parser.TaskNode, name string) *parser.FieldNode {
	for _, section := range task.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
}

func TaskFields(task *parser.TaskNode, filters ...FieldFilter) (res []*parser.FieldNode) {
	for _, section := range task.Sections {
		if section.Fields == nil {
			continue
		}

	fields:
		for _, field := range section.Fields {
			for _, filter := range filters {
				if !filter(field) {
					continue fields
				}
			}

			res = append(res, field)
		}
	}

	return res
}

func FieldHasAttribute(field *parser.FieldNode, name string) bool {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return true
		}
	}
	return false
}

// FieldGetAttribute returns the attribute of the given name on the given field,
// or nil, to signal that it doesn't have one.
func FieldGetAttribute(field *parser.FieldNode, name string) *parser.AttributeNode {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return attr
		}
	}
	return nil
}

func FieldIsUnique(field *parser.FieldNode) bool {
	attrs := []string{parser.AttributePrimaryKey, parser.AttributeUnique, parser.AttributeSequence}
	for _, v := range attrs {
		if FieldHasAttribute(field, v) {
			return true
		}
	}
	return false
}

func FieldIsComputed(field *parser.FieldNode) bool {
	return FieldHasAttribute(field, parser.AttributeComputed)
}

// CompositeUniqueFields returns the model's fields that make up a composite unique attribute.
func CompositeUniqueFields(model *parser.ModelNode, attribute *parser.AttributeNode) []*parser.FieldNode {
	if attribute.Name.Value != parser.AttributeUnique {
		return nil
	}

	fields := []*parser.FieldNode{}

	if len(attribute.Arguments) > 0 {
		operands, err := resolve.AsIdentArray(attribute.Arguments[0].Expression)
		if err != nil {
			return fields
		}

		for _, f := range operands {
			field := model.Field(f.String())
			if field != nil {
				fields = append(fields, field)
			}
		}
	}

	return fields
}

// FieldIsInCompositeUnique returns true if a field is part of a composite unique attribute.
func FieldIsInCompositeUnique(model *parser.ModelNode, field *parser.FieldNode) bool {
	for _, attribute := range ModelAttributes(model) {
		if attribute.Name.Value == parser.AttributeUnique {
			fields := CompositeUniqueFields(model, attribute)
			for _, f := range fields {
				if field == f {
					return true
				}
			}
		}
	}
	return false
}

// ActionSortableFieldNames returns the field names of the @sortable attribute.
// If no @sortable attribute exists, an empty slice is returned.
func ActionSortableFieldNames(action *parser.ActionNode) ([]string, error) {
	fields := []string{}
	var attribute *parser.AttributeNode

	for _, attr := range action.Attributes {
		if attr.Name.Value == parser.AttributeSortable {
			attribute = attr
		}
	}

	if attribute != nil {
		for _, arg := range attribute.Arguments {
			fieldName, err := resolve.AsIdent(arg.Expression)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fieldName.String())
		}
	}

	return fields, nil
}

func ModelFieldNames(model *parser.ModelNode) []string {
	names := []string{}
	for _, field := range ModelFields(model, ExcludeBuiltInFields) {
		names = append(names, field.Name.Value)
	}
	return names
}

// FieldsInModelOfType provides a list of the field names for the fields in the
// given model, that have the given type name.
func FieldsInModelOfType(model *parser.ModelNode, requiredType string) []string {
	names := []string{}
	for _, field := range ModelFields(model) {
		if field.Type.Value == requiredType {
			names = append(names, field.Name.Value)
		}
	}
	return names
}

// FieldsInModelOfType provides a list of the field names for the fields in the
// given model or task, that have the given type name.
func FieldsOfType(entity parser.Entity, typeName string) []*parser.FieldNode {
	fields := []*parser.FieldNode{}
	for _, field := range entity.Fields() {
		if field.Type.Value == typeName {
			fields = append(fields, field)
		}
	}
	return fields
}

// AllHasManyRelationFields provides a list of all the fields in the schema
// which are of type Model and which are repeated.
func AllHasManyRelationFields(asts []*parser.AST) []*parser.FieldNode {
	captured := []*parser.FieldNode{}
	for _, model := range Models(asts) {
		for _, field := range ModelFields(model) {
			if IsHasManyModelField(asts, field) {
				captured = append(captured, field)
			}
		}
	}
	return captured
}

// ResolveInputType returns a string representation of the type of the given input.
//
// If the input is explicitly typed using a built in type that type is returned
//
//	example: (foo: Text) -> Text is returned
//
// If `input` refers to a field on the parent model (or a nested field) then the type of that field is returned
//
//	example: (foo: some.field) -> The type of `field` on the model referrred to by `some` is returned
func ResolveInputType(asts []*parser.AST, input *parser.ActionInputNode, parentModel *parser.ModelNode, action *parser.ActionNode) string {
	// handle built-in type
	if parser.IsBuiltInFieldType(input.Type.ToString()) {
		return input.Type.ToString()
	}

	if action.IsArbitraryFunction() && input.Type.ToString() == parser.MessageFieldTypeAny {
		return parser.MessageFieldTypeAny
	}

	field := ResolveInputField(asts, input, parentModel)
	if field != nil {
		return field.Type.Value
	}

	// ResolveInputField above tries to resolve the fragments of an input identifier based on the input being a field
	// The below case covers explicit inputs which are enums
	if len(input.Type.Fragments) == 1 {
		// also try to match the explicit input type annotation against a known enum type
		enum := Enum(asts, input.Type.Fragments[0].Fragment)

		if enum != nil {
			return enum.Name.Value
		}
	}

	return ""
}

// ResolveInputField returns the field that the input's type references.
func ResolveInputField(asts []*parser.AST, input *parser.ActionInputNode, parentModel *parser.ModelNode) (field *parser.FieldNode) {
	// handle built-in type
	if parser.IsBuiltInFieldType(input.Type.ToString()) {
		return nil
	}

	// follow the idents of the type from the current model to wherever it leads...
	model := parentModel
	for _, fragment := range input.Type.Fragments {
		if model == nil {
			return nil
		}
		field = model.Field(fragment.Fragment)
		if field == nil {
			return nil
		}

		model = Model(asts, field.Type.Value)
	}

	return field
}

// PrimaryKey gives you the name of the primary key field on the given
// model. It favours fields that have the AttributePrimaryKey attribute,
// but drops back to the id field if none have.
func PrimaryKey(modelName string, asts []*parser.AST) *parser.FieldNode {
	model := Model(asts, modelName)
	potentialFields := ModelFields(model)

	for _, field := range potentialFields {
		if FieldHasAttribute(field, parser.AttributePrimaryKey) {
			return field
		}
	}

	for _, field := range potentialFields {
		if field.Name.Value == parser.FieldNameId {
			return field
		}
	}
	return nil
}

// IsHasOneModelField returns true if the given field can be inferred to be
// a field that references another model, and is not denoted as being repeated.
func IsHasOneModelField(asts []*parser.AST, field *parser.FieldNode) bool {
	switch {
	case !IsModel(asts, field.Type.Value):
		return false
	case field.Repeated && !field.IsScalar():
		return false
	default:
		return true
	}
}

// IsHasManyModelField returns true if the given field can be inferred to be
// a field that references another model, and is denoted as being REPEATED.
func IsHasManyModelField(asts []*parser.AST, field *parser.FieldNode) bool {
	switch {
	case !IsModel(asts, field.Type.Value):
		return false
	case !field.Repeated:
		return false
	default:
		return true
	}
}

// IsBelongsToModelField returns true if the given field refers to a model
// in which this is in a 1:1 relationship and where the other model owns the relationship.
// This means the other model's field will have @unique defined and also the other model is
// where the foreign key will exist.
func IsBelongsToModelField(asts []*parser.AST, model *parser.ModelNode, field *parser.FieldNode) bool {
	if IsModel(asts, field.Type.Value) {
		for _, v := range ModelFields(Model(asts, field.Type.Value)) {
			if v.Type.Value == model.Name.Value {
				if !v.Repeated && FieldIsUnique(v) {
					return true
				}
			}
		}
	}

	return false
}

// SubscriberNames gets a unique slice of subscriber names which have been defined in the schema.
func SubscriberNames(asts []*parser.AST) (res []string) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, section := range decl.Model.Sections {
					if section.Attribute != nil && section.Attribute.Name.Value == parser.AttributeOn {
						attribute := section.Attribute

						if len(attribute.Arguments) == 2 {
							subscriberArg := attribute.Arguments[1]

							ident, err := resolve.AsIdent(subscriberArg.Expression)
							if err == nil && ident != nil && len(ident.Fragments) == 1 {
								name := ident.String()
								if !lo.Contains(res, name) {
									res = append(res, name)
								}
							}
						}
					}
				}
			}
			if decl.Task != nil {
				for _, section := range decl.Task.Sections {
					if section.Attribute != nil && section.Attribute.Name.Value == parser.AttributeOn {
						attribute := section.Attribute

						if len(attribute.Arguments) == 2 {
							subscriberArg := attribute.Arguments[1]

							ident, err := resolve.AsIdent(subscriberArg.Expression)
							if err == nil && ident != nil && len(ident.Fragments) == 1 {
								name := ident.String()
								if !lo.Contains(res, name) {
									res = append(res, name)
								}
							}
						}
					}
				}
			}
		}
	}

	return res
}

type Relationship struct {
	Entity parser.Entity
	Field  *parser.FieldNode
}

// GetRelationshipCandidates will find all the candidates relationships that can be formed with the related model or task.
// Each relationship field should only have exactly 1 candidate, otherwise there are incorrectly defined relationships in the schema.
func GetRelationshipCandidates(asts []*parser.AST, entity parser.Entity, field *parser.FieldNode) []*Relationship {
	candidates := []*Relationship{}

	otherEntity := Entity(asts, field.Type.Value)
	if otherEntity == nil {
		return candidates
	}

	otherFields := FieldsOfType(otherEntity, entity.GetName())
	theseFields := FieldsOfType(entity, otherEntity.GetName())

	relationAttributeExists := false
	for _, otherField := range otherFields {
		// Skip when the field is the same (for self referencing models)
		if field == otherField {
			continue
		}

		if ValidOneToHasMany(field, otherField) ||
			ValidOneToHasMany(otherField, field) ||
			ValidUniqueOneToHasOne(field, otherField) ||
			ValidUniqueOneToHasOne(otherField, field) {
			// Make sure this candidate is not already being referenced by a @relation attribute on another field
			alreadyReferencedByRelation := false
			for _, f := range theseFields {
				if f.Name.Value == field.Name.Value {
					continue
				}

				attr := FieldGetAttribute(f, parser.AttributeRelation)
				if attr != nil {
					if relation, ok := RelationAttributeValue(attr); ok {
						if relation == otherField.Name.Value {
							alreadyReferencedByRelation = true
						}
					}
				}
			}

			if !alreadyReferencedByRelation {
				// This field has a new relationship candidate with the other model
				candidates = append(candidates, &Relationship{Entity: otherEntity, Field: otherField})
			}

			if FieldHasAttribute(field, parser.AttributeRelation) || FieldHasAttribute(otherField, parser.AttributeRelation) {
				relationAttributeExists = true
			}
		}
	}

	// Only use candidate relationships where an explicit @relation is used
	if relationAttributeExists {
		relationOnlyCandidates := []*Relationship{}

		for _, relationship := range candidates {
			if FieldHasAttribute(field, parser.AttributeRelation) || FieldHasAttribute(relationship.Field, parser.AttributeRelation) {
				relationOnlyCandidates = append(relationOnlyCandidates, relationship)
			}
		}

		candidates = relationOnlyCandidates
	}

	if len(candidates) == 0 && !field.Repeated {
		// When there is no inverse field provided.
		candidates = append(candidates, &Relationship{Entity: otherEntity})
	}

	return candidates
}

// GetRelationship will return the related model and field on that model which forms the relationship.
func GetRelationship(asts []*parser.AST, currentEntity parser.Entity, currentField *parser.FieldNode) (*Relationship, error) {
	candidates := GetRelationshipCandidates(asts, currentEntity, currentField)

	otherEntity := Entity(asts, currentField.Type.Value)
	if otherEntity == nil {
		return nil, nil
	}

	// There can only be exactly one candidate, since schema validation has all passed
	if len(candidates) != 1 {
		return nil, fmt.Errorf("there is not exactly one candidate relationship for %s field on %s", currentField.Name.Value, currentEntity.GetName())
	}

	return candidates[0], nil
}

// Determine if pair form a valid 1:M pattern where, for example:
//
//	belongsTo:  author Author @relation(posts)
//	hasMany:    posts Post[]
func ValidOneToHasMany(belongsTo *parser.FieldNode, hasMany *parser.FieldNode) bool {
	if FieldIsUnique(belongsTo) || FieldIsUnique(hasMany) {
		return false
	}

	if belongsTo.Repeated {
		return false
	}

	if !hasMany.Repeated {
		return false
	}

	// If belongsTo has @relation, check the field name matches hasMany
	belongsToAttribute := FieldGetAttribute(belongsTo, parser.AttributeRelation)
	if belongsToAttribute != nil {
		if relation, ok := RelationAttributeValue(belongsToAttribute); ok {
			if relation != hasMany.Name.Value {
				return false
			}
		} else {
			return false
		}
	}

	// If hasMany has @relation, then this is not a candidate
	hasManyAttribute := FieldGetAttribute(hasMany, parser.AttributeRelation)

	return hasManyAttribute == nil
}

// Determine if pair form a valid 1:! pattern where, for example:
//
//	hasOne:  	  passport Passport @unique
//	belongsTo:    person Person
func ValidUniqueOneToHasOne(hasOne *parser.FieldNode, belongsTo *parser.FieldNode) bool {
	if !FieldIsUnique(hasOne) || FieldIsUnique(belongsTo) {
		return false
	}

	if belongsTo.Repeated || hasOne.Repeated {
		return false
	}

	otherFieldAttribute := FieldGetAttribute(belongsTo, parser.AttributeRelation)
	if otherFieldAttribute != nil {
		return false
	}

	// If hasOne has @relation, check the field name matches belongsTo
	hasOneAttribute := FieldGetAttribute(hasOne, parser.AttributeRelation)
	if hasOneAttribute != nil {
		if relation, ok := RelationAttributeValue(hasOneAttribute); ok {
			if relation != belongsTo.Name.Value {
				return false
			}
		} else {
			return false
		}
	}

	// If belongsTo has @relation, then this is not a candidate
	belongsToAttribute := FieldGetAttribute(belongsTo, parser.AttributeRelation)

	return belongsToAttribute == nil
}

// RelationAttributeValue attempts to retrieve the value of the @relation attribute.
func RelationAttributeValue(attr *parser.AttributeNode) (field string, ok bool) {
	if len(attr.Arguments) != 1 {
		return "", false
	}

	operand, err := resolve.AsIdent(attr.Arguments[0].Expression)
	if err != nil {
		return "", false
	}

	if operand == nil {
		return "", false
	}

	return operand.Fragments[0], true
}

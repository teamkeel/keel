package query

import (
	"strings"

	"github.com/samber/lo"
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

type ModelFilter func(m *parser.ModelNode) bool

func ExcludeBuiltInModels(m *parser.ModelNode) bool {
	return !m.BuiltIn
}

func Models(asts []*parser.AST, filters ...ModelFilter) (res []*parser.ModelNode) {
	for _, ast := range asts {
	models:
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, filter := range filters {
					if !filter(decl.Model) {
						continue models
					}
				}

				res = append(res, decl.Model)
			}
		}
	}
	return res
}

func ModelNames(asts []*parser.AST, filters ...ModelFilter) (res []string) {
	for _, ast := range asts {

	models:
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, filter := range filters {
					if !filter(decl.Model) {
						continue models
					}
				}

				res = append(res, decl.Model.Name.Value)
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

// Field provides the field of the given name from the given model. (Or nil).
func Field(model *parser.ModelNode, name string) *parser.FieldNode {
	for _, f := range ModelFields(model) {
		if f.Name.Value == name {
			return f
		}
	}
	return nil
}

func IsModel(asts []*parser.AST, name string) bool {
	return Model(asts, name) != nil
}

func IsForeignKey(asts []*parser.AST, model *parser.ModelNode, field *parser.FieldNode) bool {
	if !field.BuiltIn {
		return false
	}
	modelField := Field(model, strings.TrimSuffix(field.Name.Value, "Id"))
	return modelField != nil && Model(asts, modelField.Type.Value) != nil
}

func IsIdentityModel(asts []*parser.AST, name string) bool {
	return name == parser.ImplicitIdentityModelName
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
	return Model(asts, name) != nil || Enum(asts, name) != nil
}

func UserDefinedTypes(asts []*parser.AST) (res []string) {
	for _, model := range Models(asts) {
		res = append(res, model.Name.Value)
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
		if section.Actions != nil && len(section.Actions) > 0 {
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

type ModelFieldFilter func(f *parser.FieldNode) bool

func ExcludeBuiltInFields(f *parser.FieldNode) bool {
	return !f.BuiltIn
}

func ModelFields(model *parser.ModelNode, filters ...ModelFieldFilter) (res []*parser.FieldNode) {
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

func ModelField(model *parser.ModelNode, name string) *parser.FieldNode {
	for _, section := range model.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
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
	return FieldHasAttribute(field, parser.AttributePrimaryKey) || FieldHasAttribute(field, parser.AttributeUnique)
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
			fieldName, err := arg.Expression.ToValue()
			if err != nil {
				return nil, err
			}
			fields = append(fields, fieldName.Ident.Fragments[0].Fragment)
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

// ResolveInputType returns a string represention of the type of the given input.
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

// ResolveInputField returns the field that the input's type references
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
		field = ModelField(model, fragment.Fragment)
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
		if field.Name.Value == parser.ImplicitFieldNameId {
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
	case field.Repeated:
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
							operand, err := subscriberArg.Expression.ToValue()
							if err == nil && operand.Ident != nil && len(operand.Ident.Fragments) == 1 {
								name := operand.Ident.Fragments[0].Fragment
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

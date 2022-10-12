package parser

func APIs(asts []*AST) (res []*APINode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.API != nil {
				res = append(res, decl.API)
			}
		}
	}
	return res
}

type ModelFilter func(m *ModelNode) bool

func ExcludeBuiltInModels(m *ModelNode) bool {
	return !m.BuiltIn
}

func Models(asts []*AST, filters ...ModelFilter) (res []*ModelNode) {
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

func ModelNames(asts []*AST, filters ...ModelFilter) (res []string) {
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

func Model(asts []*AST, name string) *ModelNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil && decl.Model.Name.Value == name {
				return decl.Model
			}
		}
	}
	return nil
}

func IsModel(asts []*AST, name string) bool {
	return Model(asts, name) != nil
}

func IsIdentityModel(asts []*AST, name string) bool {
	return name == ImplicitIdentityModelName
}

func ModelAttributes(model *ModelNode) (res []*AttributeNode) {
	for _, section := range model.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}
	return res
}

func Enums(asts []*AST) (res []*EnumNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Enum != nil {
				res = append(res, decl.Enum)
			}
		}
	}
	return res
}

func Enum(asts []*AST, name string) *EnumNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Enum != nil && decl.Enum.Name.Value == name {
				return decl.Enum
			}
		}
	}
	return nil
}

func IsEnum(asts []*AST, name string) bool {
	return Enum(asts, name) != nil
}

func Roles(asts []*AST) (res []*RoleNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Role != nil {
				res = append(res, decl.Role)
			}
		}
	}
	return res
}

func IsUserDefinedType(asts []*AST, name string) bool {
	return Model(asts, name) != nil || Enum(asts, name) != nil
}

func UserDefinedTypes(asts []*AST) (res []string) {
	for _, model := range Models(asts) {
		res = append(res, model.Name.Value)
	}
	for _, enum := range Enums(asts) {
		res = append(res, enum.Name.Value)
	}
	return res
}

func ModelActions(model *ModelNode) (res []*ActionNode) {
	return append(ModelOperations(model), ModelFunctions(model)...)
}

func ModelOperations(model *ModelNode) (res []*ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Operations...)
	}
	return res
}

func ModelFunctions(model *ModelNode) (res []*ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Functions...)
	}
	return res
}

type ModelFieldFilter func(f *FieldNode) bool

func ExcludeBuiltInFields(f *FieldNode) bool {
	return !f.BuiltIn
}

func ModelFields(model *ModelNode, filters ...ModelFieldFilter) (res []*FieldNode) {
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

func ModelField(model *ModelNode, name string) *FieldNode {
	for _, section := range model.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
}

func FieldHasAttribute(field *FieldNode, name string) bool {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return true
		}
	}
	return false
}

func FieldIsUnique(field *FieldNode) bool {
	return FieldHasAttribute(field, AttributePrimaryKey) || FieldHasAttribute(field, AttributeUnique)
}

func ModelFieldNames(model *ModelNode) []string {
	names := []string{}
	for _, field := range ModelFields(model, ExcludeBuiltInFields) {
		names = append(names, field.Name.Value)
	}
	return names
}

// ResolveInputType returns a string represention of the type of the give input
// If the input is explicitly typed using a built in type that type is returned
//
//	example: (foo: Text) -> Text is returned
//
// If `i` refers to a field on the parent model (or a nested field) then the type of that field is returned
//
//	example: (foo: some.field) -> The type of `field` on the model referrred to by `some` is returned
func ResolveInputType(asts []*AST, input *ActionInputNode, parentModel *ModelNode) string {
	// handle built-in type
	if IsBuiltInFieldType(input.Type.ToString()) {
		return input.Type.ToString()
	}

	field := ResolveInputField(asts, input, parentModel)
	if field != nil {
		return field.Type
	}

	return ""
}

// ResolveInputField returns the field that the input's type references
func ResolveInputField(asts []*AST, input *ActionInputNode, parentModel *ModelNode) (field *FieldNode) {
	// handle built-in type
	if IsBuiltInFieldType(input.Type.ToString()) {
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
		model = Model(asts, field.Type)
	}

	return field
}

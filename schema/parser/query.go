package parser

func (ast *AST) APIs() (res []*APINode) {
	for _, decl := range ast.Declarations {
		if decl.API != nil {
			res = append(res, decl.API)
		}
	}
	return res
}

func (ast *AST) Models() (res []*ModelNode) {
	for _, decl := range ast.Declarations {
		if decl.Model != nil {
			res = append(res, decl.Model)
		}
	}
	return res
}

func (ast *AST) AttributesInModel(modelName string) (res []*AttributeNode) {
	for _, decl := range ast.Declarations {
		if decl.Model != nil {
			if decl.Model.Name.Value != modelName {
				break
			}

			for _, section := range decl.Model.Sections {
				if section.Fields != nil {
					for _, field := range section.Fields {
						if field.BuiltIn {
							continue
						}
						if field.Attributes != nil {
							res = append(res, field.Attributes...)
						}
					}
				}

				if section.Functions != nil {
					for _, function := range section.Functions {
						if function.Attributes != nil {
							res = append(res, function.Attributes...)
						}
					}
				}

				if section.Operations != nil {
					for _, operation := range section.Operations {
						if operation.Attributes != nil {
							res = append(res, operation.Attributes...)
						}
					}
				}

				if section.Attribute != nil {
					res = append(res, section.Attribute)
				}
			}
		}
	}

	return res
}

func (ast *AST) Model(name string) *ModelNode {
	for _, decl := range ast.Declarations {
		if decl.Model != nil && decl.Model.Name.Value == name {
			return decl.Model
		}
	}
	return nil
}

func (model *ModelNode) Attributes() (res []*AttributeNode) {
	for _, section := range model.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}
	return res
}

func (ast *AST) Enums() (res []*EnumNode) {
	for _, decl := range ast.Declarations {
		if decl.Enum != nil {
			res = append(res, decl.Enum)
		}
	}

	return res
}

func (ast *AST) Enum(name string) *EnumNode {
	for _, decl := range ast.Declarations {
		if decl.Enum != nil && decl.Enum.Name.Value == name {
			return decl.Enum
		}
	}
	return nil
}

func (ast *AST) Roles() (res []*RoleNode) {
	for _, decl := range ast.Declarations {
		if decl.Role != nil {
			res = append(res, decl.Role)
		}
	}
	return res
}

func (ast *AST) IsUserDefinedType(name string) bool {
	return ast.Model(name) != nil || ast.Enum(name) != nil
}

func (ast *AST) UserDefinedTypes() (res []string) {
	for _, model := range ast.Models() {
		res = append(res, model.Name.Value)
	}
	for _, enum := range ast.Enums() {
		res = append(res, enum.Name.Value)
	}
	return res
}

// todo: loop over decl
func (model *ModelNode) Actions() (res []*ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Functions...)
		res = append(res, section.Operations...)
	}
	return res
}

// todo: loop over decl
func (model *ModelNode) Fields() (res []*FieldNode) {
	for _, section := range model.Sections {
		res = append(res, section.Fields...)
	}
	return res
}

// todo: loop over decl
func (model *ModelNode) Field(name string) *FieldNode {
	for _, section := range model.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
}

// todo: loop over decl
func (field *FieldNode) HasAttribute(name string) bool {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return true
		}
	}
	return false
}

func (field *FieldNode) IsUnique() bool {
	return field.HasAttribute(AttributePrimaryKey) || field.HasAttribute(AttributeUnique)
}

package query

import "github.com/teamkeel/keel/schema/parser"

func APIs(ast *parser.AST) (res []*parser.APINode) {
	for _, decl := range ast.Declarations {
		if decl.API != nil {
			res = append(res, decl.API)
		}
	}
	return res
}

func Models(ast *parser.AST) (res []*parser.ModelNode) {
	for _, decl := range ast.Declarations {
		if decl.Model != nil {
			res = append(res, decl.Model)
		}
	}
	return res
}

func AttributesInModel(ast *parser.AST, modelName string) (res []*parser.AttributeNode) {
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

func Model(ast *parser.AST, name string) *parser.ModelNode {
	for _, decl := range ast.Declarations {
		if decl.Model != nil && decl.Model.Name.Value == name {
			return decl.Model
		}
	}
	return nil
}

func ModelAttributes(model *parser.ModelNode) (res []*parser.AttributeNode) {
	for _, section := range model.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}
	return res
}

func Enums(ast *parser.AST) (res []*parser.EnumNode) {
	for _, decl := range ast.Declarations {
		if decl.Enum != nil {
			res = append(res, decl.Enum)
		}
	}

	return res
}

func Enum(ast *parser.AST, name string) *parser.EnumNode {
	for _, decl := range ast.Declarations {
		if decl.Enum != nil && decl.Enum.Name.Value == name {
			return decl.Enum
		}
	}
	return nil
}

func Roles(ast *parser.AST) (res []*parser.RoleNode) {
	for _, decl := range ast.Declarations {
		if decl.Role != nil {
			res = append(res, decl.Role)
		}
	}
	return res
}

func IsUserDefinedType(ast *parser.AST, name string) bool {
	return Model(ast, name) != nil || Enum(ast, name) != nil
}

func UserDefinedTypes(ast *parser.AST) (res []string) {
	for _, model := range Models(ast) {
		res = append(res, model.Name.Value)
	}
	for _, enum := range Enums(ast) {
		res = append(res, enum.Name.Value)
	}
	return res
}

// todo: loop over decl
func ModelActions(model *parser.ModelNode) (res []*parser.ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Functions...)
		res = append(res, section.Operations...)
	}
	return res
}

// todo: loop over decl
func ModelFields(model *parser.ModelNode) (res []*parser.FieldNode) {
	for _, section := range model.Sections {
		res = append(res, section.Fields...)
	}
	return res
}

// todo: loop over decl
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

// todo: loop over decl
func FieldHasAttribute(field *parser.FieldNode, name string) bool {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return true
		}
	}
	return false
}

func FieldIsUnique(field *parser.FieldNode) bool {
	return FieldHasAttribute(field, parser.AttributePrimaryKey) || FieldHasAttribute(field, parser.AttributeUnique)
}

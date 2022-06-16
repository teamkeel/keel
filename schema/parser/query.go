package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/teamkeel/keel/schema/node"
)

// Finds all APIs defined in an AST
func (ast *AST) APIs() (res []*APINode) {
	for _, decl := range ast.Declarations {
		if decl.API != nil {
			res = append(res, decl.API)
		}
	}
	return res
}

// Finds all models in the AST
func (ast *AST) Models() (res []*ModelNode) {
	for _, decl := range ast.Declarations {
		if decl.Model != nil {
			res = append(res, decl.Model)
		}
	}
	return res
}

// todo: do we need this anymore now we have attributes (maybe we do as we want to find by string name)
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

// Find a model by its name in an AST
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

// Find all enums defined in an AST
func (ast *AST) Enums() (res []*EnumNode) {
	for _, decl := range ast.Declarations {
		if decl.Enum != nil {
			res = append(res, decl.Enum)
		}
	}

	return res
}

// Find a specific enum by name in the AST
func (ast *AST) Enum(name string) *EnumNode {
	for _, decl := range ast.Declarations {
		if decl.Enum != nil && decl.Enum.Name.Value == name {
			return decl.Enum
		}
	}
	return nil
}

// Find all roles defined in the AST
func (ast *AST) Roles() (res []*RoleNode) {
	for _, decl := range ast.Declarations {
		if decl.Role != nil {
			res = append(res, decl.Role)
		}
	}
	return res
}

// Check if the given symbol is a user defined type
func (ast *AST) IsUserDefinedType(name string) bool {
	return ast.Model(name) != nil || ast.Enum(name) != nil
}

// Returns all valid user defined types
func (ast *AST) UserDefinedTypes() (res []string) {
	for _, model := range ast.Models() {
		res = append(res, model.Name.Value)
	}
	for _, enum := range ast.Enums() {
		res = append(res, enum.Name.Value)
	}
	return res
}

// Returns all actions defined within a given model
func (model *ModelNode) Actions() (res []*ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Functions...)
		res = append(res, section.Operations...)
	}
	return res
}

// Returns all fields defined within a model
func (model *ModelNode) Fields() (res []*FieldNode) {
	for _, section := range model.Sections {
		res = append(res, section.Fields...)
	}
	return res
}

// Finds a particular field by name within a model
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

// Checks if a field has a particular attribute
func (field *FieldNode) HasAttribute(name string) bool {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return true
		}
	}
	return false
}

// Checks if a field is marked as unique
func (field *FieldNode) IsUnique() bool {
	return field.HasAttribute(AttributePrimaryKey) || field.HasAttribute(AttributeUnique)
}

func (ast *AST) ResolveAssociation(context *ModelNode, fragments []string) (*node.Node, error) {
	if fragments[0] != strings.ToLower(context.Name.Value) {
		// e.g model is Profile
		// but expression is something.else == 123 where something should be profile (lowercased)
		return nil, errors.New("does not match model context")
	}

	targetFragments := fragments[1:]

	fmt.Print(targetFragments)
	return nil, nil
}

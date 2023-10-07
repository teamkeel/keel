package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type relatedTo struct {
	model *parser.ModelNode
	field *parser.FieldNode
}

func RelationshipsRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	relationships := map[*parser.FieldNode][]relatedTo{}

	return Visitor{
		EnterModel: func(model *parser.ModelNode) {
			relationships = map[*parser.FieldNode][]relatedTo{}
			currentModel = model
		},

		LeaveModel: func(_ *parser.ModelNode) {
			errored := map[*parser.FieldNode]bool{}

			for field, _ := range relationships {
				if len(relationships[field]) > 1 {
					for i, candidate := range relationships[field] {
						if i == 0 {
							continue
						}

						switch {
						case validOneToHasMany(field, candidate.field):
							if !errored[field] {
								errs.AppendError(makeValidationError(
									fmt.Sprintf("Cannot determine which field on the %s model to form the relationship with", candidate.model.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s[] field on the %s model which is not yet in a relationship", currentModel.Name.Value, candidate.model.Name.Value),
									field,
								))
								errored[field] = true
							}
						case validOneToHasMany(candidate.field, field):
							if !errored[candidate.field] {
								errs.AppendError(makeValidationError(
									fmt.Sprintf("Attempting to associate with '%s' on model %s but a relationship may already exist", field.Name.Value, currentModel.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s[] field on the %s model which is not yet in a relationship", candidate.model.Name.Value, currentModel.Name.Value),
									candidate.field,
								))
								errored[candidate.field] = true
							}
							if !errored[field] {
								errs.AppendError(makeValidationError(
									fmt.Sprintf("Attempting to associate with '%s' on model %s but a relationship may already exist", candidate.field.Name.Value, candidate.model.Name.Value),
									"",
									field,
								))
								errored[field] = true
							}
						case validUniqueOneToHasOne(field, candidate.field):
							if !errored[field] {
								errs.AppendError(makeValidationError(
									fmt.Sprintf("Cannot determine which field on the %s model to form the relationship with", candidate.model.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s field on the %s model which is not yet in a relationship", currentModel.Name.Value, candidate.model.Name.Value),
									field,
								))
								errored[field] = true
							}
						case validUniqueOneToHasOne(candidate.field, field):
							if !errored[candidate.field] {
								errs.AppendError(makeValidationError(
									fmt.Sprintf("Attempting to associate with '%s' on model %s but a relationship may already exist", field.Name.Value, currentModel.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s field on the %s model which is not yet in a relationship", candidate.model.Name.Value, currentModel.Name.Value),
									candidate.field,
								))
								errored[candidate.field] = true
							}
							if !errored[field] {
								errs.AppendError(makeValidationError(
									fmt.Sprintf("Attempting to associate with '%s' on model %s but a relationship may already exist", candidate.field.Name.Value, candidate.model.Name.Value),
									"",
									field,
								))
								errored[field] = true
							}
						}
					}
				}
			}

			currentModel = nil
		},
		EnterField: func(currentField *parser.FieldNode) {
			if currentModel == nil {
				// If this is not a model field, then exit.
				return
			}

			// Check that the @relation attribute, if any, is define with exactly a single identifier.
			relationAttr := query.FieldGetAttribute(currentField, parser.AttributeRelation)

			var relation string
			if relationAttr != nil {
				var ok bool
				relation, ok = attributeFirstArgAsIdentifier(relationAttr)
				if !ok {
					errs.AppendError(makeValidationError(
						"The @relation value must refer to a field on the related model",
						"For example, @relation(fieldName)",
						relationAttr,
					))
					return
				}
			}

			// Check that the field type is a model.
			otherModel := query.Model(asts, currentField.Type.Value)
			if otherModel == nil {
				if relationAttr != nil {
					errs.AppendError(makeValidationError(
						"The @relation attribute cannot be used on non-model fields",
						"",
						currentField,
					))
				}

				// If the field type is not a model, then this is not a relationship
				return
			}

			// @relation cannot be defined on a repeated field
			if relationAttr != nil && currentField.Repeated {
				errs.AppendError(makeValidationError(
					"The @relation attribute must be defined on the other side of a one to many relationship",
					"",
					relationAttr,
				))
				return
			}

			if relationAttr != nil {
				otherField := query.Field(otherModel, relation)
				if otherField == nil {
					errs.AppendError(makeValidationError(
						fmt.Sprintf("The field '%s' does not exist on the %s model", relation, otherModel.Name.Value),
						"",
						relationAttr.Arguments[0],
					))
					return
				}

				if otherField.Type.Value != currentModel.Name.Value {
					errs.AppendError(makeValidationError(
						fmt.Sprintf("The field '%s' on the %s model must be of type %s in order to establish a relationship", relation, otherModel.Name.Value, currentModel.Name.Value),
						"",
						relationAttr.Arguments[0],
					))
					return
				}

				if query.FieldIsUnique(otherField) {
					errs.AppendError(makeValidationError(
						fmt.Sprintf("Cannot create a relationship to the unique field '%s' on the %s model", relation, otherModel.Name.Value),
						"In a one to one relationship, only this side must be marked as @unique",
						relationAttr.Arguments[0],
					))
					return
				}

				if !query.FieldIsUnique(currentField) && !otherField.Repeated {
					errs.AppendError(makeValidationError(
						fmt.Sprintf("To create a one to one relationship, the '%s' field should be @unique", currentField.Name.Value),
						"In a one to one relationship, the other side must be marked as @unique",
						currentField.Name,
					))
					return
				}
			}

			otherFields := query.ModelFieldsOfType(otherModel, currentModel.Name.Value)

			matched := false
			for _, otherField := range otherFields {
				if validOneToHasMany(currentField, otherField) ||
					validOneToHasMany(otherField, currentField) ||
					validUniqueOneToHasOne(currentField, otherField) ||
					validUniqueOneToHasOne(otherField, currentField) {
					relationships[currentField] = append(relationships[currentField], relatedTo{model: otherModel, field: otherField})
					matched = true
				}
			}

			if !matched && currentField.Repeated {
				errs.AppendError(makeValidationError(
					fmt.Sprintf("The field '%s' does not have an associated field on the related %s model", currentField.Name.Value, currentField.Type.Value),
					fmt.Sprintf("To create a one to many relationship a other field must be created on the %s model", currentField.Type.Value),
					currentField,
				))
			}
		},
	}
}

// Find candidates for the 1:M pattern where:
//
//	currentField:  parent Parent @relation(children)
//	otherField:    children Children[]
func validOneToHasMany(parent *parser.FieldNode, child *parser.FieldNode) bool {
	// Neither field can be unique in a 1:M relationship
	if query.FieldIsUnique(parent) || query.FieldIsUnique(child) {
		return false
	}

	if parent.Repeated {
		return false
	}

	if !child.Repeated {
		return false
	}

	// Does the relation attribute exist and does it
	relnAttribute := query.FieldGetAttribute(parent, parser.AttributeRelation)
	if relnAttribute != nil {
		if relation, ok := attributeFirstArgAsIdentifier(relnAttribute); ok {
			if relation != child.Name.Value {
				return false
			}
		}
	}

	return true
}

// Find candidates for the 1:1 pattern where:
//
//	currentField:  contactInfo ContactInfo @unique
//	otherField:    company Company
func validUniqueOneToHasOne(currentField *parser.FieldNode, otherField *parser.FieldNode) bool {
	if !query.FieldIsUnique(currentField) || query.FieldIsUnique(otherField) {
		return false
	}

	if otherField.Repeated || currentField.Repeated {
		return false
	}

	otherFieldAttribute := query.FieldGetAttribute(otherField, parser.AttributeRelation)
	if otherFieldAttribute != nil {
		return false
	}

	// Does the relation attribute exist and does it
	currentFieldAttribute := query.FieldGetAttribute(currentField, parser.AttributeRelation)
	if currentFieldAttribute != nil {
		if relation, ok := attributeFirstArgAsIdentifier(currentFieldAttribute); ok {
			if relation != otherField.Name.Value {
				return false
			}
		}
	}

	return true
}

func makeValidationError(message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.RelationshipError,
		errorhandling.ErrorDetails{
			Message: message,
			Hint:    hint,
		},
		node,
	)
}

// attributeFirstArgAsIdentifier looks at the given attribute,
// to see if its first argument's expression is a simple identifier.
func attributeFirstArgAsIdentifier(attr *parser.AttributeNode) (theString string, ok bool) {
	if len(attr.Arguments) != 1 {
		return "", false
	}
	expr := attr.Arguments[0].Expression

	operand, err := expr.ToValue()
	if err != nil {
		return "", false
	}
	if operand.Ident == nil {
		return "", false
	}
	theString = operand.Ident.Fragments[0].Fragment
	return theString, true
}

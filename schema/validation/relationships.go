package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type relationship struct {
	model *parser.ModelNode
	field *parser.FieldNode
}

const (
	learnMore = "To learn more about relationships, visit https://docs.keel.so/models#relationships"
)

func RelationshipsRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	candidates := map[*parser.FieldNode][]*relationship{}
	alreadyErrored := map[*parser.FieldNode]bool{}

	return Visitor{
		EnterModel: func(model *parser.ModelNode) {
			// For each relationship field, we generate possible candidates of fields
			// from the other model to form the relationship.  A relationship should
			// only ever have one candidate.
			candidates = map[*parser.FieldNode][]*relationship{}
			currentModel = model
		},

		LeaveModel: func(_ *parser.ModelNode) {

			for field := range candidates {
				if len(candidates[field]) == 1 {
					otherField := candidates[field][0].field
					otherModel := candidates[field][0].model
					if len(candidates[otherField]) > 1 {
						//err - prob already covered
					} else if len(candidates[otherField]) == 1 && candidates[otherField][0].field != field {
						// field in another relationship
						errs.AppendError(makeRelationshipError(
							fmt.Sprintf("Field '%s' on model %s is already in a relationship with field '%s'", otherField.Name.Value, otherModel.Name.Value, candidates[otherField][0].field.Name.Value),
							learnMore,
							field,
						))
					}
				}

				if len(candidates[field]) > 1 {
					for i, candidate := range candidates[field] {
						// Skip the first relationship candidate match
						// since we can assume it to be valid.  For all further
						// candidates we return a validation error.  Each field
						// only have a single candidate on the other end.
						if i == 0 {
							continue
						}

						switch {
						case validOneToHasMany(field, candidate.field):
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot determine which field on the %s model to form a one to many relationship with", candidate.model.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s[] field on the %s model which is not yet in a relationship", currentModel.Name.Value, candidate.model.Name.Value),
									field,
								))
								alreadyErrored[field] = true
							}
						case validOneToHasMany(candidate.field, field):
							if !alreadyErrored[candidate.field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with field '%s' on model %s to form a one to many relationship as a relationship may already exist", field.Name.Value, currentModel.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s[] field on the %s model which is not yet in a relationship", candidate.model.Name.Value, currentModel.Name.Value),
									candidate.field,
								))
								alreadyErrored[candidate.field] = true
							}
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with field '%s' on model %s to form a one to many relationship as a relationship may already exist", candidate.field.Name.Value, candidate.model.Name.Value),
									"",
									field,
								))
								alreadyErrored[field] = true
							}
						case validUniqueOneToHasOne(field, candidate.field):
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot determine which field on the %s model to form a one to one relationship with", candidate.model.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s field on the %s model which is not yet in a relationship", currentModel.Name.Value, candidate.model.Name.Value),
									field,
								))
								alreadyErrored[field] = true
							}
						case validUniqueOneToHasOne(candidate.field, field):
							if !alreadyErrored[candidate.field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with field '%s' on model %s to form a one to one relationship as a relationship may already exist", field.Name.Value, currentModel.Name.Value),
									fmt.Sprintf("Use @relation to refer to a %s field on the %s model which is not yet in a relationship", candidate.model.Name.Value, currentModel.Name.Value),
									candidate.field,
								))
								alreadyErrored[candidate.field] = true
							}
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with field '%s' on model %s to form a one to one relationship as a relationship may already exist", candidate.field.Name.Value, candidate.model.Name.Value),
									learnMore,
									field,
								))
								alreadyErrored[field] = true
							}
						default:
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with field '%s' on model %s to form a relationship", candidate.field.Name.Value, candidate.model.Name.Value),
									learnMore,
									field,
								))
								alreadyErrored[field] = true
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
				relation, ok = relationAttributeField(relationAttr)
				if !ok {
					errs.AppendError(makeRelationshipError(
						"The @relation value must refer to a field on the related model",
						fmt.Sprintf("For example, @relation(fieldName). %s", learnMore),
						relationAttr,
					))
					return
				}
			}

			// Check that the field type is a model.
			otherModel := query.Model(asts, currentField.Type.Value)
			if otherModel == nil {
				if relationAttr != nil {
					errs.AppendError(makeRelationshipError(
						"The @relation attribute cannot be used on non-model fields",
						learnMore,
						currentField,
					))
				}

				// If the field type is not a model, then this is not a relationship
				return
			}

			// @relation cannot be defined on a repeated field
			if relationAttr != nil && currentField.Repeated {
				errs.AppendError(makeRelationshipError(
					"The @relation attribute must be defined on the other side of a one to many relationship",
					learnMore,
					relationAttr,
				))
				return
			}

			if relationAttr != nil {
				otherField := query.Field(otherModel, relation)
				if otherField == nil {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("The field '%s' does not exist on the %s model", relation, otherModel.Name.Value),
						learnMore,
						relationAttr.Arguments[0],
					))
					return
				}

				if otherField.Type.Value != currentModel.Name.Value {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("The field '%s' on the %s model must be of type %s in order to establish a relationship", relation, otherModel.Name.Value, currentModel.Name.Value),
						learnMore,
						relationAttr.Arguments[0],
					))
					return
				}

				if query.FieldIsUnique(otherField) {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("Cannot create a relationship to the unique field '%s' on the %s model", relation, otherModel.Name.Value),
						fmt.Sprintf("In a one to one relationship, only this side must be marked as @unique. %s", learnMore),
						relationAttr.Arguments[0],
					))
					return
				}

				if !query.FieldIsUnique(currentField) && !otherField.Repeated {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("A one to one relationship requires a single side to be @unique"),
						fmt.Sprintf("In a one to one relationship, the '%s' field must be @unique. %s", currentField.Name.Value, learnMore),
						currentField.Name,
					))
					return
				}
			}

			fieldCandidates := findCandidates(asts, currentModel, currentField)

			if len(fieldCandidates) > 0 {
				candidates[currentField] = fieldCandidates
			}

			if len(fieldCandidates) == 0 && currentField.Repeated {
				errs.AppendError(makeRelationshipError(
					fmt.Sprintf("The field '%s' does not have an associated field on the related %s model", currentField.Name.Value, currentField.Type.Value),
					fmt.Sprintf("In a one to many relationship, the related belongs-to field must exist on the %s model. %s", currentField.Type.Value, learnMore),
					currentField,
				))
			}
		},
	}
}

func findCandidates(asts []*parser.AST, currentModel *parser.ModelNode, currentField *parser.FieldNode) []*relationship {
	candidates := []*relationship{}

	otherModel := query.Model(asts, currentField.Type.Value)
	if otherModel == nil {
		return candidates
	}

	otherFields := query.ModelFieldsOfType(otherModel, currentModel.Name.Value)

	for _, otherField := range otherFields {
		if validOneToHasMany(currentField, otherField) ||
			validOneToHasMany(otherField, currentField) ||
			validUniqueOneToHasOne(currentField, otherField) ||
			validUniqueOneToHasOne(otherField, currentField) {
			// This field has a new relationship candidate with the other model
			candidates = append(candidates, &relationship{model: otherModel, field: otherField})
		}
	}

	return candidates
}

// Determine if pair form a valid 1:M pattern where, for example:
//
//	belongsTo:  author Author @relation(posts)
//	hasMany:    posts Post[]
func validOneToHasMany(belongsTo *parser.FieldNode, hasMany *parser.FieldNode) bool {
	// Neither field can be unique in a 1:M relationship
	if query.FieldIsUnique(belongsTo) || query.FieldIsUnique(hasMany) {
		return false
	}

	if belongsTo.Repeated {
		return false
	}

	if !hasMany.Repeated {
		return false
	}

	// If belongsTo has @relation, check the field name matches hasMany
	relnAttribute := query.FieldGetAttribute(belongsTo, parser.AttributeRelation)
	if relnAttribute != nil {
		if relation, ok := relationAttributeField(relnAttribute); ok {
			if relation != hasMany.Name.Value {
				return false
			}
		}
	}

	return true
}

// Determine if pair form a valid 1:! pattern where, for example:
//
//	hasOne:  	  passport Passport @unique
//	belongsTo:    person Person
func validUniqueOneToHasOne(hasOne *parser.FieldNode, belongsTo *parser.FieldNode) bool {
	if !query.FieldIsUnique(hasOne) || query.FieldIsUnique(belongsTo) {
		return false
	}

	if belongsTo.Repeated || hasOne.Repeated {
		return false
	}

	otherFieldAttribute := query.FieldGetAttribute(belongsTo, parser.AttributeRelation)
	if otherFieldAttribute != nil {
		return false
	}

	// If hasOne has @relation, check the field name matches belongsTo
	currentFieldAttribute := query.FieldGetAttribute(hasOne, parser.AttributeRelation)
	if currentFieldAttribute != nil {
		if relation, ok := relationAttributeField(currentFieldAttribute); ok {
			if relation != belongsTo.Name.Value {
				return false
			}
		}
	}

	return true
}

func makeRelationshipError(message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.RelationshipError,
		errorhandling.ErrorDetails{
			Message: message,
			Hint:    hint,
		},
		node,
	)
}

// relationAttributeField attempts to retrieve the value
// of the @relation attribute
func relationAttributeField(attr *parser.AttributeNode) (field string, ok bool) {
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

	return operand.Ident.Fragments[0].Fragment, true
}

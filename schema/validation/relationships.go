package validation

import (
	"fmt"
	"sort"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

const (
	learnMore = "To learn more about relationships, visit https://docs.keel.so/models#relationships"
)

func RelationshipsRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	var candidates map[*parser.FieldNode][]*query.Relationship
	var alreadyErrored map[*parser.FieldNode]bool

	return Visitor{
		EnterModel: func(model *parser.ModelNode) {
			// For each relationship field, we generate possible candidates of fields
			// from the other model to form the relationship.  A relationship should
			// only ever have one candidate.
			candidates = map[*parser.FieldNode][]*query.Relationship{}
			alreadyErrored = map[*parser.FieldNode]bool{}
			currentModel = model
		},

		LeaveModel: func(_ *parser.ModelNode) {

			// Make iterating through the map with deterministic ordering
			orderedKeys := make([]*parser.FieldNode, 0, len(candidates))
			for k := range candidates {
				orderedKeys = append(orderedKeys, k)
			}
			sort.Slice(orderedKeys, func(i, j int) bool {
				return orderedKeys[i].Name.Value < orderedKeys[j].Name.Value
			})

			for _, field := range orderedKeys {
				var pairedCandidate *query.Relationship
				if len(candidates[field]) > 1 {
					for i, candidate := range candidates[field] {
						// Skip the first relationship candidate match
						// since we can assume it to be valid.  For all further
						// candidates we return a validation error.  Each field
						// only have a single candidate on the other end.
						if i == 0 {
							pairedCandidate = candidate
							continue
						}

						if candidate.Field == nil {
							continue
						}

						switch {
						case query.ValidOneToHasMany(field, candidate.Field):
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot form a one to many relationship with field '%s' on %s as it is already associated with field '%s'", field.Name.Value, currentModel.Name.Value, pairedCandidate.Field.Name.Value),
									fmt.Sprintf("Use @relation on '%s' to explicitly create a relationship with this field. For example, %s %s @relation(%s). %s", field.Name.Value, field.Name.Value, candidate.Model.Name.Value, candidate.Field.Name.Value, learnMore),
									candidate.Field.Name,
								))
								alreadyErrored[field] = true
							}
						case query.ValidOneToHasMany(candidate.Field, field):
							if !alreadyErrored[candidate.Field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with repeated field '%s' on %s to form a one to many relationship because it is already associated with field '%s'", field.Name.Value, currentModel.Name.Value, pairedCandidate.Field.Name.Value),
									fmt.Sprintf("Use @relation to refer to another %s[] field on %s which is not yet in a relationship. %s", candidate.Model.Name.Value, currentModel.Name.Value, learnMore),
									candidate.Field.Name,
								))
								alreadyErrored[candidate.Field] = true
							}
						case query.ValidUniqueOneToHasOne(field, candidate.Field):
							if candidate.Model.Name.Value == parser.IdentityModelName {
								// We cannot show errors on the built-in Identity AST nodes, so we rather skip
								// and let the errors get picked up by the other direction.
								continue
							}
							if !alreadyErrored[field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot form a one to one relationship with field '%s' on %s as it is already associated with field '%s'", field.Name.Value, currentModel.Name.Value, pairedCandidate.Field.Name.Value),
									fmt.Sprintf("Use @relation on '%s' to explicitly create a relationship with this field. For example, %s %s @unique @relation(%s). %s", field.Name.Value, field.Name.Value, candidate.Model.Name.Value, candidate.Field.Name.Value, learnMore),
									candidate.Field.Name,
								))
								alreadyErrored[field] = true
							}
						case query.ValidUniqueOneToHasOne(candidate.Field, field):
							if !alreadyErrored[candidate.Field] {
								errs.AppendError(makeRelationshipError(
									fmt.Sprintf("Cannot associate with field '%s' on %s to form a one to one relationship because it is already associated with '%s'", field.Name.Value, currentModel.Name.Value, pairedCandidate.Field.Name.Value),
									fmt.Sprintf("Use @relation to refer to another %s field on %s which is not yet in a relationship. %s", candidate.Model.Name.Value, currentModel.Name.Value, learnMore),
									candidate.Field.Name,
								))
								alreadyErrored[candidate.Field] = true
							}
						default:
							errs.AppendError(makeRelationshipError(
								fmt.Sprintf("Cannot associate with field '%s' on model %s to form a relationship", candidate.Field.Name.Value, candidate.Model.Name.Value),
								learnMore,
								field.Name,
							))
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

			// Check that the field type is a model.
			otherModel := query.Model(asts, currentField.Type.Value)
			if otherModel == nil {
				if relationAttr != nil {
					errs.AppendError(makeRelationshipError(
						"The @relation attribute cannot be used on non-model fields",
						learnMore,
						relationAttr.Name,
					))
				}

				// If the field type is not a model, then this is not a relationship
				return
			}

			// This field is not @unique and relation field on other model is not repeated
			if query.FieldIsUnique(currentField) && currentField.Repeated {
				errs.AppendError(makeRelationshipError(
					"Cannot use @unique on a repeated model field",
					fmt.Sprintf("In a one to one relationship, there are no repeated fields. %s", learnMore),
					currentField.Name,
				))
				return
			}

			if currentField.Optional && currentField.Repeated {
				errs.AppendError(makeRelationshipError(
					"Cannot define a repeated model field as optional",
					learnMore,
					currentField.Name,
				))
				return
			}

			var relation string
			if relationAttr != nil {
				var ok bool
				relation, ok = query.RelationAttributeValue(relationAttr)
				if !ok {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("The @relation argument must refer to a field on %s", otherModel.Name.Value),
						fmt.Sprintf("For example, @relation(fieldName). %s", learnMore),
						relationAttr.Name,
					))
					return
				}
			}

			if relationAttr != nil {
				// @relation cannot be defined on a repeated field
				if currentField.Repeated {
					errs.AppendError(makeRelationshipError(
						"The @relation attribute must be defined on the other side of a one to many relationship",
						learnMore,
						relationAttr.Name,
					))
					return
				}

				// @relation field does not exist
				otherField := query.Field(otherModel, relation)
				if otherField == nil {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("The field '%s' does not exist on %s", relation, otherModel.Name.Value),
						fmt.Sprintf("The @relation argument must refer to a field on %s which is of type %s. %s", otherModel.Name.Value, currentModel.Name.Value, learnMore),
						relationAttr.Arguments[0],
					))
					return
				}

				// @relation field type is not of this model
				if otherField.Type.Value != currentModel.Name.Value {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("The field '%s' on %s must be of type %s in order to establish a relationship", relation, otherModel.Name.Value, currentModel.Name.Value),
						learnMore,
						relationAttr.Arguments[0],
					))
					return
				}

				// @relation field on other model is @unique
				if query.FieldIsUnique(otherField) {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("Cannot create a relationship to the unique field '%s' on %s", relation, otherModel.Name.Value),
						fmt.Sprintf("In a one to one relationship, only this side must be marked as @unique. %s", learnMore),
						relationAttr.Arguments[0],
					))
					return
				}

				// This field is not @unique and relation field on other model is not repeated
				if !query.FieldIsUnique(currentField) && !otherField.Repeated {
					errs.AppendError(makeRelationshipError(
						"A one to one relationship requires a single side to be @unique",
						fmt.Sprintf("In a one to one relationship, the '%s' field must be @unique. %s", currentField.Name.Value, learnMore),
						currentField.Name,
					))
					return
				}

				// This field is @unique and relation field on other model is repeated
				if query.FieldIsUnique(currentField) && otherField.Repeated {
					errs.AppendError(makeRelationshipError(
						fmt.Sprintf("A one to one relationship cannot be made with repeated field '%s' on %s", otherField.Name.Value, otherModel.Name.Value),
						fmt.Sprintf("Either make '%s' non-repeated or define a new non-repeated field on %s. %s", otherField.Name.Value, otherModel.Name.Value, learnMore),
						relationAttr.Arguments[0],
					))
					return
				}
			}

			// Determine all the possible candidate relationships between this field and the related model.
			fieldCandidates := query.GetRelationshipCandidates(asts, currentModel, currentField)

			if len(fieldCandidates) > 0 {
				candidates[currentField] = fieldCandidates
			}

			if len(fieldCandidates) == 0 && currentField.Repeated {
				errs.AppendError(makeRelationshipError(
					fmt.Sprintf("The field '%s' does not have an associated field on %s", currentField.Name.Value, currentField.Type.Value),
					fmt.Sprintf("In a one to many relationship, the related belongs-to field must exist on %s. %s", currentField.Type.Value, learnMore),
					currentField.Name,
				))
			}
		},
	}
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

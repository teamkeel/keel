package relationships

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"golang.org/x/exp/slices"
)

// Make sure the @relation attribute is used properly.
func RelationAttributeRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {

	for _, thisModel := range query.Models(asts) {

		// We accumulate a list of the related fields that are referenced
		// by @Relation attributes in THIS model.
		relatedFieldsCitedByThisModel := []*parser.FieldNode{}

		// We process here, only fields that have the @relation attribute.
		for _, thisField := range fieldsThatHaveRelationAttribute(thisModel) {

			// IMPORTANT NOTE
			//
			// This is a complex validation rule and its checks have (at least conceptually),
			// a dependency order.
			//
			// Often if one fails, the remainder become either infeasible, or would
			// create unhelpful and confusing noise.
			//
			// So as soon as one (in the dependency order) fails, we skip the remainder.

			relationAttr := query.FieldGetAttribute(thisField, parser.AttributeRelation)

			// Make sure @relation is only used on fields of type Model
			if !lo.Contains(query.ModelNames(asts), thisField.Type.Value) {
				errs.Append(
					errorhandling.ErrorRelationAttrOnWrongFieldType,
					map[string]string{
						"FieldName":  thisField.Name.Value,
						"WrongType":  thisField.Type.Value,
						"Suggestion": thisField.Name.Value,
					},
					thisField.Name)
				continue
			}

			// Make sure @relation is only used on fields that are NOT repeated.
			if thisField.Repeated {
				errs.Append(
					errorhandling.ErrorRelationAttrOnNonRepeatedField,
					map[string]string{
						"FieldName": thisField.Name.Value,
					},
					thisField.Name)
				continue
			}

			// Make sure that the attribute's argument (which is unfortunately a
			// list of expressions), boils down to just a single plain string. E.g. @relation(written)
			var relatedFieldName string
			var ok bool
			if relatedFieldName, ok = attributeFirstArgAsIdentifier(relationAttr); !ok {
				errs.Append(
					errorhandling.ErrorRelationAttributShouldBeIdentifier,
					map[string]string{
						"FieldName": thisField.Name.Value,
					},
					relationAttr.Name)

				continue
			}

			// Make sure the value of the @relation attribute (e.g. "written"), exists as a
			// field in the related model.
			relatedModelName := thisField.Type.Value
			relatedModel := query.Model(asts, relatedModelName)
			var relatedField *parser.FieldNode
			if relatedField = query.Field(relatedModel, relatedFieldName); relatedField == nil {
				fieldsAvailable := query.ModelFieldNames(relatedModel)
				suggestedNames := formatting.HumanizeList(fieldsAvailable, formatting.DelimiterOr)
				errs.Append(
					errorhandling.ErrorRelationAttributeUnrecognizedField,
					map[string]string{
						"RelatedFieldName": relatedFieldName,
						"RelatedModelName": relatedModelName,
						"SuggestedNames":   suggestedNames,
					},
					relationAttr.Name)
				continue
			}

			// Make sure the related field is of type <thisModel>
			if relatedField.Type.Value != thisModel.Name.Value {
				suitableFields := query.FieldsInModelOfType(relatedModel, thisModel.Name.Value)
				suggestedFields := formatting.HumanizeList(suitableFields, formatting.DelimiterOr)
				errs.Append(
					errorhandling.ErrorRelationAttributeRelatedFieldWrongType,
					map[string]string{
						"RelatedFieldName": relatedFieldName,
						"RelatedFieldType": relatedField.Type.Value,
						"RequiredType":     thisModel.Name.Value,
						"SuggestedNames":   suggestedFields,
					},
					relationAttr.Name)

				continue
			}

			// The related field must be a repeated field.
			if !relatedField.Repeated {
				errs.Append(
					errorhandling.ErrorRelationAttributeRelatedFieldIsNotRepeated,
					map[string]string{
						"RelatedFieldName": relatedFieldName,
					},
					relationAttr.Name)

				continue
			}

			// None of the related fields cited by this model's @Relationships, must be
			// duplicates. You CAN have more than one 1:many relationships now between model's
			// A and B, but they must use different related fields.
			if slices.Contains(relatedFieldsCitedByThisModel, relatedField) {
				errs.Append(
					errorhandling.ErrorRelationAttributeRelatedFieldIsDuplicated,
					map[string]string{
						"RelatedFieldName": relatedFieldName,
					},
					relationAttr.Name)

				continue
			}
			relatedFieldsCitedByThisModel = append(relatedFieldsCitedByThisModel, relatedField)
		}
	}

	return errs
}

func InvalidOneToOneRelationshipRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	processed := map[string]bool{}

	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			if ok := processed[fmt.Sprintf("%s-%s", model.Name.Value, field.Name.Value)]; ok {
				continue
			}

			if field.Repeated {
				continue
			}

			otherModel := query.Model(asts, field.Type.Value)

			if otherModel == nil {
				continue
			}

			otherModelFields := query.ModelFields(otherModel)

			for _, otherField := range otherModelFields {
				if otherField == field {
					continue
				}
				if otherField.Type.Value != model.Name.Value {
					continue
				}

				// If either the field on model A is repeated
				// or the corresponding field on the other side is repeated
				// then we are not interested
				if otherField.Repeated {
					continue
				}

				// if the field on the other side has a unique attribute, then we know the foreign key
				// should belong there
				if query.FieldIsUnique(otherField) && !query.FieldIsUnique(field) {
					continue
				}

				// similarly, if the field on the lhs is unique and the other side isn't, then this is fine.
				if query.FieldIsUnique(field) && !query.FieldIsUnique(otherField) {
					continue
				}

				// if fields on both sides are both marked as @unique, then this is its own validation error
				if query.FieldIsUnique(field) && query.FieldIsUnique(otherField) {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.RelationshipError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("Field '%s' on %s and '%s' on %s are both marked as @unique", field.Name.Value, model.Name.Value, otherField.Name.Value, otherModel.Name.Value),
								Hint:    "In a one-to-one relationship, only one side must be marked as @unique",
							},
							field.Name,
						),
					)
				} else {
					// check to see if a relation is defined on this attribute that
					// disambiguates the reference.
					relation := query.FieldGetAttribute(field, parser.AttributeRelation)
					if relation != nil {
						relationValue, _ := relation.Arguments[0].Expression.ToValue()
						if relationValue.ToString() != otherField.Name.Value {
							continue
						}
					}

					errs.Append(
						errorhandling.ErrorInvalidOneToOneRelationship,
						map[string]string{
							"ModelA": model.Name.Value,
							"ModelB": field.Type.Value,
						},
						field.Name,
					)
				}

				processed[fmt.Sprintf("%s-%s", model.Name.Value, field.Name.Value)] = true
				processed[fmt.Sprintf("%s-%s", otherModel.Name.Value, otherField.Name.Value)] = true
			}
		}

	}

	return
}

func InvalidImplicitBelongsToWithHasManyRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {

	for _, model := range query.Models(asts) {

	fields:
		for _, field := range query.ModelFields(model) {
			if !field.Repeated {
				continue
			}

			otherModel := query.Model(asts, field.Type.Value)

			if otherModel == nil {
				continue
			}

			otherModelFields := query.ModelFields(otherModel)

			match := false

			for _, otherField := range otherModelFields {
				if otherField.Type.Value != model.Name.Value {
					continue
				}

				if otherField.Repeated {
					continue fields
				}

				match = true

				break
			}

			if !match {
				errs.Append(
					errorhandling.ErrorMissingRelationshipField,
					map[string]string{
						"ModelA":     model.Name.Value,
						"ModelB":     field.Type.Value,
						"Suggestion": fmt.Sprintf("%s %s", casing.ToLowerCamel(model.Name.Value), model.Name.Value),
					},
					field.Name,
				)
			}
		}
	}

	return errs
}

// When ModelA has a set of HasMany relationship fields that references ModelB, then we must make
// sure that ModelB has a corresponding set of reverse HasOne relationships. We need to know
// which field in ModelB *IS* the reverse relationship in order to
// generate the SQL associated with ModelA's HasMany relationship fields.
//
// When a field in ModelB is marked with an @relation attribute - it tells us directly
// and unambiguously which hasMany field in ModelA it "reverses".
// In the code below these are called "qualified".
//
// Provided there's only ONE that is not qualified - we can
// deduce it and that's fine. So the rule is that there must be only one that does not
// carry the @relation attribute.
func MoreThanOneReverseMany(asts []*parser.AST) (errs errorhandling.ValidationErrors) {

	type hasManyField struct {
		theField  *parser.FieldNode
		belongsTo *parser.ModelNode
	}

	// First capture all the relation fields defined by the ASTs, which are of type HasMany.
	hasManyFields := []*hasManyField{}
	for _, model := range query.Models(asts) {
		for _, f := range query.ModelFields(model) {
			if query.IsHasManyModelField(asts, f) {
				hasManyFields = append(hasManyFields, &hasManyField{
					theField:  f,
					belongsTo: model,
				})
			}
		}
	}

	// Now we iterate over all the captured HasMany relation fields in order to investigate
	// the model at the HasOne end of the relationship.
	for _, hasManyF := range hasManyFields {

		singleEndModel := query.Model(asts, hasManyF.theField.Type.Value)

		if singleEndModel == nil {
			// This can be the case for invalid schemas but other rules check for that.
			// For our purposes, it just means we can't proceed with this validation
			// rule right now, for this hasMany field.
			continue
		}

		// Given access to the model at the HasOne end, how many fields does it have that
		// refer back to the model at the hasMany end - which are not *qualified*?
		reverseFields := query.ModelFields(singleEndModel, func(f *parser.FieldNode) bool {

			// It can't be a reverse relation field if it's not a hasOne relation field.
			if !query.IsHasOneModelField(asts, f) {
				return false
			}
			// It isn't a REVERSE relation field if despite it being a hasOne relation field,
			// it refers to a different model to that of the model to which the hasManyField belongs to.
			if f.Type.Value != hasManyF.belongsTo.Name.Value {
				return false
			}

			// If it is qualified - we don't count it.
			if query.FieldHasAttribute(f, parser.AttributeRelation) {
				return false
			}
			return true
		})

		// It is an error, if there are more than one un-qualified reverse fields.
		if len(reverseFields) > 1 {
			suggestedFields := lo.Map(reverseFields, func(f *parser.FieldNode, _ int) string {
				return f.Name.Value
			})
			errs.Append(
				errorhandling.ErrorAmbiguousRelationship,
				map[string]string{
					"ModelA":          singleEndModel.Name.Value,
					"ModelB":          reverseFields[0].Type.Value,
					"SuggestedFields": formatting.HumanizeList(suggestedFields, formatting.DelimiterAnd),
				},
				singleEndModel.Name,
			)
		}
	}

	return errs
}

// fieldsThatHaveRelationAttribute provides a list of all the fields in the given model,
// that have the @relation attribute.
func fieldsThatHaveRelationAttribute(model *parser.ModelNode) []*parser.FieldNode {
	allModelFields := query.ModelFields(model)
	thoseWithRelationAttr := lo.Filter(allModelFields, func(f *parser.FieldNode, _ int) bool {
		return query.FieldHasAttribute(f, parser.AttributeRelation)
	})
	return thoseWithRelationAttr
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

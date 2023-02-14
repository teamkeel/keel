package relationships

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

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

			otherModel := query.Model(asts, field.Type)

			if otherModel == nil {
				continue
			}

			otherModelFields := query.ModelFields(otherModel)

			for _, otherField := range otherModelFields {
				if otherField == field {
					continue
				}
				if otherField.Type != model.Name.Value {
					continue
				}

				// If either the field on model A is repeated
				// or the corresponding field on the other side is repeated
				// then we are not interested
				if otherField.Repeated {
					continue
				}

				errs.Append(
					errorhandling.ErrorInvalidOneToOneRelationship,
					map[string]string{
						"ModelA": model.Name.Value,
						"ModelB": field.Type,
					},
					field,
				)

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

			otherModel := query.Model(asts, field.Type)

			if otherModel == nil {
				continue
			}

			otherModelFields := query.ModelFields(otherModel)

			for _, otherField := range otherModelFields {
				if otherField.Type != model.Name.Value {
					continue
				}

				if !otherField.Repeated {
					continue fields
				}
			}

			errs.Append(
				errorhandling.ErrorMissingRelationshipField,
				map[string]string{
					"ModelA":     model.Name.Value,
					"ModelB":     field.Type,
					"Suggestion": fmt.Sprintf("%s %s", strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
				},
				field.Name,
			)

		}
	}

	return errs
}

func RelationAttributeRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		// init table of previous uses

		for _, thisField := range query.ModelFields(model) {

			// Only relevant if the field has @relation
			if !query.FieldHasAttribute(thisField, parser.AttributeRelation) {
				continue
			}

			// Make sure @relation only used on fields of type Model
			if !query.IsFieldOfTypeModel(asts, thisField.Name.Value) {
				errs.Append(
					errorhandling.ErrorRelationAttrOnWrongFieldType,
					map[string]string{
						"FieldName":  thisField.Name.Value,
						"WrongType":  thisField.Type,
						"Suggestion": thisField.Name.Value,
					},
					thisField)
				continue // Additional @relation checks for this field are meaningful in this case.
			}

			// must not be repeated
			// must be a name of a field in the related field [short circuit]
			// the related field must be type model [short circuit]
			// the related field must point to the originating field's model name [short circuit]
			// must not have been used thus previously [short circuit]
			// the related field must be multiple
			// update table of previous uses by this model
		}
	}
	return errs
}

// When ModelA has a HasMany relationship field that references ModelB, then it is invalid for
// ModelB to have more than one HasOne relation field that refers to ModelA.
//
// This is because we have no other way (at present) to infer which field of ModelB to use
// in the SQL generated associated with ModelA's field.
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

		singleEndModel := query.Model(asts, hasManyF.theField.Type)

		if singleEndModel == nil {
			// This can be the case for invalid schemas but other rules check for that.
			// For our purposes, it just means we can't proceed with this validation
			// rule right now, for this hasMany field.
			continue
		}

		// Given access to the model at the HasOne end, how many fields does it have that
		// refer back to the model at the hasMany end?
		reverseFields := query.ModelFields(singleEndModel, func(f *parser.FieldNode) bool {

			// It can't be a reverse relation field if it's not a hasOne relation field.
			if !query.IsHasOneModelField(asts, f) {
				return false
			}
			// It isn't a REVERSE relation field if despite it being a hasOne relation field,
			// it refers to a different model to that of the model to which the hasManyField belongs to.
			if f.Type != hasManyF.belongsTo.Name.Value {
				return false
			}
			return true
		})

		// It is an error, if there are more than one such reverse fields.
		if len(reverseFields) > 1 {
			errs.Append(
				errorhandling.ErrorAmbiguousRelationship,
				map[string]string{
					"ModelA": singleEndModel.Name.Value,
					"ModelB": reverseFields[0].Type,
				},
				singleEndModel,
			)
		}
	}

	return errs
}

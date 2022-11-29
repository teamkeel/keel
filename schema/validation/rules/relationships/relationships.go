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

// When ModelA has a HasMany relationship field that references ModelB, then it is invalid for
// ModelB to have more than one HasOne relation field that refers to ModelA.
//
// This is because we have no other way (at present) to infer which field of ModelB to use
// in the SQL generated associated with ModelA's field.
func MoreThanOneReverseMany(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	// First capture all the relation fields defined by the ASTs, which are of type HasMany.
	hasManyFields := []*parser.FieldNode{}
	for _, model := range query.Models(asts) {
		for _, f := range query.ModelFields(model) {
			if query.IsHasManyModelField(asts, f) {
				hasManyFields = append(hasManyFields, f)
			}
		}
	}

	// Now we iterate over all the captured HasMany relation fields in order to investigate
	// the model at the HasOne end of the relationship.
	for _, hasManyF := range hasManyFields {
		singleEndModel := query.Model(asts, hasManyF.Type)

		// Given access to the model at the HasOne end, how many fields does it have that
		// refer back to the model at the hasMany end?
		reverseFields := query.ModelFields(singleEndModel, func(f *parser.FieldNode) bool {
			return query.IsHasOneModelField(asts, f) && singleEndModel.Name.Value == hasManyF.Type
		})

		// It is an error, if there are more than one such reverse fields.
		if len(reverseFields) > 1 {
			errs.Append(
				errorhandling.ErrorNonSingularReverseField,
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

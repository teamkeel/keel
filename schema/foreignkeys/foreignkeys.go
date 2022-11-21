package foreignkeys

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// ForeignKeyInfo encapsulates everything relevant to the foreign key fields we auto
// generate into the AST for some models. The generated fields are derived from an "Owning"
// field (of type Model) defined explicitly in the keel schema, and with topology HasOne.
type ForeignKeyInfo struct {
	OwningModel               *parser.ModelNode
	OwningField               *parser.FieldNode // A field in the OwningModel that is of type MODEL, and topology HasOne.
	OwningFieldIsOptional     bool
	ReferredToModel           *parser.ModelNode
	ReferredToModelPrimaryKey *parser.FieldNode // Which field in the ReferredToModel is its Primary Key
	ForeignKeyName            string            // What name to give the generated FK.
}

// PrimaryKeys tells you the primary key field for any given model name.
type PrimaryKeys map[string]*parser.FieldNode

// NewPrimaryKeys provides a PrimaryKeys map for each model present in
// the given asts. Fields marked explicitly as primary keys take precedence,
// or it defaults to the field named "id".
func NewPrimaryKeys(asts []*parser.AST) PrimaryKeys {
	pkeys := PrimaryKeys{}
	for _, model := range query.Models(asts) {
		potentialFields := query.ModelFields(model)

		var pkField *parser.FieldNode
		for _, field := range potentialFields {
			switch {
			case query.FieldHasAttribute(field, parser.AttributePrimaryKey):
				pkField = field
			case pkField == nil && field.Name.Value == parser.ImplicitFieldNameId:
				pkField = field
			}
			pkeys[model.Name.Value] = pkField
		}
	}
	return pkeys
}

// NewForeignKeyInfo builds a list of ForeignKeyInfo that defines all the foreign key
// fields that should be auto-injected into the AST.
func NewForeignKeyInfo(asts []*parser.AST, primaryKeyMap PrimaryKeys) ([]*ForeignKeyInfo, *errorhandling.ErrorDetails) {
	fks := []*ForeignKeyInfo{}
	for _, mdl := range query.Models(asts) {
		for _, field := range query.ModelFields(mdl) {

			if !isHasOneModelField(asts, field) {
				continue
			}

			// The generated foreign key optionality follows that of the owning field.
			owningFieldIsOptional := field.Optional

			referredToModelName := strcase.ToCamel(field.Type)
			referredToModel := query.Model(asts, referredToModelName)

			if referredToModel == nil {
				errDetails := &errorhandling.ErrorDetails{
					Message: fmt.Sprintf("cannot find the model referred to (%s) by field %s, on model %s",
						referredToModelName, field.Name, mdl.Name),
					ShortMessage: "cannot find model referenced by relationship field",
					Hint:         "make sure you declare this model",
				}
				return nil, errDetails
			}

			referredToModelPKField, ok := primaryKeyMap[referredToModelName]
			if !ok {
				errDetails := &errorhandling.ErrorDetails{
					Message:      fmt.Sprintf("cannot find the primary key field on model: %s (internal error)", referredToModelName),
					ShortMessage: "cannot find primary key field",
					Hint:         "(internal error)",
				}
				return nil, errDetails
			}
			// This is the single source of truth for how we name foreign key fields.
			generatedForeignKeyName := field.Name.Value + strcase.ToCamel(referredToModelPKField.Name.Value)

			fkInfo := &ForeignKeyInfo{
				OwningModel:               mdl,
				OwningField:               field,
				OwningFieldIsOptional:     owningFieldIsOptional,
				ReferredToModel:           referredToModel,
				ReferredToModelPrimaryKey: referredToModelPKField,
				ForeignKeyName:            generatedForeignKeyName,
			}
			fks = append(fks, fkInfo)
		}
	}
	return fks, nil
}

// isHasOneModelField returns true if the given field can be inferred to be
// a field that references another model, and is not denoted as being repeated.
func isHasOneModelField(asts []*parser.AST, field *parser.FieldNode) bool {
	switch {
	case !query.IsModel(asts, field.Type):
		return false
	case field.Repeated:
		return false
	default:
		return true
	}
}

// IsModelFieldWithSiblingFK consults the given foreign key information to tell you if
// the given field name, in the context of the given model name,
// is an "owning" field of type Model, which has a corresponding (sibling) FK field.
func IsModelFieldWithSiblingFK(fkInfos []*ForeignKeyInfo, modelName string, fieldName string) bool {
	for _, fkInfo := range fkInfos {
		if fkInfo.OwningModel.Name.Value != modelName {
			continue
		}
		if fkInfo.OwningField.Name.Value != fieldName {
			continue
		}
		return true
	}
	return false
}

// IsFkField consults the given foreign key information to tell you if
// the given field name, in the context of the given model name,
// is a foreign key field associated with a sibling Owning field.
// When so, it also tells you the dotted-form alternative name.
// E.g. for AuthorId the dotted-form is "author.id".
func IsFkField(fkInfos []*ForeignKeyInfo, modelName string, fieldName string) (dottedForm string, isFKField bool) {
	for _, fkInfo := range fkInfos {
		if fkInfo.OwningModel.Name.Value != modelName {
			continue
		}
		if fkInfo.ForeignKeyName == fieldName {
			dottedForm := strings.Join([]string{
				fkInfo.OwningField.Name.Value,
				fkInfo.ReferredToModelPrimaryKey.Name.Value}, ".")

			return dottedForm, true
		}
	}
	return "", false
}

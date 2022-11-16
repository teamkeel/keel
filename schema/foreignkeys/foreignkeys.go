package foreignkeys

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

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

// ForeignKeyInfo encapsulates everything relevant to the foreign key fields we auto
// generate into the AST for some models. The generated fields are derived from an "Owning"
// field (of type Model) defined explicitly in the keel schema, and with topology HasOne.
type ForeignKeyInfo struct {
	OwningModel               *parser.ModelNode
	OwningField               *parser.FieldNode // A field in the OwningModel that is of type MODEL, and topology HasOne.
	ReferredToModel           *parser.ModelNode
	ReferredToModelPrimaryKey *parser.FieldNode // Which field in the ReferredToModel is its Primary Key
	ForeignKeyName            string            // What name to give the generated FK.
}

// NewForeignKeyInfo builds a list of ForeignKeyInfo that defines all the foreign key
// fields that should be auto-injected into the AST.
func NewForeignKeyInfo(asts []*parser.AST, primaryKeyMap PrimaryKeys) []*ForeignKeyInfo {
	fks := []*ForeignKeyInfo{}
	for _, mdl := range query.Models(asts) {
		for _, field := range query.ModelFields(mdl) {

			if !isHasOneModelField(asts, field) {
				continue
			}

			referredToModelName := strcase.ToCamel(field.Type)
			referredToModel := query.Model(asts, referredToModelName)
			if referredToModel == nil {
				// todo: proper error handling
				panic("XXXX failed to lookup model")
			}
			referredToModelPKField, ok := primaryKeyMap[referredToModelName]
			// todo: correct error handling
			if !ok {
				panic("XXXX failed to look up PK")
			}
			// This is the single source of truth for how we name foreign key fields.
			generatedForeignKeyName := field.Name.Value + strcase.ToCamel(referredToModelPKField.Name.Value)

			fkInfo := &ForeignKeyInfo{
				OwningModel:               mdl,
				OwningField:               field,
				ReferredToModel:           referredToModel,
				ReferredToModelPrimaryKey: referredToModelPKField,
				ForeignKeyName:            generatedForeignKeyName,
			}
			fks = append(fks, fkInfo)
		}
	}
	return fks
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

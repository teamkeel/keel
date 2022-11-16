package foreignkeys

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

// PrimaryKeys tells you the primary key field name for any given model name.
// E.g. primaryKeys["Post"] = "id".
type PrimaryKeys map[string]string

// NewPrimaryKeys provides a PrimaryKeys map for each model present in
// the given asts. Fields marked explicitly as primary keys take precedence,
// or it defaults to the field named "id".
func NewPrimaryKeys(asts []*parser.AST) PrimaryKeys {
	pkeys := PrimaryKeys{}
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			switch {
			case query.FieldHasAttribute(field, parser.AttributePrimaryKey):
				pkeys[model.Name.Value] = field.Name.Value
			default:
				pkeys[model.Name.Value] = parser.ImplicitFieldNameId
			}
		}
	}
	return pkeys
}

// ForeignKeys tells you the names of the implicit foreign key field names
// for any given model.
// E.g. foreignKeys["Post"] = []string{"authorId", "reviewerId"}
type ForeignKeys map[string][]string

// NewForeignKeys provides a map in which you can look up any model name,
// and it will tell you the foreign key field names it NEEDS to have.
// E.g. a Post model might need FK fields called "AuthorID", and maybe "ReviewerID" also.
func NewForeignKeys(asts []*parser.AST, primaryKeyMap PrimaryKeys) ForeignKeys {
	fkMap := ForeignKeys{}
	for _, thisModel := range query.Models(asts) {
		for _, thisModelField := range query.ModelFields(thisModel) {

			isHasOneRelation := query.IsModel(asts, thisModelField.Type) && !thisModelField.Repeated
			if !isHasOneRelation {
				continue
			}
			targetModelName := strcase.ToCamel(thisModelField.Type)
			referredToModelPK, ok := primaryKeyMap[targetModelName]
			// todo: correct error handling
			if !ok {
				panic("XXXX failed to look up PK")
			}
			// todo - need to fix up casing
			newFk := thisModelField.Name.Value + strcase.ToCamel(referredToModelPK)
			existingFks, ok := fkMap[thisModel.Name.Value]
			if ok {
				fkMap[thisModel.Name.Value] = append(existingFks, newFk)
			} else {
				fkMap[thisModel.Name.Value] = []string{newFk}
			}
		}
	}
	return fkMap
}

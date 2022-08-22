package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/sanity-io/litter"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

// List implements a Keel List Action.
// In quick overview this means generating a SQL query
// based on the List operation's Inputs and Where clause,
// running that query, and returning the results.
func List(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema,
	inputs *ListInput) (interface{}, error) {

	litter.Dump("XXXX operation passed to List action:")
	litter.Dump(operation)

	litter.Dump("XXXX inputs passed to List action:")
	litter.Dump(inputs)

	db, err := runtimectx.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	model := proto.FindModel(schema.Models, operation.ModelName)

	tableName := strcase.ToSnake(model.Name)

	// Initialise a query on the table = to which we'll add Where clauses.
	tx := db.Table(tableName)

	// Add the WHERE clauses derived from the inputs.
	tx, err = addListInputFilters(operation, inputs, tx)
	if err != nil {
		return nil, err
	}

	// todo
	// Add the WHERE clauses derived from EXPLICIT inputs (i.e. the operation's where clauses).
	// tx, err = addWhereFilters(operation, schema, args, tx)
	// if err != nil {
	// 	return nil, err
	// }

	litter.Dump("XXXX gorm generated for List action:")
	litter.Dump(tx)

	// Todo: should we validate the type of the values?, or let postgres object to them later?

	// Execute the SQL query.
	result := []map[string]any{}
	tx = tx.Find(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}
	res := toLowerCamelMaps(result)

	litter.Dump("XXXX Response being returned from List action:")
	litter.Dump(res)

	return res, nil
}

// addListInputFilters adds Where clauses to the given gorm.DB corresponding to the
// given ListInput.
func addListInputFilters(op *proto.Operation, listInput *ListInput, tx *gorm.DB) (*gorm.DB, error) {
	// We'll look at each of the fields specified as inputs by the operation in the schema,
	// and then try to find these referenced by the where filters in the given ListInput.
	for _, schemaInput := range op.Inputs {
		if schemaInput.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			return nil, errors.New("not yet supported: explicit inputs for list actions")
		}
		expectedFieldName := schemaInput.Target[0]
		var matchingWhere *Where
		for _, where := range listInput.Wheres {
			if where.Name == expectedFieldName {
				matchingWhere = where
				break
			}
		}
		if matchingWhere == nil {
			return nil, fmt.Errorf("operation expects an input named: <%s>, but none is present on the request", expectedFieldName)
		}
		switch schemaInput.Type.Type {
		case proto.Type_TYPE_STRING:
			tx, err := addStringQuery(tx, expectedFieldName, matchingWhere.StringQuery)
			if err != nil {
				return nil, err
			}
			return tx, nil
		default:
			return nil, errors.New("so far, only string field types are supported for List operation inputs")
		}
	}
	return tx, nil
}

// addStringQuery updates the given gorm.DB tx with a where clause that represents the given
// StringQuery.
func addStringQuery(tx *gorm.DB, columnName string, stringQry *StringQuery) (*gorm.DB, error) {
	switch stringQry.Operator {
	case OperatorEquals:
		stringOperand, ok := stringQry.Operand.(string)
		if !ok {
			return nil, fmt.Errorf("operand cannot be cast to string: %v", stringQry.Operand)
		}
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))
		tx := tx.Where(w, stringOperand)
		return tx, nil
	default:
		return nil, fmt.Errorf("this StringQuery.Operator is not yet supported: %v", stringQry.Operator)
	}
}

package actions

import (
	"context"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

type QueryBuilder struct {
	// inputs
	Ctx       context.Context
	Schema    *proto.Schema
	Operation *proto.Operation
	Args      map[string]any

	tx *gorm.DB

	// computed
	Constraints []*Constraint
	Values      []*Value
}

func NewQueryBuilder(ctx context.Context, schema *proto.Schema, operation *proto.Operation, args map[string]any) (*QueryBuilder, error) {
	db, err := runtimectx.GetDatabase(ctx)

	if err != nil {
		return nil, err
	}

	tableName := strcase.ToSnake(operation.ModelName)
	tx := db.Table(tableName)

	qb := &QueryBuilder{
		Ctx:       ctx,
		Schema:    schema,
		Operation: operation,
		Args:      args,
		tx:        tx,
	}

	if qb.shouldBuildConstraints() {
		qb.buildConstraints()
	}

	if qb.shouldBuildValues() {
		qb.buildValues()
	}

	return qb, nil
}

func (q *QueryBuilder) model() *proto.Model {
	return proto.FindModel(q.Schema.Models, q.Operation.ModelName)
}

func (q *QueryBuilder) shouldBuildConstraints() bool {
	return lo.SomeBy(q.Operation.Inputs, func(i *proto.OperationInput) bool {
		return i.Mode == proto.InputMode_INPUT_MODE_READ
	})
}

func (q *QueryBuilder) shouldBuildValues() bool {
	return lo.SomeBy(q.Operation.Inputs, func(i *proto.OperationInput) bool {
		return i.Mode == proto.InputMode_INPUT_MODE_WRITE
	})
}

func (q *QueryBuilder) buildConstraints() {
	args := q.Args

	switch q.Operation.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		// steps:
		// 1. validate: check arg names match either an implicit input or explicit input and is used in a set
		// 2. build up implicit inputs from args
		// 3. do we support explicit inputs for gets?
	case proto.OperationType_OPERATION_TYPE_LIST:
		// steps:
		// 1. validate: check arg names match  either an implicit input or explicit input and is used in a set
		// 2. build implicit inputs from args - args will use query fluid api
		// 3. build explicit inputs - these will override any implicit
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		// steps:
		// 1. validate args match either implicit or explicit
		// 2. build implicit
		// 3. build explicit
	case proto.OperationType_OPERATION_TYPE_DELETE:
		// 1. validate args contain only unique matching fields
		// 2. build up constrains from args
	}
}

func (q *QueryBuilder) Execute() error {
	switch q.Operation.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		q.tx.Create(q.Values)
	}
}

func (q *QueryBuilder) buildValues() {
	args := q.Args

	switch q.Operation.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		// build implicit input values first
		// then overrwrite with any clashing set attributes

		for _, input := range q.Operation.Inputs {
			if input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_EXPLICIT {
				continue
			}

			matchingValue, found := args[input.Name]

			if found {
				q.Values = append(q.Values, &Value{
					Value:  matchingValue,
					Source: input,
					Field:  proto.FindField(q.Schema.Models, input.ModelName, input.Target[0]),
				})
			}
		}
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		// build implicit input values first
		// then overwrite any clashing set attributes
		q.tx.Updates(args)
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		// build email / password hash
	}
}

func (q *QueryBuilder) ToQuery() *gorm.DB {

}

type Source = *proto.OperationInput

type Constraint struct {
	Source   Source
	Operator Operator
	Field    *proto.Field
	Value    Value
}

func (c *Constraint) ToGorm() *gorm.DB {

}

type Value struct {
	Source Source

	Field *proto.Field
	Value any
}

func (v *Value) ToMap() map[string]any {
	// call toMap

	toLowerCamelMaps(toMap())
}

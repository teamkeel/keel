package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

func Use(ctx context.Context, operation *proto.Operation, schema *proto.Schema, args map[string]any) {

	var builder CreateAction
	result, _ := builder.
		Instantiate(ctx, schema, nil).
		ApplyImplicitInputs(args).
		ApplySets(args).
		ApplyFilters(args).
		IsAuthorised(args).
		Execute()

	fmt.Println(result)
}

// the action API builds:
// - a model object used for writing data, and
// - a database object used for querying data.
type ActionBuilder interface {
	// instantiate action scope.
	Instantiate(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder
	// for each implicit input on write operation, store matching argument in underlying "Values" data structure.
	ApplyImplicitInputs(args Arguments) ActionBuilder
	// for each @set attribute on the write operation, store matching argument in the underlying "Values" data structure (for intended field).
	ApplySets(args Arguments) ActionBuilder
	// 1. for each implicit input on read operation, use matching argument to add filter clause to database object.
	// 2. for each @where attribute on the read/write operation, use matching argument to add filter clause to database object.
	ApplyFilters(args Arguments) ActionBuilder
	// use the current database query scope to perform an authorisation check on the data filter.
	// use explicit inputs where ne
	IsAuthorised(args Arguments) ActionBuilder
	// execute database query and build result.
	Execute() (*Result, error)
}

type Scope struct {
	// instantiated with context
	context   context.Context
	operation *proto.Operation
	model     *proto.Model
	schema    *proto.Schema
	table     string

	// instantiated to database
	// amended with ParseFilters as defined in each action
	// used to check authorisation using current query scope
	// used to execute action outcome using current query scope
	query *gorm.DB

	// instantiated to {}
	// modified with ParseValues and ApplySets
	values *Values
}

type Arguments map[string]any

type Values map[string]any

type Result map[string]any

type Action struct {
	Scope
}

func (c *Action) Instantiate(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder {
	c.context = ctx
	c.schema = schema
	c.operation = operation
	c.model = proto.FindModel(schema.Models, operation.ModelName)
	c.query, _ = runtimectx.GetDatabase(ctx)
	c.table = strcase.ToSnake(c.model.Name)
	c.values = &Values{}

	return c
}

func (c *Action) ApplyImplicitInputs(args Arguments) ActionBuilder {
	return c
}

func (c *Action) ApplySets(args Arguments) ActionBuilder {
	return c
}

func (c *Action) ApplyFilters(args Arguments) ActionBuilder {
	return c
}

func (c *Action) IsAuthorised(args Arguments) ActionBuilder {
	return c
}

func (c *Action) Execute() (*Result, error) {
	return &Result{}, nil
}

// ================= CREATE ACTION

type CreateAction struct {
	Action
}

func (action *CreateAction) Instantiate(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder {
	action.Instantiate(ctx, schema, operation)

	//action.Values = &Values{}

	return action
}

func (action *CreateAction) ApplyFilters(args Arguments) ActionBuilder {
	action.query, _ = addGetImplicitInputFilters(action.operation, args, action.query)
	action.query, _ = addGetExplicitInputFilters(action.operation, action.schema, args, action.query)
	return action
}

func (c *CreateAction) IsAuthorised(args Arguments) ActionBuilder {
	return c
}

func (c *CreateAction) Execute() (*Result, error) {
	// 1. create initial model
	// 2. apply sets
	return &Result{}, nil
}

// ================= GET ACTION

type GetAction struct {
	Action
}

func (action *GetAction) Instantiate(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder {
	return action
}

func (action *GetAction) ApplyImplicitInputs(args Arguments) ActionBuilder {
	return action
}

func (action *GetAction) ApplyFilters(args Arguments) ActionBuilder {
	return action
}

func (action *GetAction) IsAuthorised(args Arguments) ActionBuilder {
	return action
}

func (action *GetAction) Execute() (*Result, error) {
	result := []map[string]any{}
	action.query = action.query.Find(&result)

	if action.query.Error != nil {
		return nil, action.query.Error
	}
	n := len(result)
	if n == 0 {
		return nil, errors.New("no records found for Get() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}

	var resultMap Result
	resultMap = toLowerCamelMap(result[0])
	return &resultMap, nil
}

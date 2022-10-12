package actions

import (
	"context"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

// what are we trying to achieve by drying up the action package?
// - create an intuitive and reusable API for use across actions.
// - standardise the steps across each of the actions.
// - reusing similar logic across actions. for e.g., unifying how we parse arguments
// - simpler / less code

// how might this work against us?
// - we corner ourselves into a structure which isn't flexible enough
// - overcomplicated polymorphic solution
// - if it ain't broke, don't fix it. so why waste energy on this?

type Arguments map[string]any

type Values map[string]any

type Result map[string]any

// the ActionBuilder API fluidly constructs
// - a model object used for writing data, and
// - a database query used for querying data.
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

	// execute database query and return action-specific result.
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

type Action struct {
	Scope
}

func (action *Action) Instantiate(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder {
	action.context = ctx
	action.schema = schema
	action.operation = operation
	action.model = proto.FindModel(schema.Models, operation.ModelName)
	action.query, _ = runtimectx.GetDatabase(ctx)
	action.table = strcase.ToSnake(action.model.Name)
	action.values = &Values{}

	return action
}

func (action *Action) ApplyImplicitInputs(args Arguments) ActionBuilder {
	// todo: Default implementation for all actions types
	return action
}

func (action *Action) ApplySets(args Arguments) ActionBuilder {
	// todo: Default implementation for all actions types
	return action
}

func (action *Action) ApplyFilters(args Arguments) ActionBuilder {
	// todo: Default implementation for all actions types
	return action
}

func (action *Action) IsAuthorised(args Arguments) ActionBuilder {
	// todo: default implementation for all actions types
	return action
}

func (action *Action) Execute() (*Result, error) {
	// todo: would we ever want a default implementation or should we panic?
	return &Result{}, nil
}

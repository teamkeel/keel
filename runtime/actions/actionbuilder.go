package actions

import (
	"context"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

// what are we trying to achieve by drying up the action package?
// We hope to exploit the usual, well understood benefits of DRY code as follows:
// - provide a standardised way for Action implementation functions to
//   be coded - by providing a Go interface that provides method signatures and types for
//   the principal steps involved. These aim to help identify and separate the main high-level
//   processing steps.
// - enforcing the use of the standardised approach by making a single entry point function
//   with a signature that uses said interface.
// - replacing the casual maps we have been using for: inputs/args, queries, db records and results with
//   specific dedicated types for each context.
// - standardizing the way we build up db queries and hold their state (*gorm.DB) objects *across* the
//   main steps that wish to get involved with the *gorm.DB query.
//
// how might this work against us?
// - we corner ourselves into a structure which isn't flexible enough
// - we might discover that the problem is simply not as polymorphic as we think it is
// - if it ain't broke, don't fix it. so why waste energy on this?

// RequestArguments are input values that are provided by an incoming request. Keys are model field names
// in the case of implicit inputs, or the alias name defined in the schema in the case of explicit inputs.
type RequestArguments map[string]any

// DbValues hold the in-memory representation of a record we are going to *Write* to a database row.
// Keys are strictly model field names. (I.e. something must intervene to snake-case it before passing it on to
// a gorm.DB.Create() for example).
type DbValues map[string]any

// An ActionResult is the object returned to the caller for any of the Action functions.
// Keys are strictly model field names.
type ActionResult map[string]any

// The ActionBuilder interface governs a contract that must be used to instantiate, build-up,
// and execute any Action.

type ActionBuilder interface {

	// Initialise implementations must retain access to the given Scope - because it is the way that
	// state is shared between the interface methods. For example it contains a *gorm.DB that some of the
	// methods incrementally update.
	Initialise(scope *Scope) ActionBuilder

	// CaptureImplicitWriteInputValues implementations are expected to identify implicit
	// Action *write* input key/values in the given args, and update the the dbValues in the shared Scope
	// object accordingly.
	CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder

	// CaptureSetValues implementations are expected to reconcile the @Set expressions defined for this Action
	// by the schema with the key/values provided by the given args, and to populate the *DBValues in the
	// shared Scope accordingly.
	CaptureSetValues(args RequestArguments) ActionBuilder

	// ApplyImplicitFilters implementations are expected to reconcile the implicit read inputs defined for
	// this Action by the schema with the key/values provided by the given args, and to add corresponding
	// Where filters to the *gorm.DB in the shared Scope.
	ApplyImplicitFilters(args RequestArguments) ActionBuilder

	// ApplyExplicitFilters implementations are expected to reconcile the explicit read inputs defined for
	// this Action by the schema with the key/values provided by the given args, and to add corresponding
	// Where filters to the *gorm.DB in the shared Scope.
	ApplyExplicitFilters(args RequestArguments) ActionBuilder

	// ????? don't understand this one yet, ...
	// use the current database query scope to perform an authorisation check on the data filter.
	// use explicit inputs where ne
	IsAuthorised(args RequestArguments) ActionBuilder

	// Execute database query and return action-specific result.
	Execute() (*ActionResult, error)
}

// A Scope provides a shared single source of truth to support Action implementation code,
// plus some shared state that the ActionBuilder can update or otherwise use. For example
// the values that will be written to a database row, or the *gorm.DB that the methods will
// incrementally add to.
type Scope struct {
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
	dbValues DbValues
}

func NewScope(
	context context.Context,
	operation *proto.Operation,
	model *proto.Model,
	schema *proto.Schema,
	table string,
	query *gorm.DB) *Scope {
	return &Scope{
		context:   context,
		operation: operation,
		model:     model,
		schema:    schema,
		table:     table,
		query:     query,
		dbValues:  DbValues{},
	}
}

type Action struct {
	Scope
}

func (action *Action) Initialise(ctx context.Context, schema *proto.Schema, operation *proto.Operation) ActionBuilder {
	action.context = ctx
	action.schema = schema
	action.operation = operation
	action.model = proto.FindModel(schema.Models, operation.ModelName)
	action.query, _ = runtimectx.GetDatabase(ctx)
	action.table = strcase.ToSnake(action.model.Name)
	action.values = &DbValues{}

	return action
}

func (action *Action) ApplyImplicitInputs(args RequestArguments) ActionBuilder {
	// todo: Default implementation for all actions types
	return action
}

func (action *Action) ApplySets(args RequestArguments) ActionBuilder {
	// todo: Default implementation for all actions types
	return action
}

func (action *Action) ApplyFilters(args RequestArguments) ActionBuilder {
	// todo: Default implementation for all actions types
	return action
}

func (action *Action) IsAuthorised(args RequestArguments) ActionBuilder {
	// todo: default implementation for all actions types
	return action
}

func (action *Action) Execute() (*ActionResult, error) {
	// todo: would we ever want a default implementation or should we panic?
	return &ActionResult{}, nil
}

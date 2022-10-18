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

// Values hold the in-memory representation of a record we are going to *Write* to a database row.
// Keys are strictly model field names. (I.e. something must intervene to snake-case it before passing it on to
// a gorm.DB.Create() for example).
type WriteValues map[string]any

// An ActionResult contains the return data for an Action using generics, so that we can create an interface
// that depends ...:
type ActionResult[T any] struct {
	Value T
}

// The ActionBuilder interface governs a contract that must be used to instantiate, build-up,
// and execute any Action.
// All the following methods share a Scope object in which to accumulate query clauses and values that which
// be written to a database row and an error that has been detected. The implementation of every method below
// must short-circuit return if error is not nil and similarly set error if they encounter an error, and return.
type ActionBuilder[Result any] interface {

	// Initialise implementations must retain access to the given Scope - because it is the way that
	// state is shared between the interface methods. For example it contains a *gorm.DB that some of the
	// methods incrementally update.
	Initialise(scope *Scope) ActionBuilder[Result]

	// CaptureImplicitWriteInputValues implementations are expected to identify implicit
	// Action *write* input key/values in the given args, and update the the dbValues in the shared Scope
	// object accordingly.
	CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[Result]

	// CaptureSetValues implementations are expected to reconcile the @Set expressions defined for this Action
	// by the schema with the key/values provided by the given args, and to populate the *DBValues in the
	// shared Scope accordingly.
	CaptureSetValues(args RequestArguments) ActionBuilder[Result]

	// ApplyImplicitFilters implementations are expected to reconcile the implicit read inputs defined for
	// this Action by the schema with the key/values provided by the given args, and to add corresponding
	// Where filters to the *gorm.DB in the shared Scope.
	ApplyImplicitFilters(args RequestArguments) ActionBuilder[Result]

	// ApplyExplicitFilters implementations are expected to reconcile the explicit read inputs defined for
	// this Action by the schema with the key/values provided by the given args, and to add corresponding
	// Where filters to the *gorm.DB in the shared Scope.
	ApplyExplicitFilters(args RequestArguments) ActionBuilder[Result]

	// ????? don't understand this one yet, ...
	// use the current database query scope to perform an authorisation check on the data filter.
	// use explicit inputs where ne
	IsAuthorised(args RequestArguments) ActionBuilder[Result]

	// Execute database query and return action-specific result.
	Execute(args RequestArguments) (*ActionResult[Result], error)
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
	writeValues WriteValues

	Error error
}

func NewScope(
	ctx context.Context,
	operation *proto.Operation,
	schema *proto.Schema) (*Scope, error) {

	model := proto.FindModel(schema.Models, operation.ModelName)
	table := strcase.ToSnake(model.Name)
	query, err := runtimectx.GetDatabase(ctx)

	if err != nil {
		return nil, err
	}

	return &Scope{
		context:     ctx,
		operation:   operation,
		model:       model,
		schema:      schema,
		table:       table,
		query:       query,
		writeValues: WriteValues{},
	}, nil
}

// toLowerCamelMap returns a copy of the given map, in which all
// of the key strings are converted to LowerCamelCase.
// It is good for converting identifiers typically used as database
// table or column names, to the case requirements stipulated by the Keel schema.
func toLowerCamelMap(m map[string]any) map[string]any {
	res := map[string]any{}
	for key, value := range m {
		res[strcase.ToLowerCamel(key)] = value
	}
	return res
}

// toLowerCamelMaps is a convenience wrapper around toLowerCamelMap
// that operates on a list of input maps - rather than just a single map.
func toLowerCamelMaps(maps []map[string]any) []map[string]any {
	res := []map[string]any{}
	for _, m := range maps {
		res = append(res, toLowerCamelMap(m))
	}
	return res
}

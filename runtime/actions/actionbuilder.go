package actions

import (
	"context"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"gorm.io/gorm"
)

// RequestArguments are input values that are provided by an incoming request. Keys are model field names
// in the case of implicit inputs, or the alias name defined in the schema in the case of explicit inputs.
type RequestArguments map[string]any

// WriteValues hold the in-memory representation of a record we are going to *Write* to a database row.
// Keys are strictly model field names. (I.e. something must intervene to snake-case it before passing it on to
// a gorm.DB.Create() for example).
type WriteValues map[string]any

// An ActionResult is a parameterised Type that allows each of the specific Actions {Get,Create,List...} to define
// their own return type structure. E.g. for the List action - it can return paging information as well as
// the records in a strongly typed way.
type ActionResult[T any] struct {
	Value T
}

// The ActionBuilder interface governs a contract that must be used to instantiate, build-up,
// and execute any Action.
// All the following methods share a Scope object in which to accumulate query clauses and values that which
// be written to a database row and an error that has been detected. The implementation of every method below
// must short-circuit-return if error is not nil and similarly set error if they encounter an error, and return.
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

	// This field is connected to the database, and we use it to perform all
	// all queries and write operations on the database.
	query *gorm.DB

	// This field accumulates the values we intend to write to a database row.
	writeValues WriteValues

	// The Error field holds the current error if there is one.
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

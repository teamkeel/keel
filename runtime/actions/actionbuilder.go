package actions

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/parser"
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

// An ActionResult is the object returned to the caller for any of the Action functions.
// Keys are strictly model field names.
type ActionResult map[string]any

// The ActionBuilder interface governs a contract that must be used to instantiate, build-up,
// and execute any Action.
// All the following methods share a Scope object in which to accumulate query clauses and values that which
// be written to a database row and an error that has been detected. The implementation of every method below
// must short-circuit return if error is not nil and similarly set error if they encounter an error, and return.
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
	writeValues WriteValues

	curError error
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

type Action struct {
	*Scope
}

func (action *Action) Initialise(scope *Scope) ActionBuilder {
	action.Scope = scope

	return action
}

func (action *Action) WithError(err error) ActionBuilder {
	action.Scope.curError = err
	return action
}

func (action *Action) HasError() bool {
	return action.Scope.curError != nil
}

func (action *Action) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder {
	if action.HasError() {
		return action
	}

	// values, ok := args.(map[string]any)

	// if !ok {
	// 	return action.WithError(errors.New("values not in correct format"))
	// }

	for _, input := range action.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		if input.Mode != proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			continue
		}

		action.Scope.writeValues[fieldName] = value
	}

	return action
}

// Given an input, operator and value, this method will add a where constraint to the current
// query scope for the implicit filtering API.
// e.g operator is 'greaterThan' and value is 1, with the input being targeted to a field 'rating',
// the scope.query variable will have the following new SQL constraint added to it:
// (..existing query..) AND rating > 1
func (action *Action) addImplicitFilter(input *proto.OperationInput, operator Operator, value any) ActionBuilder {
	if action.HasError() {
		return action
	}

	inputType := input.Type.Type

	columnName := input.Target[0]

	switch operator {
	case OperatorEquals:
		w := fmt.Sprintf("%s = ?", strcase.ToSnake(columnName))

		if inputType == proto.Type_TYPE_DATE || inputType == proto.Type_TYPE_DATETIME || inputType == proto.Type_TYPE_TIMESTAMP {
			time, err := parseTimeOperand(value, inputType)

			if err != nil {
				return action.WithError(err)
			}

			action.query = action.query.Where(w, time)
		} else {
			action.query = action.query.Where(w, value)
		}
	case OperatorStartsWith:
		operandStr, ok := value.(string)

		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to a string", value))
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandStr+"%%")
	case OperatorEndsWith:
		operandStr, ok := value.(string)

		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to a string", value))
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, "%%"+operandStr)
	case OperatorContains:
		operandStr, ok := value.(string)
		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to a string", value))
		}

		w := fmt.Sprintf("%s LIKE ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, "%%"+operandStr+"%%")
	case OperatorOneOf:
		operandStrings, ok := value.([]interface{})
		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to a []interface{}", value))
		}

		w := fmt.Sprintf("%s in ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandStrings)
	case OperatorLessThan:
		operandInt, ok := value.(int)

		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to an int", value))
		}

		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandInt)
	case OperatorLessThanEquals:
		operandInt, ok := value.(int)

		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to an int", value))
		}

		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandInt)
	case OperatorGreaterThan:
		operandInt, ok := value.(int)

		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to an int", value))
		}

		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandInt)
	case OperatorGreaterThanEquals:
		operandInt, ok := value.(int)

		if !ok {
			return action.WithError(fmt.Errorf("cannot cast this: %v to an int", value))
		}

		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))

		action.query = action.query.Where(w, operandInt)
	case OperatorBefore:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return action.WithError(err)
		}

		w := fmt.Sprintf("%s < ?", strcase.ToSnake(columnName))

		action.query = action.query.Where(w, operandTime)
	case OperatorAfter:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return action.WithError(err)
		}

		w := fmt.Sprintf("%s > ?", strcase.ToSnake(columnName))

		action.query = action.query.Where(w, operandTime)
	case OperatorOnOrBefore:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return action.WithError(err)
		}

		w := fmt.Sprintf("%s <= ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandTime)
	case OperatorOnOrAfter:
		operandTime, err := parseTimeOperand(value, inputType)

		if err != nil {
			return action.WithError(err)
		}

		w := fmt.Sprintf("%s >= ?", strcase.ToSnake(columnName))
		action.query = action.query.Where(w, operandTime)
	default:
		return action.WithError(fmt.Errorf("operator: %v is not yet supported", operator))
	}

	return action
}

func (action *Action) CaptureSetValues(args RequestArguments) ActionBuilder {
	if action.HasError() {
		return action
	}

	ctx := action.Scope.context
	operation := action.operation
	schema := action.schema

	for _, setExpression := range operation.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return action.WithError(err)
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return action.WithError(err)
		}

		lhsOperandType, err := GetOperandType(assignment.LHS, operation, schema)
		if err != nil {
			return action.WithError(err)
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment

		action.Scope.writeValues[fieldName], err = evaluateOperandValue(ctx, assignment.RHS, operation, schema, args, lhsOperandType)
		if err != nil {
			return action.WithError(err)
		}
	}
	return action
}

func (action *Action) ApplyImplicitFilters(args RequestArguments) ActionBuilder {
	panic("concrete implementation required")

	// get(id, code, unqiueField): construct where from each field input (as equality) (point query)
	// update(): construct where from each inputs as equality (range query)
	// list(): construct where from many inputs with user-specified operators

	// todo: Default implementation for all actions types
	// { }
	return action
}

func (action *Action) ApplyExplicitFilters(args RequestArguments) ActionBuilder {
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

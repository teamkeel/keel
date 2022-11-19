package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

// Represents a model field.
func Field(field string) *QueryOperand {
	return &QueryOperand{
		column: strcase.ToSnake(field),
	}
}

// Represents a model field.
func ModelField(model string, field string) *QueryOperand {
	return &QueryOperand{
		table:  strcase.ToSnake(model),
		column: strcase.ToSnake(field),
	}
}

// Represents a value operand.
func Value(value any) *QueryOperand {
	return &QueryOperand{value: value}
}

// Represents a null value operand.
func Null() *QueryOperand {
	return &QueryOperand{}
}

type QueryOperand struct {
	table  string
	column string
	value  any
}

func (o *QueryOperand) IsField() bool {
	return o.column != ""
}

func (o *QueryOperand) IsValue() bool {
	return o.value != nil
}

func (o *QueryOperand) IsNull() bool {
	return o.table == "" && o.column == "" && o.value == nil
}

// The templated SQL statement and associated values.
type Statement struct {
	template string
	args     []any
}

type QueryBuilder struct {
	// The model this query building is acting on.
	Model string
	// The table name in the database.
	table string
	// The columns and clauses in SELECT.
	selection []string
	// The columns and clause in DISTINCT ON.
	distinctOn []string
	// The join clauses required for the query.
	joins []string
	// The filter fragments used to construct WHERE.
	filters []string
	// The columns and clauses in ORDER BY.
	orderBy []string
	// The columns and clauses in RETURNING.
	returning []string
	// The value for LIMIT.
	limit *int
	// The ordered slice of arguments for the SQL statement template.
	args []any
	// The columns and values to be written during an INSERT or UPDATE.
	writeValues map[string]any
}

func NewQuery(model *proto.Model) *QueryBuilder {
	return &QueryBuilder{
		Model:       model.Name,
		table:       strcase.ToSnake(model.Name),
		selection:   []string{},
		distinctOn:  []string{},
		joins:       []string{},
		filters:     []string{},
		orderBy:     []string{},
		limit:       nil,
		returning:   []string{},
		args:        []any{},
		writeValues: map[string]any{},
	}
}

// Creates a copy of the query builder.
func (query *QueryBuilder) Copy() *QueryBuilder {
	return &QueryBuilder{
		Model:       query.Model,
		table:       query.table,
		selection:   copySlice(query.selection),
		distinctOn:  copySlice(query.distinctOn),
		joins:       copySlice(query.joins),
		filters:     copySlice(query.filters),
		orderBy:     copySlice(query.orderBy),
		limit:       query.limit,
		returning:   copySlice(query.returning),
		args:        query.args,
		writeValues: query.writeValues,
	}
}

// Includes a value to be written during an INSERT or UPDATE.
func (query *QueryBuilder) AddWriteValue(fieldName string, value any) {
	query.writeValues[strcase.ToSnake(fieldName)] = value
}

// Includes values to be written during an INSERT or UPDATE.
func (query *QueryBuilder) AddWriteValues(values map[string]any) {
	for k, v := range values {
		query.AddWriteValue(k, v)
	}
}

// Includes a column in SELECT.
func (query *QueryBuilder) AppendSelect(operand *QueryOperand) {
	c := operand.toColumnString(query)

	if !lo.Contains(query.selection, c) {
		query.selection = append(query.selection, c)
	}
}

// Include a clause in SELECT.
func (query *QueryBuilder) AppendSelectClause(clause string) {
	if !lo.Contains(query.selection, clause) {
		query.selection = append(query.selection, clause)
	}
}

// Include a column in this table in DISTINCT ON.
func (query *QueryBuilder) AppendDistinctOn(operand *QueryOperand) {
	c := operand.toColumnString(query)

	if !lo.Contains(query.distinctOn, c) {
		query.distinctOn = append(query.distinctOn, c)
	}
}

// Include a WHERE condition, ANDed to the existing filters.
func (query *QueryBuilder) Where(left *QueryOperand, operator ActionOperator, right *QueryOperand) error {
	template, args, err := query.generateWhereTemplate(left, operator, right)
	if err != nil {
		return err
	}

	if len(query.filters) > 0 && query.filters[len(query.filters)-1] != "OR" {
		query.filters = append(query.filters, "AND")
	}

	query.filters = append(query.filters, template)
	query.args = append(query.args, args...)

	return nil
}

// Create a logical OR between existing filters and newly created filters.
func (query *QueryBuilder) Or() {
	query.filters = append(query.filters, "OR")
}

// Include a JOIN clause.
func (query *QueryBuilder) InnerJoin(joinField *QueryOperand, modelField *QueryOperand) {
	join := fmt.Sprintf("INNER JOIN %s ON %s = %s",
		strcase.ToSnake(joinField.table),
		joinField.toColumnString(query),
		modelField.toColumnString(query))

	if !lo.Contains(query.joins, join) {
		query.joins = append(query.joins, join)
	}
}

// Include a column in ORDER BY.
func (query *QueryBuilder) AppendOrderBy(operand *QueryOperand) {
	c := operand.toColumnString(query)

	if !lo.Contains(query.orderBy, c) {
		query.orderBy = append(query.orderBy, c)
	}
}

// Set the LIMIT to a number.
func (query *QueryBuilder) Limit(limit int) {
	query.limit = &limit
}

// Include a column in RETURNING.
func (query *QueryBuilder) AppendReturning(operand *QueryOperand) {
	c := operand.toColumnString(query)

	if !lo.Contains(query.returning, c) {
		query.returning = append(query.returning, c)
	}
}

func (operand *QueryOperand) toColumnString(query *QueryBuilder) string {
	if !operand.IsField() {
		panic("operand is not of type field")
	}

	table := operand.table
	if table == "" {
		table = query.table
	}

	return fmt.Sprintf("%s.%s", table, operand.column)
}

// Generates an executable SELECT statement with the list of arguments.
func (query *QueryBuilder) SelectStatement() *Statement {
	distinctOn := ""
	selection := ""
	joins := ""
	filters := ""
	orderBy := ""
	limit := ""

	if len(query.distinctOn) > 0 {
		distinctOn = fmt.Sprintf("DISTINCT ON(%s)", strings.Join(query.distinctOn, ", "))
	}

	if len(query.selection) == 0 {
		query.AppendSelect(Field("*"))
	}

	selection = strings.Join(query.selection, ", ")

	if len(query.joins) > 0 {
		joins = strings.Join(query.joins, " ")
	}

	if len(query.filters) > 0 {
		q := lo.DropRightWhile(query.filters, func(s string) bool { return s == "OR" || s == "AND" })
		filters = fmt.Sprintf("WHERE %s", strings.Join(q, " "))
	}

	if len(query.orderBy) > 0 {
		orderBy = fmt.Sprintf("ORDER BY %s", strings.Join(query.orderBy, ", "))
	}

	if query.limit != nil {
		limit = fmt.Sprintf("LIMIT %v", *query.limit)
	}

	sql := fmt.Sprintf("SELECT %s %s FROM %s %s %s %s %s",
		distinctOn,
		selection,
		query.table,
		joins,
		filters,
		orderBy,
		limit)

	return &Statement{
		template: sql,
		args:     query.args,
	}
}

// Generates an executable INSERT statement with the list of arguments.
func (query *QueryBuilder) InsertStatement() *Statement {
	returning := ""
	columns := []string{}
	args := []any{}

	for k, v := range query.writeValues {
		columns = append(columns, k)
		args = append(args, v)
	}

	if len(query.returning) > 0 {
		returning = fmt.Sprintf("RETURNING %s", strings.Join(query.returning, ", "))
	}

	template := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) %s",
		query.table,
		strings.Join(columns, ", "),
		strings.Join(strings.Split(strings.Repeat("?", len(query.writeValues)), ""), ", "),
		returning)

	return &Statement{
		template: template,
		args:     args,
	}
}

// Generates an executable UPDATE statement with the list of arguments.
func (query *QueryBuilder) UpdateStatement() *Statement {
	joins := ""
	filters := ""
	returning := ""
	sets := []string{}
	args := []any{}

	for k, v := range query.writeValues {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	args = append(args, query.args...)

	if len(query.joins) > 0 {
		joins = strings.Join(query.joins, " ")
	}

	if len(query.filters) > 0 {
		q := lo.DropRightWhile(query.filters, func(s string) bool { return s == "OR" || s == "AND" })
		filters = fmt.Sprintf("WHERE %s", strings.Join(q, " "))
	}

	if len(query.returning) > 0 {
		returning = fmt.Sprintf("RETURNING %s", strings.Join(query.returning, ", "))
	}

	template := fmt.Sprintf("UPDATE %s SET %s %s %s %s",
		query.table,
		strings.Join(sets, ", "),
		joins,
		filters,
		returning)

	return &Statement{
		template: template,
		args:     args,
	}
}

// Generates an executable DELETE statement with the list of arguments.
func (query *QueryBuilder) DeleteStatement() *Statement {
	joins := ""
	filters := ""
	returning := ""

	if len(query.joins) > 0 {
		joins = strings.Join(query.joins, " ")
	}

	if len(query.filters) > 0 {
		// Removes any trailing OR or AND from the where fragments
		q := lo.DropRightWhile(query.filters, func(s string) bool { return s == "OR" || s == "AND" })
		filters = fmt.Sprintf("WHERE %s", strings.Join(q, " "))
	}

	if len(query.returning) > 0 {
		returning = fmt.Sprintf("RETURNING %s", strings.Join(query.returning, ", "))
	}

	template := fmt.Sprintf("DELETE FROM %s %s %s %s",
		query.table,
		joins,
		filters,
		returning)

	return &Statement{
		template: template,
		args:     query.args,
	}
}

// Execute the SQL statement against the database, returning the number of rows affected.
func (statement *Statement) Execute(context context.Context) (int, error) {
	db, err := runtimectx.GetDatabase(context)
	if err != nil {
		return 0, err
	}

	db = db.Exec(statement.template, statement.args...)
	if db.Error != nil {
		return 0, db.Error
	}

	return int(db.RowsAffected), nil
}

// Execute the SQL statement against the database, return the rows, and  the number of rows affected.
func (statement *Statement) ExecuteWithResults(context context.Context) ([]map[string]any, int, error) {
	db, err := runtimectx.GetDatabase(context)
	if err != nil {
		return nil, 0, err
	}

	results := []map[string]any{}
	db = db.Raw(statement.template, statement.args...).Scan(&results)
	if db.Error != nil {
		return nil, 0, db.Error
	}

	return toLowerCamelMaps(results), int(db.RowsAffected), nil
}

// Builds a where conditional SQL template using the ? placeholder for values.
func (query *QueryBuilder) generateWhereTemplate(lhs *QueryOperand, operator ActionOperator, rhs *QueryOperand) (string, []any, error) {
	var template string
	var lhsSqlOperand, rhsSqlOperand any
	args := []any{}

	switch operator {
	case StartsWith:
		rhs.value = rhs.value.(string) + "%%"
	case EndsWith:
		rhs.value = "%%" + rhs.value.(string)
	case Contains, NotContains:
		rhs.value = "%%" + rhs.value.(string) + "%%"
	}

	switch {
	case lhs.IsField():
		lhsSqlOperand = lhs.toColumnString(query)
	case lhs.IsValue():
		lhsSqlOperand = "?"
		args = append(args, lhs.value)
	case lhs.IsNull():
		lhsSqlOperand = "NULL"
	default:
		return "", nil, errors.New("no handling for lhs QueryOperand type")
	}

	switch {
	case rhs.IsField():
		rhsSqlOperand = rhs.toColumnString(query)
	case rhs.IsValue():
		rhsSqlOperand = "?"
		args = append(args, rhs.value)
	case rhs.IsNull():
		rhsSqlOperand = "NULL"
	default:
		return "", nil, errors.New("no handling for rhs QueryOperand type")
	}

	switch operator {
	case Equals:
		template = fmt.Sprintf("%s IS NOT DISTINCT FROM %s", lhsSqlOperand, rhsSqlOperand)
	case NotEquals:
		template = fmt.Sprintf("%s IS DISTINCT FROM %s", lhsSqlOperand, rhsSqlOperand)
	case StartsWith, EndsWith, Contains:
		template = fmt.Sprintf("%s LIKE %s", lhsSqlOperand, rhsSqlOperand)
	case NotContains:
		template = fmt.Sprintf("%s NOT LIKE %s", lhsSqlOperand, rhsSqlOperand)
	case OneOf:
		template = fmt.Sprintf("%s in %s", lhsSqlOperand, rhsSqlOperand)
	case LessThan:
		template = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case LessThanEquals:
		template = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThan:
		template = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case GreaterThanEquals:
		template = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
	case Before:
		template = fmt.Sprintf("%s < %s", lhsSqlOperand, rhsSqlOperand)
	case After:
		template = fmt.Sprintf("%s > %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrBefore:
		template = fmt.Sprintf("%s <= %s", lhsSqlOperand, rhsSqlOperand)
	case OnOrAfter:
		template = fmt.Sprintf("%s >= %s", lhsSqlOperand, rhsSqlOperand)
	default:
		return "", nil, fmt.Errorf("operator: %v is not yet supported", operator)
	}

	return template, args, nil
}

func copySlice(a []string) []string {
	tmp := make([]string, len(a))
	copy(tmp, a)
	return tmp
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

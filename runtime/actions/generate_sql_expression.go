package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// SQLExpressionGeneratorConfig configures the SQL expression generator
type SQLExpressionGeneratorConfig struct {
	// Context for the operation
	Ctx context.Context
	// Schema being used
	Schema *proto.Schema
	// Entity (model/task) being operated on
	Entity proto.Entity
	// Action being performed (can be nil)
	Action *proto.Action
	// Inputs to the action (can be nil)
	Inputs map[string]any
	// TableAlias is the alias for the root table (e.g., "r" for computed fields, actual table name for filters)
	TableAlias string
	// ResultType is the expected result type (used for string concatenation detection)
	ResultType proto.Type
	// EmbedLiterals determines if literals should be embedded or use placeholders
	EmbedLiterals bool
	// UseJoins determines if relationships should use JOINs (true for filters/permissions) or subqueries (false for computed fields)
	UseJoins bool
}

// GenerateSQLExpression creates a visitor that generates raw SQL expressions
func GenerateSQLExpression(config SQLExpressionGeneratorConfig) resolve.Visitor[*SQLExpression] {
	return &sqlExpressionGen{
		ctx:           config.Ctx,
		schema:        config.Schema,
		entity:        config.Entity,
		action:        config.Action,
		inputs:        config.Inputs,
		tableAlias:    config.TableAlias,
		resultType:    config.ResultType,
		embedLiterals: config.EmbedLiterals,
		useJoins:      config.UseJoins,
		sql:           "",
		args:          []any{},
		joins:         []string{},
		functions:     arraystack.New(),
		arguments:     arraystack.New(),
	}
}

// SQLExpression represents a generated SQL expression with arguments
type SQLExpression struct {
	SQL   string
	Args  []any
	Joins []string // LEFT JOIN clauses (only populated when UseJoins is true)
}

// BuildSelectStatement creates a complete SELECT statement from the SQL expression
// This is used for subqueries in aggregate functions
func (e *SQLExpression) BuildSelectStatement(selectClause string, fromTable string, wherePrefix string) string {
	sql := "SELECT " + selectClause + " FROM " + sqlQuote(fromTable)

	// Add JOINs if any
	if len(e.Joins) > 0 {
		sql += " " + strings.Join(e.Joins, " ")
	}

	// Add WHERE clause
	if wherePrefix != "" && e.SQL != "" {
		sql += " WHERE " + wherePrefix + " AND " + e.SQL
	} else if wherePrefix != "" {
		sql += " WHERE " + wherePrefix
	} else if e.SQL != "" {
		sql += " WHERE " + e.SQL
	}

	return cleanSql(sql)
}

var _ resolve.Visitor[*SQLExpression] = new(sqlExpressionGen)

type sqlExpressionGen struct {
	ctx           context.Context
	schema        *proto.Schema
	entity        proto.Entity
	action        *proto.Action
	inputs        map[string]any
	tableAlias    string
	resultType    proto.Type
	embedLiterals bool
	useJoins      bool
	sql           string
	args          []any
	joins         []string
	functions     *arraystack.Stack
	arguments     *arraystack.Stack
	filter        any // Can be resolve.Visitor[*QueryBuilder] or *aggregateFilterState
}

func (v *sqlExpressionGen) StartTerm(nested bool) error {
	// If we're inside a function and have a filter, delegate to it
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.StartTerm(nested)
		case resolve.Visitor[*QueryBuilder]:
			return filter.StartTerm(nested)
		}
	}

	if nested {
		v.sql += "("
	}

	return nil
}

func (v *sqlExpressionGen) EndTerm(nested bool) error {
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.EndTerm(nested)
		case resolve.Visitor[*QueryBuilder]:
			return filter.EndTerm(nested)
		}
	}

	if nested {
		v.sql += ")"
	}
	return nil
}

func (v *sqlExpressionGen) StartFunction(name string) error {
	v.functions.Push(name)
	v.arguments.Push(0)
	return nil
}

func (v *sqlExpressionGen) EndFunction() error {
	v.functions.Pop()
	v.arguments.Pop()

	if v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			// New unified generator-based filter (used for aggregates)
			filterExpr, err := filter.filterGen.Result()
			if err != nil {
				return err
			}

			// Build the complete SELECT statement for the aggregate subquery
			sql := "SELECT " + filter.selectClause + " FROM " + sqlQuote(filter.fromTable)

			// Combine base JOINs (for aggregated field) with filter JOINs
			allJoins := append(filter.baseJoins, filterExpr.Joins...)
			if len(allJoins) > 0 {
				// Deduplicate joins
				uniqueJoins := make([]string, 0, len(allJoins))
				seen := make(map[string]bool)
				for _, join := range allJoins {
					if !seen[join] {
						uniqueJoins = append(uniqueJoins, join)
						seen[join] = true
					}
				}
				sql += " " + strings.Join(uniqueJoins, " ")
			}

			// Add WHERE clause with foreign key condition and filter expression
			whereClauses := []string{
				fmt.Sprintf("%s.%s IS NOT DISTINCT FROM %s",
					sqlQuote(filter.fromTable),
					sqlQuote(casing.ToSnake(filter.foreignKeyField)),
					filter.parentTableRef),
			}

			if filterExpr.SQL != "" {
				whereClauses = append(whereClauses, filterExpr.SQL)
			}

			sql += " WHERE " + strings.Join(whereClauses, " AND ")

			v.sql += fmt.Sprintf("(%s)", cleanSql(sql))
			v.args = append(v.args, filterExpr.Args...)

		case resolve.Visitor[*QueryBuilder]:
			// Legacy QueryBuilder-based filter (used for 1:1 relationships)
			query, err := filter.Result()
			if err != nil {
				return err
			}

			stmt := query.SelectStatement()
			v.sql += fmt.Sprintf("(%s)", stmt.SqlTemplate())
			v.args = append(v.args, stmt.SqlArgs()...)
		}

		v.filter = nil
	}

	return nil
}

func (v *sqlExpressionGen) StartArgument(num int) error {
	arg, has := v.arguments.Pop()
	if !has {
		return errors.New("argument stack is empty")
	}

	v.arguments.Push(arg.(int) + 1)
	return nil
}

func (v *sqlExpressionGen) EndArgument() error {
	return nil
}

func (v *sqlExpressionGen) VisitAnd() error {
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.VisitAnd()
		case resolve.Visitor[*QueryBuilder]:
			return filter.VisitAnd()
		}
	}

	v.sql += " AND "
	return nil
}

func (v *sqlExpressionGen) VisitOr() error {
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.VisitOr()
		case resolve.Visitor[*QueryBuilder]:
			return filter.VisitOr()
		}
	}

	v.sql += " OR "
	return nil
}

func (v *sqlExpressionGen) VisitNot() error {
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.VisitNot()
		case resolve.Visitor[*QueryBuilder]:
			return filter.VisitNot()
		}
	}

	v.sql += "NOT "
	return nil
}

func (v *sqlExpressionGen) VisitOperator(op string) error {
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.VisitOperator(op)
		case resolve.Visitor[*QueryBuilder]:
			return filter.VisitOperator(op)
		}
	}

	// Map CEL operators to SQL operators
	sqlOp := map[string]string{
		operators.Add:           "+",
		operators.Subtract:      "-",
		operators.Multiply:      "*",
		operators.Divide:        "/",
		operators.Equals:        "IS NOT DISTINCT FROM",
		operators.NotEquals:     "IS DISTINCT FROM",
		operators.Greater:       ">",
		operators.GreaterEquals: ">=",
		operators.Less:          "<",
		operators.LessEquals:    "<=",
	}[op]

	// Handle string concatenation
	if v.resultType == proto.Type_TYPE_STRING && op == operators.Add {
		sqlOp = "||"
	}

	if sqlOp == "" {
		return fmt.Errorf("unsupported operator: %s", op)
	}

	v.sql += fmt.Sprintf(" %s ", sqlOp)

	return nil
}

func (v *sqlExpressionGen) VisitLiteral(value any) error {
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.VisitLiteral(value)
		case resolve.Visitor[*QueryBuilder]:
			return filter.VisitLiteral(value)
		}
	}

	if v.embedLiterals {
		switch val := value.(type) {
		case int64:
			v.sql += fmt.Sprintf("%v", val)
		case float64:
			v.sql += fmt.Sprintf("%v", val)
		case string:
			v.sql += fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''"))
		case bool:
			v.sql += fmt.Sprintf("%t", val)
		case nil:
			v.sql += "NULL"
		default:
			return fmt.Errorf("unsupported literal type: %T", value)
		}
	} else {
		v.sql += "?"
		v.args = append(v.args, value)
	}

	return nil
}

func (v *sqlExpressionGen) VisitIdent(ident *parser.ExpressionIdent) error {
	// If we're inside a function and have a filter, delegate ident handling to it
	// This happens when processing the filter condition (second argument) of aggregate functions
	if v.functions.Size() > 0 && v.filter != nil {
		switch filter := v.filter.(type) {
		case *aggregateFilterState:
			return filter.VisitIdent(ident)
		case resolve.Visitor[*QueryBuilder]:
			return filter.VisitIdent(ident)
		}
	}

	// Check if it's an enum
	entity := v.schema.FindEntity(strcase.ToCamel(ident.Fragments[0]))
	if entity == nil {
		enum := v.schema.FindEnum(ident.Fragments[0])
		if enum == nil {
			return fmt.Errorf("model, task, or enum not found: %s", ident.Fragments[0])
		}

		var value string
		for _, enumVal := range enum.GetValues() {
			if enumVal.GetName() == ident.Fragments[1] {
				value = enumVal.GetName()
				break
			}
		}

		if value == "" {
			return fmt.Errorf("enum value not found: %s", ident.Fragments[1])
		}

		if v.embedLiterals {
			v.sql += fmt.Sprintf("'%s'", value)
		} else {
			v.sql += "?"
			v.args = append(v.args, value)
		}
		return nil
	}

	field := entity.FindField(ident.Fragments[1])

	normalised, err := NormaliseFragments(v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	// Simple field reference on the root entity
	if len(normalised) == 2 {
		// Only quote table alias when using JOINs (for filters/permissions with real table names)
		tableRef := v.tableAlias
		if v.useJoins {
			tableRef = sqlQuote(v.tableAlias)
		}
		v.sql += fmt.Sprintf("%s.%s", tableRef, sqlQuote(strcase.ToSnake(field.GetName())))
		return nil
	}

	// Complex field reference (relationship traversal)
	if len(normalised) > 2 {
		// Use JOIN mode for filters/permissions
		if v.useJoins {
			return v.handleRelationshipWithJoins(entity, normalised, ident.Fragments)
		}

		// Use subquery mode for computed fields
		isToMany, err := v.isToManyLookup(ident)
		if err != nil {
			return err
		}

		if isToMany {
			// Handle aggregate functions over 1:M relationships
			return v.handleToManyAggregate(ident, normalised, field)
		} else {
			// Handle simple 1:1 relationship lookup
			return v.handleToOneRelationship(ident, normalised, field)
		}
	}

	return nil
}

// handleRelationshipWithJoins handles relationship traversal using LEFT JOINs (for filters/permissions)
func (v *sqlExpressionGen) handleRelationshipWithJoins(entity proto.Entity, normalised []string, originalFragments []string) error {
	fieldName := ""
	currentEntity := entity

	for i, fragment := range normalised {
		switch {
		// The first fragment (entity name)
		case i == 0:
			fieldName += casing.ToSnake(fragment)

		// Remaining fragments
		default:
			field := currentEntity.FindField(fragment)
			if field == nil {
				return fmt.Errorf("model %s has no field %s", currentEntity.GetName(), fragment)
			}

			isLast := i == len(normalised)-1
			isModel := field.GetType().GetType() == proto.Type_TYPE_ENTITY
			hasFk := field.GetForeignKeyFieldName() != nil

			if isModel && (!isLast || !hasFk) {
				// Left alias is the source table
				leftAlias := fieldName

				// Append fragment to identifier
				fieldName += "$" + casing.ToSnake(fragment)

				// Right alias is the join table
				rightAlias := fieldName

				joinEntity := v.schema.FindEntity(field.GetType().GetEntityName().GetValue())
				if joinEntity == nil {
					return fmt.Errorf("model %s not found in schema", field.GetType().GetEntityName().GetValue())
				}

				leftFieldName := v.schema.GetForeignKeyFieldName(field)
				rightFieldName := joinEntity.PrimaryKeyFieldName()

				// If not belongs to then swap foreign/primary key
				if !field.IsBelongsTo() {
					leftFieldName = currentEntity.PrimaryKeyFieldName()
					rightFieldName = v.schema.GetForeignKeyFieldName(field)
				}

				// Convert field names to snake_case
				leftFieldName = casing.ToSnake(leftFieldName)
				rightFieldName = casing.ToSnake(rightFieldName)

				// Generate join with the joined table first in the ON clause
				join := fmt.Sprintf(
					"LEFT JOIN %s AS %s ON %s.%s = %s.%s",
					sqlQuote(casing.ToSnake(joinEntity.GetName())),
					sqlQuote(rightAlias),
					sqlQuote(rightAlias),
					sqlQuote(rightFieldName),
					sqlQuote(leftAlias),
					sqlQuote(leftFieldName),
				)

				v.joins = append(v.joins, join)
				currentEntity = joinEntity
			}

			if isLast {
				// Turn the table into a quoted identifier
				fieldName = sqlQuote(fieldName)

				// Then append the field name as a quoted identifier
				if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
					if field.GetForeignKeyFieldName() != nil {
						fieldName += "." + sqlQuote(field.GetForeignKeyFieldName().GetValue())
					} else {
						fieldName += "." + sqlQuote("id")
					}
				} else {
					fieldName += "." + sqlQuote(casing.ToSnake(fragment))
				}
			}
		}
	}

	v.sql += fieldName
	return nil
}

func (v *sqlExpressionGen) handleToManyAggregate(ident *parser.ExpressionIdent, normalised []string, field *proto.Field) error {
	arg, has := v.arguments.Peek()
	if !has {
		return errors.New("argument stack is empty")
	}

	entity := v.schema.FindEntity(field.GetType().GetEntityName().GetValue())

	// Only handle the first argument (the field to aggregate)
	// The second argument (filter condition) is handled automatically through delegation in Visit* methods
	if arg.(int) == 1 {
		relatedEntityField := v.entity.FindField(normalised[1])
		if relatedEntityField == nil {
			return fmt.Errorf("field %s not found on %s", normalised[1], v.entity.GetName())
		}

		foreignKeyField := v.schema.GetForeignKeyFieldName(relatedEntityField)
		if foreignKeyField == "" {
			return fmt.Errorf("foreign key field not found for %s", normalised[1])
		}

		funcName, has := v.functions.Peek()
		if !has {
			return errors.New("no function found for 1:M lookup")
		}

		// Build the SELECT field reference for the aggregate
		// For example, for invoice.item.product.price, we want "item$product"."price"
		fieldName := normalised[len(normalised)-1]
		fragments := normalised[1 : len(normalised)-1]

		selectField := sqlQuote(casing.ToSnake(strings.Join(fragments, "$"))) + "." + sqlQuote(casing.ToSnake(fieldName))
		selectClause := ""
		switch funcName {
		case typing.FunctionSum, typing.FunctionSumIf:
			selectClause = fmt.Sprintf("COALESCE(SUM(%s), 0)", selectField)
		case typing.FunctionCount, typing.FunctionCountIf:
			selectClause = fmt.Sprintf("COALESCE(COUNT(%s), 0)", selectField)
		case typing.FunctionAvg, typing.FunctionAvgIf:
			selectClause = fmt.Sprintf("COALESCE(AVG(%s), 0)", selectField)
		case typing.FunctionMedian, typing.FunctionMedianIf:
			selectClause = fmt.Sprintf("COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY %s), 0)", selectField)
		case typing.FunctionMin, typing.FunctionMinIf:
			selectClause = fmt.Sprintf("COALESCE(MIN(%s), 0)", selectField)
		case typing.FunctionMax, typing.FunctionMaxIf:
			selectClause = fmt.Sprintf("COALESCE(MAX(%s), 0)", selectField)
		default:
			return fmt.Errorf("unsupported aggregate function: %s", funcName)
		}

		// Build JOINs for the aggregated field path
		// For example, for invoice.item.product.price, we need to join from item to product
		subFragments := normalised[1:]
		subFragments[0] = strcase.ToLowerCamel(relatedEntityField.GetType().GetEntityName().GetValue())
		joins, err := v.buildJoinsForFragments(entity, subFragments)
		if err != nil {
			return err
		}

		// Create a nested SQL expression generator for the filter condition
		// The filter processes idents relative to the child entity in the subquery context
		filterGen := GenerateSQLExpression(SQLExpressionGeneratorConfig{
			Ctx:           v.ctx,
			Schema:        v.schema,
			Entity:        entity, // Use child entity for the subquery context
			Action:        v.action,
			Inputs:        v.inputs,
			TableAlias:    casing.ToSnake(entity.GetName()), // Use child entity table name as base
			EmbedLiterals: v.embedLiterals,
			UseJoins:      true, // Use JOINs for filter conditions in aggregate subqueries
		})

		// Store the aggregate subquery state
		// The prefix to strip is the parent entity name (e.g., "invoice" from "invoice.item.isDeleted")
		v.filter = &aggregateFilterState{
			filterGen:       filterGen,
			selectClause:    selectClause,
			fromTable:       casing.ToSnake(entity.GetName()),
			foreignKeyField: foreignKeyField,
			parentTableRef:  fmt.Sprintf("%s.%s", v.tableAlias, sqlQuote(parser.FieldNameId)),
			prefixToStrip:   []string{normalised[0]}, // Strip the parent entity name (e.g., ["Invoice"])
			baseJoins:       joins,                    // JOINs for the aggregated field path
		}
	}

	return nil
}

// aggregateFilterState holds state for aggregate filter processing
type aggregateFilterState struct {
	filterGen       resolve.Visitor[*SQLExpression]
	selectClause    string
	fromTable       string
	foreignKeyField string
	parentTableRef  string
	prefixToStrip   []string // Ident prefix to strip (e.g., ["invoice", "item"] → strip to start from item)
	baseJoins       []string // JOINs for the aggregated field path
}

// Implement the Visitor interface for aggregateFilterState to delegate to filterGen
func (s *aggregateFilterState) StartTerm(nested bool) error {
	return s.filterGen.StartTerm(nested)
}

func (s *aggregateFilterState) EndTerm(nested bool) error {
	return s.filterGen.EndTerm(nested)
}

func (s *aggregateFilterState) StartFunction(name string) error {
	return s.filterGen.StartFunction(name)
}

func (s *aggregateFilterState) EndFunction() error {
	return s.filterGen.EndFunction()
}

func (s *aggregateFilterState) StartArgument(num int) error {
	return s.filterGen.StartArgument(num)
}

func (s *aggregateFilterState) EndArgument() error {
	return s.filterGen.EndArgument()
}

func (s *aggregateFilterState) VisitAnd() error {
	return s.filterGen.VisitAnd()
}

func (s *aggregateFilterState) VisitOr() error {
	return s.filterGen.VisitOr()
}

func (s *aggregateFilterState) VisitNot() error {
	return s.filterGen.VisitNot()
}

func (s *aggregateFilterState) VisitOperator(op string) error {
	return s.filterGen.VisitOperator(op)
}

func (s *aggregateFilterState) VisitLiteral(value any) error {
	return s.filterGen.VisitLiteral(value)
}

func (s *aggregateFilterState) VisitIdent(ident *parser.ExpressionIdent) error {
	// Transform the ident by stripping the prefix
	// For example: ["invoice", "item", "isDeleted"] with prefix ["invoice"] becomes ["item", "isDeleted"]
	if len(s.prefixToStrip) > 0 && len(ident.Fragments) > len(s.prefixToStrip) {
		// Check if the ident starts with the prefix
		matches := true
		for i, prefix := range s.prefixToStrip {
			if ident.Fragments[i] != prefix {
				matches = false
				break
			}
		}

		if matches {
			// Create a new ident with the prefix stripped
			transformed := &parser.ExpressionIdent{
				Fragments: ident.Fragments[len(s.prefixToStrip):],
			}
			return s.filterGen.VisitIdent(transformed)
		}
	}

	return s.filterGen.VisitIdent(ident)
}

func (s *aggregateFilterState) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return s.filterGen.VisitIdentArray(idents)
}

func (s *aggregateFilterState) Result() (*QueryBuilder, error) {
	// This should never be called as we handle it in EndFunction
	return nil, errors.New("aggregateFilterState.Result should not be called")
}

// buildJoinsForFragments builds JOIN clauses for a fragment path
func (v *sqlExpressionGen) buildJoinsForFragments(startEntity proto.Entity, fragments []string) ([]string, error) {
	var joins []string
	currentEntity := startEntity
	tablePath := casing.ToSnake(startEntity.GetName())

	for i := 1; i < len(fragments); i++ {
		fieldName := fragments[i]
		field := currentEntity.FindField(fieldName)
		if field == nil {
			return nil, fmt.Errorf("field %s not found on %s", fieldName, currentEntity.GetName())
		}

		if field.GetType().GetType() != proto.Type_TYPE_ENTITY {
			// Not a relationship field, skip
			continue
		}

		// Build the join
		leftAlias := tablePath
		tablePath += "$" + casing.ToSnake(fieldName)
		rightAlias := tablePath

		joinEntity := v.schema.FindEntity(field.GetType().GetEntityName().GetValue())
		if joinEntity == nil {
			return nil, fmt.Errorf("entity %s not found", field.GetType().GetEntityName().GetValue())
		}

		leftFieldName := v.schema.GetForeignKeyFieldName(field)
		rightFieldName := joinEntity.PrimaryKeyFieldName()

		// If not belongs to then swap foreign/primary key
		if !field.IsBelongsTo() {
			leftFieldName = currentEntity.PrimaryKeyFieldName()
			rightFieldName = v.schema.GetForeignKeyFieldName(field)
		}

		// Convert field names to snake_case
		leftFieldName = casing.ToSnake(leftFieldName)
		rightFieldName = casing.ToSnake(rightFieldName)

		// Generate join with the joined table first in the ON clause
		join := fmt.Sprintf(
			"LEFT JOIN %s AS %s ON %s.%s = %s.%s",
			sqlQuote(casing.ToSnake(joinEntity.GetName())),
			sqlQuote(rightAlias),
			sqlQuote(rightAlias),
			sqlQuote(rightFieldName),
			sqlQuote(leftAlias),
			sqlQuote(leftFieldName),
		)

		joins = append(joins, join)
		currentEntity = joinEntity
	}

	return joins, nil
}

func (v *sqlExpressionGen) handleToOneRelationship(ident *parser.ExpressionIdent, normalised []string, field *proto.Field) error {
	entity := v.schema.FindEntity(field.GetType().GetEntityName().GetValue())

	var options []QueryBuilderOption
	if v.embedLiterals {
		options = append(options, EmbedLiterals())
	}
	query := NewQuery(entity, options...)

	relatedEntityField := v.entity.FindField(normalised[1])
	subFragments := normalised[1:]
	subFragments[0] = strcase.ToLowerCamel(relatedEntityField.GetType().GetEntityName().GetValue())

	err := query.AddJoinFromFragments(v.schema, subFragments)
	if err != nil {
		return err
	}

	// Select the target field
	fieldName := normalised[len(normalised)-1]
	fragments := subFragments[:len(subFragments)-1]
	query.Select(ExpressionField(fragments, fieldName, false))

	// Filter by the root entity's foreign key
	foreignKeyField := v.schema.GetForeignKeyFieldName(relatedEntityField)
	fk := fmt.Sprintf("%s.%s", v.tableAlias, sqlQuote(strcase.ToSnake(foreignKeyField)))
	err = query.Where(IdField(), Equals, Raw(fk))
	if err != nil {
		return err
	}

	stmt := query.SelectStatement()
	v.sql += fmt.Sprintf("(%s)", stmt.SqlTemplate())
	if !v.embedLiterals {
		v.args = append(v.args, stmt.SqlArgs()...)
	}

	return nil
}

func (v *sqlExpressionGen) isToManyLookup(ident *parser.ExpressionIdent) (bool, error) {
	entity := v.schema.FindEntity(strcase.ToCamel(ident.Fragments[0]))

	fragments, err := NormaliseFragments(v.schema, ident.Fragments)
	if err != nil {
		return false, err
	}

	for i := 1; i < len(fragments)-1; i++ {
		currentFragment := fragments[i]
		field := entity.FindField(currentFragment)
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY && field.GetType().GetRepeated() {
			return true, nil
		}
		entity = v.schema.FindEntity(field.GetType().GetEntityName().GetValue())
	}

	return false, nil
}

func (v *sqlExpressionGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return errors.New("ident arrays not supported in SQL expressions")
}

func (v *sqlExpressionGen) Result() (*SQLExpression, error) {
	return &SQLExpression{
		SQL:   cleanSql(v.sql),
		Args:  v.args,
		Joins: v.joins,
	}, nil
}

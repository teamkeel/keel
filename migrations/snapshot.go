package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
)

type TypeHandler interface {
	ToSQL(value interface{}, field *proto.Field) (string, error)
}

// DefaultTypeHandler handles basic types using PostgreSQL's to_json
type DefaultTypeHandler struct{}

func (h *DefaultTypeHandler) ToSQL(value interface{}, field *proto.Field) (string, error) {
	if value == nil {
		return "NULL", nil
	}

	switch v := value.(type) {
	case string:
		if field != nil && field.Type.Type == proto.Type_TYPE_OBJECT {
			// For JSONB fields, keep the JSON as is
			return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")), nil
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return fmt.Sprintf("%v", v), nil
	case map[string]interface{}:
		// For JSON/JSONB fields
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("error marshaling JSON: %w", err)
		}
		return fmt.Sprintf("'%s'", string(jsonBytes)), nil
	default:
		// For any other type, try JSON marshaling as a fallback
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("error marshaling value: %w", err)
		}
		return fmt.Sprintf("'%s'", string(jsonBytes)), nil
	}
}

type DurationTypeHandler struct{}

func (h *DurationTypeHandler) ToSQL(value interface{}, field *proto.Field) (string, error) {
	if value == nil {
		return "NULL", nil
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected string for Duration type")
	}

	return fmt.Sprintf("'%s'::interval", strings.ReplaceAll(str, "'", "''")), nil
}

type FileTypeHandler struct{}

func (h *FileTypeHandler) ToSQL(value interface{}, field *proto.Field) (string, error) {
	if value == nil {
		return "NULL", nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Create ordered map for consistent output
		orderedMap := make(map[string]interface{})
		if url, ok := v["url"].(string); ok {
			orderedMap["url"] = strings.ReplaceAll(url, "'", "''")
		} else {
			orderedMap["url"] = v["url"]
		}
		orderedMap["contentType"] = v["contentType"]
		orderedMap["size"] = v["size"]

		jsonBytes, err := json.Marshal(orderedMap)
		if err != nil {
			return "", fmt.Errorf("error marshaling File: %w", err)
		}

		return fmt.Sprintf("'%s'", string(jsonBytes)), nil
	case string:
		// If it's a string, assume it's already a JSON string
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")), nil
	default:
		return "", fmt.Errorf("unsupported type for File: %T", value)
	}
}

type ArrayTypeHandler struct{}

func (h *ArrayTypeHandler) ToSQL(value interface{}, field *proto.Field) (string, error) {
	switch v := value.(type) {
	case string:
		// If the string starts with '{' and ends with '}', it's a text array representation
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") && !strings.Contains(v, ":") {
			// Convert PostgreSQL array literal to ARRAY constructor
			elements := strings.Split(strings.Trim(v, "{}"), ",")
			arrayStr := make([]string, len(elements))
			for i, element := range elements {
				element = strings.TrimSpace(element)
				// Always quote the values, they will be cast to the appropriate type
				arrayStr[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(element, "'", "''"))
			}
			arrayExpr := fmt.Sprintf("ARRAY[%s]", strings.Join(arrayStr, ", "))
			if field != nil && field.Type.Repeated {
				switch field.Type.Type {
				case proto.Type_TYPE_INT:
					arrayExpr += "::integer[]"
				case proto.Type_TYPE_DECIMAL:
					arrayExpr += "::decimal[]"
				case proto.Type_TYPE_DATE:
					arrayExpr += "::date[]"
				case proto.Type_TYPE_TIMESTAMP:
					arrayExpr += "::timestamp[]"
				case proto.Type_TYPE_DURATION:
					arrayExpr += "::interval[]"
				}
			}
			return arrayExpr, nil
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")), nil
	case []interface{}:
		arrayStr := make([]string, len(v))
		for i, item := range v {
			switch itemVal := item.(type) {
			case string:
				arrayStr[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(itemVal, "'", "''"))
			case nil:
				arrayStr[i] = "NULL"
			case map[string]interface{}:
				jsonBytes, err := json.Marshal(itemVal)
				if err != nil {
					return "", fmt.Errorf("error marshaling JSON for array item: %w", err)
				}
				arrayStr[i] = fmt.Sprintf("'%s'", string(jsonBytes))
			case float64:
				arrayStr[i] = fmt.Sprintf("'%g'", itemVal)
			case int:
				arrayStr[i] = fmt.Sprintf("'%d'", itemVal)
			case bool:
				arrayStr[i] = fmt.Sprintf("'%t'", itemVal)
			default:
				return "", fmt.Errorf("unsupported array item type: %T", itemVal)
			}
		}
		arrayExpr := fmt.Sprintf("ARRAY[%s]", strings.Join(arrayStr, ", "))
		if field != nil && field.Type.Repeated {
			switch field.Type.Type {
			case proto.Type_TYPE_INT:
				arrayExpr += "::integer[]"
			case proto.Type_TYPE_DECIMAL:
				arrayExpr += "::decimal[]"
			case proto.Type_TYPE_DATE:
				arrayExpr += "::date[]"
			case proto.Type_TYPE_TIMESTAMP:
				arrayExpr += "::timestamp[]"
			case proto.Type_TYPE_DURATION:
				arrayExpr += "::interval[]"
			}
		}
		return arrayExpr, nil
	case []string:
		arrayStr := make([]string, len(v))
		for i, item := range v {
			arrayStr[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(item, "'", "''"))
		}
		arrayExpr := fmt.Sprintf("ARRAY[%s]", strings.Join(arrayStr, ", "))
		if field != nil && field.Type.Type == proto.Type_TYPE_DURATION {
			arrayExpr += "::interval[]"
		}
		return arrayExpr, nil
	case []int:
		arrayStr := make([]string, len(v))
		for i, item := range v {
			arrayStr[i] = fmt.Sprintf("'%d'", item)
		}
		return fmt.Sprintf("ARRAY[%s]::integer[]", strings.Join(arrayStr, ", ")), nil
	case []float64:
		arrayStr := make([]string, len(v))
		for i, item := range v {
			arrayStr[i] = fmt.Sprintf("'%g'", item)
		}
		return fmt.Sprintf("ARRAY[%s]::decimal[]", strings.Join(arrayStr, ", ")), nil
	default:
		return "", fmt.Errorf("unsupported array type: %T", value)
	}
}

type TimeTypeHandler struct{}

func (h *TimeTypeHandler) ToSQL(value interface{}, field *proto.Field) (string, error) {
	if value == nil {
		return "NULL", nil
	}

	switch v := value.(type) {
	case time.Time:
		return fmt.Sprintf("'%s'::timestamp with time zone", v.Format(time.RFC3339)), nil
	case string:
		// Try to parse the string as a time
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return "", fmt.Errorf("invalid time string format: %w", err)
		}
		return fmt.Sprintf("'%s'::timestamp with time zone", t.Format(time.RFC3339)), nil
	default:
		return "", fmt.Errorf("expected time.Time or string, got %T", value)
	}
}

func (m *Migrations) SnapshotDatabase(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "Snapshot Database V2")
	defer span.End()

	var output strings.Builder
	models := m.Schema.ModelNames()

	output.WriteString("BEGIN;\n\n")
	output.WriteString("SET CONSTRAINTS ALL DEFERRED;\n\n")

	// Get table dependencies to determine order
	tableDeps := make(map[string][]string)
	for _, model := range models {
		modelObj := m.Schema.FindModel(model)
		for _, field := range modelObj.Fields {
			if field.Type.Type == proto.Type_TYPE_MODEL && field.ForeignKeyInfo != nil {
				tableDeps[model] = append(tableDeps[model], field.ForeignKeyInfo.RelatedModelName)
			}
		}
	}

	// Sort models based on dependencies (tables with no dependencies first)
	sortedModels := make([]string, 0, len(models))
	visited := make(map[string]bool)
	var visit func(string)
	visit = func(model string) {
		if visited[model] {
			return
		}
		visited[model] = true
		for _, dep := range tableDeps[model] {
			visit(dep)
		}
		sortedModels = append(sortedModels, model)
	}
	for _, model := range models {
		visit(model)
	}

	// Reverse the order to get tables with no dependencies first
	for i, j := 0, len(sortedModels)-1; i < j; i, j = i+1, j-1 {
		sortedModels[i], sortedModels[j] = sortedModels[j], sortedModels[i]
	}

	// Add keel_storage table to the end of the list
	sortedModels = append(sortedModels, "keel_storage")

	// Register type handlers
	typeHandlers := map[proto.Type]TypeHandler{
		proto.Type_TYPE_DURATION:  &DurationTypeHandler{},
		proto.Type_TYPE_FILE:      &FileTypeHandler{},
		proto.Type_TYPE_TIMESTAMP: &TimeTypeHandler{},
	}

	for i, model := range sortedModels {
		table := Identifier(model)

		// Check if keel_storage table exists before querying
		if model == "keel_storage" {
			exists, err := m.database.ExecuteQuery(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'keel_storage')")
			if err != nil {
				return "", fmt.Errorf("error checking if keel_storage table exists: %w", err)
			}
			if !exists.Rows[0]["exists"].(bool) {
				continue
			}
		}

		// First get all columns to determine which ones to exclude
		result, err := m.database.ExecuteQuery(ctx, fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			return "", fmt.Errorf("error querying table %s: %w", table, err)
		}

		// Get column names, excluding computed fields
		columns := make([]string, 0)
		for _, col := range result.Columns {
			// For keel_storage table, include all columns without schema lookup
			if model == "keel_storage" {
				columns = append(columns, col)
				continue
			}
			field := proto.FindField(m.Schema.Models, model, casing.ToLowerCamel(col))
			if field != nil && field.ComputedExpression == nil {
				columns = append(columns, col)
			}
		}

		if len(result.Rows) == 0 {
			continue
		}

		if i > 0 {
			output.WriteString("\n")
		}

		insertStmt := fmt.Sprintf("INSERT INTO %s (\n    %s\n) VALUES\n",
			table,
			strings.Join(columns, ",\n    "),
		)
		output.WriteString(insertStmt)

		valueTuples := make([]string, len(result.Rows))

		for i, row := range result.Rows {
			values := make([]string, len(columns))

			for j, col := range columns {
				val := row[col]

				// For keel_storage table, use DefaultTypeHandler for all columns
				if model == "keel_storage" {
					handler := &DefaultTypeHandler{}
					// Special handling for the data column which is binary
					if col == "data" {
						// For binary data, we need to use the bytea type
						if val == nil {
							values[j] = "NULL"
						} else {
							// Convert the binary data to a hex string
							if bytes, ok := val.([]byte); ok {
								values[j] = fmt.Sprintf("'\\x%x'", bytes)
							} else {
								return "", fmt.Errorf("expected []byte for data column, got %T", val)
							}
						}
						continue
					}
					sql, err := handler.ToSQL(val, nil)
					if err != nil {
						return "", fmt.Errorf("error converting value for column %s: %w", col, err)
					}
					values[j] = sql
					continue
				}

				field := proto.FindField(m.Schema.Models, model, casing.ToLowerCamel(col))

				// Get the appropriate type handler
				var handler TypeHandler
				if field != nil {
					if field.Type.Repeated {
						handler = &ArrayTypeHandler{}
					} else {
						handler = typeHandlers[field.Type.Type]
					}
				}

				// If no specific handler, use default
				if handler == nil {
					handler = &DefaultTypeHandler{}
				}

				// Convert value to SQL
				sql, err := handler.ToSQL(val, field)
				if err != nil {
					return "", fmt.Errorf("error converting value for column %s: %w", col, err)
				}
				values[j] = sql
			}
			valueTuples[i] = fmt.Sprintf("    (%s)", strings.Join(values, ", "))
		}

		output.WriteString(strings.Join(valueTuples, ",\n"))
		output.WriteString(";\n")
	}

	output.WriteString("\n")
	output.WriteString("SET CONSTRAINTS ALL IMMEDIATE;\n\n")
	output.WriteString("COMMIT;\n")

	return output.String(), nil
}

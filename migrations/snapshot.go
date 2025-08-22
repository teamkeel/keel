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

// DefaultTypeHandler handles basic types using PostgreSQL's to_json.
type DefaultTypeHandler struct{}

func (h *DefaultTypeHandler) ToSQL(value interface{}, field *proto.Field) (string, error) {
	if value == nil {
		return "NULL", nil
	}

	switch v := value.(type) {
	case string:
		if field != nil && field.GetType().GetType() == proto.Type_TYPE_OBJECT {
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
		output := make(map[string]interface{})
		if url, ok := v["url"].(string); ok {
			output["url"] = strings.ReplaceAll(url, "'", "''")
		} else {
			output["url"] = v["url"]
		}
		output["contentType"] = v["contentType"]
		output["size"] = v["size"]

		jsonBytes, err := json.Marshal(output)
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
			if field != nil && field.GetType().GetRepeated() {
				switch field.GetType().GetType() {
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

	// Add keel_storage table
	models = append(models, "keel_storage")

	// Register type handlers
	typeHandlers := map[proto.Type]TypeHandler{
		proto.Type_TYPE_DURATION:  &DurationTypeHandler{},
		proto.Type_TYPE_FILE:      &FileTypeHandler{},
		proto.Type_TYPE_TIMESTAMP: &TimeTypeHandler{},
	}

	for i, entityName := range models {
		table := Identifier(entityName)

		// Check if keel_storage table exists before querying
		if entityName == "keel_storage" {
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

		if len(result.Rows) == 0 {
			continue
		}

		// Get column names, excluding computed fields and sequence fields
		columns := make([]string, 0)
		for _, col := range result.Columns {
			// For keel_storage table, include all columns without schema lookup
			if entityName == "keel_storage" {
				columns = append(columns, col)
				continue
			}

			entity := m.Schema.FindEntity(entityName)
			field := entity.FindField(casing.ToLowerCamel(col))
			if field != nil && field.GetComputedExpression() == nil && field.GetSequence() == nil {
				columns = append(columns, col)
			}
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
				// But the data column is binary, so we need to handle it differently
				if entityName == "keel_storage" {
					if col == "data" {
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
					} else {
						// For other keel_storage columns, use DefaultTypeHandler
						sql, err := (&DefaultTypeHandler{}).ToSQL(val, nil)
						if err != nil {
							return "", fmt.Errorf("error converting value for column %s: %w", col, err)
						}
						values[j] = sql
					}
					continue
				}

				entity := m.Schema.FindEntity(entityName)
				field := entity.FindField(casing.ToLowerCamel(col))

				var handler TypeHandler
				handler = &DefaultTypeHandler{}
				if field != nil {
					if field.GetType().GetRepeated() {
						handler = &ArrayTypeHandler{}
					} else if h, ok := typeHandlers[field.GetType().GetType()]; ok {
						handler = h
					}
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

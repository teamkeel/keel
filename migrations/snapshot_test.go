package migrations

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/schema"
)

var testSchema = `
model User {
	fields {
		name Text
		age Number
		salary Decimal
		isActive Boolean
		birthDate Date
		lastLogin Timestamp
		metadata Text
		tags Text[]?
		scores Number[]
		ratings Decimal[]
		preferences Text[]
		favouriteDates Date[]
		loginTimes Timestamp[]
		avatar File
		workHours Duration
		breaks Duration[]
	}
}

model Post {
	fields {
		title Text
		content Text
		author User
		isPublished Boolean @default(false)
		publishedAt Timestamp?
		views Number @default(0)
		rating Decimal @default(0.0)
		wordCount Number @computed(1)
		permalink Text @sequence("POST_")
		categories Text[]
		metadata Text
	}
}

model Comment {
	fields {
		content Text
		user User
		post Post
	}
}
`

func setupTestDatabase(t *testing.T) (*Migrations, func()) {
	// Create a simple schema for testing
	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(testSchema, config.Empty)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Create migrations instance
	ctx := context.Background()

	// Use the docker compose database
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	// Connect to the main database to create a test database
	mainDB, err := sql.Open("pgx/v5", dbConnInfo.String())
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer mainDB.Close()

	databaseName := strings.ToLower("keel_test_" + strings.ReplaceAll(t.Name(), "/", "_"))

	// Drop the database if it already exists
	_, err = mainDB.Exec("DROP DATABASE if exists " + databaseName)
	if err != nil {
		t.Fatalf("failed to drop database: %v", err)
	}

	// Create the database
	_, err = mainDB.Exec("CREATE DATABASE " + databaseName)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Connect to the test database
	database, err := db.New(ctx, dbConnInfo.WithDatabase(databaseName).String())
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create migrations instance
	migrations, err := New(ctx, schema, database)
	if err != nil {
		t.Fatalf("failed to create migrations: %v", err)
	}

	// Apply migrations
	err = migrations.Apply(ctx, false)
	if err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Insert test data
	testData := []string{
		`INSERT INTO "user" (
			name, 
			age, 
			salary, 
			is_active, 
			birth_date, 
			last_login, 
			metadata, 
			tags, 
			scores, 
			ratings, 
			preferences, 
			favourite_dates, 
			login_times,
			avatar,
			work_hours,
			breaks
		) VALUES 
		(
			'John Doe', 
			30, 
			50000.50, 
			true, 
			'1993-05-15', 
			'2024-03-24T13:39:32.805932Z', 
			'{"role": "admin"}', 
			ARRAY['tag1', 'tag2'], 
			ARRAY[85, 92, 78]::integer[], 
			ARRAY[4.5, 4.8, 4.2]::decimal[], 
			ARRAY['{"color": "blue"}', '{"color": "green"}'], 
			ARRAY['2024-01-01', '2024-12-25']::date[], 
			ARRAY['2024-03-24T10:00:00Z', '2024-03-24T11:00:00Z']::timestamp[],
			'{"url": "https://example.com/avatar1.jpg", "contentType": "image/jpeg", "size": 1024}',
			'PT8H',
			ARRAY['PT15M', 'PT30M']::interval[]
		),
		(
			'Jane Smith', 
			25, 
			45000.75, 
			true, 
			'1998-08-20', 
			'2024-03-24T13:39:32.805932Z', 
			'{"role": "user"}', 
			ARRAY['tag3'], 
			ARRAY[90, 88]::integer[], 
			ARRAY[4.7, 4.6]::decimal[], 
			ARRAY['{"color": "red"}'], 
			ARRAY['2024-02-14']::date[], 
			ARRAY['2024-03-24T09:00:00Z']::timestamp[],
			'{"url": "https://example.com/avatar2.jpg", "contentType": "image/png", "size": 2048}',
			'PT7H30M',
			ARRAY['PT20M']::interval[]
		);`,
		`INSERT INTO "post" (
			title, 
			content, 
			author_id, 
			is_published, 
			published_at, 
			views, 
			rating, 
			categories, 
			metadata
		) VALUES 
		(
			'First Post', 
			'This is a test post', 
			(SELECT id FROM "user" WHERE name = 'John Doe'), 
			true, 
			'2024-03-24T13:39:32.805932Z', 
			100, 
			4.5, 
			ARRAY['tech', 'programming'], 
			'{"tags": ["featured"]}'
		),
		(
			'Second Post', 
			'Another test post', 
			(SELECT id FROM "user" WHERE name = 'Jane Smith'), 
			false, 
			NULL, 
			50, 
			4.2, 
			ARRAY['lifestyle'], 
			'{"tags": []}'
		);`,
		`INSERT INTO "user" (
			name,
			age,
			salary,
			is_active,
			birth_date,
			last_login,
			metadata,
			tags,
			scores,
			ratings,
			preferences,
			favourite_dates,
			login_times,
			avatar,
			work_hours,
			breaks
		) VALUES (
			'O''Connor',
			35,
			60000.00,
			true,
			'1989-01-01',
			'2024-03-24T13:39:32.805932Z',
			'{"quote": "''"}',
			ARRAY['tag''4'],
			ARRAY[95]::integer[],
			ARRAY[4.9]::decimal[],
			ARRAY['{"color": "yellow"}'],
			ARRAY['2024-03-01']::date[],
			ARRAY['2024-03-24T12:00:00Z']::timestamp[],
			'{"url": "https://example.com/avatar3.jpg", "contentType": "image/gif", "size": 512}',
			'PT9H',
			ARRAY['PT45M']::interval[]
		);`,
		`INSERT INTO "comment" (
			content,
			user_id,
			post_id
		) VALUES 
		(
			'Great post!',
			(SELECT id FROM "user" WHERE name = 'Jane Smith'),
			(SELECT id FROM "post" WHERE title = 'First Post')
		),
		(
			'Thanks for sharing!',
			(SELECT id FROM "user" WHERE name = 'John Doe'),
			(SELECT id FROM "post" WHERE title = 'First Post')
		),
		(
			'Interesting perspective',
			(SELECT id FROM "user" WHERE name = 'John Doe'),
			(SELECT id FROM "post" WHERE title = 'Second Post')
		);`,
	}

	for _, sql := range testData {
		if _, err := database.ExecuteStatement(ctx, sql); err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}
	}

	return migrations, func() { database.Close() }
}

func TestSnapshotDatabase(t *testing.T) {
	// Setup test database
	migrations, cleanup := setupTestDatabase(t)
	defer cleanup()

	t.Run("generates_correct_SQL_structure", func(t *testing.T) {
		sql, err := migrations.SnapshotDatabase(context.Background())
		require.NoError(t, err)

		// Check for transaction and constraint management
		require.Contains(t, sql, "BEGIN;")
		require.Contains(t, sql, "SET CONSTRAINTS ALL DEFERRED;")
		require.Contains(t, sql, "SET CONSTRAINTS ALL IMMEDIATE;")
		require.Contains(t, sql, "COMMIT;")
	})

	t.Run("excludes_computed_fields", func(t *testing.T) {
		sql, err := migrations.SnapshotDatabase(context.Background())
		require.NoError(t, err)

		// Post table should not include wordCount field
		require.NotContains(t, sql, "wordCount")
	})

	t.Run("excludes_sequence_fields", func(t *testing.T) {
		sql, err := migrations.SnapshotDatabase(context.Background())
		require.NoError(t, err)

		// Post table should not include permalink field
		require.NotContains(t, sql, "permalink")
	})

	t.Run("handles_special_data_types", func(t *testing.T) {
		sql, err := migrations.SnapshotDatabase(context.Background())
		require.NoError(t, err)

		// Check for proper handling of JSON
		require.Contains(t, sql, "'{\"role\": \"admin\"}'")

		// Check for proper handling of arrays with correct types
		require.Contains(t, sql, "ARRAY['tag1', 'tag2']")                                     // text array
		require.Contains(t, sql, "ARRAY['85', '92', '78']::integer[]")                        // integer array
		require.Contains(t, sql, "ARRAY['4.5', '4.8', '4.2']::decimal[]")                     // decimal array
		require.Contains(t, sql, "ARRAY['2024-01-01', '2024-12-25']::date[]")                 // date array
		require.Contains(t, sql, "'{\"2024-03-24 10:00:00+00\",\"2024-03-24 11:00:00+00\"}'") // timestamp array
		require.Contains(t, sql, "ARRAY['PT15M', 'PT30M']::interval[]")                       // interval array
	})

	t.Run("handles_special_characters_in_strings", func(t *testing.T) {
		ctx := context.Background()

		// Create a test user with special characters
		_, err := migrations.database.ExecuteStatement(ctx, `
			INSERT INTO "user" (
				name,
				age,
				salary,
				is_active,
				birth_date,
				last_login,
				metadata,
				tags,
				scores,
				ratings,
				preferences,
				favourite_dates,
				login_times,
				avatar,
				work_hours,
				breaks,
				id,
				created_at,
				updated_at
			) VALUES (
				'Test O''''Connor',
				'35',
				60000,
				true,
				'1989-01-01T00:00:00Z',
				'2024-03-24T12:00:00Z',
				'{"quote": "''"}',
				ARRAY['tag''4'],
				ARRAY['95']::integer[],
				ARRAY['4.9']::decimal[],
				ARRAY['{"color": "yellow"}']::jsonb[],
				ARRAY['2024-03-01']::date[],
				ARRAY['2024-03-24 12:00:00+00']::timestamp[],
				'{"url": "https://example.com/avatar''''test.jpg", "contentType": "image/jpeg", "size": 1024}',
				'PT8H30M',
				ARRAY['PT15M', 'PT30M']::interval[],
				'2ulZOW3TTZ1r6uImq3zuBVPqqKG',
				'2025-03-24T14:04:49.200409Z',
				'2025-03-24T14:04:49.200409Z'
			)
		`)
		require.NoError(t, err)

		// Get the SQL
		sql, err := migrations.SnapshotDatabase(ctx)
		require.NoError(t, err)

		// Verify the SQL contains the escaped values
		require.Contains(t, sql, "Test O''''Connor")
		require.Contains(t, sql, "tag''4")
		require.Contains(t, sql, `{"quote": "''"}`)
		require.Contains(t, sql, `{"url": "https://example.com/avatar''''test.jpg", "size": 1024, "contentType": "image/jpeg"}`)
		require.Contains(t, sql, "ARRAY['PT15M', 'PT30M']::interval[]")

		// Clean up
		_, err = migrations.database.ExecuteStatement(ctx, `DELETE FROM "user" WHERE name IN ('Test O''''Connor', 'O''''Connor')`)
		require.NoError(t, err)
	})

	t.Run("can_be_reapplied_to_recreate_the_database_state", func(t *testing.T) {
		ctx := context.Background()

		// Get the initial snapshot
		initialSnapshot, err := migrations.SnapshotDatabase(ctx)
		require.NoError(t, err)

		// Clear the database
		_, err = migrations.database.ExecuteStatement(ctx, `TRUNCATE TABLE "comment", "post", "user" CASCADE;`)
		require.NoError(t, err)

		// Reapply the snapshot
		_, err = migrations.database.ExecuteStatement(ctx, initialSnapshot)
		require.NoError(t, err)

		// Verify the data was restored by checking counts
		result, err := migrations.database.ExecuteQuery(ctx, `SELECT COUNT(*) FROM "user";`)
		require.NoError(t, err)
		userCount := result.Rows[0]["count"]
		switch v := userCount.(type) {
		case int64:
			require.Equal(t, int64(3), v)
		case float64:
			require.Equal(t, float64(3), v)
		default:
			t.Fatalf("unexpected type for count: %T", userCount)
		}

		result, err = migrations.database.ExecuteQuery(ctx, `SELECT COUNT(*) FROM "post";`)
		require.NoError(t, err)
		postCount := result.Rows[0]["count"]
		switch v := postCount.(type) {
		case int64:
			require.Equal(t, int64(2), v)
		case float64:
			require.Equal(t, float64(2), v)
		default:
			t.Fatalf("unexpected type for count: %T", postCount)
		}

		result, err = migrations.database.ExecuteQuery(ctx, `SELECT COUNT(*) FROM "comment";`)
		require.NoError(t, err)
		commentCount := result.Rows[0]["count"]
		switch v := commentCount.(type) {
		case int64:
			require.Equal(t, int64(3), v)
		case float64:
			require.Equal(t, float64(3), v)
		default:
			t.Fatalf("unexpected type for count: %T", commentCount)
		}

		// Take another snapshot and verify it matches the initial one
		finalSnapshot, err := migrations.SnapshotDatabase(ctx)
		require.NoError(t, err)

		// Compare the snapshots
		require.Equal(t, initialSnapshot, finalSnapshot, "snapshots before and after reapplication should be identical")
	})

	t.Run("handles_null_array_values", func(t *testing.T) {
		ctx := context.Background()

		// Insert a user with NULL array fields
		_, err := migrations.database.ExecuteStatement(ctx, `
			INSERT INTO "user" (
				name,
				age,
				salary,
				is_active,
				birth_date,
				last_login,
				metadata,
				tags,
				scores,
				ratings,
				preferences,
				favourite_dates,
				login_times,
				avatar,
				work_hours,
				breaks
			) VALUES (
				'Null Array User',
				'35',
				60000,
				true,
				'1989-01-01T00:00:00Z',
				'2024-03-24T12:00:00Z',
				'{"quote": "''"}',
				NULL,
				ARRAY['95']::integer[],
				ARRAY['4.9']::decimal[],
				ARRAY['{"color": "yellow"}']::jsonb[],
				ARRAY['2024-03-01']::date[],
				ARRAY['2024-03-24 12:00:00+00']::timestamp[],
				'{"url": "https://example.com/avatar''''test.jpg", "contentType": "image/jpeg", "size": 1024}',
				'PT8H30M',
				ARRAY['PT15M', 'PT30M']::interval[]
			)
		`)
		require.NoError(t, err)

		// Get snapshot
		sql, err := migrations.SnapshotDatabase(ctx)
		require.NoError(t, err)

		// Verify NULL arrays are represented as NULL in the snapshot
		require.Contains(t, sql, "Null Array User")
		// The snapshot should have NULL values for the array fields, not error out

		// Clear the database
		_, err = migrations.database.ExecuteStatement(ctx, `TRUNCATE TABLE "comment", "post", "user" CASCADE;`)
		require.NoError(t, err)

		// Reapply the snapshot to ensure NULL arrays can be restored
		_, err = migrations.database.ExecuteStatement(ctx, sql)
		require.NoError(t, err)

		// Verify the user with NULL arrays was restored
		result, err := migrations.database.ExecuteQuery(ctx, `SELECT name, tags, scores, ratings FROM "user" WHERE name = 'Null Array User'`)
		require.NoError(t, err)
		require.Len(t, result.Rows, 1)
		require.Equal(t, "Null Array User", result.Rows[0]["name"])
		require.Nil(t, result.Rows[0]["tags"])
	})

	t.Run("handles_keel_storage_table", func(t *testing.T) {
		ctx := context.Background()

		// Create keel_storage table if it doesn't exist
		_, err := migrations.database.ExecuteStatement(ctx, `
			CREATE TABLE IF NOT EXISTS keel_storage (
				id text NOT NULL DEFAULT ksuid(),
				filename text NOT NULL,
				content_type text NOT NULL,
				data bytea NOT NULL,
				created_at timestamptz NOT NULL DEFAULT now(),
				PRIMARY KEY (id)
			);
		`)
		require.NoError(t, err)

		// Insert test data with binary content
		testData := []byte("Hello, World!")
		_, err = migrations.database.ExecuteStatement(ctx, `
			INSERT INTO keel_storage (id, filename, content_type, data)
			VALUES ('test1', 'test.txt', 'text/plain', $1)
		`, testData)
		require.NoError(t, err)

		// Get snapshot
		sql, err := migrations.SnapshotDatabase(ctx)
		require.NoError(t, err)

		// Clear all tables
		_, err = migrations.database.ExecuteStatement(ctx, `
			TRUNCATE TABLE keel_storage, "user", "post", "comment" CASCADE;
		`)
		require.NoError(t, err)

		// Apply the snapshot
		_, err = migrations.database.ExecuteStatement(ctx, sql)
		require.NoError(t, err)

		// Verify data was restored correctly by comparing the actual bytes
		result, err := migrations.database.ExecuteQuery(ctx, `SELECT data FROM keel_storage WHERE id = 'test1'`)
		require.NoError(t, err)
		require.Len(t, result.Rows, 1)
		require.Equal(t, testData, result.Rows[0]["data"].([]byte))
	})
}

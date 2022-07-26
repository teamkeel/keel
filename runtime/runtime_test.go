package runtime

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dbConnString = "host=localhost port=8001 user=postgres password=postgres dbname=%s sslmode=disable"

func TestRuntime(t *testing.T) {
	// We connect to the "main" database here only so we can create a new database
	// for each sub-test
	mainDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, "keel")),
		&gorm.Config{})
	require.NoError(t, err)

	// Make a database name for this test
	re := regexp.MustCompile(`[^\w]`)
	dbName := strings.ToLower(re.ReplaceAllString("experimentTest", ""))

	// Drop the database if it already exists. The normal dropping of it at the end of the
	// test case is bypassed if you quit a debug run of the test in VS Code.
	require.NoError(t, mainDB.Exec("DROP DATABASE if exists "+dbName).Error)

	// Create the database and drop at the end of the test
	err = mainDB.Exec("CREATE DATABASE " + dbName).Error
	require.NoError(t, err)
	defer func() {
		require.NoError(t, mainDB.Exec("DROP DATABASE "+dbName).Error)
	}()

	// Connect to the newly created test database and close connection
	// at the end of the test. We need to explicitly close the connection
	// so the mainDB connection can drop the database.
	testDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, dbName)),
		&gorm.Config{})
	require.NoError(t, err)
	defer func() {
		conn, err := testDB.DB()
		require.NoError(t, err)
		conn.Close()
	}()

	schema := protoSchema(t, experimentalSchema)

	// Migrate the database to this schema, in readiness for the Create Action.
	m := migrations.New(schema, nil)
	require.NoError(t, m.Apply(testDB))

	// todo This call should'nt need testDB
	handler := NewHandler(testDB, schema)
	reqBody := queryAsJSONPayload(t, experimentMutation, map[string]any{"name": "fred"})
	request := Request{
		Context: runtimectx.NewContext(testDB),
		URL: url.URL{
			Path: "/Test",
		},
		Body: []byte(reqBody),
	}

	response, err := handler(&request)
	respString := string(response.Body)
	_ = respString
	require.NoError(t, err)

	_ = response

	// Do some assertions
}

func protoSchema(t *testing.T, s string) *proto.Schema {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromInputs(&reader.Inputs{
		SchemaFiles: []reader.SchemaFile{
			{
				Contents: s,
			},
		},
	})
	require.NoError(t, err)
	return schema
}

var experimentalSchema string = `
model Person {
    fields {
        name Text
    }

    operations {
        get getPerson(id)
        create createPerson() with (name)
    }
}

api Test {
    @graphql

    models {
        Person
    }
}
`

const experimentMutation string = `
mutation someMutation($name: String!) {
	createPerson(input: {name: $name}) {
	  name
	}
}
`

func queryAsJSONPayload(t *testing.T, mutationString string, vars map[string]any) (asJSON string) {
	d := map[string]any{
		"query":     mutationString,
		"variables": vars,
	}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	return string(b)
}

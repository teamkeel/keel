package jsonschema_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/schema"
)

func TestJSONSchemaGeneration(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	type testCase struct {
		keelSchema string
		jsonSchema string
	}

	cases := map[string]*testCase{}

	for _, e := range entries {
		caseName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		c, ok := cases[caseName]
		if !ok {
			c = &testCase{}
			cases[caseName] = c
		}
		b, err := os.ReadFile(filepath.Join("./testdata", e.Name()))
		require.NoError(t, err)
		switch filepath.Ext(e.Name()) {
		case ".keel":
			c.keelSchema = string(b)
		case ".json":
			// bit of a dance here just to make sure the JSON from the fixture file
			// is formatted the same as the JSON from the jsonschema package
			m := jsonschema.JSONSchema{}
			err = json.Unmarshal(b, &m)
			require.NoError(t, err)
			b, err = json.Marshal(m)
			require.NoError(t, err)
			c.jsonSchema = string(b)
		}
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := schema.Builder{}
			schema, err := builder.MakeFromString(c.keelSchema)
			require.NoError(t, err)

			action := proto.FindAction(schema, "testAction")
			require.NotNil(t, action, "action with name testAction could not be found")

			jsonSchema := jsonschema.JSONSchemaForActionInput(context.Background(), schema, action)
			jsonSchemaBytes, err := json.Marshal(jsonSchema)
			require.NoError(t, err)

			opts := jsondiff.DefaultConsoleOptions()
			diff, explanation := jsondiff.Compare([]byte(c.jsonSchema), jsonSchemaBytes, &opts)

			if diff != jsondiff.FullMatch {
				t.Errorf("actual JSON schema does not match expected: %s", explanation)

				fmt.Println("Actual:")
				fmt.Print(string(jsonSchemaBytes))
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	type fixture struct {
		name    string
		opName  string
		request string
		errors  map[string]string
	}

	type fixtureGroup struct {
		name   string
		schema string
		cases  []fixture
	}

	fixtures := []fixtureGroup{
		{
			name: "get action",
			schema: `
				model Person {
					fields {
						name Text @unique
					}
					actions {
						get getPerson(id)
						get getBestBeatle() {
							@where(person.name == "John Lennon")
						}
					}
				}
			`,
			cases: []fixture{
				{
					name:    "valid - with input",
					request: `{"id": "1234"}`,
					opName:  "getPerson",
				},
				{
					name:    "valid - without input",
					request: `{}`,
					opName:  "getBestBeatle",
				},

				// errors
				{
					name:    "missing input",
					request: `{}`,
					opName:  "getPerson",
					errors: map[string]string{
						"(root)": "id is required",
					},
				},
				{
					name:    "wrong type",
					request: `{"id": 1234}`,
					opName:  "getPerson",
					errors: map[string]string{
						"id": "Invalid type. Expected: string, given: integer",
					},
				},
				{
					name:    "null",
					request: `{"id": null}`,
					opName:  "getPerson",
					errors: map[string]string{
						"id": "Invalid type. Expected: string, given: null",
					},
				},
				{
					name:    "valid inputs with additional properties",
					request: `{"id": "1234", "foo": "bar"}`,
					opName:  "getPerson",
					errors: map[string]string{
						"(root)": "Additional property foo is not allowed",
					},
				},
				{
					name:    "additional properties when no inputs expected",
					request: `{"id": "1234"}`,
					opName:  "getBestBeatle",
					errors: map[string]string{
						"(root)": "Additional property id is not allowed",
					},
				},
			},
		},
		{
			name: "delete action",
			schema: `
				model Person {
					fields {
						name Text @unique
					}
					actions {
						delete deletePerson(id)
						delete deleteBob() {
							@where(person.name == "Bob")
						}
					}
				}
			`,
			cases: []fixture{
				{
					name:    "valid - with input",
					request: `{"id": "1234"}`,
					opName:  "deletePerson",
				},
				{
					name:    "valid - without input",
					request: `{}`,
					opName:  "deleteBob",
				},

				// errors
				{
					name:    "missing input",
					request: `{}`,
					opName:  "deletePerson",
					errors: map[string]string{
						"(root)": "id is required",
					},
				},
				{
					name:    "wrong type",
					request: `{"id": 1234}`,
					opName:  "deletePerson",
					errors: map[string]string{
						"id": "Invalid type. Expected: string, given: integer",
					},
				},
				{
					name:    "null",
					request: `{"id": null}`,
					opName:  "deletePerson",
					errors: map[string]string{
						"id": "Invalid type. Expected: string, given: null",
					},
				},
				{
					name:    "valid inputs with additional properties",
					request: `{"id": "1234", "foo": "bar"}`,
					opName:  "deletePerson",
					errors: map[string]string{
						"(root)": "Additional property foo is not allowed",
					},
				},
				{
					name:    "additional properties when no inputs expected",
					request: `{"id": "1234"}`,
					opName:  "deleteBob",
					errors: map[string]string{
						"(root)": "Additional property id is not allowed",
					},
				},
			},
		},
		{
			name: "create action",
			schema: `
				enum Hobby {
					Tennis
					Chess
				}
				model Person {
					fields {
						name Text
						birthday Date?
						hobby Hobby?
					}
					actions {
						create createPerson() with (name)
						create createPersonWithDob() with (name, birthday)
						create createPersonWithOptionalDob() with (name, birthday?)
						create createPersonWithEnum() with (hobby) {
							@set(person.name = "")
						}
					}
				}
			`,
			cases: []fixture{
				{
					name:    "valid - basic",
					request: `{"name": "Jon"}`,
					opName:  "createPerson",
				},
				{
					name:    "valid - input for optional field provided",
					request: `{"name": "Jon", "birthday": "1986-03-18"}`,
					opName:  "createPersonWithDob",
				},
				{
					name:    "valid - input for optional field provided as null",
					request: `{"name": "Jon", "birthday": null}`,
					opName:  "createPersonWithDob",
				},
				{
					name:    "valid - ommitting optional input",
					request: `{"name": "Jon"}`,
					opName:  "createPersonWithOptionalDob",
				},
				{
					name:    "valid - providing optional input for optional field as null",
					request: `{"name": "Jon", "birthday": null}`,
					opName:  "createPersonWithOptionalDob",
				},
				{
					name:    "valid - providing optional input for optional field",
					request: `{"name": "Jon", "birthday": "1986-03-18"}`,
					opName:  "createPersonWithOptionalDob",
				},
				{
					name:    "valid - providing optional input for optional enum field as null",
					request: `{"hobby": null}`,
					opName:  "createPersonWithEnum",
				},
				{
					name:    "valid - providing optional input for optional enum field",
					request: `{"hobby": "Chess"}`,
					opName:  "createPersonWithEnum",
				},

				// errors
				{
					name:    "missing input",
					request: `{}`,
					opName:  "createPerson",
					errors: map[string]string{
						"(root)": "name is required",
					},
				},
				{
					name:    "missing required enum on optional field",
					request: `{}`,
					opName:  "createPersonWithEnum",
					errors: map[string]string{
						"(root)": "hobby is required",
					},
				},
				{
					name:    "missing required input on optional field",
					request: `{"name": "Jon"}`,
					opName:  "createPersonWithDob",
					errors: map[string]string{
						"(root)": "birthday is required",
					},
				},
				{
					name:    "null",
					request: `{"name": null}`,
					opName:  "createPerson",
					errors: map[string]string{
						"name": "Invalid type. Expected: string, given: null",
					},
				},
				{
					name:    "wrong type",
					request: `{"name": 1234}`,
					opName:  "createPerson",
					errors: map[string]string{
						"name": "Invalid type. Expected: string, given: integer",
					},
				},
				{
					name:    "wrong format for date",
					request: `{"name": "Jon", "birthday": "18th March 1986"}`,
					opName:  "createPersonWithDob",
					errors: map[string]string{
						"birthday": "Does not match format 'date'",
					},
				},
				{
					name:    "providing ISO8601 format for a Date",
					request: `{"name": "Jon", "birthday": "1986-03-18T00:00:00.000Z"}`,
					opName:  "createPersonWithOptionalDob",
					errors: map[string]string{
						"birthday": "Does not match format 'date'",
					},
				},
				{
					name:    "additional properties",
					request: `{"name": "Bob", "foo": "bar"}`,
					opName:  "createPerson",
					errors: map[string]string{
						"(root)": "Additional property foo is not allowed",
					},
				},
			},
		},
		{
			name: "update action",
			schema: `
				model Person {
					fields {
						identity Identity @unique
						name Text
						nickName Text?
					}
					actions {
						update updateName(id) with (name)
						update updateNameAndNickname(id) with (name, nickName)
						update updateNameOrNickname(id) with (name?, nickName?)
						update updateMyPerson() {
							@where(person.identity == ctx.identity)
							@set(person.name = "Hello")
						}
						update updateMyPersonWithName() with (name) {
							@where(person.identity == ctx.identity)
						}
					}
				}
			`,
			cases: []fixture{
				{
					name:    "valid - one input",
					request: `{"where": {"id": "1234"}, "values": {"name": "Jon"}}`,
					opName:  "updateName",
				},
				{
					name:    "valid - two inputs",
					request: `{"where": {"id": "1234"}, "values": {"name": "Jon", "nickName": "Johnny"}}`,
					opName:  "updateNameAndNickname",
				},
				{
					name:    "valid - two inputs - null for optional field",
					request: `{"where": {"id": "1234"}, "values": {"name": "Jon", "nickName": null}}`,
					opName:  "updateNameAndNickname",
				},
				{
					name:    "valid - two inputs - both optional - both provided",
					request: `{"where": {"id": "1234"}, "values": {"name": "Jon", "nickName": "Johnny"}}`,
					opName:  "updateNameOrNickname",
				},
				{
					name:    "valid - two inputs - both optional - one provided",
					request: `{"where": {"id": "1234"}, "values": {"nickName": "Johnny"}}`,
					opName:  "updateNameOrNickname",
				},
				{
					name:    "valid - two inputs - both optional - neither provided",
					request: `{"where": {"id": "1234"}, "values": {}}`,
					opName:  "updateNameOrNickname",
				},
				{
					name:    "valid - no inputs - empty request is ok",
					request: `{}`,
					opName:  "updateMyPerson",
				},
				{
					name:    "valid - no inputs - empty where and values is ok",
					request: `{"where": {}, "values": {}}`,
					opName:  "updateMyPerson",
				},
				{
					name:    "valid - values but no where",
					request: `{"values": {"name": "Jon"}}`,
					opName:  "updateMyPersonWithName",
				},

				// errors
				{
					name:    "missing required value",
					request: `{"where": {"id": "1234"}, "values": {}}`,
					opName:  "updateName",
					errors: map[string]string{
						"values": "name is required",
					},
				},
				{
					name:    "missing required where",
					request: `{"where": {}, "values": {"name": "Jon"}}`,
					opName:  "updateName",
					errors: map[string]string{
						"where": "id is required",
					},
				},
				{
					name:    "incorrect type for value",
					request: `{"where": {"id": "1234"}, "values": {"name": true}}`,
					opName:  "updateName",
					errors: map[string]string{
						"values.name": "Invalid type. Expected: string, given: boolean",
					},
				},
				{
					name:    "incorrect type for where",
					request: `{"where": {"id": 1234}, "values": {"name": "Jon"}}`,
					opName:  "updateName",
					errors: map[string]string{
						"where.id": "Invalid type. Expected: string, given: integer",
					},
				},
			},
		},
		{
			name: "list action",
			schema: `
				enum Genre {
					Romance
					Horror
				}
				model Publisher {
					fields {
						name Text
						dateFounded Date?
					}
				}
				model Author {
					fields {
						publisher Publisher
						name Text
					}
				}
				model Book {
					fields {
						author Author
						title Text
						genre Genre
						price Number
						available Boolean
						releaseDate Date
					}
					actions {
						list listBooks(id?, title?, genre?, price?, available?, createdAt?, releaseDate?) {
							@sortable(title, genre)
						}
						list booksByTitleAndGenre(title: Text, genre: Genre, minPrice: Number?) {
							@where(book.title == title)
							@where(book.genre == genre)
							@where(book.price > minPrice)
						}
						list listBooksByPublisherName(author.publisher.name)
						list listBooksByPublisherDateFounded(author.publisher.dateFounded)
						list listBooksByPublisherOptionalDateFounded(author.publisher.dateFounded?)
					}
				}
			`,
			cases: []fixture{
				{
					name:    "valid - no inputs",
					opName:  "listBooks",
					request: `{"where": {}}`,
				},
				{
					name:    "valid - text equals",
					opName:  "listBooks",
					request: `{"where": {"title": {"equals": "Great Gatsby"}}}`,
				},
				{
					name:    "valid - text not equals",
					opName:  "listBooks",
					request: `{"where": {"title": {"notEquals": "Great Gatsby"}}}`,
				},
				{
					name:    "valid - text starts with",
					opName:  "listBooks",
					request: `{"where": {"title": {"startsWith": "Great Gatsby"}}}`,
				},
				{
					name:    "valid - text ends with",
					opName:  "listBooks",
					request: `{"where": {"title": {"startsWith": "Great Gatsby"}}}`,
				},
				{
					name:    "valid - text contains",
					opName:  "listBooks",
					request: `{"where": {"title": {"startsWith": "Great Gatsby"}}}`,
				},
				{
					name:    "valid - text one of",
					opName:  "listBooks",
					request: `{"where": {"title": {"oneOf": ["Great Gatsby", "Ulysses"]}}}`,
				},
				{
					name:    "valid - text multi",
					opName:  "listBooks",
					request: `{"where": {"title": {"startsWith": "Great", "endsWith": "Gatsby"}}}`,
				},
				{
					name:    "valid - enum equals",
					opName:  "listBooks",
					request: `{"where": {"genre": {"equals": "Romance"}}}`,
				},
				{
					name:    "valid - enum one of",
					opName:  "listBooks",
					request: `{"where": {"genre": {"oneOf": ["Romance", "Horror"]}}}`,
				},
				{
					name:    "valid - number equals",
					opName:  "listBooks",
					request: `{"where": {"price": {"equals": 10}}}`,
				},
				{
					name:    "valid - number not equals",
					opName:  "listBooks",
					request: `{"where": {"price": {"notEquals": 10}}}`,
				},
				{
					name:    "valid - number less than",
					opName:  "listBooks",
					request: `{"where": {"price": {"lessThan": 10}}}`,
				},
				{
					name:    "valid - number greater than",
					opName:  "listBooks",
					request: `{"where": {"price": {"greaterThan": 10}}}`,
				},
				{
					name:    "valid - number less than or equals",
					opName:  "listBooks",
					request: `{"where": {"price": {"lessThanOrEquals": 10}}}`,
				},
				{
					name:    "valid - number greater than or equals",
					opName:  "listBooks",
					request: `{"where": {"price": {"greaterThanOrEquals": 10}}}`,
				},
				{
					name:    "valid - boolean equals",
					opName:  "listBooks",
					request: `{"where": {"available": {"equals": true}}}`,
				},
				{
					name:    "valid - timestamp before",
					opName:  "listBooks",
					request: `{"where": {"createdAt": {"before": "2022-12-02T12:28:29.609Z"}}}`,
				},
				{
					name:    "valid - timestamp after",
					opName:  "listBooks",
					request: `{"where": {"createdAt": {"after": "2022-12-02T12:28:29.609Z"}}}`,
				},
				{
					name:    "valid - date equals",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"equals": "2022-12-02"}}}`,
				},
				{
					name:    "valid - date before",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"before": "2022-12-02"}}}`,
				},
				{
					name:    "valid - date on or before",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"onOrBefore": "2022-12-02"}}}`,
				},
				{
					name:    "valid - date after",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"after": "2022-12-02"}}}`,
				},
				{
					name:    "valid - date on or after",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"onOrAfter": "2022-12-02"}}}`,
				},
				{
					name:    "valid - id equals",
					opName:  "listBooks",
					request: `{"where": {"id": {"equals": "123456789"}}}`,
				},
				{
					name:    "valid - id one of",
					opName:  "listBooks",
					request: `{"where": {"id": {"oneOf": ["123456789"]}}}`,
				},
				{
					name:    "valid - empty orderby",
					opName:  "listBooks",
					request: `{"orderBy": []}`,
				},
				{
					name:    "valid - orderby title and genre",
					opName:  "listBooks",
					request: `{"orderBy": [{"title":"asc"},{"genre":"desc"}]}`,
				},
				{
					name:    "valid - non-query types",
					opName:  "booksByTitleAndGenre",
					request: `{"where": {"title": "Some title", "genre": "Horror", "minPrice": 10}}`,
				},
				{
					name:    "valid - nested model field",
					opName:  "listBooksByPublisherName",
					request: `{"where": {"author": { "publisher": { "name": {"equals": "Jim"}}}}}`,
				},
				{
					name:    "valid - missing optional where",
					opName:  "listBooksByPublisherOptionalDateFounded",
					request: `{"where": { }}`,
				},
				{
					name:    "valid - missing optional input",
					opName:  "listBooksByPublisherOptionalDateFounded",
					request: `{}`,
				},

				// errors
				{
					name:    "invalid sort direction",
					opName:  "listBooks",
					request: `{"orderBy": [{"title":"asc"},{"genre":"down"}]}`,
					errors: map[string]string{
						"orderBy.1":       `Must validate one and only one schema (oneOf)`,
						"orderBy.1.genre": `orderBy.1.genre must be one of the following: "asc", "desc"`,
					},
				},
				{
					name:    "nullable nested model list field as null",
					opName:  "listBooksByPublisherDateFounded",
					request: `{"where": {"author": { "publisher": { "dateFounded": null}}}}`,
					errors: map[string]string{
						"where.author.publisher.dateFounded": `Invalid type. Expected: object, given: null`,
					},
				},
				{
					name:    "missing required input targeting optional field",
					opName:  "listBooksByPublisherDateFounded",
					request: `{"where": { }}`,
					errors: map[string]string{
						"where": `author is required`,
					},
				},
				{
					name:    "optional inputs cannot be null when fields are required",
					opName:  "listBooks",
					request: `{"where": {"id": null, "title": null, "genre": null, "price": null, "available": null, "createdAt": null}}`,
					errors: map[string]string{
						"where.id":        `Invalid type. Expected: object, given: null`,
						"where.title":     `Invalid type. Expected: object, given: null`,
						"where.genre":     `Invalid type. Expected: object, given: null`,
						"where.price":     `Invalid type. Expected: object, given: null`,
						"where.available": `Invalid type. Expected: object, given: null`,
						"where.createdAt": `Invalid type. Expected: object, given: null`,
					},
				},
				{
					name:    "text unknown filter",
					opName:  "listBooks",
					request: `{"where": {"title": {"isSimilarTo": "Sci-fi"}}}`,
					errors: map[string]string{
						"where.title": `Additional property isSimilarTo is not allowed`,
					},
				},
				{
					name:    "enum equals not valid enum",
					opName:  "listBooks",
					request: `{"where": {"genre": {"equals": "Sci-fi"}}}`,
					errors: map[string]string{
						"where.genre.equals": `where.genre.equals must be one of the following: "Romance", "Horror", null`,
					},
				},
				{
					name:    "enum one of not valid enum",
					opName:  "listBooks",
					request: `{"where": {"genre": {"oneOf": ["Sci-fi"]}}}`,
					errors: map[string]string{
						"where.genre.oneOf.0": `where.genre.oneOf.0 must be one of the following: "Romance", "Horror"`,
					},
				},
				{
					name:    "timestamp invalid format",
					opName:  "listBooks",
					request: `{"where": {"createdAt": {"after": "not-a-date-time"}}}`,
					errors: map[string]string{
						"where.createdAt.after": `Does not match format 'date-time'`,
					},
				},
				{
					name:    "date invalid format with time component",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"after": "1986-03-18T00:00:00.000Z"}}}`,
					errors: map[string]string{
						"where.releaseDate.after": `Does not match format 'date'`,
					},
				},
				{
					name:    "date invalid format",
					opName:  "listBooks",
					request: `{"where": {"releaseDate": {"after": "not-a-date-time"}}}`,
					errors: map[string]string{
						"where.releaseDate.after": `Does not match format 'date'`,
					},
				},
				{
					name:    "using invalid query types for explicit filters",
					opName:  "booksByTitleAndGenre",
					request: `{"where": {"title": {"contains": "Some title"}, "genre": {"equals": "Romance"}}}`,
					errors: map[string]string{
						"where.title": `Invalid type. Expected: string, given: object`,
						"where.genre": `where.genre must be one of the following: "Romance", "Horror"`,
					},
				},
			},
		},
		{
			name:   "authenticate",
			schema: `model Whatever {}`,
			cases: []fixture{
				{
					name:    "valid",
					opName:  "authenticate",
					request: `{"emailPassword": {"email": "foo@bar.com", "password": "pa33w0rd"}}`,
				},
			},
		},
		{
			name: "arbitrary functions",
			schema: `
			    message In {}
				model Whatever {
					actions {
						read getWhatever(Any) returns(Any)
						read noInputs() returns(Any)
						read emptyInputMessage(In) returns(Any)
					}
				}
			`,
			cases: []fixture{
				{
					name:   "valid - object",
					opName: "getWhatever",
					request: `{
						"string": "hey",
						"number": 1,
						"simpleNumberArray": [1, 2, 3],
						"simpleStringArray": ["one", "two", "three"],
						"nestedObject": {
							"foo": "bar"
						},
						"arrayOfObjects": [
							{
								"foo": "bar"
							}
						]
					}`,
				},
				{
					name:   "valid - array",
					opName: "getWhatever",
					request: `
						[1, 2, 3]
					`,
				},
				{
					name:    "valid - string",
					opName:  "getWhatever",
					request: `"hello world"`,
				},
				{
					name:    "valid - number",
					opName:  "getWhatever",
					request: "1234",
				},
				{
					name:    "valid - boolean",
					opName:  "getWhatever",
					request: "true",
				},
				{
					name:    "valid - null",
					opName:  "getWhatever",
					request: "null",
				},
				{
					name:    "valid - object",
					opName:  "getWhatever",
					request: `{ "name": "Arnold" }`,
				},
				{
					name:    "valid - no arguments",
					opName:  "noInputs",
					request: "{}",
				},
				{
					name:    "valid - no arguments for empty message",
					opName:  "emptyInputMessage",
					request: "{}",
				},
			},
		},
	}

	for _, group := range fixtures {
		for _, f := range group.cases {
			group := group
			f := f
			t.Run(group.name+"/"+f.name, func(t *testing.T) {

				builder := schema.Builder{}
				schema, err := builder.MakeFromString(group.schema)
				require.NoError(t, err)

				var req any
				err = json.Unmarshal([]byte(f.request), &req)
				require.NoError(t, err)

				action := proto.FindAction(schema, f.opName)

				result, err := jsonschema.ValidateRequest(context.Background(), schema, action, req)
				require.NoError(t, err)
				require.NotNil(t, result)

				if len(f.errors) == 0 {
					assert.True(t, result.Valid(), "expected request to be valid")
				}

				for _, e := range result.Errors() {
					jsonPath := e.Field()
					expected, ok := f.errors[jsonPath]
					if !ok {
						assert.Fail(t, "unexpected error", "%s - %s", jsonPath, e.Description())
						continue
					}

					assert.Equal(t, expected, e.Description(), "error for path %s did not match expected", jsonPath)
					delete(f.errors, jsonPath)
				}

				// f.errors should now be empty, if not mark test as failed
				for path, description := range f.errors {
					assert.Fail(t, "expected error was not returned", "%s - %s", path, description)
				}
			})
		}
	}

}

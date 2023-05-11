package completions_test

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema/completions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"gopkg.in/yaml.v3"
)

type testCase struct {
	name     string
	schema   string
	expected []string
}

func TestRootCompletions(t *testing.T) {

	cases := []testCase{
		{
			name:     "top-level-keyword",
			schema:   "mod<Cursor>",
			expected: []string{"api", "enum", "message", "model", "role"},
		},
		{
			name: "top-level-keyword-not-first",
			schema: `
			model A {

            }

            m<Cursor>`,
			expected: []string{"api", "enum", "message", "model", "role"},
		},
		{
			name:     "top-level-keyword-whitespace",
			schema:   `<Cursor>`,
			expected: []string{"api", "enum", "message", "model", "role"},
		},
		{
			name: "top-level-keyword-whitespace-partial-schema",
			schema: `
			model A {}

			<Cursor>

			model B {}
			`,
			expected: []string{"api", "enum", "message", "model", "role"},
		},
	}

	runTestsCases(t, cases)
}

func TestModelCompletions(t *testing.T) {

	cases := []testCase{
		// name tests
		{
			name:     "model-name-no-completions",
			schema:   "model Per<Cursor>",
			expected: []string{},
		},
		{
			name: "model-name-completion",
			schema: `
			model Person {
				fields {
					author Author
				}
			}

			model A<Cursor>
			`,
			expected: []string{"Author"},
		},
		{
			name: "model-name-completion-predefined-enum",
			schema: `
			enum Author {

			}
			model Person {
				fields {
					author Author
				}
			}

			model A<Cursor>
			`,
			expected: []string{},
		},
		// block tests
		{
			name: "model-block-keywords",
			schema: `
			model A {
              f<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		{
			name: "model-block-keywords-whitespace",
			schema: `
			model A {
			  <Cursor>
			}`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		// attributes tests
		{
			name: "model-attributes",
			schema: `
			model A {
              @<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
	}

	runTestsCases(t, cases)
}

func TestFieldCompletions(t *testing.T) {

	cases := []testCase{
		// name tests
		{
			name: "field-name-no-completions",
			schema: `
			model A {
				fields {
					te<Cursor>
				}
			}
			`,
			expected: []string{},
		},
		{
			name: "field-name-no-completions-previous-optional",
			schema: `
			model A {
				fields {
					some Number?
					te<Cursor>
				}
			}
			`,
			expected: []string{},
		},
		{
			name: "field-name-no-completions-previous-list",
			schema: `
			model A {
				fields {
					some Number[]
					te<Cursor>
				}
			}
			`,
			expected: []string{},
		},
		{
			name: "field-name-no-completions-whitespace",
			schema: `
			model A {
				fields {
					<Cursor>
				}
			}
			`,
			expected: []string{},
		},
		// keyword tests
		{
			name: "fields-keyword",
			schema: `
			model A {
              fi<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		// type tests
		{
			name: "field-type",
			schema: `
			model Foo {
              fields {
                name Te<Cursor>
			  }
            }`,
			expected: []string{"Foo", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		{
			name: "field-type-whitespace",
			schema: `
			model Foo {
              fields {
                name <Cursor>
			  }
            }`,
			expected: []string{"Foo", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		{
			name: "field-type-previous-optional",
			schema: `
			model Foo {
              fields {
				optionalNumber Number?
				myField <Cursor>
			  }
            }`,
			expected: []string{"Foo", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		{
			name: "field-type-previous-list",
			schema: `
			model Foo {
              fields {
				optionalNumber Number[]
				myField <Cursor>
			  }
            }`,
			expected: []string{"Foo", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		{
			name: "field-type-complex",
			schema: `
			model Foo {
	          fields {
				optionalNumber Number[]
				things Text[]
				some Boolean @default(false)
				other Boolean {
					@default(1 != 1)
				}
				myField <Cursor>
			  }
            }`,
			expected: []string{"Foo", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		{
			name: "field-type-model",
			schema: `
			model Author {}
			
			model Book {
				fields {
					author Au<Cursor>
				}	
			}`,
			expected: []string{"Author", "Book", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		{
			name: "field-type-enum",
			schema: `
			model Book {
				fields {
					category Ca<Cursor>
				}
			}
			
			enum Category {
				Romance
				Horror	
			}`,
			expected: []string{"Book", "Category", "Identity", "ID", "Text", "Number", "Boolean", "Date", "Timestamp", "Secret", "Password"},
		},
		// attributes tests
		{
			name: "field-attributes",
			schema: `
			model A {
              fields {
                name Text @u<Cursor>
			  }
            }`,
			expected: []string{"@unique", "@default", "@relation"},
		},
		{
			name: "field-attributes-block",
			schema: `
			model A {
              fields {
                name Text {
                  @un<Cursor>
				}
			  }
            }`,
			expected: []string{"@unique", "@default", "@relation"},
		},
		{
			name: "field-attributes-block-whitespace",
			schema: `
			model A {
				fields {
					name Text {
						<Cursor>
					}
				}
			}`,
			expected: []string{"@unique", "@default", "@relation"},
		},
		{
			name: "field-attributes-bare-at",
			schema: `model Person {
				fields {
					name Text @<Cursor>
				}
			}`,
			expected: []string{"@unique", "@default", "@relation"},
		},
		{
			name: "field-attributes-whitespace",
			schema: `
			model Person {
				fields {
					name Text <Cursor>
				}
			}`,
			expected: []string{"@unique", "@default", "@relation"},
		},
	}

	runTestsCases(t, cases)
}

func TestOperationCompletions(t *testing.T) {
	cases := []testCase{
		// actions tests
		{
			name: "operations-action-type-completions",
			schema: `
			model A {
				operations {
					<Cursor>
				}
			}`,
			expected: parser.OperationActionTypes,
		},
		{
			name: "create-keyword",
			schema: `
			model A {
              operations {
                c<Cursor>
			  }
            }`,
			expected: parser.OperationActionTypes,
		},
		{
			name: "create-keyword-not-first",
			schema: `
			model A {
              operations {
                get getA(id)
                crea<Cursor>
			  }
            }`,
			expected: parser.OperationActionTypes,
		},
		{
			name: "get-keyword",
			schema: `
			model A {
              operations {
                get getA(id)
                g<Cursor>
			  }
            }`,
			expected: parser.OperationActionTypes,
		},
		// with tests
		{
			name: "with-keyword",
			schema: `
			model A {
              operations {
                create createA() wi<Cursor>
			  }
            }`,
			expected: []string{"with"},
		},
		{
			name: "with-keyword-whitespace",
			schema: `
			model A {
              operations {
                create createA() <Cursor>
			  }
	        }`,
			expected: []string{"with"},
		},
		{
			name: "with-end-of-line",
			schema: `
			model A {
			  fields {
				name Text
			  }
	          operations {
	            create createA() with (name) <Cursor>
			  }
            }`,
			expected: []string{},
		},
		// attribute tests
		{
			name: "action-attributes",
			schema: `
			model A {
			  fields {
				name Text
			  }
              operations {
                create createA() with (name) {
                  @se<Cursor>
				}
			  }
            }`,
			expected: []string{"@permission", "@set", "@validate", "@where"},
		},
		{
			name: "action-attributes-bare-at",
			schema: `
			model A {
			  fields {
				name Text
			  }
              operations {
                create createA() with (name) {
                  @<Cursor>
				}
			  }
            }`,
			expected: []string{"@permission", "@set", "@validate", "@where"},
		},
		{
			name: "action-attributes-whitespace",
			schema: `
			model A {
			  fields {
				name Text
			  }
              operations {
                create createA() with (name) {
                  <Cursor>
				}
			  }
            }`,
			expected: []string{"@permission", "@set", "@validate", "@where"},
		},
	}

	runTestsCases(t, cases)
}

func TestFunctionCompletions(t *testing.T) {

	cases := []testCase{
		// block tests
		{
			name: "functions-keyword",
			schema: `
			model A {
              fields {
				name Text
			  }

              fun<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		// name tests
		{
			name: "suggested-function-name-completion",
			schema: `
			model Post {
				fields {
					title Text
				}
			}
			model PostExtended {
				fields {
					title Text
				}
			
				functions {
					create c<Cursor>
				}
			}
			`,
			expected: []string{"createPostExtended"},
		},
		// action type tests
		{
			name: "arbitrary-function-action-type-completions",
			schema: `
			model Person {
				functions {
					<Cursor>
				}
			}
			`,
			expected: parser.FunctionActionTypes,
		},
		// input tests
		{
			name: "arbitrary-function-input-completions",
			schema: `
			message GetPersonInput {}
			model Person {
				functions {
					read getPerson(<Cursor>
				}
			}
			`,
			expected: []string{"GetPersonInput", "Any", "createdAt", "id", "updatedAt"},
		},
		// returns keyword tests
		{
			name: "arbitrary-function-returns-completions",
			schema: `
			message GetPersonInput {}
			message GetPersonResponse {}
			model Person {
				functions {
					read getPerson(GetPersonInput) returns(<Cursor>)
				}
			}
			`,
			expected: []string{"GetPersonInput", "GetPersonResponse"},
		},
		{
			name: "arbitrary-function-returns-keyword-completions",
			schema: `
			message GetPersonInput {}
			message GetPersonResponse {}
			model Person {
				functions {
					read getPerson(GetPersonInput) <Cursor>
				}
			}
			`,
			expected: []string{"returns"},
		},
		// Any keyword tests
		{
			name: "arbitrary-function-any-completions",
			schema: `
			model Person {
				functions {
					read getPerson(<Cursor>)
				}
			}
			`,
			expected: []string{"Any", "createdAt", "id", "updatedAt"},
		},
		{
			name: "arbitrary-function-create-with-completion",
			schema: `
			model Person {
				functions {
					create createPerson() <Cursor>
				}
			}`,
			expected: []string{"with"},
		},
		// With keyword tests
		{
			name: "arbitrary-function-create-partial-with-completion",
			schema: `
			model Person {
				functions {
					create createPerson() w<Cursor>
				}
			}`,
			expected: []string{"with"},
		},
		{
			name: "arbitrary-function-update-with-completion",
			schema: `
			model Person {
				functions {
					update updatePerson() <Cursor>
				}
			}`,
			expected: []string{"with"},
		},
		{
			name: "arbitrary-function-update-partial-with-completion",
			schema: `
			model Person {
				functions {
					update updatePerson() w<Cursor>
				}
			}`,
			expected: []string{"with"},
		},
	}

	runTestsCases(t, cases)
}

func TestPermissionCompletions(t *testing.T) {

	cases := []testCase{
		{
			name: "permission-attribute",
			schema: `
			model A {
              @p<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		{
			name: "model-permission-attribute-labels",
			schema: `
			model Person {
				@permission(<Cursor>)
			}	
			`,
			expected: []string{"expression", "roles", "actions"},
		},
		{
			name: "action-permission-attribute-labels",
			schema: `
			model Person {
				operations {
					create createPerson() {
						@permission(<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"expression", "roles"},
		},
		{
			name: "permission-attribute-expression",
			schema: `
			model Person {
				@permission(
					expression: p<Cursor>
				)
			}
			`,
			expected: []string{"person", "ctx", "env", "secrets"},
		},
		{
			name: "permission-attribute-expression-whitespace",
			schema: `
			model Person {
				@permission(
					expression: <Cursor>
				)
			}
			`,
			expected: []string{"person", "ctx", "env", "secrets"},
		},
		{
			name: "permission-attribute-model-fields",
			schema: `
			model Person {
				@permission(
					expression: person.<Cursor>
				)
			}
			`,
			expected: []string{"id", "createdAt", "updatedAt"},
		},
		{
			name: "permission-attribute-actions",
			schema: `
			model Person {
				@permission(
					actions: [<Cursor>]
				)
			}
			`,
			expected: parser.FunctionActionTypes,
		},
		{
			name: "permission-attribute-actions-many",
			schema: `
			model Person {
				@permission(
					actions: [get, update, cr<Cursor>]
				)
			}
			`,
			expected: parser.FunctionActionTypes,
		},
		{
			name: "permission-attribute-roles",
			schema: `
			model Person {
				@permission(
					roles: [<Cursor>]
				)
			}

			role Staff {}
			`,
			expected: []string{"Staff"},
		},
	}

	runTestsCases(t, cases)
}

func TestRoleCompletions(t *testing.T) {

	cases := []testCase{
		{
			name: "role-keyword",
			schema: `
			r<Cursor>`,
			expected: []string{"api", "enum", "message", "model", "role"},
		},
		{
			name: "domains-keyword",
			schema: `
			role Staff {
				dom<Cursor>	
			}`,
			expected: []string{"domains", "emails"},
		},
		{
			name: "emails-keyword",
			schema: `
			role Staff {
				e<Cursor>	
			}`,
			expected: []string{"domains", "emails"},
		},
		{
			name: "role-name-completion-nadda",
			schema: `
			model Person {
				fields {
					author Author
				}
			}

			role A<Cursor>
			`,
			expected: []string{},
		},
	}

	runTestsCases(t, cases)
}

func TestAPICompletions(t *testing.T) {

	cases := []testCase{
		{
			name: "api-keyword",
			schema: `
			a<Cursor>`,
			expected: []string{"api", "enum", "message", "model", "role"},
		},
		{
			name: "models-keyword",
			schema: `
			api Test {
				mo<Cursor>
			}
			`,
			expected: []string{"models"},
		},
		{
			name: "api-attributes",
			schema: `
			api Test {
				@<Cursor>
			}
			`,
			expected: []string{"models"},
		},
		{
			name: "api-model-names",
			schema: `
			model Person {}

			api Test {
				models {
					P<Cursor>
				}
			}
			`,
			expected: []string{"Person"},
		},
	}

	runTestsCases(t, cases)
}

func TestEnumCompletions(t *testing.T) {

	cases := []testCase{
		{
			name: "enum-name-completion",
			schema: `
			model Person {
				fields {
					author Author
				}
			}

			enum A<Cursor>
			`,
			expected: []string{"Author"},
		},
		{
			name: "enum-name-completion-predefined-model",
			schema: `
			model Author {

			}
			model Person {
				fields {
					author Author
				}
			}

			enum A<Cursor>
			`,
			expected: []string{},
		},
	}

	runTestsCases(t, cases)
}

func TestMessageCompletions(t *testing.T) {

	cases := []testCase{
		{
			name: "message-field-completions",
			schema: `
			message AnotherMessage {}
			message MyMessage {
				foo <Cursor>
			}
			`,
			expected: []string{"AnotherMessage", "Boolean", "Date", "ID", "Identity", "MyMessage", "Number", "Password", "Secret", "Text", "Timestamp"},
		},
	}

	runTestsCases(t, cases)
}

func runTestsCases(t *testing.T, cases []testCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var pos *node.Position

			// find position of <Cursor> marker
			lines := strings.Split(tc.schema, "\n")
			for i, line := range lines {
				idx := strings.Index(line, "<Cursor>")
				if idx == -1 {
					continue
				}
				pos = &node.Position{
					Line:   i + 1,
					Column: idx + 1,
				}
				break
			}

			if pos == nil {
				t.Fatal("no <Cursor> marker in schema")
			}

			// remove cursor marker from schema
			schema := strings.Replace(tc.schema, "<Cursor>", "", 1)

			dir, err := os.Getwd()
			assert.NoError(t, err)

			configFile, err := config.Load(dir + "/fixtures")
			assert.NoError(t, err)

			configString, err := yaml.Marshal(configFile)
			assert.NoError(t, err)

			results := completions.Completions(schema, pos, string(configString))
			values := []string{}
			for _, r := range results {
				values = append(values, r.Label)
			}

			// we don't care about order, just that the values match
			sort.Strings(tc.expected)
			sort.Strings(values)

			assert.EqualValues(t, tc.expected, values)
		})
	}
}

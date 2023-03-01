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

func TestCompletions(t *testing.T) {

	type testCase struct {
		name     string
		schema   string
		expected []string
	}

	cases := []testCase{
		{
			name:     "model-name-no-completions",
			schema:   "model Per<Cursor>",
			expected: []string{},
		},
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
		{
			name: "fields-keyword",
			schema: `
			model A {
              fi<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		{
			name: "model-attributes",
			schema: `
			model A {
              @<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
		{
			name: "permission-attribute",
			schema: `
			model A {
              @p<Cursor>
            }`,
			expected: []string{"@permission", "fields", "functions", "operations"},
		},
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
		{
			name: "field-attributes",
			schema: `
			model A {
              fields {
                name Text @u<Cursor>
			  }
            }`,
			expected: []string{"@unique", "@default"},
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
			expected: []string{"@unique", "@default"},
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
			expected: []string{"@unique", "@default"},
		},
		{
			name: "field-attributes-bare-at",
			schema: `model Person {
				fields {
					name Text @<Cursor>
				}
			}`,
			expected: []string{"@unique", "@default"},
		},
		{
			name: "field-attributes-whitespace",
			schema: `
			model Person {
				fields {
					name Text <Cursor>
				}
			}`,
			expected: []string{"@unique", "@default"},
		},
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
		{
			name: "create-keyword",
			schema: `
			model A {
              operations {
                c<Cursor>
			  }
            }`,
			expected: append([]string{"with"}, parser.OperationActionTypes...),
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
			expected: append([]string{"with"}, parser.OperationActionTypes...),
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
			expected: append([]string{"with"}, parser.OperationActionTypes...),
		},
		{
			name: "with-keyword",
			schema: `
			model A {
              operations {
                create createA() wi<Cursor>
			  }
            }`,
			expected: append([]string{"with"}, parser.OperationActionTypes...),
		},
		{
			name: "with-keyword-whitespace",
			schema: `
			model A {
              operations {
                create createA() <Cursor>
			  }
            }`,
			expected: append([]string{"with"}, parser.OperationActionTypes...),
		},
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
		{
			name: "actions-input-field-name",
			schema: `
			model A {
			  fields {
				name Text @unique
			  }
              operations {
                get getA(na<Cursor>)
			  }
            }`,
			expected: []string{"createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "actions-input-with-field-name",
			schema: `
			model A {
			  fields {
				name Text @unique
				nickName Text
			  }
              operations {
                create createA() with (n<Cursor>`,
			expected: []string{"createdAt", "id", "name", "nickName", "updatedAt"},
		},
		{
			name: "actions-input-not-first",
			schema: `
			model A {
			  fields {
				name Text @unique
				birthday Date
			  }
              operations {
                create createA() with (name, bi<Cursor>`,
			expected: []string{"birthday", "createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "actions-input-built-in-fields",
			schema: `
			model A {
              operations {
                get getA(i<Cursor>`,
			expected: []string{"createdAt", "id", "updatedAt"},
		},
		{
			name: "actions-input-long-form-type",
			schema: `
			model Person {
				fields {
					name Text
				}
				operations {
					create createPerson() with (name: Te<Cursor>)
				}
			}
			`,
			expected: []string{
				"Boolean", "Date", "ID", "Identity", "Number", "Text", "Timestamp", "Password", "Secret", "createdAt", "id", "name", "updatedAt",
			},
		},
		{
			name: "actions-input-nested-field",
			schema: `
			model Author {}

			model Book {
				fields {
					author Author
				}
				operations {
					list booksByAuthor(author.<Cursor>)
				}
			}
			`,
			expected: []string{"createdAt", "id", "updatedAt"},
		},
		{
			name: "actions-input-nested-field",
			schema: `
			model Publisher {
				fields {
					name Text
				}
			}

			model Author {
				fields {
					publisher Publisher
				}
			}

			model Book {
				fields {
					author Author
				}
				operations {
					list booksByPublisher(author.publisher.<Cursor>)
				}
			}
			`,
			expected: []string{"createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "actions-input-nested-field-partial",
			schema: `
			model Publisher {
				fields {
					name Text
				}
			}

			model Author {
				fields {
					publisher Publisher
				}
			}

			model Book {
				fields {
					author Author
				}
				operations {
					list booksByPublisher(author.publisher.na<Cursor>)
				}
			}
			`,
			expected: []string{"createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "actions-input-unresolvable",
			schema: `
			model Foo {
				operations {
					list listFoos(no.idea.what.im.doing<Cursor>)
				}
			}
			`,
			expected: []string{},
		},
		{
			name: "set-expression",
			schema: `
			model Person {
				operations {
					create createPerson() {
						@set(p<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"person", "ctx", "env", "secrets"},
		},
		{
			name: "set-expression-model-attribute",
			schema: `
			model Person {
				fields {
					name Text
				}
				operations {
					create createPerson() {
						@set(person.<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"id", "createdAt", "updatedAt", "name"},
		},
		{
			name: "set-expression-ctx-fields",
			schema: `
			model Person {
				fields {
					identity Identity
				}
				operations {
					create createPerson() {
						@set(person.identity = ctx.<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"env", "identity", "now", "secrets"},
		},
		{
			name: "set-expression-ctx-env-vars",
			schema: `
			model Person {
				fields {
					identity Identity
				}
				operations {
					create createPerson() {
						@set(person.identity = ctx.env.<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"TEST", "TEST_2"},
		},
		{
			name: "set-expression-ctx-secrets",
			schema: `
			model Person {
				fields {
					identity Identity
					apikey Secret
				}
				operations {
					create createPerson() {
						@set(person.apikey = ctx.secrets.<Cursor>)
					}
				}
			}
			`,
			expected: []string{"API_KEY"},
		},
		{
			name: "set-expression-unresolvable",
			schema: `
			model Person {
				fields {
					name Text
				}
				operations {
					create createPerson() {
						@set(total.nonsense.<Cursor>)
					}
				}
			}	
			`,
			expected: []string{},
		},
		{
			name: "where-expression",
			schema: `
			model Person {
				operations {
					get getPerson() {
						@where(p<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"person", "ctx", "env", "secrets"},
		},
		{
			name: "validate-expression",
			schema: `
			model Person {
				fields {
					name Text
				}
				operations {
					update updatePerson(id) with (name) {
						@validate(p<Cursor>)
					}
				}
			}	
			`,
			expected: []string{"person", "ctx", "env", "secrets"},
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
			expected: []string{"create"},
		},
		{
			name: "permission-attribute-actions-many-complete-token",
			schema: `
			model Person {
				@permission(
					actions: [get, update, create<Cursor>]
				)
			}
			`,
			expected: []string{","},
		},
		{
			name: "permission-attribute-actions-many-complete-token",
			schema: `
			model Person {
				@permission(
					actions: [get, update, updated<Cursor>]
				)
			}
			`,
			expected: []string{","},
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
			expected: []string{"GetPersonInput", "createdAt", "id", "updatedAt"},
		},
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
	}

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

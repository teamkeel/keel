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
	"github.com/teamkeel/keel/schema/reader"
)

type testCase struct {
	name        string
	schema      string
	otherSchema string
	expected    []string
}

func TestRootCompletions(t *testing.T) {

	cases := []testCase{
		{
			name:     "top-level-keyword",
			schema:   "mod<Cursor>",
			expected: []string{"api", "enum", "message", "model", "role", "job"},
		},
		{
			name: "top-level-keyword-not-first",
			schema: `
			model A {

            }

            m<Cursor>`,
			expected: []string{"api", "enum", "message", "model", "role", "job"},
		},
		{
			name:     "top-level-keyword-whitespace",
			schema:   `<Cursor>`,
			expected: []string{"api", "enum", "message", "model", "role", "job"},
		},
		{
			name: "top-level-keyword-whitespace-partial-schema",
			schema: `
			model A {}

			<Cursor>

			model B {}
			`,
			expected: []string{"api", "enum", "message", "model", "role", "job"},
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
			expected: []string{"fields", "actions"},
		},
		{
			name: "model-block-keywords-whitespace",
			schema: `
			model A {
			  <Cursor>
			}`,
			expected: []string{"@permission", "@unique", "@on", "fields", "actions"},
		},
		// attributes tests
		{
			name: "model-attributes",
			schema: `
			model A {
              @<Cursor>
            }`,
			expected: []string{"@permission", "@unique", "@on", "fields", "actions"},
		},
	}

	runTestsCases(t, cases)
}

func TestCompositeUniqueCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "array-notation",
			schema: `
			model A {
				@unique(<Cursor>
			}
			`,
			expected: []string{"["},
		},
		{
			name: "available-fields",
			schema: `
			model A {
				fields {
					title Text
					subTitle Text
					date Date
					timestamp Timestamp
				}
				@unique([<Cursor>
			}
			`,
			expected: []string{"subTitle", "title", "date"},
		},
		{
			name: "existing-composite",
			schema: `
			model A {
				fields {
					title Text
					subTitle Text
				}
				@unique([title,<Cursor>
			}
			`,
			expected: []string{"subTitle"},
		},
		{
			name: "closing-array-notation",
			schema: `
			model A {
				fields {
					title Text
					subTitle Text
				}
				@unique([title,subTitle<Cursor>
			}
			`,
			expected: []string{"]"},
		},
		{
			name: "closing-paren",
			schema: `
			model A {
				fields {
					title Text
					subTitle Text
				}
				@unique([title,subTitle]<Cursor>
			}
			`,
			expected: []string{")"},
		},
		{
			name: "model-field-exclusion",
			schema: `
			model B {}
			model A {
				fields {
					title Text
					relation B
					subTitle Text
				}
				@unique([<Cursor>
			}
			`,
			expected: []string{"subTitle", "title"},
		},
		{
			name: "relationship",
			schema: `
			model B {}
			model A {
				fields {
					title Text
					relation B
					subTitle Text
				}
				@unique([relation.<Cursor>
			}
			`,
			expected: []string{},
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
			expected: []string{"actions", "fields"},
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
			name: "field-type-model-multi-file",
			schema: `
			model Book {
				fields {
					author Au<Cursor>
				}	
			}`,
			otherSchema: `
			model Author {}
			`,
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
			name: "field-type-enum-multi-file",
			schema: `
			model Book {
				fields {
					category Ca<Cursor>
				}
			}
			`,
			otherSchema: `
			enum Category {
				Romance
				Horror	
			}
			`,
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
			expected: []string{"@unique"},
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
			expected: []string{"@unique"},
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

func TestActionCompletions(t *testing.T) {
	cases := []testCase{
		// actions tests
		{
			name: "action-type-completions",
			schema: `
			model A {
				actions {
					<Cursor>
				}
			}`,
			expected: parser.ActionTypes,
		},
		{
			name: "create-keyword",
			schema: `
			model A {
				actions {
				c<Cursor>
				}
			}`,
			expected: parser.ActionTypes,
		},
		{
			name: "create-keyword-not-first",
			schema: `
			model A {
				actions {
				get getA(id)
				crea<Cursor>
				}
			}`,
			expected: parser.ActionTypes,
		},
		{
			name: "get-keyword",
			schema: `
			model A {
		      actions {
		        get getA(id)
		        g<Cursor>
			  }
		    }`,
			expected: parser.ActionTypes,
		},
		// input tests
		{
			name: "model-field-inputs",
			schema: `
			model A {
			  fields {
				something Text
			  }
		      actions {
		        get getA(<Cursor>)
			  }
		    }`,
			expected: []string{"something", "id", "createdAt", "updatedAt"},
		},
		{
			name: "model-field-inputs-relationship",
			schema: `
			model B {
				fields {
					foo Text
				}
			}
			model A {
			  fields {
				other B
			  }
		      actions {
		        get getA(other.<Cursor>)
			  }
		    }
			`,
			expected: []string{"foo", "id", "createdAt", "updatedAt"},
		},
		{
			name: "with-nested-relationship",
			schema: `
			enum Sex {
				Male
				Female
			}

			model Author {
				fields {
					name Text
				}
			}

			model Person {
				fields {
					title Text
						author Author
				}

				actions {
					create createPerson() with (title, author.<Cursor>) {
							@permission(expression: true)
					}
				}
			}
			`,
			expected: []string{"createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "model-field-inputs-relationship-multi-file",
			schema: `
			model A {
			  fields {
				other B
			  }
		      actions {
		        get getA(other.deeper.<Cursor>)
			  }
		    }
			`,
			otherSchema: `
			model B {
				fields {
					deeper C
				}
			}
			model C {
				fields {
					bar Text
				}
			}
			`,
			expected: []string{"bar", "id", "createdAt", "updatedAt"},
		},
		// with tests
		{
			name: "with-keyword",
			schema: `
			model A {
		      actions {
		        create createA() wi<Cursor>
			  }
		    }`,
			expected: []string{"with"},
		},
		{
			name: "with-keyword-whitespace",
			schema: `
			model A {
		      actions {
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
		      actions {
		        create createA() with (name) <Cursor>
			  }
		    }`,
			expected: []string{},
		},
		// attribute tests
		{
			name: "action-attributes-prefixed",
			schema: `
			model A {
			  fields {
				name Text
			  }
		      actions {
		        create createA() with (name) {
		          @s<Cursor>
				}
			  }
		    }`,
			expected: []string{"@set", "@sortable"},
		},
		{
			name: "action-attributes-bare-at",
			schema: `
			model A {
			  fields {
				name Text
			  }
		      actions {
		        create createA() with (name) {
		          @<Cursor>
				}
			  }
		    }`,
			expected: []string{"@function", "@orderBy", "@permission", "@set", "@sortable", "@validate", "@where"},
		},
		{
			name: "action-attributes-whitespace",
			schema: `
			model A {
			  fields {
				name Text
			  }
		      actions {
		        create createA() with (name) {
		          <Cursor>
				}
			  }
		    }`,
			expected: []string{"@function", "@orderBy", "@permission", "@set", "@sortable", "@validate", "@where"},
		},
	}

	runTestsCases(t, cases)
}

func TestWhereAttributeCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "where-attribute-ctx-env",
			schema: `
			model Post {
				fields {
					text Text
				}
				actions {
					list Posts() {
						@where(record.text == ctx.env.<Cursor>)
					}
				}
			}`,
			expected: []string{"TEST", "TEST_2"},
		},
		{
			name: "where-attribute-ctx-env-no-completions",
			schema: `
			model Post {
				actions {
					fields {
						text Text
					}
					list Posts() {
						@where(post.text == ctx.env.TEST.<Cursor>)
					}
				}
			}`,
			expected: []string{},
		},
		{
			name: "where-attribute-ctx-secrets",
			schema: `
			model Post {
				fields {
					text Text
				}
				actions {
					list Posts() {
						@where(post.text == ctx.secrets.<Cursor>)
					}
				}
			}`,
			expected: []string{"API_KEY"},
		},
		{
			name: "where-attribute-ctx-isauthenticated-no-completions",
			schema: `
			model Post {
				fields {
					text Text
				}
				actions {
					list Posts() {
						@where(post.text == ctx.isAuthenticated.<Cursor>)
					}
				}
			}`,
			expected: []string{},
		},
		{
			name: "where-attribute-ctx",
			schema: `
			model CompanyUser {
				fields {
					identity Identity @unique @relation(user)
				}
			}
			model Record {
				fields {
					owner CompanyUser
				}
				actions {
					list listRecords() {
						@where(record.owner == ctx.<Cursor>)
					}
				}
			}`,
			expected: []string{"env", "headers", "identity", "isAuthenticated", "now", "secrets"},
		},
		{
			name: "where-attribute-ctx-identity",
			schema: `
			model CompanyUser {
				fields {
					name Text
					identity Identity @unique @relation(user)
				}
			}
			model Record {
				fields {
					owner CompanyUser
				}
				actions {
					list listRecords() {
						@where(record.owner == ctx.identity.<Cursor>)
					}
				}
			}`,
			expected: []string{"createdAt", "email", "emailVerified", "externalId", "id", "issuer", "password", "updatedAt", "user"},
		},

		{
			name: "where-attribute-ctx-identity-user",
			schema: `
			model CompanyUser {
				fields {
					name Text
					identity Identity @unique @relation(user)
					company Company
				}
			}
			model Company {
				fields {
					name Text
				}
			}
			model Record {
				fields {
					owner CompanyUser
				}
				actions {
					list listRecords() {
						@where(record.owner  == ctx.identity.user.<Cursor>)
					}
				}
			}`,
			expected: []string{"company", "createdAt", "id", "identity", "name", "updatedAt"},
		},
		{
			name: "where-attribute-ctx-identity-user-company",
			schema: `
			model CompanyUser {
				fields {
					name Text
					identity Identity @unique @relation(user)
					company Company
				}
			}
			model Company {
				fields {
					name Text
				}
			}
			model Record {
				fields {
					owner CompanyUser
				}
				actions {
					list listRecords() {
						@where(record.owner == ctx.identity.user.company.<Cursor>)
					}
				}
			}`,
			expected: []string{"createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "where-attribute-ctx-identity-user-company-name-no-completions",
			schema: `
			model CompanyUser {
				fields {
					name Text
					identity Identity @unique @relation(user)
					company Company
				}
			}
			model Company {
				fields {
					name Text
				}
			}
			model Record {
				fields {
					owner CompanyUser
				}
				actions {
					list listRecords() {
						@where(record.owner == ctx.identity.user.company.name.<Cursor>)
					}
				}
			}`,
			expected: []string{},
		},
		{
			name: "where-attribute-model-fields",
			schema: `
			model Person {
				fields {
					name Text
					age Number
				}
				actions {
					list people() {
						@where(person.<Cursor>)
					}
				}
			}`,
			expected: []string{"name", "age", "id", "createdAt", "updatedAt"},
		},
		{
			name: "where-attribute-model-fields-relationships",
			schema: `
			model Dog {
				fields {
					breed Text
					owner Person
				}
			}
			model Person {
				fields {
					dogs Dog[]
				}
				actions {
					list people() {
						@where(person.dogs.<Cursor>)
					}
				}
			}`,
			expected: []string{"breed", "id", "createdAt", "owner", "updatedAt"},
		},
		{
			name: "where-attribute-model-fields-relationships-multi-file",
			schema: `
			model Person {
				fields {
					dogs Dog[]
				}
				actions {
					list people() {
						@where(person.dogs.<Cursor>)
					}
				}
			}`,
			otherSchema: `
			model Dog {
				fields {
					breed Text
					owner Person
				}
			}
			`,
			expected: []string{"breed", "id", "createdAt", "owner", "updatedAt"},
		},
		{
			name: "where-attribute-enums",
			schema: `
			model Pet {
				fields {
					species Animal
				}
				actions {
					list pets() {
						@where(pet.species == <Cursor>)
					}
				}
			}`,
			otherSchema: `
			enum Animal {
				Dog
				Cat
				Rabbit
			}
			`,
			expected: []string{"Animal", "ctx", "pet"},
		},
		{
			name: "where-attribute-enum-values",
			schema: `
			model Pet {
				fields {
					species Animal
				}
				actions {
					list pets() {
						@where(pet.species == Animal.<Cursor>)
					}
				}
			}`,
			otherSchema: `
			enum Animal {
				Dog
				Cat
				Rabbit
			}
			`,
			expected: []string{"Dog", "Cat", "Rabbit"},
		},
	}

	runTestsCases(t, cases)
}

func TestSetAttributeCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "set-expression",
			schema: `
			model User {
				fields {
					identity Identity
				}
			}
			model Team {
				fields {
					name Text
				}
			}
			model UserTeam {
				fields {
					user User
					team Team
				}
				actions {
					create createTeam() with (team.name) {
						@set(<Cursor>)
					}
				}
			}`,
			expected: []string{"ctx", "userTeam"},
		},
		{
			name: "set-attribute-ctx",
			schema: `
			model User {
				fields {
					name Text
					identity Identity @unique @relation(user)
				}
			}
			model Post {
				fields {
					owner User
				}
				actions {
					create create() {
						@set(post.owner = ctx.<Cursor>)
					}
				}
			}`,
			expected: []string{"env", "headers", "identity", "isAuthenticated", "now", "secrets"},
		},
		{
			name: "set-attribute-ctx-identity",
			schema: `
			model User {
				fields {
					name Text
					identity Identity @unique @relation(user)
				}
			}
			model Post {
				fields {
					owner User
				}
				actions {
					list create() {
						@set(post.owner = ctx.identity.<Cursor>)
					}
				}
			}`,
			expected: []string{"createdAt", "email", "emailVerified", "externalId", "id", "issuer", "password", "updatedAt", "user"},
		},
		{
			name: "set-attribute-ctx-identity-user",
			schema: `
			model User {
				fields {
					name Text
					identity Identity @unique @relation(user)
				}
			}
			model Post {
				fields {
					owner User
				}
				actions {
					list create() {
						@set(post.owner = ctx.identity.user.<Cursor>)
					}
				}
			}`,
			expected: []string{"createdAt", "id", "identity", "name", "updatedAt"},
		},
		{
			name: "set-attribute-first-operand-ctx-identity",
			schema: `
			model User {
				fields {
					name Text
					identity Identity @unique @relation(user)
				}
			}
			model Post {
				fields {
					owner User
				}
				actions {
					list create() {
						@set(user.owner = ctx.identity.<Cursor>)
					}
				}
			}`,
			expected: []string{"createdAt", "email", "emailVerified", "externalId", "id", "issuer", "password", "updatedAt", "user"},
		},
	}

	runTestsCases(t, cases)
}

func TestFunctionCompletions(t *testing.T) {
	cases := []testCase{
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
			
				actions {
					create c<Cursor>
				}
			}
			`,
			expected: []string{"createPostExtended"},
		},
		// input tests
		{
			name: "arbitrary-function-input-completions",
			schema: `
			message GetPersonInput {}
			model Person {
				actions {
					read getPerson(<Cursor>
				}
			}
			`,
			expected: []string{"GetPersonInput", "Any", "createdAt", "id", "updatedAt"},
		},
		{
			name: "arbitrary-function-input-completions-multi-file",
			schema: `
			model Person {
				actions {
					read getPerson(<Cursor>
				}
			}
			`,
			otherSchema: `
			message GetPersonInput {}
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
				actions {
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
				actions {
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
				actions {
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
				actions {
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
				actions {
					create createPerson() w<Cursor>
				}
			}`,
			expected: []string{"with"},
		},
		{
			name: "arbitrary-function-update-with-completion",
			schema: `
			model Person {
				actions {
					update updatePerson() <Cursor>
				}
			}`,
			expected: []string{"with"},
		},
		{
			name: "arbitrary-function-update-partial-with-completion",
			schema: `
			model Person {
				actions {
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
			expected: []string{"@permission", "actions", "fields"},
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
				actions {
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
			expected: []string{"person", "ctx"},
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
			expected: []string{"person", "ctx"},
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
			name: "permission-attribute-ctx-fields",
			schema: `
			model Person {
				@permission(
					expression: ctx.<Cursor>
				)
			}
			`,
			expected: []string{"env", "headers", "identity", "isAuthenticated", "now", "secrets"},
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
			expected: parser.ActionTypes,
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
			expected: parser.ActionTypes,
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

func TestOrderByCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "orderby-attribute-labels",
			schema: `
			enum Country {
				UK
				ZA
			}
			model Company {}
			model Person {
				fields {
					name Text
					age Number
					nationality Country
					employer Company
				}
				actions {
					list people() {
						@orderBy(<Cursor>)
					}
				}
			}`,
			expected: []string{"age", "createdAt", "id", "name", "nationality", "updatedAt"},
		},
		{
			name: "orderby-attribute-labels-next-arg",
			schema: `
			model Person {
				fields {
					name Text
					age Number
				}
				actions {
					list people() {
						@orderBy(name: asc, <Cursor>
					}
				}
			}`,
			expected: []string{"age", "createdAt", "id", "name", "updatedAt"},
		},
		{
			name: "orderby-attribute-values",
			schema: `
			model Person {
				fields {
					name Text
					age Number
				}
				actions {
					list people() {
						@orderBy(name: <Cursor>
					}
				}
			}`,
			expected: []string{"asc", "desc"},
		},
		{
			name: "orderby-attribute-values-prefix",
			schema: `
			model Person {
				fields {
					name Text
					age Number
				}
				actions {
					list people() {
						@orderBy(name: as<Cursor>)
					}
				}
			}`,
			expected: []string{"asc", "desc"},
		},
		{
			name: "orderby-attribute-values-next-arg",
			schema: `
			model Person {
				fields {
					name Text
					age Number
				}
				actions {
					list people() {
						@orderBy(name: desc, age: <Cursor>
					}
				}
			}`,
			expected: []string{"asc", "desc"},
		},
	}

	runTestsCases(t, cases)
}

func TestSortableCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "sortable-attribute-values",
			schema: `
			enum Country {
				UK
				ZA
			}
			model Company {}
			model Person {
				fields {
					name Text
					age Number
					nationality Country
					employer Company
				}
				actions {
					list people() {
						@sortable(<Cursor>
					}
				}
		    }`,
			expected: []string{"age", "createdAt", "id", "name", "nationality", "updatedAt"},
		},
		{
			name: "sortable-attribute-model-fields-second-arg",
			schema: `
			model Person {
			  fields {
				name Text
				age Number
			  }
		      actions {
		        list people() {
					@sortable(name, <Cursor>
				}
			  }
		    }`,
			expected: []string{"age", "createdAt", "id", "name", "updatedAt"},
		},
	}

	runTestsCases(t, cases)
}

func TestOnCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "on-attribute-action-args-without-array",
			schema: `
			model Person {
				@on(<Cursor>
		    }`,
			expected: []string{"["},
		},
		{
			name: "on-attribute-action-args",
			schema: `
			model Person {
				@on([<Cursor>
		    }`,
			expected: []string{"create", "delete", "update"},
		},
		{
			name: "on-attribute-action-args-second",
			schema: `
			model Person {
				@on([update, <Cursor>
		    }`,
			expected: []string{"create", "delete", "update"},
		},
		{
			name: "on-attribute-subscriber-arg",
			schema: `
			model Person {
				@on([create, delete],<Cursor>
		    }
			`,
			expected: []string{},
		},
		{
			name: "on-attribute-subscriber-arg-suggest-existing",
			schema: `
			model Employee {
				@on([update], verifyDetails)
				@on([create], verifydetails) // Different casing
				@on([delete], sendGoodbyeMail)
			}
			model Person {
				@on([create, update], verifyDetails)
				@on([create, delete], <Cursor>
		    }
			`,
			expected: []string{"sendGoodbyeMail", "verifyDetails", "verifydetails"},
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
			expected: []string{"api", "enum", "message", "model", "role", "job"},
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
			expected: []string{"api", "enum", "message", "model", "role", "job"},
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
			expected: []string{"Identity", "Person"},
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

func TestJobCompletions(t *testing.T) {
	cases := []testCase{
		{
			name: "job-completions",
			schema: `
			job MyJob {
				<Cursor>
			}
			`,
			expected: []string{"inputs", "@permission", "@schedule"},
		},
		{
			name: "job-block-keywords",
			schema: `
			job MyJob {
              i<Cursor>
            }`,
			expected: []string{"inputs"},
		},
		{
			name: "job-input-completions",
			schema: `
			job MyJob {
			  inputs {
				input1 <Cursor>
			  }
			}
			`,
			expected: []string{"Boolean", "Date", "ID", "Identity", "Number", "Password", "Secret", "Text", "Timestamp"},
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
					Filename: "schema.keel",
					Line:     i + 1,
					Column:   idx + 1,
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

			schemaFiles := []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: schema,
				},
			}
			if tc.otherSchema != "" {
				schemaFiles = append(schemaFiles, &reader.SchemaFile{
					FileName: "other.keel",
					Contents: tc.otherSchema,
				})
			}

			results := completions.Completions(schemaFiles, pos, configFile)
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

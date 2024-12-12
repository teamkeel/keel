package definitions_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/definitions"
	"github.com/teamkeel/keel/schema/reader"
)

func TestDefinitions(t *testing.T) {
	type TestCase struct {
		Name                   string
		Files                  []*reader.SchemaFile
		ExpectSchemaDefinition bool
		ExpectedFunctionName   string
	}

	// In these tests the source position is indicated by <Pos> and
	// the definition position by <Def>.
	cases := []TestCase{
		{
			Name: "field name - no definition",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model Post {
	fields {
		ti<Pos>tle Text
	}
}
`,
				},
			},
			ExpectSchemaDefinition: false,
		},
		{
			Name: "built-in type - no definition",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model Post {
	fields {
		title T<Pos>ext
	}
}
`,
				},
			},
			ExpectSchemaDefinition: false,
		},
		{
			Name: "go to Model from field type",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model <Def>Profile {}

model Post {
	fields {
		profile P<Pos>rofile
	}
}
`,
				},
			},
			ExpectSchemaDefinition: true,
		},
		{
			Name: "go to Enum from field type",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model Post {
	fields {
		title Text
		category Cate<Pos>gory
	}
}

enum <Def>Category {
	Sports
	Finance
}
		`,
				},
			},
			ExpectSchemaDefinition: true,
		},
		{
			Name: "go to field from action input",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model Post {
	fields {
		<Def>title Text
	}
	actions {
		create createPost() with (<Pos>title)
	}
}
		`,
				},
			},
			ExpectSchemaDefinition: true,
		},
		{
			Name: "go to field from action input using relationship",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model Author {
	fields {
		publisher Publisher
	}
}
model Post {
	fields {
		author Author
	}
	actions {
		list listBooks(author.publisher.nam<Pos>e)
	}
}
model Publisher {
	fields {
		<Def>name Text
	}
}
		`,
				},
			},
			ExpectSchemaDefinition: true,
		},
		// 		{
		// 			Name: "go to field from @set expression",
		// 			Files: []*reader.SchemaFile{
		// 				{
		// 					FileName: "schema.keel",
		// 					Contents: `
		// model Author {
		// 	fields {
		// 		<Def>identity Identity
		// 	}
		// 	actions {
		// 		create newAuthor() {
		// 			@set(author.i<Pos>dentity == ctx.identity)
		// 		}
		// 	}
		// }
		// 		`,
		// 				},
		// 			},
		// 			ExpectSchemaDefinition: true,
		// 		},
		// 		{
		// 			Name: "go to field from @where expression",
		// 			Files: []*reader.SchemaFile{
		// 				{
		// 					FileName: "schema.keel",
		// 					Contents: `
		// model Book {
		// 	fields {
		// 		<Def>published Boolean
		// 	}
		// 	actions {
		// 		list books() {
		// 			@where(book.p<Pos>ublished == true)
		// 		}
		// 	}
		// }
		// 		`,
		// 				},
		// 			},
		// 			ExpectSchemaDefinition: true,
		// 		},
		// 		{
		// 			Name: "go to field from @where expression with relationship in different file",
		// 			Files: []*reader.SchemaFile{
		// 				{
		// 					FileName: "publisher.keel",
		// 					Contents: `
		// model Publisher {
		// 	fields {
		// 		<Def>isActive Boolean
		// 	}
		// }
		// 		`,
		// 				},
		// 				{
		// 					FileName: "author.keel",
		// 					Contents: `
		// model Author {
		// 	fields {
		// 		publisher Publisher
		// 	}
		// }
		// 		`,
		// 				},
		// 				{
		// 					FileName: "book.keel",
		// 					Contents: `
		// model Book {
		// 	fields {
		// 		author Author
		// 	}
		// 	actions {
		// 		list books() {
		// 			@where(book.author.publisher.isActiv<Pos>e == true)
		// 		}
		// 	}
		// }
		// 		`,
		// 				},
		// 			},
		// 			ExpectSchemaDefinition: true,
		// 		},
		{
			Name: "go to function",
			Files: []*reader.SchemaFile{
				{
					FileName: "schema.keel",
					Contents: `
model Book {
	actions {
		read get<Pos>Book(Any)
	}
}
		`,
				},
			},
			ExpectedFunctionName: "getBook",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var pos definitions.Position
			var expected *definitions.Definition

			for _, f := range c.Files {
				contents := strings.ReplaceAll(f.Contents, "\t", "    ")
				lines := []string{}

				for i, line := range strings.Split(contents, "\n") {
					posIdx := strings.Index(line, "<Pos>")
					defIdx := strings.Index(line, "<Def>")

					line = strings.ReplaceAll(line, "<Pos>", "")
					line = strings.ReplaceAll(line, "<Def>", "")

					switch {
					case posIdx != -1:
						pos.Filename = f.FileName
						pos.Line = i + 1
						pos.Column = posIdx + 1
					case defIdx != -1:
						expected = &definitions.Definition{
							Schema: &definitions.Position{
								Filename: f.FileName,
								Line:     i + 1,
								Column:   defIdx + 1,
							},
						}
					}

					lines = append(lines, line)
				}

				f.Contents = strings.Join(lines, "\n")
			}

			if c.ExpectedFunctionName != "" && expected != nil {
				t.Fatal("cannot provide both ExpectedFunctionName and a source definition")
			}

			if c.ExpectedFunctionName != "" {
				expected = &definitions.Definition{
					Function: &definitions.FunctionDefinition{
						Name: c.ExpectedFunctionName,
					},
				}
			}

			actual := definitions.GetDefinition(c.Files, pos)
			if c.ExpectSchemaDefinition {
				require.NotNil(t, actual)
				require.NotNil(t, actual.Schema)
			}
			assert.EqualValues(t, expected, actual)
		})
	}
}

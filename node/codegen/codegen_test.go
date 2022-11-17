package codegenerator_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	codegenerator "github.com/teamkeel/keel/node/codegen"
	"github.com/teamkeel/keel/schema"
)

type TestCase struct {
	Name          string
	Schema        string
	ExpectedFiles []*codegenerator.GeneratedFile
}

func TestSdk(t *testing.T) {
	cases := []TestCase{
		{
			Name: "model-generation-simple",
			Schema: `
			model Person {
				fields {
					name Text
					age Number
				}
			}
			`,
			ExpectedFiles: []*codegenerator.GeneratedFile{
				{
					Path: "index.js",
					Contents: `
						export class PersonApi {
							constructor() {
								this.create = async (inputs) => {
									return this.db.create(inputs);
								};
								this.where = (conditions) => {
									return this.db.where(conditions);
								};
								this.delete = (id) => {
									return this.db.delete(id);
								};
								this.findOne = (query) => {
									return this.db.findOne(query);
								};
								this.update = (id, inputs) => {
									return this.db.update(id, inputs);
								};
								this.findMany = (query) => {
									return this.db.where(query).all();
								};
								this.db = new Query({
									tableName: 'person',
									queryResolver: queryResolverFromEnv(process.env),
									logger: queryLogger
								});
							}
						}
						export class IdentityApi {
							constructor() {
								this.create = async (inputs) => {
									return this.db.create(inputs);
								};
								this.where = (conditions) => {
									return this.db.where(conditions);
								};
								this.delete = (id) => {
									return this.db.delete(id);
								};
								this.findOne = (query) => {
									return this.db.findOne(query);
								};
								this.update = (id, inputs) => {
									return this.db.update(id, inputs);
								};
								this.findMany = (query) => {
									return this.db.where(query).all();
								};
								this.db = new Query({
									tableName: 'identity',
									queryResolver: queryResolverFromEnv(process.env),
									logger: queryLogger
								});
							}
						}
						export const api = {
							models: {
								person: new PersonApi(),
								identity: new IdentityApi(),
							}
						}`,
				},
				{
					Path:     "index.d.ts",
					Contents: ``,
				},
			},
		},
		{
			Name: "model-generation-custom-function",
			Schema: `model Person {
				fields {
					name Text
					age Number
				}

				functions {
					create createPerson() with(name, age)
					update updatePerson(id) with(name, age)
					delete deletePerson(id)
					list listPerson()
					get getPerson(id)
				}
			}
			`,
			ExpectedFiles: []*codegenerator.GeneratedFile{
				{
					Path: "index.js",
					Contents: `
					export class PersonApi {
						constructor() {
							this.create = async (inputs) => {
								return this.db.create(inputs);
							};
							this.where = (conditions) => {
								return this.db.where(conditions);
							};
							this.delete = (id) => {
								return this.db.delete(id);
							};
							this.findOne = (query) => {
								return this.db.findOne(query);
							};
							this.update = (id, inputs) => {
								return this.db.update(id, inputs);
							};
							this.findMany = (query) => {
								return this.db.where(query).all();
							};
							this.db = new Query({
								tableName: 'person',
								queryResolver: queryResolverFromEnv(process.env),
								logger: queryLogger
							});
						}
					}
					export class IdentityApi {
						constructor() {
							this.create = async (inputs) => {
								return this.db.create(inputs);
							};
							this.where = (conditions) => {
								return this.db.where(conditions);
							};
							this.delete = (id) => {
								return this.db.delete(id);
							};
							this.findOne = (query) => {
								return this.db.findOne(query);
							};
							this.update = (id, inputs) => {
								return this.db.update(id, inputs);
							};
							this.findMany = (query) => {
								return this.db.where(query).all();
							};
							this.db = new Query({
								tableName: 'identity',
								queryResolver: queryResolverFromEnv(process.env),
								logger: queryLogger
							});
						}
					}
					export const api = {
						models: {
						  person: new PersonApi(),
						  identity: new IdentityApi(),
						}
					}
					export const createPerson = (callback) => (inputs, api) => {
						return callback(inputs, api);
					};
					export const updatePerson = (callback) => (inputs, api) => {
						return callback(inputs, api);
					};
					export const deletePerson = (callback) => (inputs, api) => {
						return callback(inputs, api);
					};
					export const listPerson = (callback) => (inputs, api) => {
						return callback(inputs, api);
					};
					export const getPerson = (callback) => (inputs, api) => {
						return callback(inputs, api);
					};`,
				},
				{
					Path:     "index.d.ts",
					Contents: "",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			builder := schema.Builder{}

			sch, err := builder.MakeFromString(tc.Schema)

			require.NoError(t, err)

			tmpDir, err := os.MkdirTemp("", tc.Name)

			require.NoError(t, err)

			cg := codegenerator.NewGenerator(sch, tmpDir)

			generatedFiles, err := cg.GenerateSDK()

			require.NoError(t, err)

			compareFiles(t, tc, generatedFiles)
		})
	}
}

// Normalises differences between actual and expected strings (replaces tabs with 2 spaces)
func normaliseString(str string) string {
	lines := strings.Split(str, "\n")

	lines = lo.Filter(lines, func(line string, idx int) bool {
		if line == "" {
			return idx > 0 && idx < len(lines)-1
		}

		return true
	})

	if len(lines) == 0 {
		return ""
	}

	firstLine := lines[0]

	firstLineChars := strings.Split(firstLine, "")

	indentDepth := 0

	for _, char := range firstLineChars {
		if char != " " && char != "\t" {
			break
		}

		indentDepth += 1
	}

	newLines := []string{}

	for _, line := range lines {
		newLines = append(newLines, line[indentDepth:])
	}

	tabReplacer := strings.NewReplacer(
		"\t", "  ",
	)

	newStr := strings.Join(newLines, "\n")

	newStr = tabReplacer.Replace(newStr)

	return newStr
}

func compareFiles(t *testing.T, tc TestCase, generatedFiles []*codegenerator.GeneratedFile) {
actual:
	for _, actualFile := range generatedFiles {
		for _, expectedFile := range tc.ExpectedFiles {
			if actualFile.Path == expectedFile.Path {
				assert.Equal(t, normaliseString(actualFile.Contents), normaliseString(expectedFile.Contents))

				continue actual
			}
		}

		assert.Fail(t, fmt.Sprintf("no matching expectated fike for actual file %s", actualFile.Path))
	}

expected:
	for _, expectedFile := range tc.ExpectedFiles {
		for _, actualFile := range generatedFiles {
			if actualFile.Path == expectedFile.Path {
				assert.Equal(t, normaliseString(actualFile.Contents), normaliseString(expectedFile.Contents))

				continue expected
			}
		}

		assert.Fail(t, fmt.Sprintf("no matching actual file for expected file %s", expectedFile.Path))
	}
}

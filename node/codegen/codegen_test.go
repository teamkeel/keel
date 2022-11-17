package codegenerator_test

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
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
	// When comparing actual vs expected src code contents, the expected value is
	// evaluated partially [e.g expect(actual).toInclude(expected)] so you do not need
	// to specify the full code to be generated, and instead match partially on whatever
	// you want to expect to be added for a given test.
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
					Path: "index.d.ts",
					Contents: `
						import {
							QueryConstraints,
							ChainableQuery,
							Query,
							queryResolverFromEnv,
							// todo: the response types below will change
							FunctionError,
							FunctionCreateResponse,
							FunctionGetResponse,
							FunctionDeleteResponse,
							FunctionListResponse,
							FunctionUpdateResponse,
							FunctionAuthenticateResponse,
						} from '@teamkeel/functions-runtime';
						
						// todo: examine this type
						export type Timestamp = string

						export interface Person {
						  name: string
						  age: number
						  id: string
						  createdAt: Date
						  updatedAt: Date
						}
						export interface Identity {
						  email: string
						  password: string
						  id: string
						  createdAt: Date
						  updatedAt: Date
						}

						// query types
						
						// func wrapper types
						
						// api typings
						
						// top level api typing
					`,
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
					Path: "index.d.ts",
					Contents: `
						import {
							QueryConstraints,
							ChainableQuery,
							Query,
							queryResolverFromEnv,
							// todo: the response types below will change
							FunctionError,
							FunctionCreateResponse,
							FunctionGetResponse,
							FunctionDeleteResponse,
							FunctionListResponse,
							FunctionUpdateResponse,
							FunctionAuthenticateResponse,
						} from '@teamkeel/functions-runtime';
						
						// todo: examine this type
						export type Timestamp = string

						export interface Person {
							name: string
							age: number
							id: string
							createdAt: Date
							updatedAt: Date
						}
						export interface Identity {
							email: string
							password: string
							id: string
							createdAt: Date
							updatedAt: Date
						}
						
						
						// query types
						
						// func wrapper types
						
						// api typings
						
						// top level api typing
					`,
				},
			},
		},
		{
			Name: "enum-generation",
			Schema: `
				enum TheBeatles {
					John
					Paul
					Ringo
					George
				}
			`,
			ExpectedFiles: []*codegenerator.GeneratedFile{
				{
					Path:     "index.js",
					Contents: "",
				},
				{
					Path: "index.d.ts",
					Contents: `
						export enum TheBeatles {
							John = "John",
							Paul = "Paul",
							Ringo = "Ringo",
							George = "George",
						}
					`,
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

			o, err := typecheck(t, generatedFiles)

			if err != nil {
				fmt.Print(string(o))
				t.Fail()
			}
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
		lineLength := len(line)

		if lineLength <= indentDepth {
			// blank line in body of lines
			continue
		} else {
			newLines = append(newLines, line[indentDepth:])
		}
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
				assert.Contains(t, normaliseString(actualFile.Contents), normaliseString(expectedFile.Contents))

				actual := normaliseString(actualFile.Contents)
				expected := normaliseString(expectedFile.Contents)

				if !strings.Contains(actual, expected) {
					t.Errorf("Result not as expected:\n%v", diff.LineDiff(expected, actual))
					t.Fail()
				}

				continue actual
			}
		}

		assert.Fail(t, fmt.Sprintf("no matching expectated fike for actual file %s", actualFile.Path))
	}

expected:
	for _, expectedFile := range tc.ExpectedFiles {
		for _, actualFile := range generatedFiles {
			if actualFile.Path == expectedFile.Path {
				continue expected
			}
		}

		assert.Fail(t, fmt.Sprintf("no matching actual file for expected file %s", expectedFile.Path))
	}
}

//go:embed tsconfig.json
var sampleTsConfig string

func typecheck(t *testing.T, generatedFiles []*codegenerator.GeneratedFile) (output string, err error) {
	// todo: we need to generate a tsconfig to be able to run tsc for typechecking
	// however, when we come to use the testing package in real projects, there may already
	// be a tsconfig file that we need to respect

	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	f, err := os.Create(filepath.Join(tmpDir, "tsconfig.json"))

	if err != nil {
		return "", err
	}

	f.WriteString(sampleTsConfig)

	for _, file := range generatedFiles {
		f, err := os.Create(filepath.Join(tmpDir, file.Path))

		assert.NoError(t, err)

		_, err = f.WriteString(file.Contents)

		assert.NoError(t, err)
	}

	defer f.Close()
	cmd := exec.Command("npx", "tsc", "--noEmit", "--pretty", "--skipLibCheck", "--incremental", "--project", filepath.Base(f.Name()))
	cmd.Dir = tmpDir

	b, e := cmd.CombinedOutput()

	if e != nil {
		err = e
	}

	return string(b), err
}

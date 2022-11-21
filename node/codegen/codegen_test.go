package codegenerator_test

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/samber/lo"
	"github.com/sergi/go-diff/diffmatchpatch"
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

const (
	DIVIDER string = "=========================================================="
)

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
						export interface Person {
						  name: string
						  age: number
						  id: ID
						  createdAt: Date
						  updatedAt: Date
						}
						export declare type PersonQuery = {
						  name?: QueryConstraints.StringConstraint
						  age?: QueryConstraints.NumberConstraint
						  id?: QueryConstraints.IdConstraint
						  createdAt?: QueryConstraints.DateConstraint
						  updatedAt?: QueryConstraints.DateConstraint
						}
						export declare type PersonUniqueFields = {
							id?: QueryConstraints.IdConstraint
						}
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
					Path:     "index.d.ts",
					Contents: ``,
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
						export declare enum TheBeatles {
							John = "John",
							Paul = "Paul",
							Ringo = "Ringo",
							George = "George",
						}
					`,
				},
			},
		},
		{
			Name: "query-types",
			Schema: `
				model Person {
					fields {
						title Text
						subTitle Text?
					}
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
						export declare type PersonQuery = {
							title?: QueryConstraints.StringConstraint
							subTitle?: QueryConstraints.StringConstraint
							id?: QueryConstraints.IdConstraint
							createdAt?: QueryConstraints.DateConstraint
							updatedAt?: QueryConstraints.DateConstraint
						}
						export declare type PersonUniqueFields = {
							id?: QueryConstraints.IdConstraint
						}
					`,
				},
			},
		},
		{
			Name: "input-types",
			Schema: `
				model Person {
					fields {
						title Text
						subTitle Text?
						age Number?
					}

					operations {
						create createPerson() with(title, subTitle?)
						delete deletePerson(id)
						list listPeople(title, age)
						get getPerson(id)
						update updatePerson(id) with(title, subTitle?)
					}
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
						export interface CreatePersonInput {
							title: string
							subTitle?: string
						}
						export interface DeletePersonInput {
							id: ID
						}
						export interface ListPeopleInput {
							where: {
								title: StringConstraint
								age: NumberConstraint
							}
						}
						export interface GetPersonInput {
							id: ID
						}
						export interface UpdatePersonInput {
							where: {
								id: ID
							}
							values: {
								title: string
								subTitle?: string
							}
						}
						export interface AuthenticateInput {
							createIfNotExists?: boolean
							emailPassword: unknown
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

// Removes inconsequential differences between actual and expected strings:
// - Normalises tab / 2 space differences
// - Normalises indentation between actual vs expected
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

				if expected == "" {
					// was not asserted

					continue actual
				}

				if !strings.Contains(actual, expected) {
					// Attempting to perform a unified diff between a partial substring and a much larger actual string creates
					// some problems with the diff output - comprehension of what has changed within the substring is particularly tricky with the unified
					// diff output
					// Therefore, the call to matchPartial will try to match the expected string against the actual, returning only the relevant
					// portion of the actual string for diff display.
					actualPartial := matchPartial(actual, expected)
					diff := diffmatchpatch.New()
					diffs := diff.DiffMain(expected, actualPartial, true)

					fmt.Printf("Test case '%s' failed.\n%s\nContextual Diff:\n%s\n%s\nActual:\n%s\n%s\n\nExpected:\n%s\n%s\n",
						t.Name(),
						DIVIDER,
						DIVIDER,
						diff.DiffPrettyText(diffs),
						DIVIDER,
						actual,
						DIVIDER,
						expected,
					)
					t.Fail()
				}

				continue actual
			}
		}

		assert.Fail(t, fmt.Sprintf("no matching expectated file for actual file %s", actualFile.Path))
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

// After we have asserted that the actual and expected values match, we want to typecheck the outputted d.ts
// files using tsc to make sure that it is valid typescript!
func typecheck(t *testing.T, generatedFiles []*codegenerator.GeneratedFile) (output string, err error) {
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

	str := string(b)

	// tsc outputs ansi escape codes in its typechecking output
	// which don't render correctly in the vscode test output channel
	// (most likely not handled in the vscode-go plugin)
	str = stripansi.Strip(str)

	if e != nil {
		err = e
	}

	return str, err
}

// given a large body of text, and a partial string we want to match against
// returns the relevant location in the larger string where a match was found
func matchPartial(full, partial string) string {
	fullLines := strings.Split(full, "\n")
	partialLines := strings.Split(partial, "\n")

	firstPartial := partialLines[0]

	loc := 0
	match := false

	for _, fl := range fullLines {
		if firstPartial == fl {
			match = true
			break
		}

		loc++
	}

	// we found a match for the first line, so do the diff on this to avoid confusing
	// diffs between huge lhs and tiny rhs
	if match {
		subset := strings.Join(fullLines[loc:], "\n")

		return subset
	}

	return full
}

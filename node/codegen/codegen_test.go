package codegenerator_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	codegenerator "github.com/teamkeel/keel/node/codegen"
	"github.com/teamkeel/keel/schema"
)

type TestCase struct {
	Name                       string
	Schema                     string
	TypeScriptDefinitionOutput string
	JavaScriptOutput           string
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
			JavaScriptOutput: `
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
}`,
			TypeScriptDefinitionOutput: "",
		},
		{
			Name: "model-generation-custom-function",
			Schema: `
			model Person {
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
			JavaScriptOutput: `
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
			TypeScriptDefinitionOutput: "",
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

func normaliseString(str string) string {
	replacer := strings.NewReplacer(
		"  ", "\t",
	)

	return replacer.Replace(str)
}

func compareFiles(t *testing.T, tc TestCase, generatedFiles []*codegenerator.GeneratedFile) {
	for _, f := range generatedFiles {
		actual := normaliseString(f.Contents)

		expected := ""

		switch f.Type {
		case codegenerator.SourceCodeTypeJavaScript:
			expected = normaliseString(tc.JavaScriptOutput)
		case codegenerator.SourceCodeTypeDefinition:
			expected = normaliseString(tc.TypeScriptDefinitionOutput)
		}

		assert.Equal(t, expected, actual)
	}
}

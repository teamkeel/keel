package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

const testSchema = `
enum Gender {
	Male
	Female
}

model Person {
	fields {
		firstName Text @unique
		lastName Text?
		age Number
		dateOfBirth Date
		gender Gender
		hasChildren Boolean
	}
}
`

func TestWriteTableInterface(t *testing.T) {
	expected := `
export interface PersonTable {
	first_name: string
	last_name: string | null
	age: number
	date_of_birth: Date
	gender: Gender
	has_children: boolean
	id: Generated<string>
	created_at: Generated<Date>
	updated_at: Generated<Date>
}
`
	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeTableInterface(w, m)
	})
}

func TestWriteModelInterface(t *testing.T) {
	expected := `
export interface Person {
	firstName: string
	lastName: string | null
	age: number
	dateOfBirth: Date
	gender: Gender
	hasChildren: boolean
	id: string
	createdAt: Date
	updatedAt: Date
}
`
	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeModelInterface(w, m)
	})
}

func TestWriteCreateValuesInterface(t *testing.T) {
	expected := `
export interface PersonCreateValues {
	firstName: string
	lastName?: string | null
	age: number
	dateOfBirth: Date
	gender: Gender
	hasChildren: boolean
	id?: string
	createdAt?: Date
	updatedAt?: Date
}
`
	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeCreateValuesInterface(w, m)
	})
}

func TestWriteCreateValuesInterfaceWithRelationships(t *testing.T) {
	schema := `
	model Author {}
	model Post {
		fields {
			author Post
		}
	}
	`
	expected := `
export interface PostCreateValues {
	id?: string
	createdAt?: Date
	updatedAt?: Date
	authorId: string
}
`
	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Post")
		writeCreateValuesInterface(w, m)
	})
}

func TestWriteWhereConditionsInterface(t *testing.T) {
	expected := `
export interface PersonWhereConditions {
	firstName?: string | runtime.StringWhereCondition
	lastName?: string | runtime.StringWhereCondition | null
	age?: number | runtime.NumberWhereCondition
	dateOfBirth?: Date | runtime.DateWhereCondition
	gender?: Gender | GenderWhereCondition
	hasChildren?: boolean | runtime.BooleanWhereCondition
	id?: string | runtime.IDWhereCondition
	createdAt?: Date | runtime.DateWhereCondition
	updatedAt?: Date | runtime.DateWhereCondition
}
`
	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeWhereConditionsInterface(w, m)
	})
}

func TestWriteUniqueConditionsInterface(t *testing.T) {
	expected := `
export type PersonUniqueConditions = 
    | {firstName: string}
	| {id: string}`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeUniqueConditionsInterface(w, m)
	})
}

func TestWriteModelAPIDeclaration(t *testing.T) {
	expected := `
export type PersonAPI = {
	create(values: PersonCreateValues): Promise<Person>;
	update(where: PersonUniqueConditions, values: Partial<Person>): Promise<Person>;
	delete(where: PersonUniqueConditions): Promise<string>;
	findOne(where: PersonUniqueConditions): Promise<Person | null>;
	findMany(where: PersonWhereConditions): Promise<Person[]>;
	where(where: PersonWhereConditions): PersonQueryBuilder;
}`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeModelAPIDeclaration(w, m)
	})
}

func TestWriteEnum(t *testing.T) {
	expected := `
export enum Gender {
	Male = "Male",
	Female = "Female",
}`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		writeEnum(w, s.Enums[0])
	})
}

func TestWriteEnumWhereCondition(t *testing.T) {
	expected := `
export interface GenderWhereCondition {
	equals?: Gender
	oneOf?: Gender[]
}`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		writeEnumWhereCondition(w, s.Enums[0])
	})
}

func TestWriteDatabaseInterface(t *testing.T) {
	expected := `
interface database {
	person: PersonTable;
	identity: IdentityTable;
}
export declare function getDatabase(): Kysely<database>;`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		writeDatabaseInterface(w, s)
	})
}

func TestWriteAPIFactory(t *testing.T) {
	expected := `
function createFunctionAPI(headers) {
	const models = {
		person: new runtime.ModelAPI("person", personDefaultValues, null, tableConfigMap),
		identity: new runtime.ModelAPI("identity", identityDefaultValues, null, tableConfigMap),
	};
    return {models, headers};
}
function createContextAPI(meta) {
	const headers = new runtime.RequestHeaders(meta.headers);
	const identity = meta.identity;
    return {headers, identity};
}
module.exports.createFunctionAPI = createFunctionAPI;
module.exports.createContextAPI = createContextAPI;`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		writeAPIFactory(w, s.Models)
	})
}

func TestWriteAPIDeclarations(t *testing.T) {
	expected := `
export type ModelsAPI = {
    person: PersonAPI;
    identity: IdentityAPI;
}
export type FunctionAPI = {
	models: ModelsAPI;
	headers: Headers;
}
export interface ContextAPI extends runtime.ContextAPI {
	identity: Identity;
}`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		writeAPIDeclarations(w, s.Models)
	})
}

func TestWriteModelDefaultValuesFunction(t *testing.T) {
	schema := `
model Person {
	fields {
		name Text @default
		isAdmin Boolean @default
		counter Number @default
	}
}
	`
	expected := `
function personDefaultValues() {
	const r = {};
	r.name = "";
	r.isAdmin = false;
	r.counter = 0;
	r.id = runtime.ksuid();
	r.createdAt = new Date();
	r.updatedAt = new Date();
	return r;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeModelDefaultValuesFunction(w, m)
	})
}

func TestWriteActionInputTypesGet(t *testing.T) {
	schema := `
model Person {
	functions {
		get getPerson(id)
	}
}
	`
	expected := `
export interface GetPersonInput {
	id: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteActionInputTypesCreate(t *testing.T) {
	schema := `
model Person {
	fields {
		name Text
	}
	functions {
		create createPerson() with (name)
	}
}
	`
	expected := `
export interface CreatePersonInput {
	name: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteActionInputTypesCreateWithNull(t *testing.T) {
	schema := `
model Person {
	fields {
		name Text?
	}
	functions {
		create createPerson() with (name)
	}
}
	`
	expected := `
export interface CreatePersonInput {
	name: string | null;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteActionInputTypesUpdate(t *testing.T) {
	schema := `
model Person {
	fields {
		name Text
	}
	functions {
		update updatePerson(id) with (name)
	}
}
	`
	expected := `
export interface UpdatePersonInputWhere {
	id: string;
}
export interface UpdatePersonInputValues {
	name: string;
}
export interface UpdatePersonInput {
	where: UpdatePersonInputWhere;
	values: UpdatePersonInputValues;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteActionInputTypesList(t *testing.T) {
	schema := `
model Person {
	fields {
		name Text
	}
	functions {
		list listPeople(name, some: Boolean?)
	}
}
	`
	expected := `
export interface ListPeopleInputWhere {
	name: string;
	some?: boolean | null;
}
export interface ListPeopleInput {
	where: ListPeopleInputWhere;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteActionInputTypesListOperation(t *testing.T) {
	schema := `
enum Sport {
	Football
	Tennis
}
model Person {
	fields {
		name Text
		favouriteSport Sport
	}
	operations {
		list listPeople(name, favouriteSport)
	}
}
	`
	expected := `
export interface ListPeopleInputWhere {
	name: runtime.StringWhereCondition;
	favouriteSport: SportWhereCondition;
}
export interface ListPeopleInput {
	where: ListPeopleInputWhere;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteActionInputTypesDelete(t *testing.T) {
	schema := `
model Person {
	functions {
		delete deletePerson(id)
	}
}
	`
	expected := `
export interface DeletePersonInput {
	id: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeActionInputTypes(w, s, m.Operations[0], false)
	})
}

func TestWriteCustomFunctionWrapperType(t *testing.T) {
	schema := `
model Person {
	functions {
		get getPerson(id)
		create createPerson()
		update updatePerson()
		delete deletePerson()
		list listPeople()
	}
}
	`
	expected := `
export declare function GetPerson(fn: (inputs: GetPersonInput, api: FunctionAPI, ctx: ContextAPI) => Promise<Person | null>): Promise<Person | null>;
export declare function CreatePerson(fn: (inputs: CreatePersonInput, api: FunctionAPI, ctx: ContextAPI) => Promise<Person>): Promise<Person>;
export declare function UpdatePerson(fn: (inputs: UpdatePersonInput, api: FunctionAPI, ctx: ContextAPI) => Promise<Person>): Promise<Person>;
export declare function DeletePerson(fn: (inputs: DeletePersonInput, api: FunctionAPI, ctx: ContextAPI) => Promise<string>): Promise<string>;
export declare function ListPeople(fn: (inputs: ListPeopleInput, api: FunctionAPI, ctx: ContextAPI) => Promise<Person[]>): Promise<Person[]>;`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		for _, op := range m.Operations {
			writeCustomFunctionWrapperType(w, m, op)
		}
	})
}

func TestWriteTestingTypes(t *testing.T) {
	schema := `
model Person {
	operations {
		get getPerson(id)
		create createPerson()
	}
	functions {
		update updatePerson()
		delete deletePerson()
		list listPeople()
	}
}
	`
	expected := `
import * as sdk from "@teamkeel/sdk";
import * as runtime from "@teamkeel/functions-runtime";
import "@teamkeel/testing-runtime";

export interface GetPersonInput {
	id: string;
}
export interface CreatePersonInput {
}
export interface UpdatePersonInput {
}
export interface DeletePersonInput {
}
export interface ListPeopleInput {
}
export interface AuthenticateInput {
}
declare class ActionExecutor {
	withIdentity(identity: sdk.Identity): ActionExecutor;
	withAuthToken(token: string): ActionExecutor;
	getPerson(i: GetPersonInput): Promise<sdk.Person | null>;
	createPerson(i: CreatePersonInput): Promise<sdk.Person>;
	updatePerson(i: UpdatePersonInput): Promise<sdk.Person>;
	deletePerson(i: DeletePersonInput): Promise<string>;
	listPeople(i: ListPeopleInput): Promise<{results: sdk.Person[], hasNextPage: boolean}>;
	authenticate(i: AuthenticateInput): Promise<any>;
}
export declare const actions: ActionExecutor;
export declare const models: sdk.ModelsAPI;
export declare function resetDatabase(): Promise<void>;`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeTestingTypes(w, s)
	})
}

func TestWriteTableConfig(t *testing.T) {
	schema := `
model Publisher {
	fields {
		authors Author[]
	}
}
model Author {
	fields {
		publisher Publisher
		books Book[]
	}
}
model Book {
	fields {
		author Author
	}
}`
	expected := `
const tableConfigMap = {
	"author": {
		"books": {
			"foreignKey": "author_id",
			"referencesTable": "book",
			"relationshipType": "hasMany"
		},
		"publisher": {
			"foreignKey": "publisher_id",
			"referencesTable": "publisher",
			"relationshipType": "belongsTo"
		}
	},
	"book": {
		"author": {
			"foreignKey": "author_id",
			"referencesTable": "author",
			"relationshipType": "belongsTo"
		}
	},
	"publisher": {
		"authors": {
			"foreignKey": "publisher_id",
			"referencesTable": "author",
			"relationshipType": "hasMany"
		}
	}
};`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeTableConfig(w, s.Models)
	})
}

func TestWriteTestingTypesEnums(t *testing.T) {
	schema := `
enum Hobby {
	Tennis
	Chess
}
model Person {
	fields {
		hobby Hobby
	}
	operations {
		list peopleByHobby(hobby)
	}
}
	`
	expected := `
import * as sdk from "@teamkeel/sdk";
import * as runtime from "@teamkeel/functions-runtime";
import "@teamkeel/testing-runtime";

export interface PeopleByHobbyInputWhere {
	hobby: sdk.HobbyWhereCondition;
}
export interface PeopleByHobbyInput {
	where: PeopleByHobbyInputWhere;
}
export interface AuthenticateInput {
}
declare class ActionExecutor {
	withIdentity(identity: sdk.Identity): ActionExecutor;
	withAuthToken(token: string): ActionExecutor;
	peopleByHobby(i: PeopleByHobbyInput): Promise<{results: sdk.Person[], hasNextPage: boolean}>;
	authenticate(i: AuthenticateInput): Promise<any>;
}
export declare const actions: ActionExecutor;
export declare const models: sdk.ModelsAPI;
export declare function resetDatabase(): Promise<void>;`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeTestingTypes(w, s)
	})
}

func TestTestingActionExecutor(t *testing.T) {
	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = Bootstrap(tmpDir, WithPackagesPath(filepath.Join(wd, "../packages")))
	require.NoError(t, err)

	err = GeneratedFiles{
		{
			Contents: `
			model Person {
				functions {
					get getPerson(id)
				}
			}
			`,
			Path: filepath.Join(tmpDir, "schema.keel"),
		},
		{
			Contents: `
			import { actions } from "@teamkeel/testing";
			import { test, expect } from "vitest";

			test("action execution", async () => {
				const res = await actions.getPerson({id: "1234"});
				expect(res).toEqual({
					name: "Barney",
				});
			});

			test("toHaveAuthorizationError custom matcher", async () => {
				const p = Promise.reject({code: "ERR_PERMISSION_DENIED"});
				await expect(p).toHaveAuthorizationError();
			});
			`,
			Path: filepath.Join(tmpDir, "code.test.ts"),
		},
	}.Write()
	require.NoError(t, err)

	files, err := Generate(context.Background(), tmpDir)
	require.NoError(t, err)

	err = files.Write()
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		assert.True(t, strings.HasSuffix(r.URL.Path, "/getPerson"))

		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		type Payload struct {
			ID string
		}
		var payload Payload
		err = json.Unmarshal(b, &payload)
		assert.NoError(t, err)
		assert.Equal(t, "1234", payload.ID)

		_, err = w.Write([]byte(`{"name": "Barney"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	cmd := exec.Command("npx", "tsc", "--noEmit")
	cmd.Dir = tmpDir
	b, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		fmt.Println(string(b))
		t.FailNow()
	}

	cmd = exec.Command("npx", "vitest", "run", "--config", "./node_modules/@teamkeel/testing-runtime/vitest.config.mjs")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), []string{
		"DB_CONN_TYPE=pg",
		"DB_CONN=postgresql://postgres:postgres@localhost:8001/keel",
		fmt.Sprintf("KEEL_TESTING_ACTIONS_API_URL=%s", server.URL),
	}...)

	b, err = cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		fmt.Println(string(b))
	}
}

func TestSDKTypings(t *testing.T) {
	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = Bootstrap(tmpDir, WithPackagesPath(filepath.Join(wd, "../packages")))
	require.NoError(t, err)

	err = GeneratedFiles{
		{
			Path: filepath.Join(tmpDir, "schema.keel"),
			Contents: `
				model Person {
					fields {
						name Text
						lastName Text?
					}
					functions {
						get getPerson(id: Number)
					}
				}`,
		},
	}.Write()
	require.NoError(t, err)

	type fixture struct {
		name  string
		code  string
		error string
	}

	fixtures := []fixture{
		{
			name: "findOne",
			code: `
				import { GetPerson } from "@teamkeel/sdk";
		
				export default GetPerson((inputs, api) => {
					return api.models.person.findOne({
						id: inputs.id,
					});
				});
			`,
			error: "code.ts(6,7): error TS2322: Type 'number' is not assignable to type 'string'",
		},
		{
			name: "findOne - can return null",
			code: `
				import { GetPerson } from "@teamkeel/sdk";
		
				export default GetPerson(async (inputs, api) => {
					const r = await api.models.person.findOne({
						id: "1234",
					});
					console.log(r.id);
					return r;
				});
			`,
			error: "code.ts(8,18): error TS18047: 'r' is possibly 'null'",
		},
		{
			name: "findMany - correct typings on where condition",
			code: `
				import { GetPerson } from "@teamkeel/sdk";
		
				export default GetPerson(async (inputs, api) => {
					const r = await api.models.person.findMany({
						name: {
							startsWith: true,
						}
					});
					return r.length > 0 ? r[0] : null;
				});
			`,
			error: "code.ts(7,8): error TS2322: Type 'boolean' is not assignable to type 'string'",
		},
		{
			name: "optional model fields are typed as nullable",
			code: `
				import { GetPerson } from "@teamkeel/sdk";
		
				export default GetPerson(async (inputs, api) => {
					const person = await api.models.person.findOne({
						id: "1234",
					});
					if (person) {
						person.lastName.toUpperCase();
					}
					return person;
				});
			`,
			error: "code.ts(9,7): error TS18047: 'person.lastName' is possibly 'null'",
		},
		{
			name: "testing actions executor - input types",
			code: `
				import { actions } from "@teamkeel/testing";
		
				async function foo() {
					await actions.getPerson({
						id: "1234",
					});
				}
			`,
			error: "code.ts(6,7): error TS2322: Type 'string' is not assignable to type 'number'",
		},
		{
			name: "testing actions executor - return types",
			code: `
				import { actions } from "@teamkeel/testing";
		
				async function foo() {
					const p = await actions.getPerson({
						id: 1234,
					});
					console.log(p.id);
				}
			`,
			error: "code.ts(8,18): error TS18047: 'p' is possibly 'null'",
		},
		{
			name: "testing actions executor - withIdentity",
			code: `
				import { actions } from "@teamkeel/testing";
		
				async function foo() {
					await actions.withIdentity(null).getPerson({
						id: 1234,
					});
				}
			`,
			error: "code.ts(5,33): error TS2345: Argument of type 'null' is not assignable to parameter of type 'Identity'",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			err := GeneratedFiles{
				{
					Path:     filepath.Join(tmpDir, "code.ts"),
					Contents: fixture.code,
				},
			}.Write()
			require.NoError(t, err)

			files, err := Generate(context.Background(), tmpDir)
			require.NoError(t, err)

			err = files.Write()
			require.NoError(t, err)

			c := exec.Command("npx", "tsc", "--noEmit")
			c.Dir = tmpDir
			b, _ := c.CombinedOutput()
			assert.Contains(t, string(b), fixture.error)
		})
	}
}

func normalise(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\t", "    ")
}

func runWriterTest(t *testing.T, schemaString string, expected string, fn func(s *proto.Schema, w *Writer)) {
	b := schema.Builder{}
	s, err := b.MakeFromString(schemaString)
	require.NoError(t, err)
	w := &Writer{}
	fn(s, w)
	diff := diffmatchpatch.New()
	diffs := diff.DiffMain(normalise(expected), normalise(w.String()), true)
	if lo.SomeBy(diffs, func(d diffmatchpatch.Diff) bool {
		return d.Type != diffmatchpatch.DiffEqual
	}) {
		t.Errorf("generated code does not match expected:\n%s", diffPrettyText(diffs))
	}
}

// diffPrettyText is a port of the same function from the diffmatchpatch
// lib but with better handling of whitespace diffs (by using background colours)
func diffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff strings.Builder

	green := color.New(color.FgGreen)
	green.EnableColor()
	red := color.New(color.FgRed)
	red.EnableColor()
	bgGreen := color.New(color.BgGreen, color.FgWhite)
	bgGreen.EnableColor()
	bgRed := color.New(color.BgRed, color.FgWhite)
	bgRed.EnableColor()

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			if strings.TrimSpace(diff.Text) == "" {
				buff.WriteString(bgGreen.Sprint(diff.Text))
			} else {
				buff.WriteString(green.Sprint(diff.Text))
			}
		case diffmatchpatch.DiffDelete:
			if strings.TrimSpace(diff.Text) == "" {
				buff.WriteString(bgRed.Sprintf("%s", diff.Text))
			} else {
				buff.WriteString(red.Sprint(diff.Text))
			}
		case diffmatchpatch.DiffEqual:
			buff.WriteString(diff.Text)
		}
	}

	return buff.String()
}

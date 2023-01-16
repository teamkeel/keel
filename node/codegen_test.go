package node

import (
	"context"
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
	oneOf?: [Gender]
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
function createFunctionAPI() {
	const models = {
		person: new runtime.ModelAPI("person", personDefaultValues),
		identity: new runtime.ModelAPI("identity", identityDefaultValues),
	};
    return {models};
}
module.exports.createFunctionAPI = createFunctionAPI;`

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
	r.id = KSUID.randomSync().string;
	r.createdAt = new Date();
	r.updatedAt = new Date();
	return r;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeModelDefaultValuesFunction(w, m)
	})
}

func TestWriteCustomFunctionInputTypesGet(t *testing.T) {
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
		writeCustomFunctionInputTypes(w, m.Operations[0])
	})
}

func TestWriteCustomFunctionInputTypesCreate(t *testing.T) {
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
		writeCustomFunctionInputTypes(w, m.Operations[0])
	})
}

func TestWriteCustomFunctionInputTypesUpdate(t *testing.T) {
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
		writeCustomFunctionInputTypes(w, m.Operations[0])
	})
}

func TestWriteCustomFunctionInputTypesList(t *testing.T) {
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
	some?: boolean;
}
export interface ListPeopleInput {
	where: ListPeopleInputWhere;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeCustomFunctionInputTypes(w, m.Operations[0])
	})
}

func TestWriteCustomFunctionInputTypesDelete(t *testing.T) {
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
		writeCustomFunctionInputTypes(w, m.Operations[0])
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
export declare function GetPerson(fn: (inputs: GetPersonInput, api: FunctionAPI) => Promise<Person | null>): Promise<Person | null>;
export declare function CreatePerson(fn: (inputs: CreatePersonInput, api: FunctionAPI) => Promise<Person>): Promise<Person>;
export declare function UpdatePerson(fn: (inputs: UpdatePersonInput, api: FunctionAPI) => Promise<Person>): Promise<Person>;
export declare function DeletePerson(fn: (inputs: DeletePersonInput, api: FunctionAPI) => Promise<string>): Promise<string>;
export declare function ListPeople(fn: (inputs: ListPeopleInput, api: FunctionAPI) => Promise<Person[]>): Promise<Person[]>;`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		for _, op := range m.Operations {
			writeCustomFunctionWrapperType(w, m, op)
		}
	})
}

func TestSDKTypings(t *testing.T) {
	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = Bootstrap(tmpDir, WithPackagesPath(filepath.Join(wd, "../packages")))
	require.NoError(t, err)

	schema := []byte(`
		model Person {
			fields {
				name Text
				lastName Text?
			}
			functions {
				get getPerson(id: Number)
			}
		}
	`)
	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), schema, 0666)
	require.NoError(t, err)

	type fixture struct {
		name     string
		function string
		error    string
	}

	fixtures := []fixture{
		{
			name: "findOne",
			function: `
				import { GetPerson } from "@teamkeel/sdk";
		
				export default GetPerson((inputs, api) => {
					return api.models.person.findOne({
						id: inputs.id,
					});
				});
			`,
			error: "myFunction.ts(6,7): error TS2322: Type 'number' is not assignable to type 'string'",
		},
		{
			name: "findOne - can return null",
			function: `
				import { GetPerson } from "@teamkeel/sdk";
		
				export default GetPerson(async (inputs, api) => {
					const r = await api.models.person.findOne({
						id: "1234",
					});
					console.log(r.id);
					return r;
				});
			`,
			error: "myFunction.ts(8,18): error TS18047: 'r' is possibly 'null'",
		},
		{
			name: "findMany - correct typings on where condition",
			function: `
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
			error: "myFunction.ts(7,8): error TS2322: Type 'boolean' is not assignable to type 'string'",
		},
		{
			name: "optional model fields are typed as nullable",
			function: `
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
			error: "myFunction.ts(9,7): error TS18047: 'person.lastName' is possibly 'null'",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {

			function := []byte(fixture.function)
			err = os.WriteFile(filepath.Join(tmpDir, "myFunction.ts"), function, 0666)
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

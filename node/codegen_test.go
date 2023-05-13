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

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/colors"
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
	firstName?: string | runtime.StringWhereCondition | null;
	lastName?: string | runtime.StringWhereCondition | null;
	age?: number | runtime.NumberWhereCondition | null;
	dateOfBirth?: Date | runtime.DateWhereCondition | null;
	gender?: Gender | GenderWhereCondition | null;
	hasChildren?: boolean | runtime.BooleanWhereCondition | null;
	id?: string | runtime.IDWhereCondition | null;
	createdAt?: Date | runtime.DateWhereCondition | null;
	updatedAt?: Date | runtime.DateWhereCondition | null;
}`
	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Person")
		writeWhereConditionsInterface(w, m)
	})
}

func TestWriteUniqueConditionsInterface(t *testing.T) {
	schema := `
	model Author {
		fields {
			books Book[]
		}
	}
	model Book {
		fields {
			title Text @unique
			author Author
		}
	}
	`

	// You can't find a single book by author, because an author
	// writes many books
	expectedBookType := `
export type BookUniqueConditions = 
	| {title: string}
	| {id: string};
	`

	// You can find a single author by a book, because a book
	// is written by a single author. So we include the
	// BookUniqueConditions type within AuthorUniqueConditions
	expectedAuthorType := `
export type AuthorUniqueConditions = 
	| {books: BookUniqueConditions}
	| {id: string};
	`

	runWriterTest(t, schema, expectedBookType, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Book")
		writeUniqueConditionsInterface(w, m)
	})

	runWriterTest(t, schema, expectedAuthorType, func(s *proto.Schema, w *Writer) {
		m := proto.FindModel(s.Models, "Author")
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
	equals?: Gender | null;
	oneOf?: Gender[] | null;
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

func TestWriteDevelopmentServer(t *testing.T) {
	expected := `
const functions = {
	createPost: function_createPost,
	updatePost: function_updatePost,
}
const actionTypes = {
	createPost: "OPERATION_TYPE_CREATE",
	updatePost: "OPERATION_TYPE_UPDATE",
}
	`

	schema := `
		model Post {
			fields {
				title Text
			}

			functions {
				create createPost() with(title)
				update updatePost(id) with(title)
			}
		}
	`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		files := generateDevelopmentServer(s)

		serverJs := files[0]

		w.Write(serverJs.Contents)
	})
}

func TestWriteAPIFactory(t *testing.T) {
	expected := `
function createFunctionAPI({ db, meta }) {
	const models = {
		person: new runtime.ModelAPI("person", personDefaultValues, db, tableConfigMap),
		identity: new runtime.ModelAPI("identity", identityDefaultValues, db, tableConfigMap),
	};
	const { permissionState } = meta || { permissionState: { status: 'unknown' } };
	return { models, permissions: new runtime.Permissions(permissionState) };
};
function createContextAPI({ responseHeaders, meta }) {
	const headers = new runtime.RequestHeaders(meta.headers);
	const response = { headers: responseHeaders }
	const now = () => { return new Date(); };
	const { identity } = meta;
	const isAuthenticated = identity != null;
	const env = {
		TEST: process.env["TEST"] || "",
	};
	const secrets = {
		SECRET_KEY: meta.secrets.SECRET_KEY || "",
	};
	return { headers, response, identity, env, now, secrets, isAuthenticated };
}
module.exports.createFunctionAPI = createFunctionAPI;
module.exports.createContextAPI = createContextAPI;`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		s.EnvironmentVariables = append(s.EnvironmentVariables, &proto.EnvironmentVariable{
			Name: "TEST",
		})
		s.Secrets = append(s.Secrets, &proto.Secret{
			Name: "SECRET_KEY",
		})

		writeAPIFactory(w, s)
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
	fetch(input: RequestInfo | URL, init?: RequestInit | undefined): Promise<Response>;
	permissions: runtime.Permissions;
}
type Environment = {
	TEST: string;
}
type Secrets = {
	SECRET_KEY: string;
}

export interface ContextAPI extends runtime.ContextAPI {
	secrets: Secrets;
	env: Environment;
	identity?: Identity;
	now(): Date;
}`

	runWriterTest(t, testSchema, expected, func(s *proto.Schema, w *Writer) {
		s.EnvironmentVariables = append(s.EnvironmentVariables, &proto.EnvironmentVariable{
			Name: "TEST",
		})
		s.Secrets = append(s.Secrets, &proto.Secret{
			Name: "SECRET_KEY",
		})

		writeAPIDeclarations(w, s)
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
		writeMessages(w, s, false)
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
		writeMessages(w, s, false)
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
	name?: string | null;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
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
export interface UpdatePersonWhere {
	id: string;
}
export interface UpdatePersonValues {
	name: string;
}
export interface UpdatePersonInput {
	where: UpdatePersonWhere;
	values: UpdatePersonValues;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
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
export interface StringQueryInput {
	equals?: string | null;
	notEquals?: string | null;
	startsWith?: string | null;
	endsWith?: string | null;
	contains?: string | null;
	oneOf?: string[] | null;
}
export interface ListPeopleWhere {
	name: StringQueryInput;
	some?: boolean | null;
}
export interface ListPeopleInput {
	where: ListPeopleWhere;
	first?: number | null;
	after?: string | null;
	last?: number | null;
	before?: string | null;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
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
export interface StringQueryInput {
	equals?: string | null;
	notEquals?: string | null;
	startsWith?: string | null;
	endsWith?: string | null;
	contains?: string | null;
	oneOf?: string[] | null;
}
export interface SportQueryInput {
	equals?: Sport | null;
	notEquals?: Sport | null;
	oneOf?: Sport[] | null;
}
export interface ListPeopleWhere {
	name: StringQueryInput;
	favouriteSport: SportQueryInput;
}
export interface ListPeopleInput {
	where: ListPeopleWhere;
	first?: number | null;
	after?: string | null;
	last?: number | null;
	before?: string | null;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
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
		writeMessages(w, s, false)
	})
}

func TestWriteActionInputTypesInlineInputRead(t *testing.T) {
	schema := `
message PersonNameResponse {
	name Text
}

model Person {
	functions {
		read getPersonName(id) returns (PersonNameResponse)
	}
}`
	expected := `
export interface GetPersonNameInput {
	id: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionInputTypesMessageInputRead(t *testing.T) {
	schema := `
message PersonNameResponse {
	name Text
}

message GetInput {
	id ID
}

model Person {
	functions {
		read deletePerson(GetInput) returns (PersonNameResponse)
	}
}
	`
	expected := `
export interface GetInput {
	id: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionResponseTypesRead(t *testing.T) {
	schema := `
message PersonNameResponse {
	name Text
}

message GetInput {
	id ID
}

model Person {
	functions {
		read deletePerson(GetInput) returns (PersonNameResponse)
	}
}
	`
	expected := `
export interface PersonNameResponse {
	name: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionInputTypesInlineInputWrite(t *testing.T) {
	schema := `
message DeleteResponse {
	isDeleted Boolean
}

model Person {
	functions {
		write deletePerson(id) returns (DeleteResponse)
	}
}`
	expected := `
export interface DeletePersonInput {
	id: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionInputTypesMessageInputWrite(t *testing.T) {
	schema := `
message DeleteResponse {
	isDeleted Boolean
}

message DeleteInput {
	id ID
}

model Person {
	functions {
		write deletePerson(DeleteInput) returns (DeleteResponse)
	}
}
	`
	expected := `
export interface DeleteInput {
	id: string;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionResponseTypesWrite(t *testing.T) {
	schema := `
message DeleteResponse {
	isDeleted Boolean
}

message DeleteInput {
	id ID
}

model Person {
	functions {
		read deletePerson(DeleteInput) returns (DeleteResponse)
	}
}
	`
	expected := `
export interface DeleteResponse {
	isDeleted: boolean;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionInputTypesArrayField(t *testing.T) {
	schema := `
message PeopleInput {
	ids ID[]
}

message People {
	names Text[]
}

model Person {
	functions {
		read readPerson(PeopleInput) returns (People)
	}
}`
	expected := `
export interface PeopleInput {
	ids: string[];
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestMessageFieldAnyType(t *testing.T) {
	schema := `
	message Foo {
		bar Any
	}

	model Person {
		functions {
			read getPerson(Foo) returns(Foo)
		}
	}
	`
	expected := `
export interface Foo {
    bar: any;
}
	`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionTypesEnumField(t *testing.T) {
	schema := `
message Input {
	sports Sport[]
	favouriteSport Sport?
}

message Response {
	sports Sport[]
	favouriteSport Sport?
}

model Person {
	functions {
		write writeSportInterests(Input) returns (Response)
	}
}

enum Sport {
	Cricket
	Rugby
	Soccer
}`
	inputExpected := `
export interface Input {
	sports: Sport[];
	favouriteSport?: Sport | null;
}`
	responseExpected := `
export interface Response {
	sports: Sport[];
	favouriteSport?: Sport | null;
}`

	runWriterTest(t, schema, inputExpected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})

	runWriterTest(t, schema, responseExpected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionResponseTypesArrayField(t *testing.T) {
	schema := `
message People {
	names Text[]
}

model Person {
	functions {
		read readPerson(name: Text) returns (People)
	}
}`
	expected := `
export interface People {
	names: string[];
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionResponseTypesArrayNestedMessage(t *testing.T) {
	schema := `
message People {
	names Details[]
}

message Details {
	names Text
}

model Person {
	functions {
		read readPerson(name: Text) returns (People)
	}
}`
	expected := `
export interface People {
	names: Details[];
}
export interface Details {
	names: string;
}
`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
	})
}

func TestWriteActionResponseTypesNestedModels(t *testing.T) {
	schema := `
message PersonResponse {
	sales Sale[]
	person Person
	topSale Sale?
}

model Person {
	functions {
		read readPerson(id) returns (PersonResponse)
	}
}

model Sale {

}
	`
	expected := `
export interface PersonResponse {
	sales: Sale[];
	person: Person;
	topSale?: Sale | null;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *Writer) {
		writeMessages(w, s, false)
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
export declare function GetPerson(fn: (ctx: ContextAPI, inputs: GetPersonInput, api: FunctionAPI) => Promise<Person | null>): Promise<Person | null>;
export declare function CreatePerson(fn: (ctx: ContextAPI, inputs: CreatePersonInput, api: FunctionAPI) => Promise<Person>): Promise<Person>;
export declare function UpdatePerson(fn: (ctx: ContextAPI, inputs: UpdatePersonInput, api: FunctionAPI) => Promise<Person>): Promise<Person>;
export declare function DeletePerson(fn: (ctx: ContextAPI, inputs: DeletePersonInput, api: FunctionAPI) => Promise<string>): Promise<string>;
export declare function ListPeople(fn: (ctx: ContextAPI, inputs: ListPeopleInput, api: FunctionAPI) => Promise<Person[]>): Promise<Person[]>;`

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
}`

	expected := `
import * as sdk from "@teamkeel/sdk";
import * as runtime from "@teamkeel/functions-runtime";
import "@teamkeel/testing-runtime";

export interface GetPersonInput {
	id: string;
}
export interface CreatePersonInput {
}
export interface UpdatePersonWhere {
}
export interface UpdatePersonValues {
}
export interface UpdatePersonInput {
	where?: UpdatePersonWhere | null;
	values?: UpdatePersonValues | null;
}
export interface DeletePersonInput {
}
export interface ListPeopleWhere {
}
export interface ListPeopleInput {
	where?: ListPeopleWhere | null;
	first?: number | null;
	after?: string | null;
	last?: number | null;
	before?: string | null;
}
export interface EmailPasswordInput {
	email: string;
	password: string;
}
export interface AuthenticateInput {
	createIfNotExists?: boolean | null;
	emailPassword: EmailPasswordInput;
}
export interface AuthenticateResponse {
	identityCreated: boolean;
	token: string;
}
export interface RequestPasswordResetInput {
	email: string;
	redirectUrl: string;
}
export interface RequestPasswordResetResponse {
}
export interface ResetPasswordInput {
	token: string;
	password: string;
}
export interface ResetPasswordResponse {
}
declare class ActionExecutor {
	withIdentity(identity: sdk.Identity): ActionExecutor;
	withAuthToken(token: string): ActionExecutor;
	getPerson(i: GetPersonInput): Promise<sdk.Person | null>;
	createPerson(i?: CreatePersonInput): Promise<sdk.Person>;
	updatePerson(i?: UpdatePersonInput): Promise<sdk.Person>;
	deletePerson(i?: DeletePersonInput): Promise<string>;
	listPeople(i?: ListPeopleInput): Promise<{results: sdk.Person[], pageInfo: runtime.PageInfo}>;
	authenticate(i: AuthenticateInput): Promise<AuthenticateResponse>;
	requestPasswordReset(i: RequestPasswordResetInput): Promise<RequestPasswordResetResponse>;
	resetPassword(i: ResetPasswordInput): Promise<ResetPasswordResponse>;
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

export interface HobbyQueryInput {
	equals?: Hobby | null;
	notEquals?: Hobby | null;
	oneOf?: Hobby[] | null;
}
export interface PeopleByHobbyWhere {
	hobby: HobbyQueryInput;
}
export interface PeopleByHobbyInput {
	where: PeopleByHobbyWhere;
	first?: number | null;
	after?: string | null;
	last?: number | null;
	before?: string | null;
}
export interface EmailPasswordInput {
	email: string;
	password: string;
}
export interface AuthenticateInput {
	createIfNotExists?: boolean | null;
	emailPassword: EmailPasswordInput;
}
export interface AuthenticateResponse {
	identityCreated: boolean;
	token: string;
}
export interface RequestPasswordResetInput {
	email: string;
	redirectUrl: string;
}
export interface RequestPasswordResetResponse {
}
export interface ResetPasswordInput {
	token: string;
	password: string;
}
export interface ResetPasswordResponse {
}
declare class ActionExecutor {
	withIdentity(identity: sdk.Identity): ActionExecutor;
	withAuthToken(token: string): ActionExecutor;
	peopleByHobby(i: PeopleByHobbyInput): Promise<{results: sdk.Person[], pageInfo: runtime.PageInfo}>;
	authenticate(i: AuthenticateInput): Promise<AuthenticateResponse>;
	requestPasswordReset(i: RequestPasswordResetInput): Promise<RequestPasswordResetResponse>;
	resetPassword(i: ResetPasswordInput): Promise<ResetPasswordResponse>;
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
			Path: "schema.keel",
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
			Path: "code.test.ts",
		},
	}.Write(tmpDir)
	require.NoError(t, err)

	builder := schema.Builder{}
	schema, err := builder.MakeFromDirectory(tmpDir)
	require.NoError(t, err)

	files, err := Generate(context.Background(), schema)
	require.NoError(t, err)

	err = files.Write(tmpDir)
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

	cmd = exec.Command("npx", "vitest", "run", "--config", ".build/vitest.config.mjs")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), []string{
		"KEEL_DB_CONN_TYPE=pg",
		"KEEL_DB_CONN=postgresql://postgres:postgres@localhost:8001/keel",
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
			Path: "schema.keel",
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
	}.Write(tmpDir)
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
		
				export default GetPerson((_, inputs, api) => {
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
		
				export default GetPerson(async (_, inputs, api) => {
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
		
				export default GetPerson(async (_, inputs, api) => {
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
		
				export default GetPerson(async (_, inputs, api) => {
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
					Path:     "code.ts",
					Contents: fixture.code,
				},
			}.Write(tmpDir)
			require.NoError(t, err)

			builder := schema.Builder{}
			schema, err := builder.MakeFromDirectory(tmpDir)
			require.NoError(t, err)

			files, err := Generate(context.Background(), schema)
			require.NoError(t, err)

			err = files.Write(tmpDir)
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
	if !strings.Contains(normalise(w.String()), normalise(expected)) {
		t.Errorf("generated code does not match expected:\n%s", diffPrettyText(diffs))

		t.Errorf("\nExpected:\n---------\n%s", normalise(expected))
		t.Errorf("\nActual:\n---------\n%s", normalise(w.String()))
	}
}

// diffPrettyText is a port of the same function from the diffmatchpatch
// lib but with better handling of whitespace diffs (by using background colours)
func diffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff strings.Builder

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			if strings.TrimSpace(diff.Text) == "" {
				buff.WriteString(colors.Green(fmt.Sprint(diff.Text)).String())
			} else {
				buff.WriteString(colors.Green(fmt.Sprint(diff.Text)).Highlight().String())
			}
		case diffmatchpatch.DiffDelete:
			if strings.TrimSpace(diff.Text) == "" {
				buff.WriteString(colors.Red(diff.Text).String())
			} else {
				buff.WriteString(colors.Red(fmt.Sprint(diff.Text)).Highlight().String())
			}
		case diffmatchpatch.DiffEqual:
			buff.WriteString(diff.Text)
		}
	}

	return buff.String()
}

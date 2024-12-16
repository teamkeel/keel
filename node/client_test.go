package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

func TestClientActionsByApi(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	actions {
		get getPerson(id)
	}
}

api Api {
	models {
		Person
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		getPerson: (i: GetPersonInput) => Promise<APIResult<Person | null>>;
	},
	mutations: {
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionsByApiExcludedAction(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	actions {
		get getPerson(id)
	}
}
model Company {
	actions {
		delete deleteCompany(id)
		get getCompany(id)
	}
}

api Api {
	models {
		Person
		Company {
			actions {
				getCompany
			}
		}
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		getPerson: (i: GetPersonInput) => Promise<APIResult<Person | null>>;
		getCompany: (i: GetCompanyInput) => Promise<APIResult<Company | null>>;
	},
	mutations: {
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionsByDifferentApi(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	actions {
		get getPerson(id)
	}
}

api Web {
	models {
		Person
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		getPerson: (i: GetPersonInput) => Promise<APIResult<Person | null>>;
	},
	mutations: {
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Web")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionGet(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		get getPerson(id)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		getPerson: (i: GetPersonInput) => Promise<APIResult<Person | null>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionCreate(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		create createPerson() with (name)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
	},
	mutations: {
		createPerson: (i: CreatePersonInput) => Promise<APIResult<Person>>;
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionUpdate(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		update updatePerson(id) with (name)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
	},
	mutations: {
		updatePerson: (i: UpdatePersonInput) => Promise<APIResult<Person>>;
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionDelete(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		delete deletePerson(id)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
	},
	mutations: {
		deletePerson: (i: DeletePersonInput) => Promise<APIResult<string>>;
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionList(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		list listPeople(name)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		listPeople: (i: ListPeopleInput) => Promise<APIResult<{ results: Person[], pageInfo: PageInfo }>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

// TODO: Fix this.  It should be:
//
//	createPerson: () => ...
func TestClientActionNoInputs(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	actions {
		create createPerson()
		list listPeople()
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		listPeople: (i?: ListPeopleInput) => Promise<APIResult<{ results: Person[], pageInfo: PageInfo }>>;
	},
	mutations: {
		createPerson: (i?: CreatePersonInput) => Promise<APIResult<Person>>;
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionAllOptionalInputs(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
		age Number?
	}
	actions {
		list listPeople(name?, age?)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		listPeople: (i?: ListPeopleInput) => Promise<APIResult<{ results: Person[], pageInfo: PageInfo }>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionNotAllOptionalInputs(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
		age Number?
	}
	actions {
		list listPeople(name, age?)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		listPeople: (i: ListPeopleInput) => Promise<APIResult<{ results: Person[], pageInfo: PageInfo }>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionRead(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
		age Number?
	}
	actions {
		read readPeople(name, age) returns (People)
	}
}

message People {
	people Person[]
}`

	expected := `
interface KeelAPI {
	queries: {
		readPeople: (i: ReadPeopleInput) => Promise<APIResult<People>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionWrite(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
		age Number?
	}
	actions {
		write writePeople(name, age) returns (People)
	}
}

message People {
	people Person[]
}`

	expected := `
interface KeelAPI {
	queries: {
	},
	mutations: {
		writePeople: (i: WritePeopleInput) => Promise<APIResult<People>>;
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionMessageInput(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	actions {
		read readPeople(SearchParams) returns (People)
	}
}

message SearchParams {
	names Text[]
}

message People {
	people Person[]
}`

	expected := `
interface KeelAPI {
	queries: {
		readPeople: (i: SearchParams) => Promise<APIResult<People>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientActionAny(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	actions {
		read readPeople(Any) returns (Any)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		readPeople: (i?: any) => Promise<APIResult<any>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientApiDefinition(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		get getPerson(id)
		create createPerson() with (name)
		update updatePerson(id) with (name)
		delete deletePerson(id)
		list listPeople(name)
		read readPeople(Any) returns (Any)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		getPerson: (i: GetPersonInput) => Promise<APIResult<Person | null>>;
		listPeople: (i: ListPeopleInput) => Promise<APIResult<{ results: Person[], pageInfo: PageInfo }>>;
		readPeople: (i?: any) => Promise<APIResult<any>>;
	},
	mutations: {
		createPerson: (i: CreatePersonInput) => Promise<APIResult<Person>>;
		updatePerson: (i: UpdatePersonInput) => Promise<APIResult<Person>>;
		deletePerson: (i: DeletePersonInput) => Promise<APIResult<string>>;
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiInterface(w, s, api)
	})
}

func TestClientApiTypes(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
		age Number
		isRegistered Boolean
	}
	actions {
		get getPerson(id)
		list listPeople(name)
	}
}`

	expected := `
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
export interface GetPersonInput {
	id: string;
}
export interface StringQueryInput {
	equals?: string | null;
	notEquals?: string | null;
	startsWith?: string;
	endsWith?: string;
	contains?: string;
	oneOf?: string[];
}
export interface ListPeopleWhere {
	name: StringQueryInput;
}
export interface ListPeopleInput {
	where: ListPeopleWhere;
	first?: number;
	after?: string;
	last?: number;
	before?: string;
}
export interface Person {
	name: string
	age: number
	isRegistered: boolean
	id: string
	createdAt: Date
	updatedAt: Date
}
export interface Identity {
	email: string | null
	emailVerified: boolean
	password: any | null
	externalId: string | null
	issuer: string | null
	name: string | null
	givenName: string | null
	familyName: string | null
	middleName: string | null
	nickName: string | null
	profile: string | null
	picture: string | null
	website: string | null
	gender: string | null
	zoneInfo: string | null
	locale: string | null
	id: string
	createdAt: Date
	updatedAt: Date
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientTypes(w, s, api)
	})
}

func TestClientCoreClass(t *testing.T) {
	t.Parallel()
	schema := `
model Person {
	fields {
		name Text
	}
	actions {
		get getPerson(id)
	}
}`

	expected := `
interface KeelAPI {
	queries: {
		getPerson: (i: GetPersonInput) => Promise<APIResult<Person | null>>;
	},
	mutations: {
		requestPasswordReset: (i: RequestPasswordResetInput) => Promise<APIResult<RequestPasswordResetResponse>>;
		resetPassword: (i: ResetPasswordInput) => Promise<APIResult<ResetPasswordResponse>>;
	}
}

export class APIClient extends Core {
	constructor(config: Config) {
		super(config);
	}

	api = {
		queries: new Proxy({}, {
			get: (_, fn: string) => (i: any) => this.client.rawRequest(fn, i),
		}),
		mutations: new Proxy({}, {
			get: (_, fn: string) => (i: any) => this.client.rawRequest(fn, i),
		})
	} as KeelAPI;
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiClass(w, s, api)
	})
}

func TestGenerateClientFiles(t *testing.T) {
	t.Parallel()
	schemaString := `
model Person {
	actions {
		get getPerson(id)
	}
}`

	b := schema.Builder{}
	schema, err := b.MakeFromString(schemaString, config.Empty)
	require.NoError(t, err)

	files, err := GenerateClient(context.Background(), schema, false, "Api")
	require.NoError(t, err)

	require.Len(t, files, 1)
	require.Equal(t, files[0].Path, "keelClient.ts")
}

func TestGenerateClientPackagesFiles(t *testing.T) {
	t.Parallel()
	schemaString := `
model Person {
	actions {
		get getPerson(id)
	}
}`

	b := schema.Builder{}
	schema, err := b.MakeFromString(schemaString, config.Empty)
	require.NoError(t, err)

	files, err := GenerateClient(context.Background(), schema, true, "Api")
	require.NoError(t, err)

	require.Len(t, files, 4)
	require.Equal(t, files[0].Path, "@teamkeel/client/core.ts")
	require.Equal(t, files[1].Path, "@teamkeel/client/index.ts")
	require.Equal(t, files[2].Path, "@teamkeel/client/types.ts")
	require.Equal(t, files[3].Path, "@teamkeel/client/package.json")
}

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
getPerson: (i: GetPersonInput) => {
	return this.client.rawRequest<Person | null>("getPerson", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionsByApiExcludedAction(t *testing.T) {
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
getPerson: (i: GetPersonInput) => {
	return this.client.rawRequest<Person | null>("getPerson", i);
},
getCompany: (i: GetCompanyInput) => {
	return this.client.rawRequest<Company | null>("getCompany", i);
},`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionsByDifferentApi(t *testing.T) {
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
getPerson: (i: GetPersonInput) => {
	return this.client.rawRequest<Person | null>("getPerson", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Web")
		writeClientActions(w, s, api)
	})
}

func TestClientActionGet(t *testing.T) {
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
getPerson: (i: GetPersonInput) => {
	return this.client.rawRequest<Person | null>("getPerson", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionCreate(t *testing.T) {
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
createPerson: (i: CreatePersonInput) => {
	return this.client.rawRequest<Person>("createPerson", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionUpdate(t *testing.T) {
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
updatePerson: (i: UpdatePersonInput) => {
	return this.client.rawRequest<Person>("updatePerson", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionDelete(t *testing.T) {
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
deletePerson: (i: DeletePersonInput) => {
	return this.client.rawRequest<string>("deletePerson", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionList(t *testing.T) {
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
listPeople: (i: ListPeopleInput) => {
	return this.client.rawRequest<{results: Person[], pageInfo: PageInfo}>("listPeople", i);
},`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

// TODO: Fix this.  It should be:
//
//	createPerson: () => ...
func TestClientActionNoInputs(t *testing.T) {
	schema := `
model Person {
	actions {
		create createPerson()
		list listPeople()
	}
}`

	expected := `
createPerson: (i?: CreatePersonInput) => {
	return this.client.rawRequest<Person>("createPerson", i);
},
listPeople: (i?: ListPeopleInput) => {
	return this.client.rawRequest<{results: Person[], pageInfo: PageInfo}>("listPeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionAllOptionalInputs(t *testing.T) {
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
listPeople: (i?: ListPeopleInput) => {
	return this.client.rawRequest<{results: Person[], pageInfo: PageInfo}>("listPeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionNotAllOptionalInputs(t *testing.T) {
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
listPeople: (i: ListPeopleInput) => {
	return this.client.rawRequest<{results: Person[], pageInfo: PageInfo}>("listPeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionRead(t *testing.T) {
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
readPeople: (i: ReadPeopleInput) => {
	return this.client.rawRequest<People>("readPeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionWrite(t *testing.T) {
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
writePeople: (i: WritePeopleInput) => {
	return this.client.rawRequest<People>("writePeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionMessageInput(t *testing.T) {
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
readPeople: (i: SearchParams) => {
	return this.client.rawRequest<People>("readPeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientActionAny(t *testing.T) {
	schema := `
model Person {
	actions {
		read readPeople(Any) returns (Any)
	}
}`

	expected := `
readPeople: (i?: any) => {
	return this.client.rawRequest<any>("readPeople", i);
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientActions(w, s, api)
	})
}

func TestClientApiDefinition(t *testing.T) {
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
queries: {
	getPerson: this.actions.getPerson,
	listPeople: this.actions.listPeople,
	readPeople: this.actions.readPeople,
},
mutations: {
	createPerson: this.actions.createPerson,
	updatePerson: this.actions.updatePerson,
	deletePerson: this.actions.deletePerson,
	authenticate: this.actions.authenticate,
	requestPasswordReset: this.actions.requestPasswordReset,
	resetPassword: this.actions.resetPassword,
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiDefinition(w, s, api)
	})
}

func TestClientApiTypes(t *testing.T) {
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
export interface EmailPasswordInput {
	email: string;
	password: string;
}
export interface AuthenticateInput {
	createIfNotExists?: boolean;
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
	id: string
	createdAt: Date
	updatedAt: Date
}
export type SortDirection = "asc" | "desc" | "ASC" | "DESC";

type PageInfo = {
	count: number;
	endCursor: string;
	hasNextPage: boolean;
	startCursor: string;
	totalCount: number;
};`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientTypes(w, s, api)
	})
}

func TestClientCoreClass(t *testing.T) {
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
export class APIClient extends Core {
	constructor(config: RequestConfig, refreshTokenStore: TokenStore = new InMemoryTokenStore()) {
		super(config, refreshTokenStore);
	}

	private actions = {
		getPerson: (i: GetPersonInput) => {
			return this.client.rawRequest<Person | null>("getPerson", i);
		},
		authenticate: (i: AuthenticateInput) => {
			return this.client.rawRequest<AuthenticateResponse>("authenticate", i).then((res) => {
				if (res.data && res.data.token) this.client.setToken(res.data.token);
				return res;
			});
		},
		requestPasswordReset: (i: RequestPasswordResetInput) => {
			return this.client.rawRequest<RequestPasswordResetResponse>("requestPasswordReset", i);
		},
		resetPassword: (i: ResetPasswordInput) => {
			return this.client.rawRequest<ResetPasswordResponse>("resetPassword", i);
		},
	};

	api = {
		queries: {
			getPerson: this.actions.getPerson,
		},
		mutations: {
			authenticate: this.actions.authenticate,
			requestPasswordReset: this.actions.requestPasswordReset,
			resetPassword: this.actions.resetPassword,
		}
	};
}`

	runWriterTest(t, schema, expected, func(s *proto.Schema, w *codegen.Writer) {
		api := proto.FindApi(s, "Api")
		writeClientApiClass(w, s, api)
	})
}

func TestGenerateClientFiles(t *testing.T) {
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

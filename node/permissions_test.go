package node_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
)

type fixture struct {
	name     string
	schema   string
	expected string
}

func TestPermissionFnBuilder(t *testing.T) {
	fixtures := []fixture{
		{
			name: "True expression",
			schema: `
				model Person {
					fields {
						name Text
					}

					functions {
						create createPost() with(name)
					}

					@permission(expression: true, actions: [list, update, get, create, delete])
				}
			`,
			expected: `
// @permission(expression: true)
const permissionFn_1 = async (records, ctx, db) => {
	// operand: true
	const operand_1 = [true];
	return operand_1.every(r => r);
};
const permissionFns = {
	createPost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "False expression",
			schema: `
				model Person {
					fields {
						name Text
					}

					functions {
						create createPost() with(name)
					}

					@permission(expression: false, actions: [list, update, get, create, delete])
				}
			`,
			expected: `
// @permission(expression: false)
const permissionFn_1 = async (records, ctx, db) => {
	// operand: false
	const operand_1 = [false];
	return operand_1.every(r => r);
};
const permissionFns = {
	createPost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "Value expression",
			schema: `
				model Post {
					fields {
						title Text
						isPublic Boolean
					}
					functions {
						update updatePost() with(title) {
							@permission(expression: post.isPublic)
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.isPublic)
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.isPublic
	let operand_1 = await db.selectFrom("post")
		.where('post.id', 'in', records.map((r) => r.id))
		.select('post.is_public as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	return operand_1.every(r => r);
};
const permissionFns = {
	updatePost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "Multi line expression",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					functions {
						update updatePost() with(title) {
							@permission(expression:
								post.identity.email == "adam@keel.xyz"
								and
								post.identity.email == "jon@keel.xyz"
							)
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.identity.email == "adam@keel.xyz" and post.identity.email == "jon@keel.xyz")
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.identity.email
	let operand_1 = await db.selectFrom("post")
		.innerJoin('identity', 'identity.id', 'post.identity_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('identity.email as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	// operand: "adam@keel.xyz"
	const operand_2 = ["adam@keel.xyz"];
	// operand: post.identity.email
	let operand_3 = await db.selectFrom("post")
		.innerJoin('identity', 'identity.id', 'post.identity_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('identity.email as v')
		.execute();

	operand_3 = operand_3.map(x => x.v);
	// operand: "jon@keel.xyz"
	const operand_4 = ["jon@keel.xyz"];
	return (operand_1.every(x => operand_2.every(y => y === x)) && operand_3.every(x => operand_4.every(y => y === x)));
};
const permissionFns = {
	updatePost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "Literal",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					functions {
						update updatePost() with(title) {
							@permission(expression: post.identity.email == "adam@keel.xyz")
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.identity.email == "adam@keel.xyz")
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.identity.email
	let operand_1 = await db.selectFrom("post")
		.innerJoin('identity', 'identity.id', 'post.identity_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('identity.email as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	// operand: "adam@keel.xyz"
	const operand_2 = ["adam@keel.xyz"];
	return operand_1.every(x => operand_2.every(y => y === x));
};
const permissionFns = {
	updatePost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "ctx.identity",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					functions {
						update updatePost() with(title) {
							@permission(expression: post.identity == ctx.identity)
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.identity == ctx.identity)
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.identity
	let operand_1 = await db.selectFrom("post")
		.innerJoin('identity', 'identity.id', 'post.identity_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('identity.id as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	// operand: ctx.identity
	const operand_2 = [ctx.identity.id];
	return operand_1.every(x => operand_2.every(y => y === x));
};
const permissionFns = {
	updatePost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "ctx.identity.id",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					functions {
						update updatePost() with(title) {
							@permission(expression: post.identity.id == ctx.identity.id)
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.identity.id == ctx.identity.id)
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.identity.id
	let operand_1 = await db.selectFrom("post")
		.innerJoin('identity', 'identity.id', 'post.identity_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('identity.id as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	// operand: ctx.identity.id
	const operand_2 = [ctx.identity.id];
	return operand_1.every(x => operand_2.every(y => y === x));
};
const permissionFns = {
	updatePost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "ctx.identity.email",
			schema: `
				model Post {
					fields {
						title Text
						identity Identity
					}
					functions {
						update updatePost() with(title) {
							@permission(expression: post.identity.email == ctx.identity.email)
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.identity.email == ctx.identity.email)
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.identity.email
	let operand_1 = await db.selectFrom("post")
		.innerJoin('identity', 'identity.id', 'post.identity_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('identity.email as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	// operand: ctx.identity.email
	const operand_2 = [ctx.identity.email];
	return operand_1.every(x => operand_2.every(y => y === x));
};
const permissionFns = {
	updatePost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
		{
			name: "Complex nested expression",
			schema: `
				model Publisher {
					fields {
						name Text
					}
				}
				model Author {
					fields {
						name Text
						publisher Publisher
					}
				}
				model Post {
					fields {
						name Text
						author Author
					}
					functions {
						create createPost() with(name, author.id) {
							@permission(expression: post.author.publisher.name == "Jim" or (post.name == "123" and post.name == "123"))
						}
					}
				}
			`,
			expected: `
// @permission(expression: post.author.publisher.name == "Jim" or (post.name == "123" and post.name == "123"))
const permissionFn_1 = async (records, ctx, db) => {
	// operand: post.author.publisher.name
	let operand_1 = await db.selectFrom("post")
		.innerJoin('author', 'author.id', 'post.author_id')
		.innerJoin('publisher', 'publisher.id', 'author.publisher_id')
		.where('post.id', 'in', records.map((r) => r.id))
		.select('publisher.name as v')
		.execute();

	operand_1 = operand_1.map(x => x.v);
	// operand: "Jim"
	const operand_2 = ["Jim"];
	// operand: post.name
	let operand_3 = await db.selectFrom("post")
		.where('post.id', 'in', records.map((r) => r.id))
		.select('post.name as v')
		.execute();

	operand_3 = operand_3.map(x => x.v);
	// operand: "123"
	const operand_4 = ["123"];
	// operand: post.name
	let operand_5 = await db.selectFrom("post")
		.where('post.id', 'in', records.map((r) => r.id))
		.select('post.name as v')
		.execute();

	operand_5 = operand_5.map(x => x.v);
	// operand: "123"
	const operand_6 = ["123"];
	return operand_1.every(x => operand_2.every(y => y === x)) || (operand_3.every(x => operand_4.every(y => y === x)) && operand_5.every(x => operand_6.every(y => y === x)));
};
const permissionFns = {
	createPost: [permissionFn_1],
}
module.exports.permissionFns = permissionFns;
			`,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			w := node.Writer{}
			builder := schema.Builder{}
			schema, err := builder.MakeFromString(fixture.schema)

			require.NoError(t, err)
			node.GeneratePermissionFunctions(&w, schema)

			assert.Equal(t, normalise(fixture.expected), normalise(w.String()))
		})
	}
}

func normalise(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\t", "    ")
}

package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
)

type testCase struct {
	name       string
	keelSchema string
	expected   string
}

// todo: defaults

var testCases = []testCase{
	{
		name: "base_keel_types",
		keelSchema: `
model Post {
	fields {
		slackId ID
		title Text
		rating Number
		published Boolean
		publishedDate Date
		publishedAt Timestamp
		passwordToView Password
		secretPassword Secret
	}
}
`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Post {
	slackId String @map("slack_id")
	title String @map("title")
	rating Int @map("rating")
	published Boolean @map("published")
	publishedDate DateTime @map("published_date")
	publishedAt DateTime @map("published_at")
	passwordToView String @map("password_to_view")
	secretPassword String @map("secret_password")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}`,
	},
	{
		name: "belongs_to_with_implicit_has_many",
		keelSchema: `
			model Author {
				fields {
					name Text
				}
			}
			model Post {
				fields {
					title Text
					author Author
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Author {
	name String @map("name")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	post_By_Author Post[] @relation("AuthorPostauthorpost_By_Author")
	@@map("author")
}
model Post {
	title String @map("title")
	author Author  @relation("AuthorPostauthorpost_By_Author", fields: [authorId], references: [id])
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")
	authorId String @map("author_id")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}`,
	},
	{
		name: "enum_field",
		keelSchema: `
			enum Status {
				Published
				Draft
			}

			model Post {
				fields {
					status Status
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Post {
	status Status 
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}
enum Status {
	Published
	Draft
}`,
	},
	{
		name: "field_optionality",
		keelSchema: `
			model Post {
				fields {
					slackId ID?
					title Text?
					rating Number?
					published Boolean?
					publishedDate Date?
					publishedAt Timestamp?
					passwordToView Password?
					secretPassword Secret?
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Post {
	slackId String? @map("slack_id")
	title String? @map("title")
	rating Int? @map("rating")
	published Boolean? @map("published")
	publishedDate DateTime? @map("published_date")
	publishedAt DateTime? @map("published_at")
	passwordToView String? @map("password_to_view")
	secretPassword String? @map("secret_password")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}`,
	},
	{
		name: "belongs_to_with_explicit_has_many",
		keelSchema: `
			model Author {
				fields {
					name Text
					posts Post[]
				}
			}
			model Post {
				fields {
					title Text
					author Author
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Author {
	name String @map("name")
	posts Post[]  @relation("AuthorPostauthorposts")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("author")
}
model Post {
	title String @map("title")
	author Author  @relation("AuthorPostauthorposts", fields: [authorId], references: [id])
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")
	authorId String @map("author_id")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}`,
	},
	{
		name: "explicit_relation_attribute_disambiguation",
		keelSchema: `
			model Author {
				fields {
					name Text
					written Post[]
					coWritten Post[]
				}
			}
			model Post {
				fields {
					title Text
					author Author @relation(written)
					coAuthor Author @relation(coWritten)
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Author {
	name String @map("name")
	written Post[]  @relation("AuthorPostauthorwritten")
	coWritten Post[]  @relation("AuthorPostcoAuthorcoWritten")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("author")
}
model Post {
	title String @map("title")
	author Author  @relation("AuthorPostauthorwritten", fields: [authorId], references: [id])
	coAuthor Author  @relation("AuthorPostcoAuthorcoWritten", fields: [coAuthorId], references: [id])
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")
	authorId String @map("author_id")
	coAuthorId String @map("co_author_id")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}`,
	},
	{
		name: "implicit_disambiguation",
		keelSchema: `
			model Author {
				fields {
					name Text
				}
			}
			model Post {
				fields {
					title Text
					author Author
					coAuthor Author
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model Author {
	name String @map("name")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	post_By_Author Post[] @relation("AuthorPostauthorpost_By_Author")
	post_By_CoAuthor Post[] @relation("AuthorPostcoAuthorpost_By_CoAuthor")
	@@map("author")
}
model Post {
	title String @map("title")
	author Author  @relation("AuthorPostauthorpost_By_Author", fields: [authorId], references: [id])
	coAuthor Author  @relation("AuthorPostcoAuthorpost_By_CoAuthor", fields: [coAuthorId], references: [id])
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")
	authorId String @map("author_id")
	coAuthorId String @map("co_author_id")

	@@map("post")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	@@map("identity")
}`,
	},
	{
		name: "has_one_unique",
		keelSchema: `
			model User {
				fields {
					identity Identity @unique
				}
			}
		`,
		expected: `
datasource db {
	provider = "postgresql"
	url      = env("KEEL_DB_CONN") 
} 

generator client {
	provider = "prisma-client-js"
    previewFeatures = ["jsonProtocol", "tracing"]
    binaryTargets = ["native", "rhel-openssl-1.0.x"]
}

model User {
	identity Identity  @relation("IdentityUseridentityuser_By_Identity", fields: [identityId], references: [id])
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")
	identityId String @map("identity_id") @unique

	@@map("user")
}
model Identity {
	email String? @map("email") @unique
	password String? @map("password")
	externalId String? @map("external_id")
	createdBy String? @map("created_by")
	id String @map("id") @id
	createdAt DateTime @map("created_at")
	updatedAt DateTime @map("updated_at")

	user_By_Identity User? @relation("IdentityUseridentityuser_By_Identity")
	@@map("identity")
}`,
	},
}

func TestPrismaSchemaGeneration(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := schema.Builder{}
			s, err := b.MakeFromString(tc.keelSchema)
			require.NoError(t, err)

			files := generatePrismaSchema(s)

			actual := files[0].Contents

			tmpDir := os.TempDir()
			err = os.WriteFile(filepath.Join(tmpDir, "schema.prisma"), []byte(actual), os.ModePerm)
			require.NoError(t, err)

			// validate the generated schema using the prisma validate cmd
			cmd := exec.Command("npx", "prisma", "validate", "--schema", "schema.prisma")
			cmd.Dir = tmpDir
			cmd.Env = os.Environ()

			// This env var isn't actually used but prisma validate will fail without it as the prisma schema references it
			cmd.Env = append(cmd.Env, "KEEL_DB_CONN=postgresql://postgres:postgres@localhost:8001/keel")

			bytes, err := cmd.CombinedOutput()

			if err != nil {
				cleanMsg := stripansi.Strip(string(bytes))
				assert.Fail(t, cleanMsg)
			}

			sanitizedExpected := normalise(tc.expected)
			sanitizedActual := normalise(actual)

			if sanitizedExpected != sanitizedActual {
				assert.Fail(t, fmt.Sprintf(`
Expected:
%s
Actual:
%s`, sanitizedExpected, sanitizedActual))
			}
		})
	}
}

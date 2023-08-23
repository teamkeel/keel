package node

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/schema"
)

func TestScaffold(t *testing.T) {
	tmpDir := t.TempDir()

	schemaString := `
	model Post {
		fields {
			title Text
		}
		actions {
			create createPost() with(title) @function
			list listPosts() @function
			update updatePost(id) with(title) @function
			get getPost(id) @function
			delete deletePost(id) @function
			write customFunctionWrite(Any) returns(Any)
			read customFunctionRead(Any) returns(Any)
		}
	}
	job MyJobWithInputs {
		inputs {
		  name Text
		}
		@permission(roles: [Developer])
	}
	job MyJobNoInputs {
		@permission(roles: [Developer])
	}

	role Developer {
		domains {
			"keel.dev"
		}
	}
`

	builder := schema.Builder{}

	schema, err := builder.MakeFromString(schemaString)

	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(schemaString), 0777)
	require.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir, schema)

	// If you enable this litter.Dump during development, it produces output that can be
	// pasted without change into the expectedFiles literal below. Obviously to do that, you have
	// to be confident by other means that the generated content is now correct.

	// litter.Dump(actualFiles)

	require.NoError(t, err)

	expectedFiles := codegen.GeneratedFiles{
		&codegen.GeneratedFile{
			Contents: "import { CreatePost, CreatePostHooks } from '@teamkeel/sdk';\n\n// To learn more about what you can do with hooks,\n// visit https://docs.keel.so/functions\nconst hooks : CreatePostHooks = {};\n\nexport default CreatePost(hooks);\n\t",
			Path:     "functions/createPost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { ListPosts, ListPostsHooks } from '@teamkeel/sdk';\n\n// To learn more about what you can do with hooks,\n// visit https://docs.keel.so/functions\nconst hooks : ListPostsHooks = {};\n\nexport default ListPosts(hooks);\n\t",
			Path:     "functions/listPosts.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { UpdatePost, UpdatePostHooks } from '@teamkeel/sdk';\n\n// To learn more about what you can do with hooks,\n// visit https://docs.keel.so/functions\nconst hooks : UpdatePostHooks = {};\n\nexport default UpdatePost(hooks);\n\t",
			Path:     "functions/updatePost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { GetPost, GetPostHooks } from '@teamkeel/sdk';\n\n// To learn more about what you can do with hooks,\n// visit https://docs.keel.so/functions\nconst hooks : GetPostHooks = {};\n\nexport default GetPost(hooks);\n\t",
			Path:     "functions/getPost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { DeletePost, DeletePostHooks } from '@teamkeel/sdk';\n\n// To learn more about what you can do with hooks,\n// visit https://docs.keel.so/functions\nconst hooks : DeletePostHooks = {};\n\nexport default DeletePost(hooks);\n\t",
			Path:     "functions/deletePost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { CustomFunctionWrite } from '@teamkeel/sdk';\nexport default CustomFunctionWrite(async (ctx, inputs) => {\n\n})",
			Path:     "functions/customFunctionWrite.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { CustomFunctionRead } from '@teamkeel/sdk';\nexport default CustomFunctionRead(async (ctx, inputs) => {\n\n})",
			Path:     "functions/customFunctionRead.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { MyJobWithInputs, models } from '@teamkeel/sdk';\nexport default MyJobWithInputs(async (ctx, inputs) => {\n\t// Build something cool\n});\n\t",
			Path:     "jobs/myJobWithInputs.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { MyJobNoInputs, models } from '@teamkeel/sdk';\nexport default MyJobNoInputs(async (ctx) => {\n\t// Build something cool\n});\n\t",
			Path:     "jobs/myJobNoInputs.ts",
		},
	}

	for _, f := range expectedFiles {
		matchingActualFile, found := lo.Find(actualFiles, func(a *codegen.GeneratedFile) bool {
			return a.Path == f.Path
		})

		if !found {
			assert.Fail(t, fmt.Sprintf("%s not found in actual files", f.Path))
		} else {
			assert.Equal(t, normalise(f.Contents), normalise(matchingActualFile.Contents))
		}
	}

	for _, f := range actualFiles {
		_, found := lo.Find(expectedFiles, func(e *codegen.GeneratedFile) bool {
			return f.Path == e.Path
		})

		if !found {
			assert.Fail(t, fmt.Sprintf("%s not found in expected files", f.Path))
		}
	}
}

func TestExistingFunction(t *testing.T) {
	tmpDir := t.TempDir()

	schemaString := `
	model Post {
		fields {
			title Text
		}
		actions {
			create existingCreatePost() with(title) @function
		}
	}
`
	builder := schema.Builder{}
	schema, err := builder.MakeFromString(schemaString)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(schemaString), 0777)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, "functions"), os.ModePerm)

	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "functions", "existingCreatePost.ts"), []byte(`import { ExistingCreatePost } from '@teamkeel/sdk';

	export default ExistingCreatePost(async (inputs, api, ctx) => {
		const post = await api.models.post.create(inputs);
		return post;
	});`), 0777)

	assert.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir, schema)

	assert.NoError(t, err)

	assert.Len(t, actualFiles, 0)
}

func TestExistingJob(t *testing.T) {
	tmpDir := t.TempDir()

	schemaString := `
	model Post {
	}
	job MyJobNoInputs {
		@permission(roles: [Developer])
	}

	role Developer {
		domains {
			"keel.dev"
		}
	}
`
	builder := schema.Builder{}
	schema, err := builder.MakeFromString(schemaString)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(schemaString), 0777)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, "jobs"), os.ModePerm)

	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "jobs", "myJobNoInputs.ts"), []byte(`unused garbage`), 0777)

	assert.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir, schema)

	assert.NoError(t, err)

	assert.Len(t, actualFiles, 0)
}

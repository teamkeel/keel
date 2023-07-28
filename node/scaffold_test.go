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
		functions {
			create createPost() with(title)
			list listPosts()
			update updatePost(id) with(title)
			get getPost(id)
			delete deletePost(id)
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
			Contents: "import { CreatePost, models } from '@teamkeel/sdk';\n\nexport default CreatePost(async (ctx, inputs) => {\n\tconst post = await models.post.create(inputs);\n\treturn post;\n});\n\t",
			Path:     "functions/createPost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { ListPosts, models } from '@teamkeel/sdk';\n\nexport default ListPosts(async (ctx, inputs) => {\n\tconst posts = await models.post.findMany(inputs);\n\treturn posts;\n});\n\t",
			Path:     "functions/listPosts.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { UpdatePost, models } from '@teamkeel/sdk';\n\nexport default UpdatePost(async (ctx, inputs) => {\n\tconst post = await models.post.update(inputs.where, inputs.values);\n\treturn post;\n});\n\t",
			Path:     "functions/updatePost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { GetPost, models } from '@teamkeel/sdk';\n\nexport default GetPost(async (ctx, inputs) => {\n\tconst post = await models.post.findOne(inputs);\n\treturn post;\n});\n\t",
			Path:     "functions/getPost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { DeletePost, models } from '@teamkeel/sdk';\n\nexport default DeletePost(async (ctx, inputs) => {\n\tconst post = await models.post.delete(inputs);\n\treturn post;\n});\n\t",
			Path:     "functions/deletePost.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { CustomFunctionWrite } from '@teamkeel/sdk';\n\nexport default CustomFunctionWrite(async (ctx, inputs) => {\n\t// Build something cool\n});\n\t",
			Path:     "functions/customFunctionWrite.ts",
		},
		&codegen.GeneratedFile{
			Contents: "import { CustomFunctionRead } from '@teamkeel/sdk';\n\nexport default CustomFunctionRead(async (ctx, inputs) => {\n\t// Build something cool\n});\n\t",
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
		functions {
			create existingCreatePost() with(title)
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

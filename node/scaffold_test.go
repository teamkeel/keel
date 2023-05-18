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
`

	builder := schema.Builder{}

	schema, err := builder.MakeFromString(schemaString)

	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(schemaString), 0777)
	require.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir, schema)

	assert.NoError(t, err)

	expectedFiles := []codegen.GeneratedFile{
		{
			Path: "deletePost.ts",
			Contents: `import { DeletePost, models } from '@teamkeel/sdk';

export default DeletePost(async (ctx, inputs) => {
	const post = await models.post.delete(inputs);
	return post;
});`,
		},
		{
			Path: "createPost.ts",
			Contents: `import { CreatePost, models } from '@teamkeel/sdk';

export default CreatePost(async (ctx, inputs) => {
	const post = await models.post.create(inputs);
	return post;
});`,
		},
		{
			Path: "updatePost.ts",
			Contents: `import { UpdatePost, models } from '@teamkeel/sdk';

export default UpdatePost(async (ctx, inputs) => {
	const post = await models.post.update(inputs.where, inputs.values);
	return post;
});`,
		},
		{
			Path: "getPost.ts",
			Contents: `import { GetPost, models } from '@teamkeel/sdk';

export default GetPost(async (ctx, inputs) => {
	const post = await models.post.findOne(inputs);
	return post;
});`,
		},
		{
			Path: "listPosts.ts",
			Contents: `import { ListPosts, models } from '@teamkeel/sdk';

export default ListPosts(async (ctx, inputs) => {
	const posts = await models.post.findMany(inputs.where!);
	return posts;
});`,
		},
		{
			Path: "customFunctionRead.ts",
			Contents: `import { CustomFunctionRead } from '@teamkeel/sdk';

export default CustomFunctionRead(async (ctx, inputs) => {
	// Build something cool
});`,
		},
		{
			Path: "customFunctionWrite.ts",
			Contents: `import { CustomFunctionWrite } from '@teamkeel/sdk';

export default CustomFunctionWrite(async (ctx, inputs) => {
	// Build something cool
});`,
		},
	}

	for _, f := range expectedFiles {
		matchingActualFile, found := lo.Find(actualFiles, func(a *codegen.GeneratedFile) bool {
			base := filepath.Base(a.Path)

			return base == f.Path
		})

		if !found {
			assert.Fail(t, fmt.Sprintf("%s not found in actual files", f.Path))
		} else {
			assert.Equal(t, normalise(f.Contents), normalise(matchingActualFile.Contents))
		}
	}

	for _, f := range actualFiles {
		base := filepath.Base(f.Path)

		_, found := lo.Find(expectedFiles, func(e codegen.GeneratedFile) bool {
			return base == e.Path
		})

		if !found {
			assert.Fail(t, fmt.Sprintf("%s not found in expected files", base))
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

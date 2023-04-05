package node

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaffold(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
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
	`), 0777)
	require.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir)

	assert.NoError(t, err)

	expectedFiles := []GeneratedFile{
		{
			Path: "deletePost.ts",
			Contents: `import { DeletePost } from '@teamkeel/sdk';

export default DeletePost(async (inputs, api, ctx) => {
	const post = await api.models.post.delete(inputs);
	return post;
});`,
		},
		{
			Path: "createPost.ts",
			Contents: `import { CreatePost } from '@teamkeel/sdk';

export default CreatePost(async (inputs, api, ctx) => {
	const post = await api.models.post.create(inputs);
	return post;
});`,
		},
		{
			Path: "updatePost.ts",
			Contents: `import { UpdatePost } from '@teamkeel/sdk';

export default UpdatePost(async (inputs, api, ctx) => {
	const post = await api.models.post.update(inputs.where, inputs.values);
	return post;
});`,
		},
		{
			Path: "getPost.ts",
			Contents: `import { GetPost } from '@teamkeel/sdk';

export default GetPost(async (inputs, api, ctx) => {
	const post = await api.models.post.findOne(inputs);
	return post;
});`,
		},
		{
			Path: "listPosts.ts",
			Contents: `import { ListPosts } from '@teamkeel/sdk';

export default ListPosts(async (inputs, api, ctx) => {
	const posts = await api.models.post.findMany(inputs.where!);
	return posts;
});`,
		},
		{
			Path: "customFunctionRead.ts",
			Contents: `import { CustomFunctionRead } from '@teamkeel/sdk';

export default CustomFunctionRead(async (inputs, api, ctx) => {
	// Build something cool
});`,
		},
		{
			Path: "customFunctionWrite.ts",
			Contents: `import { CustomFunctionWrite } from '@teamkeel/sdk';

export default CustomFunctionWrite(async (inputs, api, ctx) => {
	// Build something cool
});`,
		},
	}

	for _, f := range expectedFiles {
		matchingActualFile, found := lo.Find(actualFiles, func(a *GeneratedFile) bool {
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

		_, found := lo.Find(expectedFiles, func(e GeneratedFile) bool {
			return base == e.Path
		})

		if !found {
			assert.Fail(t, fmt.Sprintf("%s not found in expected files", base))
		}
	}
}

func TestExistingFunction(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
	model Post {
		fields {
			title Text
		}
		functions {
			create existingCreatePost() with(title)
		}
	}
`), 0777)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, "functions"), os.ModePerm)

	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "functions", "existingCreatePost.ts"), []byte(`import { ExistingCreatePost } from '@teamkeel/sdk';

	export default ExistingCreatePost(async (inputs, api, ctx) => {
		const post = await api.models.post.create(inputs);
		return post;
	});`), 0777)

	assert.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir)

	assert.NoError(t, err)

	assert.Len(t, actualFiles, 0)
}

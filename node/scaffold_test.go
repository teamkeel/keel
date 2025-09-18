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
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
)

func TestScaffold(t *testing.T) {
	t.Parallel()
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

		@on([create, update], doSomething)
		@on([update], doSomethingElse)

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
	flow MyFlow {
		inputs {
			name Text
			age Number
		}
	}
	flow MyFlowWithoutInputs {}
	task MyTask {}`

	cfg := `
auth:
  hooks: [afterAuthentication]`

	builder := schema.Builder{}

	schema, err := builder.MakeFromString(schemaString, cfg)

	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(schemaString), 0777)
	require.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir, schema, builder.Config)

	require.NoError(t, err)

	expectedFiles := codegen.GeneratedFiles{
		&codegen.GeneratedFile{
			Contents: `
import { AfterAuthentication } from '@teamkeel/sdk';

// This synchronous hook will execute after authentication has been concluded
export default AfterAuthentication(async (ctx) => {

});`,
			Path: "functions/auth/afterAuthentication.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { CreatePost, CreatePostHooks } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : CreatePostHooks = {};

export default CreatePost(hooks);`,
			Path: "functions/createPost.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { ListPosts, ListPostsHooks } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : ListPostsHooks = {};

export default ListPosts(hooks);`,
			Path: "functions/listPosts.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { UpdatePost, UpdatePostHooks } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : UpdatePostHooks = {};

export default UpdatePost(hooks);`,
			Path: "functions/updatePost.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { GetPost, GetPostHooks } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : GetPostHooks = {};

export default GetPost(hooks);`,
			Path: "functions/getPost.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { DeletePost, DeletePostHooks } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : DeletePostHooks = {};

export default DeletePost(hooks);`,
			Path: "functions/deletePost.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { CustomFunctionWrite } from '@teamkeel/sdk';

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default CustomFunctionWrite(async (ctx, inputs) => {

});`,
			Path: "functions/customFunctionWrite.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { CustomFunctionRead } from '@teamkeel/sdk';

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default CustomFunctionRead(async (ctx, inputs) => {

});`,
			Path: "functions/customFunctionRead.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { MyJobWithInputs } from '@teamkeel/sdk';

// To learn more about jobs, visit https://docs.keel.so/jobs
export default MyJobWithInputs(async (ctx, inputs) => {

});`,
			Path: "jobs/myJobWithInputs.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { MyJobNoInputs } from '@teamkeel/sdk';

// To learn more about jobs, visit https://docs.keel.so/jobs
export default MyJobNoInputs(async (ctx) => {

});`,
			Path: "jobs/myJobNoInputs.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { DoSomething } from '@teamkeel/sdk';

// To learn more about events and subscribers, visit https://docs.keel.so/events
export default DoSomething(async (ctx, event) => {

});`,
			Path: "subscribers/doSomething.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { DoSomethingElse } from '@teamkeel/sdk';

// To learn more about events and subscribers, visit https://docs.keel.so/events
export default DoSomethingElse(async (ctx, event) => {

});`,
			Path: "subscribers/doSomethingElse.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { MyFlow, FlowConfig } from '@teamkeel/sdk';

const config = {
	// See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default MyFlow(config, async (ctx, inputs) => {

});`,
			Path: "flows/myFlow.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { MyFlowWithoutInputs, FlowConfig } from '@teamkeel/sdk';

const config = {
	// See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default MyFlowWithoutInputs(config, async (ctx) => {

});`,
			Path: "flows/myFlowWithoutInputs.ts",
		},
		&codegen.GeneratedFile{
			Contents: `
import { MyTask, FlowConfig } from '@teamkeel/sdk';

const config = {
	// See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default MyTask(config, async (ctx) => {

});`,
			Path: "tasks/myTask.ts",
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
	t.Parallel()
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
	schema, err := builder.MakeFromString(schemaString, config.Empty)
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

	actualFiles, err := Scaffold(tmpDir, schema, &config.ProjectConfig{})

	assert.NoError(t, err)

	assert.Len(t, actualFiles, 0)
}

func TestExistingJob(t *testing.T) {
	t.Parallel()
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
	schema, err := builder.MakeFromString(schemaString, config.Empty)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(schemaString), 0777)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, "jobs"), os.ModePerm)

	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "jobs", "myJobNoInputs.ts"), []byte(`unused garbage`), 0777)

	assert.NoError(t, err)

	actualFiles, err := Scaffold(tmpDir, schema, &config.ProjectConfig{})

	assert.NoError(t, err)

	assert.Len(t, actualFiles, 0)
}

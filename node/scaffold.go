package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
)

const (
	FunctionsDir   = "functions"
	AuthHooksDir   = "functions/auth"
	JobsDir        = "jobs"
	FlowsDir       = "flows"
	TasksDir       = "tasks"
	SubscribersDir = "subscribers"
	RoutesDir      = "routes"
)

func Scaffold(dir string, schema *proto.Schema, cfg *config.ProjectConfig) (codegen.GeneratedFiles, error) {
	functions := schema.FilterActions(func(op *proto.Action) bool {
		return op.GetImplementation() == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
	})

	type Dir struct {
		path     string
		required bool
	}

	dirs := []Dir{
		{
			path:     FunctionsDir,
			required: len(functions) > 0,
		},
		{
			path:     AuthHooksDir,
			required: len(cfg.Auth.Hooks) > 0,
		},
		{
			path:     JobsDir,
			required: len(schema.GetJobs()) > 0,
		},
		{
			path:     FlowsDir,
			required: len(schema.GetFlows()) > 0,
		},
		{
			path:     TasksDir,
			required: len(schema.GetTasks()) > 0,
		},
		{
			path:     SubscribersDir,
			required: len(schema.GetSubscribers()) > 0,
		},
		{
			path:     RoutesDir,
			required: len(schema.GetRoutes()) > 0,
		},
	}

	for _, d := range dirs {
		if !d.required {
			continue
		}
		err := ensureDir(filepath.Join(dir, d.path))
		if err != nil {
			return nil, err
		}
	}

	generatedFiles := codegen.GeneratedFiles{}

	for _, hook := range cfg.Auth.EnabledHooks() {
		var contents string
		switch hook {
		case config.HookAfterAuthentication:
			contents = `import { AfterAuthentication } from '@teamkeel/sdk';

// This synchronous hook will execute after authentication has been concluded
export default AfterAuthentication(async (ctx) => {

});`
		case config.HookAfterIdentityCreated:
			contents = `import { AfterIdentityCreated } from '@teamkeel/sdk';

// This synchronous hook will execute after a new identity record is created during an authentication flow
export default AfterIdentityCreated(async (ctx) => {

});`
		}

		path := filepath.Join(AuthHooksDir, fmt.Sprintf("%s.ts", string(hook)))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: contents,
			})
		}
	}

	for _, fn := range functions {
		path := filepath.Join(FunctionsDir, fmt.Sprintf("%s.ts", fn.GetName()))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeFunctionWrapper(fn),
			})
		}
	}

	for _, job := range schema.GetJobs() {
		path := filepath.Join(JobsDir, fmt.Sprintf("%s.ts", casing.ToLowerCamel(job.GetName())))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeJobWrapper(job),
			})
		}
	}

	for _, flow := range schema.GetFlows() {
		subdir := FlowsDir
		if flow.GetTaskName() != nil {
			subdir = TasksDir
		}

		path := filepath.Join(subdir, fmt.Sprintf("%s.ts", casing.ToLowerCamel(flow.GetName())))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeFlowWrapper(flow),
			})
		}
	}

	for _, subscriber := range schema.GetSubscribers() {
		path := filepath.Join(SubscribersDir, fmt.Sprintf("%s.ts", subscriber.GetName()))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeSubscriberWrapper(subscriber),
			})
		}
	}

	for _, route := range schema.GetRoutes() {
		path := filepath.Join(RoutesDir, fmt.Sprintf("%s.ts", route.GetHandler()))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path: path,
				Contents: `import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  return {
    body: '',
  };
};

export default handler;
`,
			})
		}
	}

	return generatedFiles, nil
}

func ensureDir(dirName string) error {
	err := os.Mkdir(dirName, 0700)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func writeFunctionWrapper(function *proto.Action) string {
	functionName := casing.ToCamel(function.GetName())

	if function.IsArbitraryFunction() {
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default %s(async (ctx, inputs) => {

});`, functionName, functionName)
	}

	hookType := fmt.Sprintf("%sHooks", casing.ToCamel(function.GetName()))

	return fmt.Sprintf(`import { %s, %s } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : %s = {};

export default %s(hooks);
	`, functionName, hookType, hookType, functionName)
}

func writeJobWrapper(job *proto.Job) string {
	// The "inputs" argument for the function signature is only
	// wanted there are some.
	switch {
	case job.GetInputMessageName() == "":
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about jobs, visit https://docs.keel.so/jobs
export default %s(async (ctx) => {

});`, job.GetName(), job.GetName())

	default:
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about jobs, visit https://docs.keel.so/jobs
export default %s(async (ctx, inputs) => {

});`, job.GetName(), job.GetName())
	}
}

func writeFlowWrapper(flow *proto.Flow) string {
	// The "inputs" argument for the function signature is only
	// wanted when there are some.
	switch {
	case flow.GetInputMessageName() == "":
		return fmt.Sprintf(`import { %s, FlowConfig } from '@teamkeel/sdk';

const config = {
	// See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default %s(config, async (ctx) => {

});`, flow.GetName(), flow.GetName())

	default:
		return fmt.Sprintf(`import { %s, FlowConfig } from '@teamkeel/sdk';

const config = {
	// See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default %s(config, async (ctx, inputs) => {

});`, flow.GetName(), flow.GetName())
	}
}

func writeSubscriberWrapper(subscriber *proto.Subscriber) string {
	return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about events and subscribers, visit https://docs.keel.so/events
export default %s(async (ctx, event) => {

});`, strcase.ToCamel(subscriber.GetName()), strcase.ToCamel(subscriber.GetName()))
}

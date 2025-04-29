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
	SubscribersDir = "subscribers"
	RoutesDir      = "routes"
)

func Scaffold(dir string, schema *proto.Schema, cfg *config.ProjectConfig) (codegen.GeneratedFiles, error) {
	functions := schema.FilterActions(func(op *proto.Action) bool {
		return op.Implementation == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
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
			required: len(schema.Jobs) > 0,
		},
		{
			path:     FlowsDir,
			required: len(schema.Flows) > 0,
		},
		{
			path:     SubscribersDir,
			required: len(schema.Subscribers) > 0,
		},
		{
			path:     RoutesDir,
			required: len(schema.Routes) > 0,
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
		path := filepath.Join(FunctionsDir, fmt.Sprintf("%s.ts", fn.Name))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeFunctionWrapper(fn),
			})
		}
	}

	for _, job := range schema.Jobs {
		path := filepath.Join(JobsDir, fmt.Sprintf("%s.ts", casing.ToLowerCamel(job.Name)))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeJobWrapper(job),
			})
		}
	}

	for _, flow := range schema.Flows {
		path := filepath.Join(FlowsDir, fmt.Sprintf("%s.ts", casing.ToLowerCamel(flow.Name)))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeFlowWrapper(flow),
			})
		}
	}

	for _, subscriber := range schema.Subscribers {
		path := filepath.Join(SubscribersDir, fmt.Sprintf("%s.ts", subscriber.Name))
		_, err := os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeSubscriberWrapper(subscriber),
			})
		}
	}

	for _, route := range schema.Routes {
		path := filepath.Join(RoutesDir, fmt.Sprintf("%s.ts", route.Handler))
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
	functionName := casing.ToCamel(function.Name)

	if function.IsArbitraryFunction() {
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default %s(async (ctx, inputs) => {

});`, functionName, functionName)
	}

	hookType := fmt.Sprintf("%sHooks", casing.ToCamel(function.Name))

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
	case job.InputMessageName == "":
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about jobs, visit https://docs.keel.so/jobs
export default %s(async (ctx) => {

});`, job.Name, job.Name)

	default:
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about jobs, visit https://docs.keel.so/jobs
export default %s(async (ctx, inputs) => {

});`, job.Name, job.Name)
	}
}

func writeFlowWrapper(flow *proto.Flow) string {
	// The "inputs" argument for the function signature is only
	// wanted when there are some.
	switch {
	case flow.InputMessageName == "":
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about flows, visit https://docs.keel.so/flows
export default %s({ title: "%s" }, async (ctx) => {

});`, flow.Name, flow.Name, flow.Name)

	default:
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about flows, visit https://docs.keel.so/flows
export default %s({ title: "%s" }, async (ctx, inputs) => {

});`, flow.Name, flow.Name, flow.Name)
	}
}

func writeSubscriberWrapper(subscriber *proto.Subscriber) string {
	return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

// To learn more about events and subscribers, visit https://docs.keel.so/events
export default %s(async (ctx, event) => {

});`, strcase.ToCamel(subscriber.Name), strcase.ToCamel(subscriber.Name))
}

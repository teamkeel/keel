package node

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/proto"
)

const (
	FUNCTIONS_DIR = "functions"
	JOBS_DIR      = "jobs"
)

func Scaffold(dir string, schema *proto.Schema) (codegen.GeneratedFiles, error) {
	files, err := Generate(context.TODO(), schema)

	if err != nil {
		return nil, err
	}

	err = files.Write(dir)

	if err != nil {
		return nil, err
	}

	functionsDir := filepath.Join(dir, FUNCTIONS_DIR)
	if err := ensureDir(functionsDir); err != nil {
		return nil, err
	}
	jobsDir := filepath.Join(dir, JOBS_DIR)
	if err := ensureDir(jobsDir); err != nil {
		return nil, err
	}

	generatedFiles := codegen.GeneratedFiles{}

	functions := proto.FilterActions(schema, func(op *proto.Action) bool {
		return op.Implementation == proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
	})

	for _, fn := range functions {
		path := filepath.Join(FUNCTIONS_DIR, fmt.Sprintf("%s.ts", fn.Name))

		_, err = os.Stat(filepath.Join(dir, path))

		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeFunctionWrapper(fn),
			})
		}

	}

	for _, job := range schema.Jobs {
		path := filepath.Join(JOBS_DIR, fmt.Sprintf("%s.ts", casing.ToLowerCamel(job.Name)))
		_, err = os.Stat(filepath.Join(dir, path))
		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &codegen.GeneratedFile{
				Path:     path,
				Contents: writeJobWrapper(job),
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

	if proto.ActionIsArbitraryFunction(function) {
		return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';
export default %s(async (ctx, inputs) => {

})`, functionName, functionName)
	}

	hookType := fmt.Sprintf("%sHooks", casing.ToCamel(function.Name))

	return fmt.Sprintf(`import { %s, %s } from '@teamkeel/sdk';

// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks : %s = {};

export default %s(hooks);
	`, functionName, hookType, hookType, functionName)
}

func writeJobWrapper(job *proto.Job) string {
	extraImports := ", models"
	suggestedImplementation := "// Build something cool"

	// The "inputs" argument for the function signature is only
	// wanted there are some.
	switch {
	case job.InputMessageName == "":
		return fmt.Sprintf(`import { %s%s } from '@teamkeel/sdk';
export default %s(async (ctx) => {
	%s
});
	`, job.Name, extraImports, job.Name, suggestedImplementation)

	default:
		return fmt.Sprintf(`import { %s%s } from '@teamkeel/sdk';
export default %s(async (ctx, inputs) => {
	%s
});
	`, job.Name, extraImports, job.Name, suggestedImplementation)
	}
}

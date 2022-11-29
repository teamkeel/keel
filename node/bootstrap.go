package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	codegenerator "github.com/teamkeel/keel/node/codegen"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

func Bootstrap(dir string) error {
	_, err := os.Stat(filepath.Join(dir, "package.json"))

	if err == nil {
		return nil
	}

	// Make a package JSON for the temp dir that references my-package
	// as a local package
	err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(fmt.Sprintf(
		`
	{
		"name": "%s",
		"dependencies": {
			"@teamkeel/testing": "*",
			"@teamkeel/sdk":     "*",
			"@teamkeel/runtime": "*",
			"ts-node":           "*",
			// https://typestrong.org/ts-node/docs/swc/
			"@swc/core":           "*",
			"regenerator-runtime": "*",
		}
	}`, filepath.Base(dir),
	)), 0644)

	if err != nil {
		return err
	}

	npmInstall := exec.Command("npm", "install", "--progress=false", "--no-audit")
	npmInstall.Dir = dir

	o, err := npmInstall.CombinedOutput()

	if err != nil {
		fmt.Print(string(o))
		return err
	}

	return nil
}

// Generates @teamkeel/sdk and @teamkeel/testing
func GeneratePackages(dir string) error {
	builder := schema.Builder{}

	schema, err := builder.MakeFromDirectory(dir)

	if err != nil {
		return err
	}

	// Dont do any code generation if there are no functions in the schema
	// or any Keel tests defined
	if !hasFunctions(schema) && !hasTests(dir) {
		return nil
	}

	cg := codegenerator.NewGenerator(schema, dir)

	_, err = cg.GenerateSDK()

	if err != nil {
		return err
	}

	_, err = cg.GenerateTesting()

	if err != nil {
		return err
	}

	return nil
}

func GenerateDevelopmentServer(dir string) error {
	// 1. make a single js file inside .build directory
	// that imports custom function handler code from node_modules/@teamkeel/functions-runtime
	// 2. bootstrap code to start a node server for the custom function runtime

	builder := schema.Builder{}

	schema, err := builder.MakeFromDirectory(dir)

	if err != nil {
		return err
	}

	// Dont do any code generation if there are no functions in the schema
	// or any Keel tests defined
	if !hasFunctions(schema) && !hasTests(dir) {
		return nil
	}

	cg := codegenerator.NewGenerator(schema, dir)

	_, err = cg.GenerateDevelopmentHandler()

	if err != nil {
		return err
	}

	return nil
}

type RuntimeServer interface {
	Kill() error
}

type DevelopmentServer struct {
	proc *os.Process
}

func (ds *DevelopmentServer) Kill() error {
	return ds.proc.Kill()
}

// RunDevelopmentServer will start a new node runtime server
// serving custom function requests
func RunDevelopmentServer(dir string, envVars map[string]any) (RuntimeServer, error) {
	// 1. run dev server with ts-node.
	handlerPath := filepath.Join(dir, codegenerator.BUILD_DIR_NAME, "index.js")

	cmd := exec.Command("./node_modules/.bin/ts-node", handlerPath)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return &DevelopmentServer{
		proc: cmd.Process,
	}, nil
}

func hasFunctions(sch *proto.Schema) bool {
	var ops []*proto.Operation

	for _, model := range sch.Models {
		ops = append(ops, model.Operations...)
	}

	return lo.SomeBy(ops, func(o *proto.Operation) bool {
		return o.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
	})
}

func hasTests(dir string) bool {
	fs := os.DirFS(dir)

	// the only potential error returned from glob here is bad pattern,
	// which we know not to be true
	testFiles, _ := doublestar.Glob(fs, "**/*.test.ts")

	// there could be other *.test.ts files unrelated to the Keel testing framework,
	// so for each test, we do a naive check that the file contents includes a match
	// for the string "@teamkeel/testing"
	return lo.SomeBy(testFiles, func(path string) bool {
		b, err := os.ReadFile(path)

		if err != nil {
			return false
		}

		// todo: improve this check as its pretty naive
		return strings.Contains(string(b), "@teamkeel/testing")
	})
}

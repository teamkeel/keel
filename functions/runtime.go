package functions

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

type Runtime struct {
	Schema     *proto.Schema
	WorkingDir string
	generator  CodeGenerator
}

var SCHEMA_FILE = "schema.keel"
var FUNCTIONS_DIRECTORY = "functions"

func NewRuntime(workingDir string) (*Runtime, error) {
	schema, err := buildSchema(workingDir)

	if err != nil {
		return nil, err
	}

	return &Runtime{
		WorkingDir: workingDir,
		Schema:     schema,
		generator:  *NewCodeGenerator(schema),
	}, nil
}

// Generate generates the typescript codefiles based on the schema
func (r *Runtime) GenerateClient() error {
	src := r.generator.GenerateClientCode()

	_, err := r.makeModule(filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "index.ts"), src)

	return err
}

func (r *Runtime) GenerateHandler() error {
	src := r.generator.GenerateEntryPoint()

	_, err := r.makeModule(filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "handler.ts"), src)

	return err
}

func (r *Runtime) ReconcilePackageJson() error {
	packageJson, err := NewPackageJson(filepath.Join(r.WorkingDir, "package.json"))

	if err != nil {
		return err
	}

	err = packageJson.Bootstrap()

	if err != nil {
		return err
	}

	return nil
}

// Bundle transpiles all TypeScript in a working directory using
// esbuild, and outputs the JavaScript equivalent to the OutDir
func (r *Runtime) Bundle(write bool) (api.BuildResult, []error) {
	// Run esbuild on the generated entrypoint code
	// The entrypoint references the users custom functions
	// so these will be bundled in addition to any generated code
	buildResult := api.Build(api.BuildOptions{
		EntryPoints: []string{
			path.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "index.ts"),
			path.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "handler.ts"),
		},
		Bundle:         true,
		Outdir:         filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "dist"),
		Write:          write,
		AllowOverwrite: true,
		Platform:       api.PlatformNode,
		LogLevel:       api.LogLevelError,
	})

	if len(buildResult.Errors) > 0 {
		return buildResult, r.buildResultErrors(buildResult.Errors)
	}

	return buildResult, nil
}

func (r *Runtime) Scaffold() error {
	generator := NewCodeGenerator(r.Schema)

	for _, model := range r.Schema.Models {
		for _, op := range model.Operations {
			if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
				path := filepath.Join(r.WorkingDir, FUNCTIONS_DIRECTORY, fmt.Sprintf("%s.ts", op.Name))

				if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
					src := generator.GenerateFunction(model.Name)

					if err != nil {
						return err
					}

					err = os.WriteFile(path, []byte(src), 0644)

					fmt.Printf("Scaffolded function %s (%s)", op.Name, path)

					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (r *Runtime) RunServer(port int, onBoot func(process *os.Process)) error {
	serverDistPath := filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "dist", "handler.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		fmt.Print(err)
		return err
	}

	cmd := exec.Command("node", filepath.Join("node_modules", "@teamkeel", "client", "dist", "handler.js"))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
	cmd.Dir = r.WorkingDir
	err := cmd.Start()

	if err != nil {
		return err
	}

	onBoot(cmd.Process)

	return nil
}

func (r *Runtime) buildResultErrors(errs []api.Message) (e []error) {
	for _, err := range errs {
		e = append(e, errors.New(err.Text))
	}
	return e
}

func (r *Runtime) makeModule(path string, code string) (string, error) {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModePerm)

		if err != nil {
			return "", err
		}
	}

	err := ioutil.WriteFile(path, []byte(code), 0644)

	if err != nil {
		return "", err
	}

	return path, nil
}

func buildSchema(workingDir string) (*proto.Schema, error) {
	builder := schema.Builder{}

	proto, err := builder.MakeFromDirectory(workingDir)

	return proto, err
}

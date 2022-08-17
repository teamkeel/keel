package functions

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/proto"
)

type Runtime struct {
	Schema     *proto.Schema
	WorkingDir string
	generator  CodeGenerator
}

type ScaffoldResult struct {
	FunctionsCount int

	CreatedFunctions []string
}

type FunctionImplementation struct {
	Op    *proto.Operation
	Model *proto.Model
}

var SCHEMA_FILE = "schema.keel"
var FUNCTIONS_DIRECTORY = "functions"

func NewRuntime(schema *proto.Schema, workDir string) (*Runtime, error) {
	return &Runtime{
		WorkingDir: workDir,
		Schema:     schema,
		generator:  *NewCodeGenerator(schema),
	}, nil
}

// Generate generates the typescript codefiles based on the schema
func (r *Runtime) GenerateClient() error {
	src := r.generator.GenerateClientCode()

	_, err := r.makeModule(filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "src", "index.ts"), src)

	if err != nil {
		return err
	}

	err = r.GenerateClientTypings()

	if err != nil {
		return err
	}

	err = r.GenerateClientPackageJson()

	if err != nil {
		return err
	}

	err = r.GenerateHandler()

	if err != nil {
		return err
	}

	_, errs := r.Bundle(true)

	if len(errs) > 0 {
		var errors []string

		for _, err := range errs {
			errors = append(errors, err.Error())
		}

		return fmt.Errorf(strings.Join(errors, ","))
	}

	return nil
}

func (r *Runtime) GenerateClientTypings() error {
	src := r.generator.GenerateClientTypings()

	_, err := r.makeModule(filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "dist", "index.d.ts"), src)

	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) GenerateHandler() error {
	src := r.generator.GenerateEntryPoint()

	_, err := r.makeModule(filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "src", "handler.js"), src)

	return err
}

//go:embed client-package.json
var clientPackageJson string

func (r *Runtime) GenerateClientPackageJson() error {
	packageJsonPath := filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "package.json")
	clientDir := filepath.Dir(packageJsonPath)

	if _, err := os.Stat(clientDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(clientDir, os.ModePerm)

		if err != nil {
			return err
		}
	}

	f, err := os.Create(packageJsonPath)

	if err != nil {
		return err
	}

	_, err = f.WriteString(clientPackageJson)

	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) ReconcilePackageJson() error {
	packageJson, err := nodedeps.NewPackageJson(filepath.Join(r.WorkingDir, "package.json"), nodedeps.PackageManagerPnpm)

	if err != nil {
		return err
	}

	err = packageJson.Bootstrap()

	if err != nil {
		return err
	}

	return nil
}

// Bundle transpiles all generated TypeScript files in a working directory using
// esbuild, and outputs the JavaScript equivalent to the OutDir
func (r *Runtime) Bundle(write bool) (api.BuildResult, []error) {
	// Run esbuild on the generated entrypoint code
	// The entrypoint references the users custom functions
	// so these will be bundled in addition to any generated code
	buildResult := api.Build(api.BuildOptions{
		EntryPoints: []string{
			path.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "src", "index.ts"),
			path.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "src", "handler.js"),
		},
		Bundle:   true,
		Outdir:   filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "dist"),
		Write:    write,
		Platform: api.PlatformNode,
		LogLevel: api.LogLevelError,
		External: []string{
			"@teamkeel/sdk",
		},
	})

	if len(buildResult.Errors) > 0 {
		return buildResult, r.buildResultErrors(buildResult.Errors)
	}

	return buildResult, nil
}

func (r *Runtime) Scaffold() (s *ScaffoldResult, e error) {
	generator := NewCodeGenerator(r.Schema)

	functionsDir := filepath.Join(r.WorkingDir, FUNCTIONS_DIRECTORY)

	if _, err := os.Stat(functionsDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(functionsDir, os.ModePerm)

		if err != nil {
			return nil, err
		}
	}

	funcs := lo.FlatMap(r.Schema.Models, func(m *proto.Model, _ int) (ops []*FunctionImplementation) {
		for _, op := range m.Operations {
			if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
				ops = append(ops, &FunctionImplementation{
					Model: m,
					Op:    op,
				})
			}
		}

		return ops
	})

	if len(funcs) == 0 {
		return &ScaffoldResult{
			FunctionsCount: 0,
		}, nil
	}

	sr := &ScaffoldResult{
		FunctionsCount: len(funcs),
	}
	for _, f := range funcs {
		path := filepath.Join(r.WorkingDir, FUNCTIONS_DIRECTORY, fmt.Sprintf("%s.ts", f.Op.Name))

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {

			src := generator.GenerateFunction(f.Op.Name)
			err = os.WriteFile(path, []byte(src), 0644)

			if err != nil {
				return sr, err
			}

			sr.CreatedFunctions = append(sr.CreatedFunctions, path)
		}
	}

	return sr, nil
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

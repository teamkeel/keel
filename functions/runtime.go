package functions

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	_ "embed"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
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

	_, err := r.makeModule(filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "src", "index.ts"), src)

	if err != nil {
		return err
	}

	err = r.GenerateClientTypings()

	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) GenerateClientTypings() error {
	cmd := exec.Command(
		"tsc",
		"index.ts",
		"--declaration",
		"--emitDeclarationOnly",
		"--declarationDir",
		"../dist",
	)

	cmd.Dir = filepath.Join(r.WorkingDir, "node_modules", "@teamkeel", "client", "src")

	o, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Print(string(o))
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
		LogLevel: api.LogLevelSilent,
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
		}, errors.New("no functions to generate")
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

package runtime

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/teamkeel/keel/functions/codegen"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

type Runtime struct {
	OutDir     string
	Schema     *proto.Schema
	WorkingDir string
	generator  codegen.CodeGenerator
}

var DEV_DIRECTORY = ".keel_generated"
var FUNCTIONS_DIRECTORY = "functions"

func NewRuntime(schema *proto.Schema, workingDir string, outDir string) (*Runtime, error) {
	return &Runtime{
		WorkingDir: workingDir,
		OutDir:     outDir,
		Schema:     schema,
		generator:  *codegen.NewCodeGenerator(schema),
	}, nil
}

// Generate generates the typescript codefiles based on the schema
func (r *Runtime) Generate() (filePath string, err error) {
	src := r.generator.GenerateClientCode()

	filePath, err = r.makeModule(path.Join(r.OutDir, "index.ts"), src)

	return filePath, err
}

func (r *Runtime) InstallDeps() error {
	// NPM install all dependencies from the working directories'
	// package.json file so we can bundle the code
	npmInstall := exec.Command("npm", "install")

	// The location where we want to install is the working directory path of the target app
	npmInstall.Dir = r.WorkingDir

	// .Run() waits for the npm install command to complete
	err := npmInstall.Run()

	if err != nil {
		return err
	}

	return nil
}

// Bundle transpiles all TypeScript in a working directory using
// esbuild, and outputs the JavaScript equivalent to the OutDir
func (r *Runtime) Bundle(write bool) (errs []error) {
	err := r.InstallDeps()

	if err != nil {
		return []error{err}
	}

	// Run esbuild on the generated entrypoint code
	// The entrypoint references the users custom functions
	// so these will be bundled in addition to any generated code
	buildResult := api.Build(api.BuildOptions{
		EntryPoints: []string{
			path.Join(r.WorkingDir, DEV_DIRECTORY, "index.ts"),
		},
		Bundle:   true,
		Outdir:   filepath.Join(r.OutDir, "dist"),
		Write:    write,
		Platform: api.PlatformNode,
		LogLevel: api.LogLevelError,
	})

	if len(buildResult.Errors) > 0 {
		return r.buildResultErrors(buildResult.Errors)
	}

	return nil
}

func (r *Runtime) Scaffold() error {
	generator := codegen.NewCodeGenerator(r.Schema)

	functionsDir := filepath.Join(r.WorkingDir, FUNCTIONS_DIRECTORY)

	if _, err := os.Stat(functionsDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(functionsDir, 0700)

		if err != nil {
			return err
		}
	}

	for _, model := range r.Schema.Models {
		for _, op := range model.Operations {
			if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
				path, err := filepath.Abs(filepath.Join(r.WorkingDir, FUNCTIONS_DIRECTORY, fmt.Sprintf("%s.ts", op.Name)))

				if err != nil {
					return err
				}

				if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
					src := generator.GenerateFunction(model.Name)

					f, err := os.Create(path)

					if err != nil {
						return err
					}

					_, err = f.WriteString(src)

					if err != nil {
						return err
					}

				}
			}
		}
	}

	return nil
}

func (r *Runtime) RunServer(port int, onBoot func(process *os.Process)) (*os.Process, error) {
	serverDistPath := filepath.Join(r.OutDir, "dist", "index.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		panic(".keel_generated/dist/index.js has not been generated")
	}

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {

		fmt.Print(err)
	}

	cmd := exec.Command("node", filepath.Join(DEV_DIRECTORY, "dist", "index.js"))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
	cmd.Dir = r.WorkingDir
	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	onBoot(cmd.Process)

	return cmd.Process, nil
}

func (r *Runtime) buildResultErrors(errs []api.Message) (e []error) {
	for _, err := range errs {
		e = append(e, errors.New(err.Text))
	}
	return e
}

func (r *Runtime) makeModule(path string, code string) (string, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(filepath.Dir(path), os.ModePerm)

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

func buildSchema(schemaPath string) (*proto.Schema, error) {
	if _, err := os.Stat(schemaPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	schemaBytes, err := ioutil.ReadFile(schemaPath)

	if err != nil {
		return nil, err
	}

	builder := schema.Builder{}

	proto, err := builder.MakeFromString(string(schemaBytes))

	return proto, err
}

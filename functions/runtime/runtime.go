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

var SCHEMA_FILE = "schema.keel"
var DEV_DIRECTORY = ".keel"

func NewRuntime(workingDir string, outDir string) (*Runtime, error) {
	schema, err := buildSchema(workingDir)

	if err != nil {
		return nil, err
	}

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

// Bundle transpiles all TypeScript in a working directory using
// esbuild, and outputs the JavaScript equivalent to the OutDir
func (r *Runtime) Bundle(write bool) (errs []error) {
	// NPM install all dependencies from the working directories'
	// package.json file so we can bundle the code
	npmInstall := exec.Command("npm", "install")

	// The location where we want to install is the working directory path of the target app
	npmInstall.Dir = r.WorkingDir

	// .Run() waits for the npm install command to complete
	err := npmInstall.Run()

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
		LogLevel: api.LogLevelInfo,
	})

	if len(buildResult.Errors) > 0 {
		return r.buildResultErrors(buildResult.Errors)
	}

	return nil
}

func (r *Runtime) RunServer(port int, onBoot func(process *os.Process)) error {
	serverDistPath := filepath.Join(r.OutDir, "dist", "index.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		panic(".keel/dist/index.js has not been generated")
	}

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {

		fmt.Print(err)
	}

	cmd := exec.Command("node", filepath.Join(DEV_DIRECTORY, "dist", "index.js"))
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

func buildSchema(workingDir string) (*proto.Schema, error) {
	if _, err := os.Stat(filepath.Join(workingDir, SCHEMA_FILE)); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	schemaBytes, err := ioutil.ReadFile(filepath.Join(workingDir, SCHEMA_FILE))

	if err != nil {
		return nil, err
	}

	builder := schema.Builder{}

	proto, err := builder.MakeFromString(string(schemaBytes))

	return proto, err
}

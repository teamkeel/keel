package runtime

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

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

func (r *Runtime) Generate() (filePath string, err error) {
	src := r.generator.GenerateClientCode()

	filePath, err = r.makeModule(path.Join(r.OutDir, "index.ts"), src)

	return filePath, err
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

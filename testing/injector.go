package testing

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/proto"
)

type Injector struct {
	workDir string
	schema  *proto.Schema
}

func NewInjector(workDir string, schema *proto.Schema) *Injector {
	return &Injector{
		workDir: workDir,
		schema:  schema,
	}
}

func (inj *Injector) Inject() error {
	generator := codegen.NewGenerator(inj.schema)

	src := generator.GenerateTesting()

	_, err := inj.makeModule(filepath.Join(inj.workDir, "node_modules", "@teamkeel", "testing", "src", "generated.ts"), src)

	if err != nil {
		return err
	}

	return nil
}

func (r *Injector) makeModule(path string, code string) (string, error) {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dir, os.ModePerm)

		if err != nil {
			return "", err
		}
	}

	err := os.WriteFile(path, []byte(code), 0644)

	if err != nil {
		return "", err
	}

	return path, nil
}

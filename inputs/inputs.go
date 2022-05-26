package inputs

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/teamkeel/keel/model"
)

// Assemble constructs an Inputs instance by selecting relevant
// files from the given directory.
//
// So far it only looks for *.keel files and puts those in the SchemaFiles field.
func Assemble(dirName string) (*model.Inputs, error) {
	inputs := &model.Inputs{
		Directory:   dirName,
		SchemaFiles: []model.SchemaFile{},
	}
	globPattern := filepath.Join(dirName, "*.keel")
	schemaFileNames, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("filepath.Glob errored with: %v", err)
	}
	if len(schemaFileNames) < 1 {
		return nil, fmt.Errorf("No files matching: <%s> were found:", globPattern)
	}
	for _, fName := range schemaFileNames {
		fileBytes, err := ioutil.ReadFile(fName)
		if err != nil {
			return nil, fmt.Errorf("Error reading file: %v", err)
		}
		inputs.SchemaFiles = append(inputs.SchemaFiles, model.SchemaFile{
			FileName: fName,
			Contents: string(fileBytes),
		})
	}
	return inputs, nil
}

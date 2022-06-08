package reader

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
)

// Inputs models a set of files (Schema files and other files) that have been found in a
// given directory.
type Inputs struct {
	Directory   string
	SchemaFiles []SchemaFile
}

type SchemaFile struct {
	FileName string
	Contents string
}

// FromDir constructs an Inputs instance by selecting relevant
// files from the given directory.
//
// So far it only looks for *.keel files and puts those in the SchemaFiles field.
func FromDir(dirName string) (*Inputs, error) {
	inputs := &Inputs{
		Directory:   dirName,
		SchemaFiles: []SchemaFile{},
	}
	globPattern := filepath.Join(dirName, "*.keel")
	schemaFileNames, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("filepath.Glob errored with: %v", err)
	}
	if len(schemaFileNames) < 1 {
		return nil, fmt.Errorf("no files matching: <%s> were found", globPattern)
	}
	for _, fName := range schemaFileNames {
		fileBytes, err := ioutil.ReadFile(fName)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %v", err)
		}
		inputs.SchemaFiles = append(inputs.SchemaFiles, SchemaFile{
			FileName: fName,
			Contents: string(fileBytes),
		})
	}
	return inputs, nil
}

func FromFile(filename string) (*Inputs, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	schemaFile := SchemaFile{
		FileName: filename,
		Contents: string(fileBytes),
	}
	return &Inputs{
		Directory:   path.Dir(filename),
		SchemaFiles: []SchemaFile{schemaFile},
	}, nil
}

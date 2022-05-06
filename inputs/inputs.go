package inputs

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Inputs models a set of files (like Schema files etc.) that have been found in a directory.
type Inputs struct {
	Directory string
	SchemaFiles []InputFile
	OtherTypesOfFiles int // Placeholder for illustration
}

type InputFile struct{
	FileName string
	Contents string
}

// Assemble constructs an Inputs instance from the files in the given
// directory that are relevant.
//
// So far it only looks for *.keel files and puts those in the SchemaFiles field.
func Assemble(dirName string) (*Inputs, error) {
	inputs := &Inputs{
		Directory: dirName,
		SchemaFiles: []InputFile{},
	}
	schemaFileNames, err := filepath.Glob(filepath.Join(dirName, "*.keel"))
	if err != nil {
		return nil, fmt.Errorf("filepath.Glob errored with: %v", err)
	}
	for _, fName := range schemaFileNames {
		fileBytes, err := ioutil.ReadFile(fName)
		if err != nil {
			return nil, fmt.Errorf("Error reading file: %v", err)
		}
		inputs.SchemaFiles = append(inputs.SchemaFiles, InputFile{
			FileName: fName,
			Contents: string(fileBytes),
		})
	}
	return inputs, nil
}
package reader

import (
	"os"
	"path"
	"path/filepath"
)

// Inputs models a set of files (Schema files and other files) that have been found in a
// given directory.
type Inputs struct {
	Directory   string
	SchemaFiles []*SchemaFile
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
		SchemaFiles: []*SchemaFile{},
	}

	// Search current directory and schemas subfolder if exists
	patterns := []string{
		filepath.Join(dirName, "*.keel"),
		filepath.Join(dirName, "schemas", "*.keel"),
	}

	for _, pattern := range patterns {
		schemaFileNames, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, fName := range schemaFileNames {
			fileBytes, err := os.ReadFile(fName)
			if err != nil {
				return nil, err
			}
			inputs.SchemaFiles = append(inputs.SchemaFiles, &SchemaFile{
				FileName: fName,
				Contents: string(fileBytes),
			})
		}
	}

	return inputs, nil
}

func FromFile(filename string) (*Inputs, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	schemaFile := &SchemaFile{
		FileName: filename,
		Contents: string(fileBytes),
	}
	return &Inputs{
		Directory:   path.Dir(filename),
		SchemaFiles: []*SchemaFile{schemaFile},
	}, nil
}

package codegen

import (
	"fmt"
	"os"
	"path/filepath"
)

type GeneratedFile struct {
	Contents string
	Path     string
}

// GeneratedFiles represents a collection of files due to be generated as part of code generation
// You can append any number of files to an instance of GeneratedFiles like so:
// files := GeneratedFiles{}
// files = append(files, &GeneratedFile{...})
// And then once you are ready to write these files to disk, you can call write:
// files.Write(dir).
type GeneratedFiles []*GeneratedFile

func (files GeneratedFiles) Write(dir string) error {
	for _, f := range files {
		path := filepath.Join(dir, f.Path)
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		err = os.WriteFile(path, []byte(f.Contents), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}

	return nil
}

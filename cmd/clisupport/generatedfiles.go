package clisupport

import (
	"fmt"
	"os"
	"path/filepath"
)

type GeneratedFile struct {
	Contents string
	Path     string
}

type GeneratedFiles []*GeneratedFile

func (files GeneratedFiles) Write(dir string) error {
	for _, f := range files {
		path := filepath.Join(dir, f.Path)
		err := os.MkdirAll(filepath.Dir(path), 0777)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		err = os.WriteFile(path, []byte(f.Contents), 0777)
		if err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}
	return nil
}

package main

// A cmd that finds all the *.keel files in the current working dir, and
// reports on those found whose contents match some condition.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cwd, err := os.Getwd()
	panicOnErr(err)

	err = filepath.WalkDir(cwd, func(path string, d os.DirEntry, unused error) error {
		if d.IsDir() {
			return nil
		}
		fName := d.Name()
		if !strings.HasSuffix(strings.ToLower(fName), ".keel") {
			return nil
		}
		b, err := os.ReadFile(path)
		panicOnErr(err)
		schemaAsString := string(b)
		lines := strings.Split(schemaAsString, "\n")

		applyTests(path, lines, schemaAsString)
		return nil
	})
	panicOnErr(err)
}

func panicOnErr(err error) {
	if err == nil {
		return
	}
	panic(fmt.Sprintf("stopping on err: %s", err))
}

// You can change this function to suit your purpose.
func applyTests(filePath string, lines []string, schemaAsString string) {

	// Example does the schema have any references to permissions?
	if strings.Contains(schemaAsString, "@permission") {
		fmt.Printf("This file uses permission attribute: %s\n", filePath)
	}
}

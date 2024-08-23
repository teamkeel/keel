package main

// A cmd that searches for all the *.keel files in the current working dir and
// any subdirectories (recursively), and
// reports on schema files found whose contents match some condition (of your choice).
//
// It helped when we switched the default permission to no-permission and thus
// had to edit all our existing schemas, but not in a perfectly universal way.
// And will likely be useful again for something similar.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	// Example 1) capture the location of all schema files that have at least one references to permissions?
	if strings.Contains(schemaAsString, "@permission") {
		fmt.Printf("XXXX This file uses permission attribute: %s\n", filePath)
	}

	// Example 2) capture the first line of every model declaration in all schema files.
	modelLines := []string{}
	re := regexp.MustCompile(`model\s`)
	for _, line := range lines {
		matched := re.MatchString(line)
		if matched {
			modelLines = append(modelLines, line)
		}
	}
	if len(modelLines) > 0 {
		fmt.Printf("XXXX this file contains the following model definitions\n%s\n", filePath)
		for _, ln := range modelLines {
			fmt.Printf("  %s\n", ln)
		}
	}
}

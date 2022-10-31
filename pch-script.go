package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

func main() {
	for _, dir := range []string{"./integration/testdata"} {
		doDir(dir)
	}
}

func doDir(dir string) {
	subDirs, err := os.ReadDir(dir)
	panicOnErr(err)

	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}
		schemaPath := path.Join(dir, subDir.Name(), "schema.keel")
		b, err := os.ReadFile(schemaPath)
		if err != nil {
			continue
		}
		schemaAsString := string(b)

		// Ignore this schema if it already has some Permissions in.
		if strings.Contains(schemaAsString, "@permission") {
			continue
		}

		fmt.Printf("XXXX processing: %s\n", schemaPath)

		// Build a replacement schema file, line by line,
		// sticking in a default permissions block below each model declaration line.
		lines := strings.Split(schemaAsString, "\n")
		newLines := []string{}
		for _, line := range lines {
			newLines := append(newLines, line)
			if isModelDeclaration(line) {
				fmt.Printf("XXXX found: %s\n", line)
				//newLines := append(newLines, permissionLines...)
			}
			_ = newLines
		}
	}
}

func panicOnErr(err error) {
	if err == nil {
		return
	}
	panic(fmt.Sprintf("stopping on err: %s", err))
}

func isModelDeclaration(line string) bool {
	matched, err := regexp.MatchString(`model\s`, line)
	panicOnErr(err)
	return matched
}

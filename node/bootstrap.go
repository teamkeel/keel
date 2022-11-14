package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/teamkeel/keel/schema"
)

func Bootstrap(dir string) error {
	_, err := os.Stat(filepath.Join(dir, "package.json"))

	if err == nil {
		return nil
	}

	// Make a package JSON for the temp dir that references my-package
	// as a local package
	err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(fmt.Sprintf(
		`
	{
		"name": "%s",
		"dependencies": {
			"@teamkeel/testing": "*",
			"@teamkeel/sdk":     "*",
			"@teamkeel/runtime": "*",
			"ts-node":           "*",
			// https://typestrong.org/ts-node/docs/swc/
			"@swc/core":           "*",
			"regenerator-runtime": "*",
		}
	}`, filepath.Base(dir),
	)), 0644)

	if err != nil {
		return err
	}

	npmInstall := exec.Command("npm", "install", "--progress=false", "--no-audit")
	npmInstall.Dir = dir

	o, err := npmInstall.CombinedOutput()

	if err != nil {
		fmt.Print(string(o))
		return err
	}

	return nil
}

// Generates @teamkeel/sdk and @teamkeel/testing
func GeneratePackages(dir string) error {
	// todo: if no functions or test files, then dont do anything.

	builder := schema.Builder{}

	_, err := builder.MakeFromDirectory(dir)

	if err != nil {
		return err
	}

	// sdk (client)
	// sdk typings d.ts
	// package.json for sdk (*peer* dependency on functions-runtime)
	// testing library
	// testing library d.t.s
	// package.json for testing (*peer* dependency on functions-runtime)

	panic("sksnsn")
}

func GenerateDevelopmentServer(dir string) error {
	// 1. make a single js file inside .build directory
	// that imports custom function handler code from node_modules/@teamkeel/functions-runtime
	// 2. bootstrap code to start a node server for the custom function runtime

	return nil
}

// maybe return a custom type over os.Process
func RunDevelopmentServer(dir string, envVars map[string]any) (*os.Process, error) {
	// 1. run dev server with ts-node.
	return nil, nil
}

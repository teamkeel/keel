package node

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/runtime"
)

const minimumRequiredNodeVersion = "18.0.0"

func getRequiredDependencies(options *bootstrapOptions) map[string]string {
	functionsRuntimeVersion := runtime.GetVersion()
	testingRuntimeVersion := runtime.GetVersion()

	// It is possible to reference a local version of our NPM modules rather than a version
	// from the NPM registry, by utilizing the --node-packages-path on the CLI. This flag is only applicable to the 'run' cmd at the moment, not 'generate'.
	if options.packagesPath != "" {
		functionsRuntimeVersion = filepath.Join(options.packagesPath, "functions-runtime")
		testingRuntimeVersion = filepath.Join(options.packagesPath, "testing-runtime")
	}

	return map[string]string{
		"@teamkeel/functions-runtime": functionsRuntimeVersion,
		"@teamkeel/testing-runtime":   testingRuntimeVersion,
		"@types/node":                 "18.11.18",
		"kysely":                      "0.23.4",
		"tsx":                         "3.12.6",
		"typescript":                  "4.9.4",
		"vitest":                      "0.27.2",
		"node-fetch":                  "3.3.0",
	}
}

type bootstrapOptions struct {
	packagesPath string
}

// WithPackagesPath causes any @teamkeel packages to be installed
// from this path. The path should point to the directory that contains
// all the different @teamkeel packages.
func WithPackagesPath(p string) BootstrapOption {
	return func(o *bootstrapOptions) {
		o.packagesPath = p
	}
}

type BootstrapOption func(o *bootstrapOptions)

// Bootstrap ensures that the target directory has a package.json file with the correct dependencies
// required by the Keel Custom Functions runtime, and additionally it ensures that a tsconfig.json
// file has been generated
func Bootstrap(dir string, opts ...BootstrapOption) (codegen.GeneratedFiles, error) {
	packageJsonPath := filepath.Join(dir, "package.json")

	_, err := os.Stat(packageJsonPath)

	// if the package.json doesn't exist, then we need to generate with npm init -y
	if errors.Is(err, os.ErrNotExist) {
		npmInit := exec.Command("npm", "init", "--yes")
		npmInit.Dir = dir

		err := npmInit.Run()

		if err != nil {
			return codegen.GeneratedFiles{}, err
		}
	} else if err != nil {
		// any other type of error
		return codegen.GeneratedFiles{}, err
	}

	// Once we know the package.json exists, we need to install all of the required dependencies using npm install --save {deps}
	// if the dependencies are already satisfied in the package.json this will be a noop or if an older version
	// of the dependency is present, then it will be updated (this is particularly relevant for our @teamkeel dependencies that take their version from the runtime version of the CLI via ldflags).

	options := &bootstrapOptions{}
	for _, o := range opts {
		o(options)
	}

	files := codegen.GeneratedFiles{}

	// the args to pass to the npm cmd
	args := []string{}

	// the first arg is obviously install (because that's what this is all about)
	args = append(args, "install")

	requiredDeps := getRequiredDependencies(options)

	// due to the way that exec.Command handles args, we can't just pass a concatenated string
	// of dependences to install to npm install. Instead we need to pass the whole args array as a second
	// argument to the exec.Command call.
	for key, value := range requiredDeps {
		if value == "" {
			args = append(args, key)
		} else {
			args = append(args, fmt.Sprintf("%s@%s", key, value))
		}
	}

	installCmd := exec.Command("npm", args...)
	installCmd.Dir = dir

	err = installCmd.Run()

	if err != nil {
		return codegen.GeneratedFiles{}, fmt.Errorf("Could not install required dependencies")
	}

	files = append(files, &codegen.GeneratedFile{
		Path: "tsconfig.json",
		Contents: `{
			"compilerOptions": {
				"lib": ["ES2016"],
				"target": "ES2016",
				"esModuleInterop": true,
				"moduleResolution": "node",
				"skipLibCheck": true,
				"strictNullChecks": true,
				"types": ["node"],
				"allowJs": true,
				"resolveJsonModule": true,
				"paths": {
					"@teamkeel/sdk": ["./.build/sdk"],
					"@teamkeel/testing": ["./.build/testing"]
				}
			},
			"include": ["**/*.ts"],
			"exclude": ["node_modules"]
		}`,
	})

	err = files.Write(dir)

	if err != nil {
		return codegen.GeneratedFiles{}, err
	}

	return files, nil
}

func CheckNodeVersion() error {
	_, err := exec.LookPath("node")
	if errors.Is(err, exec.ErrNotFound) {
		return &NodeNotFoundError{}
	}

	output, err := exec.Command("node", "--version").Output()
	if err != nil {
		return err
	}

	nodeVersion := strings.TrimPrefix(string(output), "v")
	nodeVersion = strings.TrimSuffix(nodeVersion, "\n")

	validVersionNumber, err := regexp.MatchString(`(\d+)\.(\d+)\.(\d+)`, nodeVersion)
	if err != nil {
		return err
	}
	if !validVersionNumber {
		return fmt.Errorf("unexpected output from node -v: '%s'", nodeVersion)
	}

	current, err := version.NewVersion(nodeVersion)
	if err != nil {
		return err
	}
	minimum, err := version.NewVersion(minimumRequiredNodeVersion)
	if err != nil {
		return err
	}

	if current.LessThan(minimum) {
		return &IncorrectNodeVersionError{
			Current: nodeVersion,
			Minimum: minimumRequiredNodeVersion,
		}
	}

	return nil
}

type NpmInstallError struct {
	err    error
	Output string
}

func (n *NpmInstallError) Error() string {
	return fmt.Sprintf("npm install error (%s): %s", n.err.Error(), n.Output)
}

type IncorrectNodeVersionError struct {
	Current string
	Minimum string
}

func (n *IncorrectNodeVersionError) Error() string {
	return fmt.Sprintf("incorrect node version. requires %s", n.Minimum)
}

type NodeNotFoundError struct{}

func (n *NodeNotFoundError) Error() string {
	return "node command not found"
}

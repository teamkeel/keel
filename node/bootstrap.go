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

// Bootstrap sets dir up to use either custom functions or write tests. It will do nothing
// if there is already a package.json present in the directory.
func Bootstrap(dir string, opts ...BootstrapOption) (codegen.GeneratedFiles, error) {
	_, err := os.Stat(filepath.Join(dir, "package.json"))
	// todo: this probably isn't what we want
	// No error - we have a package.json so we're done
	if err == nil {
		return codegen.GeneratedFiles{}, nil
	}

	// A "not exists" error is fine, that means we're generating a fresh package.json
	// Bail on all other errors
	if !os.IsNotExist(err) {
		return codegen.GeneratedFiles{}, nil
	}

	options := &bootstrapOptions{}
	for _, o := range opts {
		o(options)
	}

	functionsRuntimeVersion := runtime.GetVersion()
	testingRuntimeVersion := runtime.GetVersion()

	if options.packagesPath != "" {
		functionsRuntimeVersion = filepath.Join(options.packagesPath, "functions-runtime")
		testingRuntimeVersion = filepath.Join(options.packagesPath, "testing-runtime")
	}

	files := codegen.GeneratedFiles{}

	files = append(files, &codegen.GeneratedFile{
		Path: "package.json",
		Contents: fmt.Sprintf(`{
			"name": "%s",
			"dependencies": {
				"@teamkeel/functions-runtime": "%s",
				"@teamkeel/testing-runtime": "%s",
				"@types/node": "^18.11.18",
				"kysely": "^0.23.4",
				"tsx": "^3.12.6",
				"typescript": "^4.9.4",
				"vitest": "^0.27.2",
				"node-fetch": "^3.3.0",
				"@prisma/client": "^4.14.1"
			}
		}`, filepath.Base(dir), functionsRuntimeVersion, testingRuntimeVersion),
	})

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
				"allowJs": true
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

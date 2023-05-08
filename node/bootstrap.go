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
func Bootstrap(dir string, opts ...BootstrapOption) error {
	_, err := os.Stat(filepath.Join(dir, "package.json"))
	// No error - we have a package.json so we're done
	if err == nil {
		return nil
	}

	// A "not exists" error is fine, that means we're generating a fresh package.json
	// Bail on all other errors
	if !os.IsNotExist(err) {
		return err
	}

	options := &bootstrapOptions{}
	for _, o := range opts {
		o(options)
	}

	functionsRuntimeVersion := "*"
	testingRuntimeVersion := "*"
	if options.packagesPath != "" {
		functionsRuntimeVersion = filepath.Join(options.packagesPath, "functions-runtime")
		testingRuntimeVersion = filepath.Join(options.packagesPath, "testing-runtime")
	}

	err = os.WriteFile(filepath.Join(dir, "package.json"), []byte(fmt.Sprintf(`{
		"name": "%s",
		"dependencies": {
			"@teamkeel/functions-runtime": "%s",
			"@teamkeel/testing-runtime": "%s",
			"@types/node": "^18.11.18",
			"kysely": "^0.23.4",
			"tsx": "^3.12.6",
			"typescript": "^4.9.4",
			"vitest": "^0.27.2",
			"node-fetch": "^3.3.0"
		}
	}`, filepath.Base(dir), functionsRuntimeVersion, testingRuntimeVersion)), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{
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
	}`), 0644)
	if err != nil {
		return err
	}

	npmInstall := exec.Command("npm", "install", "--progress=false", "--no-audit")
	npmInstall.Dir = dir

	o, err := npmInstall.CombinedOutput()
	if err != nil {
		return &NpmInstallError{
			Output: string(o),
			err:    err,
		}
	}

	return nil
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

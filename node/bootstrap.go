package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/go-version"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/runtime"
)

const minimumRequiredNodeVersion = "18.0.0"

func GetDependencies(options *bootstrapOptions) (map[string]string, map[string]string) {
	functionsRuntimeVersion := runtime.GetVersion()
	testingRuntimeVersion := runtime.GetVersion()

	// If doing a snapshot build with goreleaser the function will be something like <version>-SNAPSHOT-<git_hash> which
	// won't work, so we remove the "-SNAPSHOt-<git_hash>" bit
	if strings.Contains(functionsRuntimeVersion, "-SNAPSHOT-") {
		functionsRuntimeVersion = strings.Split(functionsRuntimeVersion, "-SNAPSHOT-")[0]
	}
	if strings.Contains(testingRuntimeVersion, "-SNAPSHOT-") {
		testingRuntimeVersion = strings.Split(testingRuntimeVersion, "-SNAPSHOT-")[0]
	}

	// It is possible to reference a local version of our NPM modules rather than a version
	// from the NPM registry, by utilizing the --node-packages-path on the CLI. This flag is only applicable to the 'run' cmd at the moment, not 'generate'.
	if options.packagesPath != "" {
		functionsRuntimeVersion = filepath.Join(options.packagesPath, "functions-runtime")
		testingRuntimeVersion = filepath.Join(options.packagesPath, "testing-runtime")
	}

	deps := map[string]string{
		"@teamkeel/functions-runtime": functionsRuntimeVersion,
		"@teamkeel/testing-runtime":   testingRuntimeVersion,
		"kysely":                      "0.27.4",
		"node-fetch":                  "3.3.2",
	}

	devDeps := map[string]string{
		"@types/node": "22.13.5",
		"tsx":         "4.19.3",
		"typescript":  "5.7.3",
		"vitest":      "^3.0.7",
	}

	return deps, devDeps
}

type bootstrapOptions struct {
	packagesPath   string
	packageManager string
	logger         func(string)
	output         io.Writer
}

// WithPackagesPath causes any @teamkeel packages to be installed
// from this path. The path should point to the directory that contains
// all the different @teamkeel packages.
func WithPackagesPath(p string) BootstrapOption {
	return func(o *bootstrapOptions) {
		o.packagesPath = p
	}
}

func WithLogger(l func(string)) BootstrapOption {
	return func(o *bootstrapOptions) {
		o.logger = l
	}
}

func WithOutputWriter(w io.Writer) BootstrapOption {
	return func(o *bootstrapOptions) {
		o.output = w
	}
}

func WithPackageManager(p string) BootstrapOption {
	return func(o *bootstrapOptions) {
		o.packageManager = p
	}
}

type BootstrapOption func(o *bootstrapOptions)

type PackageJson struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// Bootstrap ensures that the target directory has a package.json file with the correct dependencies
// required by the Keel Custom Functions runtime, and additionally it ensures that a tsconfig.json
// file has been generated
func Bootstrap(dir string, opts ...BootstrapOption) error {
	options := &bootstrapOptions{
		packageManager: "npm",
		logger: func(s string) {
			fmt.Println(s)
		},
		output: io.Discard,
	}
	for _, o := range opts {
		o(options)
	}

	if !lo.Contains([]string{"npm", "pnpm"}, options.packageManager) {
		return fmt.Errorf("unsupported package manager: %s", options.packageManager)
	}

	packageJsonPath := filepath.Join(dir, "package.json")
	tsConfigPath := filepath.Join(dir, "tsconfig.json")

	_, err := os.Stat(packageJsonPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if errors.Is(err, fs.ErrNotExist) {
		options.logger("Creating package.json")
		b, _ := json.MarshalIndent(map[string]string{"name": filepath.Base(dir)}, "", "  ")

		err = os.WriteFile(packageJsonPath, b, os.ModePerm)
		if err != nil {
			return err
		}
	}

	b, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return err
	}

	var pkgJson PackageJson
	err = json.Unmarshal(b, &pkgJson)
	if err != nil {
		return err
	}

	deps, devDeps := GetDependencies(options)

	toInstall, err := getDepsToInstall(deps, pkgJson.Dependencies)
	if err != nil {
		return err
	}
	if len(toInstall) > 0 {
		options.logger("Installing dependencies...")
		err = installDeps(dir, toInstall, false, options)
		if err != nil {
			return err
		}
	}

	toInstall, err = getDepsToInstall(devDeps, pkgJson.DevDependencies)
	if err != nil {
		return err
	}
	if len(toInstall) > 0 {
		options.logger("Installing dev dependencies...")
		err = installDeps(dir, toInstall, true, options)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(tsConfigPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if errors.Is(err, fs.ErrNotExist) {
		options.logger("Creating tsconfig.json")
		tsConfig := map[string]any{
			"compilerOptions": map[string]any{
				"lib":               []string{"ES2016", "DOM"},
				"target":            "ES2016",
				"esModuleInterop":   true,
				"moduleResolution":  "node",
				"skipLibCheck":      true,
				"strictNullChecks":  true,
				"types":             []string{"node"},
				"allowJs":           true,
				"resolveJsonModule": true,
				"paths": map[string]any{
					"@teamkeel/sdk":     []string{"./.build/sdk"},
					"@teamkeel/testing": []string{"./.build/testing"},
				},
			},
			"include": []string{"**/*.ts"},
			"exclude": []string{"node_modules"},
		}
		b, err := json.MarshalIndent(tsConfig, "", "  ")
		if err != nil {
			return err
		}
		err = os.WriteFile(tsConfigPath, b, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDepsToInstall(required map[string]string, existing map[string]string) ([]string, error) {
	toInstall := []string{}
	for dep, version := range required {
		if version == "" {
			version = "latest"
		}

		withVersion := fmt.Sprintf("%s@%s", dep, version)

		v, ok := existing[dep]
		if !ok {
			toInstall = append(toInstall, withVersion)
			continue
		}

		requiredVersion, _ := semver.NewVersion(version)
		if requiredVersion != nil {
			constraint, err := semver.NewConstraint(v)
			if err != nil {
				return nil, err
			}
			if constraint.Check(requiredVersion) {
				continue
			}
		}

		toInstall = append(toInstall, withVersion)
	}

	return toInstall, nil
}

func installDeps(dir string, deps []string, dev bool, options *bootstrapOptions) error {
	args := []string{}

	switch options.packageManager {
	case "npm":
		args = append(args, "install")
	case "pnpm":
		args = append(args, "add")
	}

	args = append(args, deps...)
	args = append(args, lo.Ternary(dev, "--save-dev", "--save-prod"))

	cmd := exec.Command(options.packageManager, args...)
	cmd.Stdout = options.output
	cmd.Stderr = options.output
	cmd.Dir = dir

	return cmd.Run()
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

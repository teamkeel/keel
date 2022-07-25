package runtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/aybabtme/orderedjson"
)

type Dependencies = map[string]string

type PackageJson struct {
	// Meta fields
	Path     string `json:"-"`
	Contents string `json:"-"`

	// Dev + normal dependencies defined in the json file
	Dependencies    Dependencies `json:"dependencies"`
	DevDependencies Dependencies `json:"devDependencies"`
}

// Instantiates an in memory representation of a package.json file.
// The relevant entries (devDependencies / dependencies) we are
// interested in are unmarshalled into memory
func NewPackageJson(path string) (*PackageJson, error) {
	p := PackageJson{
		Path: path,
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No package.json found, creating...")

		cmd := exec.Command("npm", "init", "-y")
		cmd.Dir = filepath.Dir(path)

		err := cmd.Run()

		if err != nil {
			return nil, err
		}

		err = p.ReadIntoMemory()

		if err != nil {
			return nil, err
		}

		p.Dependencies = map[string]string{}
		p.DevDependencies = map[string]string{}

		err = p.Write()

		if err != nil {
			return nil, err
		}
	}

	return &p, nil
}

// Runs npm install on the current *written* state of the package.json file, causing node_modules to be populated, and the lockfile to be updated
// Call .Write() beforehand to persist any changes made.
func (p *PackageJson) Install() error {
	npmInstall := exec.Command("npm", "install")

	workDir := path.Dir(p.Path)
	npmInstall.Dir = workDir

	err := npmInstall.Run()

	if err != nil {
		return err
	}

	return nil
}

func (p *PackageJson) ReadIntoMemory() error {
	bytes, err := os.ReadFile(p.Path)

	if err != nil {
		return err
	}

	p.Contents = string(bytes)

	err = json.Unmarshal(bytes, &p)

	if err != nil {
		return err
	}

	return nil
}

// Inject devDependencies into the package.json file
// Where there are matching packages already, the version we inject overwrites the original
func (p *PackageJson) Inject(devDeps map[string]string, deps map[string]string) error {
	if p.DevDependencies != nil {
		d := p.DevDependencies

		for packageName, version := range devDeps {
			if originalVersion, found := d[packageName]; found {
				d[packageName] = originalVersion
			} else {
				d[packageName] = version
			}
		}
	} else {
		var d = map[string]string{}

		for packageName, version := range devDeps {
			d[packageName] = version
		}

		p.DevDependencies = d
	}

	if p.Dependencies != nil {
		d := p.Dependencies

		for packageName, version := range deps {
			if originalVersion, found := d[packageName]; found {
				d[packageName] = originalVersion
			} else {
				d[packageName] = version
			}
		}
	} else {
		var d = map[string]string{}

		for packageName, version := range deps {
			d[packageName] = version
		}

		p.Dependencies = d
	}

	err := p.Write()

	if err != nil {
		return err
	}

	fmt.Println("Injected dev dependencies")
	return nil
}

var (
	KeyDependencies    = "dependencies"
	KeyDevDependencies = "devDependencies"
)

// Write will inject any changes made in memory to the target
// package.json
// Using standard Marshal/Unmarshal into a map[string]interface{}
// does not guarantee that the keys will be serialized in the order originally specified
// so we need to use a special orderjson.Map objec to ensure the order is not disturbed
func (p *PackageJson) Write() error {
	var originalPackageJson orderedjson.Map
	var mutatedPackageJson orderedjson.Map

	err := json.Unmarshal([]byte(p.Contents), &originalPackageJson)

	if err != nil {
		return err
	}

	for _, entry := range originalPackageJson {
		k, err := strconv.Unquote(string(entry.Key))

		if err != nil {
			continue
		}

		switch k {
		case KeyDevDependencies:
			b, err := json.Marshal(p.DevDependencies)

			if err != nil {
				return err
			}

			entry.Value = json.RawMessage(b)
		case KeyDependencies:
			b, err := json.Marshal(p.Dependencies)

			if err != nil {
				return err
			}

			entry.Value = json.RawMessage(b)
		}

		mutatedPackageJson = append(mutatedPackageJson, entry)
	}

	marshalled, err := mutatedPackageJson.MarshalJSON()
	if err != nil {
		return err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, marshalled, "", "  ")

	if err != nil {
		return err
	}
	err = os.WriteFile(p.Path, prettyJSON.Bytes(), 0644)

	if err != nil {
		return err
	}

	// Update the lockfile
	err = p.Install()

	if err != nil {
		return err
	}

	p.Contents = prettyJSON.String()

	fmt.Println("Wrote changes to package.json")
	return nil
}

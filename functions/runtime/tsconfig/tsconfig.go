package tsconfig

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/aybabtme/orderedjson"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

type CompilerOptions = map[string]interface{}

type Include = []string

type TSConfig struct {
	Path     string `json:"-"`
	Contents string `json:"-"`

	CompilerOptions CompilerOptions `json:"compilerOptions"`
	Include         Include         `json:"include"`
}

type PathAlias struct {
	Alias string
	Paths []string
}

type PathAliases = []PathAlias

//go:embed tsconfig.json
var defaultTsConfig string

func DefaultConfig() (*TSConfig, error) {
	c := TSConfig{}
	err := json.Unmarshal([]byte(defaultTsConfig), &c)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func NewTSConfig(path string) (*TSConfig, error) {
	t := TSConfig{
		Path: path,
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No tsconfig.json found, creating...")

		file, err := os.Create(path)

		if err != nil {
			return nil, err
		}

		_, err = file.WriteString(defaultTsConfig)

		if err != nil {
			return nil, err
		}

		t.Contents = defaultTsConfig

		err = t.ReadIntoMemory()

		if err != nil {
			return nil, err
		}

		return &t, nil
	}

	err := t.ReadIntoMemory()

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *TSConfig) ReadIntoMemory() error {
	bytes, err := os.ReadFile(t.Path)

	if err != nil {
		return err
	}

	t.Contents = string(bytes)

	err = json.Unmarshal(bytes, t)

	if err != nil {
		return err
	}

	return nil
}

func (t *TSConfig) Write() error {
	var originalTsConfig orderedjson.Map
	// var mutatedTsConfig orderedjson.Map

	err := json.Unmarshal([]byte(t.Contents), &originalTsConfig)

	if err != nil {
		return err
	}

	b, err := originalTsConfig.MarshalJSON()

	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, b, "", "  ")

	if err != nil {
		return err
	}

	return os.WriteFile(t.Path, prettyJSON.Bytes(), 0644)
}

func (t *TSConfig) Reconcile() error {
	// check includes
	defaultConfig, err := DefaultConfig()

	if err != nil {
		return err
	}

	// compare lib entries

	dInterface, ok := defaultConfig.CompilerOptions["lib"].([]interface{})

	if !ok {
		return errors.New("could not parse default tsconfig.json")
	}

	defaultLibs := []string{}

	for _, item := range dInterface {
		if str, ok := item.(string); ok {
			defaultLibs = append(defaultLibs, str)
		}
	}

	aInterface, ok := t.CompilerOptions["lib"].([]interface{})

	if !ok {
		return errors.New("could not parse tsconfig.json")
	}

	actualLibs := []string{}

	for _, item := range aInterface {
		if str, ok := item.(string); ok {
			actualLibs = append(actualLibs, str)
		}
	}

	intersection := lo.Intersect(actualLibs, defaultLibs)

	if slices.Equal(intersection, defaultLibs) {
		// nothing to do
		fmt.Print("compilerOptions.lib satisfied")
	} else {
		t.CompilerOptions["lib"] = append(actualLibs, defaultLibs...)
	}

	// paths
	pDInterface, ok := defaultConfig.CompilerOptions["paths"].(map[string]interface{})

	if !ok {
		return errors.New("could not parse default tsconfig.json")
	}

	defaultAliases := PathAliases{}

	for key, paths := range pDInterface {
		ps := paths.([]interface{})

		strPaths := []string{}

		for _, p := range ps {
			if str, ok := p.(string); ok {
				strPaths = append(strPaths, str)
			}
		}

		defaultAliases = append(defaultAliases, PathAlias{Alias: key, Paths: strPaths})
	}

	aDInterface, ok := t.CompilerOptions["paths"].(map[string]interface{})

	if !ok {
		return errors.New("could not parse default tsconfig.json")
	}

	actualAliases := PathAliases{}

	for key, paths := range aDInterface {
		ps := paths.([]interface{})

		strPaths := []string{}

		for _, p := range ps {
			if str, ok := p.(string); ok {
				strPaths = append(strPaths, str)
			}
		}

		actualAliases = append(actualAliases, PathAlias{Alias: key, Paths: strPaths})
	}

	for _, defaultAlias := range defaultAliases {
		match, found := lo.Find(actualAliases, func(a PathAlias) bool {
			return a.Alias == defaultAlias.Alias
		})

		if found {
			combined := lo.Union(match.Paths, defaultAlias.Paths)
			t.CompilerOptions["paths"].(map[string]interface{})[match.Alias] = combined
		}
	}

	// Write the changes
	err = t.Write()

	if err != nil {
		return err
	}

	return nil
}

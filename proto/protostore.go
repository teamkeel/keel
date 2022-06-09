package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

func SaveToLocalStorage(p *Schema, schemaDir string) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("could not save protobuf to local storage (json marshal): %v", err)
	}
	privateDir, err := accessPrivateDir(schemaDir)
	if err != nil {
		return fmt.Errorf("could not save protobuf to local storage (directory access): %v", err)
	}
	protoFile := path.Join(privateDir, protoFileBaseName)
	if err := os.WriteFile(protoFile, b, 0644); err != nil {
		return fmt.Errorf("could not save protobuf to local storage (file write error): %v", err)
	}
	return nil
}

// FetchFromLocalStorage returns the protobuf.Schema that has been serialized into
// the .keel private directory inside the given schema directory. If that file does not
// exist it returns a valid (but empty) schema.
func FetchFromLocalStorage(schemaDir string) (*Schema, error) {
	privateDir, err := accessPrivateDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("could not fetch protobuf from local storage (directory access): %v", err)
	}
	protoFile := path.Join(privateDir, protoFileBaseName)

	// Detect the first ever pass through this code, for any given schemaDir, and
	// return a valid, but empty protobuf.
	if !fileExists(protoFile) {
		return &Schema{}, nil
	}
	b, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, fmt.Errorf("could not fetch protobuf from local storage (reading file): %v", err)
	}
	proto := Schema{}
	if err = json.Unmarshal(b, &proto); err != nil {
		return nil, fmt.Errorf("could not fetch protobuf from local storage (json unmarshal): %v", err)
	}

	return &proto, nil
}

// accessPrivateDir returns the full path of the .keel private directory
// inside the given schema directory. It creates it also if it does not
// already exist.
func accessPrivateDir(schemaDir string) (string, error) {
	privateDir := path.Join(schemaDir, privateDirBasename)
	if err := os.MkdirAll(privateDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating private directory: %v", err)
	}
	return privateDir, nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

const privateDirBasename string = ".keel-state"
const protoFileBaseName string = "last-known-proto"

package migrations

import (
	"fmt"
	"os"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/protoqry"
)

func MakeMigrationsFromSchemaDifference(oldProto, newProto *proto.Schema) (theSQL string, err error) {
	differ := NewProtoDiffer(oldProto, newProto)
	differences, err := differ.Analyse() // todo compress the constructor and method into a plain function.
	if err != nil {
		return "", fmt.Errorf("could not analyse differences: %v", err)
	}
	// Create tables for any new models
	for _, newModelName := range differences.ModelsAdded {
		model := protoqry.FindModel(newProto.Models, newModelName)
		theSQL += createTable(model)
	}

	if os.Getenv("DEBUG") != "" {
		fmt.Printf("\n%s\n\n", theSQL)
	}

	return theSQL, nil
}

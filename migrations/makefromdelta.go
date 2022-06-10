package migrations

import (
	"fmt"
	"os"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/protoqry"
)

func MakeMigrationsFromSchemaDifference(oldProto, newProto *proto.Schema) (theSQL string, err error) {
	differences, err := ProtoDeltas(oldProto, newProto)
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

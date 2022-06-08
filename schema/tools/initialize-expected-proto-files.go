package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nsf/jsondiff"
	"github.com/teamkeel/keel/schema"
)

/*
This command !!!ASSUMES!!! that the Schema.MakeFromDirectory function produces
correct results, and uses it to initialize "expected" test data output.

Specificially it populates all the directories whose name is testdata/proto<something>
with a proto.json file.
*/
func main() {
	testdataDir := os.Args[1]
	subDirs, err := os.ReadDir(testdataDir)
	if err != nil {
		panic(fmt.Errorf("cannot read the testdata directory: %v", err))
	}

	var stats stats

	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}
		if !strings.HasPrefix(subDir.Name(), "proto") {
			continue
		}

		outputFile := filepath.Join(testdataDir, subDir.Name(), "proto.json")
		originalContents := getFileContents(outputFile)

		s2m := schema.Builder{}
		protoSchema, err := s2m.MakeFromDirectory(testdataDir + "/" + subDir.Name())
		if err != nil {
			panic(fmt.Errorf("failed to make schema from directory: %v", err))
		}

		opts := protojson.MarshalOptions{Indent: "  "}
		newFileContents, err := opts.Marshal(protoSchema)
		if err != nil {
			panic(fmt.Errorf("could not marshal protobuf structure into json: %v", err))
		}

		if len(originalContents) == 0 || filesDiffer(originalContents, newFileContents) {
			err = os.WriteFile(outputFile, newFileContents, 0666)
			if err != nil {
				panic(fmt.Errorf("could not save proto.json file: %v", err))
			}
		}

		// Update statistics
		switch {
		case len(originalContents) == 0:
			stats.created = append(stats.created, subDir.Name())
		case len(originalContents) != 0 && filesDiffer(originalContents, newFileContents):
			stats.changed = append(stats.changed, subDir.Name())
		default:
			stats.unchanged++
		}
	}
	outputStats(stats)
}

type stats struct {
	created   []string
	unchanged int
	changed   []string
}

func getFileContents(fileName string) []byte {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return []byte{}
	}
	return contents
}

func filesDiffer(a, b []byte) bool {
	opts := jsondiff.DefaultConsoleOptions()
	diff, _ := jsondiff.Compare(a, b, &opts)
	switch diff {
	case jsondiff.FullMatch:
		return false
	case jsondiff.SupersetMatch, jsondiff.NoMatch:
		return true
	default:
		panic("jsondiff.Compare() thinks that one or other of the given files are invalid JSON")
	}
}

func outputStats(stats stats) {
	if stats.unchanged > 0 {
		fmt.Printf("%d files were unchanged\n\n", stats.unchanged)
	}

	if len(stats.created) > 0 {
		fmt.Println("The following files were created...")
		for _, c := range stats.created {
			fmt.Printf(" - %s\n", c)
		}
		fmt.Println("")
	}

	if len(stats.changed) > 0 {
		fmt.Println("The following files changed...")
		for _, c := range stats.changed {
			fmt.Printf(" - %s\n", c)
		}
		fmt.Println("")
	}
}

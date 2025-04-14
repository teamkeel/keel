package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nsf/jsondiff"
	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/tools"
)

func main() {
	var stats stats

	testdataDir := os.Args[1]
	subDirs, err := os.ReadDir(testdataDir)
	if err != nil {
		panic(fmt.Errorf("cannot read the testdata directory: %v", err))
	}
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}

		currContents, err := os.ReadFile(filepath.Join(testdataDir, subDir.Name(), "tools.json"))
		if err != nil {
			panic(err)
		}

		builder := schema.Builder{}
		schema, err := builder.MakeFromDirectory(filepath.Join(testdataDir, subDir.Name()))
		if err != nil {
			panic(err)
		}
		tools, err := tools.GenerateTools(context.Background(), schema, builder.Config)
		if err != nil {
			panic(err)
		}
		response := &rpc.ListToolsResponse{Tools: tools}
		opts := protojson.MarshalOptions{Indent: "  "}
		b, err := opts.Marshal(response)
		if err != nil {
			panic(fmt.Errorf("cannot generate tools for testdata directory %s: %v", subDir.Name(), err))
		}
		var dest bytes.Buffer
		_ = json.Indent(&dest, b, "", "  ")
		outputContents := dest.Bytes()

		outputFileName := filepath.Join(testdataDir, subDir.Name(), "tools.json")

		if len(currContents) == 0 || filesDiffer(currContents, outputContents) {
			err = os.WriteFile(outputFileName, outputContents, 0666)
			if err != nil {
				panic(fmt.Errorf("could not save file: %v", err))
			}
		}
		// Update statistics
		switch {
		case len(currContents) == 0:
			stats.created = append(stats.created, subDir.Name())
		case len(currContents) != 0 && filesDiffer(currContents, outputContents):
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

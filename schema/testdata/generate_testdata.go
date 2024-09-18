package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nsf/jsondiff"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

/*
This command !!!ASSUMES!!! that the Schema.MakeFromDirectory function produces
correct results, and uses it to initialize "expected" test data output.
The logic of this command is:
  - if MakeFromDirectory returns validation errors then they are written to an errors.json file
  - if MakeFromDirectory returns no errors then the proto message is written to a proto.json file
  - if an errors.json file would be written but a proto.json file already exists a warning is printed
  - if a proto.json file would be written but an errors.json file alrady exists a warning is printed

This command should be run via the make command `testdata` e.g. `make testdata`
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

		files, err := os.ReadDir(filepath.Join(testdataDir, subDir.Name()))
		if err != nil {
			panic(err)
		}

		var hasCurrErrors bool
		var hasCurrProto bool
		var currContents []byte
		var defaultConfig *config.ProjectConfig

		if len(files) < 1 {
			panic(fmt.Sprintf("No files present in directory %s", subDir.Name()))
		}

		for _, file := range files {
			if file.Name() == "errors.json" {
				b, err := os.ReadFile(filepath.Join(testdataDir, subDir.Name(), "errors.json"))
				if err != nil {
					panic(err)
				}
				currContents = b
				hasCurrErrors = true
			}
			if file.Name() == "proto.json" {
				b, err := os.ReadFile(filepath.Join(testdataDir, subDir.Name(), "proto.json"))
				if err != nil {
					panic(err)
				}
				currContents = b
				hasCurrProto = true
			}
		}

		if hasCurrErrors && hasCurrProto {
			fmt.Printf("WARNING: Test case %s has both an errors.json file and a proto.json file. Test cases can only be valid or invalid, delete one of these files and re-run this command.\n", subDir.Name())
			continue
		}

		var hasNewErrors bool
		var outputFileName string
		var outputContents []byte

		s2m := schema.Builder{}

		if defaultConfig != nil {
			s2m.Config = defaultConfig
		}

		protoSchema, err := s2m.MakeFromDirectory(filepath.Join(testdataDir, subDir.Name()))
		if err != nil {
			verrs, ok := err.(*errorhandling.ValidationErrors)
			if !ok {
				panic(fmt.Errorf("failed to make schema from directory %s: %v", subDir.Name(), err))
			}

			b, _ := json.MarshalIndent(verrs, "", "  ")
			outputContents = b
			outputFileName = filepath.Join(testdataDir, subDir.Name(), "errors.json")
			hasNewErrors = true

		} else {
			opts := protojson.MarshalOptions{Indent: "  "}
			b, err := opts.Marshal(protoSchema)
			if err != nil {
				panic(err)
			}

			// protojson does some slightly weird things with whitespace so we run
			// the output through go's default indenter to fix this

			//var dest bytes.Buffer
			//_ = json.Indent(&dest, b, "", "  ")
			outputContents = b //dest.Bytes()

			outputFileName = filepath.Join(testdataDir, subDir.Name(), "proto.json")
		}

		if hasCurrErrors && !hasNewErrors {
			fmt.Printf("WARNING: Test case %s has an errors.json file but produced a valid proto. If this is correct delete the errors.json file and re-run this command.\n", subDir.Name())
			continue
		}

		if hasCurrProto && hasNewErrors {
			fmt.Printf("WARNING: Test case %s has a proto.json file but produced validation errors. If this is correct then delete the proto.json file and re-run this command.\n", subDir.Name())
			continue
		}

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

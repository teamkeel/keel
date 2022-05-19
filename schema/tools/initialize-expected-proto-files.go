package main

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/teamkeel/keel/schema"
)

/*
This command !!!ASSUMES!!! that the Schema.MakeFromDirectory function produces
correct results, and uses it to initialize "expected" test data output.

Specificially it populates all the directories whose name is testdata/proto<something>
with a proto.json file.
*/
func main() {
	testdataDir := "../testdata"
	subDirs, err := os.ReadDir(testdataDir)
	if err != nil {
		panic(fmt.Errorf("cannot read the testdata directory: %v", err))
	}

	nFilesWritten := 0
	for _, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}
		if !strings.HasPrefix(subDir.Name(), "proto") {
			continue
		}

		s2m := schema.Schema{}
		protoSchema, err := s2m.MakeFromDirectory(testdataDir + "/" + subDir.Name())
		if err != nil {
			panic(fmt.Errorf("failed to make schema from directory: %v", err))
		}

		opts := protojson.MarshalOptions{Indent: "  "}
		asJSON, err := opts.Marshal(protoSchema)
		if err != nil {
			panic(fmt.Errorf("could not marshal protobuf structure into json: %v", err))
		}
		
		err = os.WriteFile("../testdata/" + subDir.Name() + "/proto.json", asJSON, 0666)
		if err != nil {
			panic(fmt.Errorf("could not save proto.json file: %v", err))
		}
		nFilesWritten++
	}
	fmt.Printf("Success, %d files written\n", nFilesWritten)
}
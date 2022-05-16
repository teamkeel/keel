package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/teamkeel/keel/schema"
)

/*
This command automatically PRE-CREATES our tests' expected proto.json files by
running the protobuf generation stage and ASSUMING the files generated are
correct.

It operates on all of the sub directories inside the testdata directory whose
name begins with "proto".
*/
func main() {
	testdataDir := ".."
	fileNodes, err := ioutil.ReadDir(testdataDir)
	if err != nil {
		panic(fmt.Errorf("ioutil.ReadDir() failed with: %v", err))
	}

	nFilesWritten := 0
	for _, fileNode := range fileNodes {
		if !fileNode.IsDir() {
			continue
		}
		dirFullPathName := fileNode.Name()
		if !strings.HasPrefix(dirFullPathName, "proto") {
			continue
		}

		s2m := schema.Schema{}
		protoSchema, err := s2m.MakeFromDirectory(dirFullPathName)
		if err != nil {
			panic(fmt.Errorf("MakeFromDirectory() failed with: %v", err))
		}

		opts := protojson.MarshalOptions{Indent: "  "}
		asJSON, err := opts.Marshal(protoSchema)
		if err != nil {
			panic(fmt.Errorf("Marshal() failed with: %v", err))
		}
		
		err = os.WriteFile(dirFullPathName + "/proto.json", asJSON, 0666)
		if err != nil {
			panic(fmt.Errorf("Marshal() failed with: %v", err))
		}
		nFilesWritten++
	}
	fmt.Printf("Success, %d files written\n", nFilesWritten)
}
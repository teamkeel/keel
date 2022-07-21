package runtime_test

import (
	"io/fs"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions/runtime"
)

// Tests:
// - Run tsc against generated "whole thing"
// - Run code with node.js
// - Throw some requests at the server and assert that the right response is returned

func TestAllCases(t *testing.T) {
	testCases, err := ioutil.ReadDir("testdata")
	require.NoError(t, err)

	toRun := []fs.FileInfo{}

	for _, testCase := range testCases {
		if strings.HasSuffix(testCase.Name(), ".only") {
			toRun = append(toRun, testCase)
		}
	}

	if len(toRun) > 0 {
		testCases = toRun
	}

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			continue
		}

		t.Run(testCase.Name(), func(t *testing.T) {
			workingDir := filepath.Join("testdata", testCase.Name())
			outDir := filepath.Join("testdata", testCase.Name(), runtime.DEV_DIRECTORY)
			runtime, err := runtime.NewRuntime(workingDir, outDir)
			require.NoError(t, err)

			_, err = runtime.Generate()
			require.NoError(t, err)

			typecheckResult := typecheck(workingDir)

			assert.True(t, typecheckResult)
		})
	}
}

func typecheck(workingDir string) bool {
	command := exec.Command("tsc", "-p", "tsconfig.json", "--noEmit")
	command.Dir = workingDir
	err := command.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == 0
		}
	}
	return true
}

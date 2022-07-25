package tsconfig_test

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions/runtime/tsconfig"
)

func TestCases(t *testing.T) {
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
		workingDir := filepath.Join("testdata", testCase.Name())
		tsconfigPath := filepath.Join(workingDir, "tsconfig.json")

		config, err := tsconfig.NewTSConfig(tsconfigPath)

		require.NoError(t, err)

		err = config.Reconcile()

		require.NoError(t, err)
	}
}

package testhelpers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

// WithTmpDir copies the contents of the src dir to a new temporary directory
// The second parameter is a function where you can run your tests against the new tmp directory
func WithTmpDir(dir string, testBody func(workingDir string)) (string, error) {
	base := filepath.Base(dir)

	tmpDir, err := ioutil.TempDir("", base)

	if err != nil {
		return "", err
	}

	err = cp.Copy(dir, tmpDir)

	if err != nil {
		return "", err
	}

	testBody(tmpDir)

	defer os.RemoveAll(tmpDir)

	return "", nil
}

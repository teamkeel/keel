package codegen

import (
	"fmt"
	"os"
)

// This function takes an array of test file paths, and will inject
// the shim that allows tests to be run automatically at the bottom of the file
func WrapTestFilesWithShim(parentPort string, testFiles []string) error {
	for _, f := range testFiles {
		file, err := os.OpenFile(f, os.O_APPEND|os.O_WRONLY, 0644)

		if err != nil {
			return err
		}

		defer file.Close()

		if _, err := file.WriteString(
			fmt.Sprintf(
				`
					import { runAllTests } from '@teamkeel/testing';

					runAllTests({ parentPort: %s })
				`,
				parentPort,
			),
		); err != nil {
			return err
		}
	}

	return nil
}

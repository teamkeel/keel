package testing

import (
	"fmt"
	"os"
)

// This function takes an array of test file paths, and will inject
// the shim that allows tests to be run automatically at the bottom of the file
func WrapTestFileWithShim(parentPort string, filePath string, pattern string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := file.WriteString(
		fmt.Sprintf(
			`
					import { runAllTests } from '@teamkeel/testing';

					runAllTests({ parentPort: %s, host: 'localhost', pattern: '%s' })
				`,
			parentPort,
			fmt.Sprintf("^%s$", pattern),
		),
	); err != nil {
		return err
	}
	return nil
}

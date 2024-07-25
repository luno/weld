package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luno/jettison/errors"
)

const (
	weldGenFileName     = "weld_gen.go"
	backendsGenFileName = "backends_gen.go"
	testingGenFileName  = "testing_gen.go"
)

// RemoveGenFiles removes previously generated files. During Generate these
// files can cause issues if they contain syntax or other compilation errors.
func RemoveGenFiles(workDir string) error {
	target := filepath.Join(workDir, weldGenFileName)
	err := os.Remove(target)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	target = filepath.Join(workDir, testingGenFileName)
	err = os.Remove(target)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	target = filepath.Join(workDir, backendsGenFileName)
	err = os.Remove(target)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// WriteGenFiles outputs the results from Generate to *_gen.go files.
func WriteGenFiles(res *Result, workDir string, verbose bool) error {
	if verbose {
		fmt.Println("Writing", weldGenFileName)
	}
	target := filepath.Join(workDir, weldGenFileName)
	err := os.WriteFile(target, res.WeldOutput, 0o644)
	if err != nil {
		return err
	}

	if len(res.TestingOutput) > 0 {
		target = filepath.Join(workDir, testingGenFileName)

		err := os.WriteFile(target, res.TestingOutput, 0o644)
		if err != nil {
			return err
		}

		if verbose {
			fmt.Println("Writing", testingGenFileName)
		}
	}

	if len(res.BackendsOutput) == 0 {
		return nil
	}

	if verbose {
		fmt.Println("Writing", backendsGenFileName)
	}
	target = filepath.Join(workDir, backendsGenFileName)
	return os.WriteFile(target, res.BackendsOutput, 0o644)
}

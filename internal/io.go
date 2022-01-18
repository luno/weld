package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	weldGenFileName     = "weld_gen.go"
	backendsGenFileName = "backends_gen.go"
)

// RemoveGenFiles removes previously generated files. During Generate these
// files can cause issues if they contain syntax or other compilation errors.
func RemoveGenFiles(workDir string) error {
	target := filepath.Join(workDir, weldGenFileName)
	err := os.Remove(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	target = filepath.Join(workDir, backendsGenFileName)
	err = os.Remove(target)
	if err != nil && !os.IsNotExist(err) {
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
	err := ioutil.WriteFile(target, res.WeldOutput, 0644)
	if err != nil {
		return err
	}

	if len(res.BackendsOutput) == 0 {
		return nil
	}

	if verbose {
		fmt.Println("Writing", backendsGenFileName)
	}
	target = filepath.Join(workDir, backendsGenFileName)
	return ioutil.WriteFile(target, res.BackendsOutput, 0644)
}

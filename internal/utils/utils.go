package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime/pprof"
)

func ReadDirRecursively(path string) ([]string, error) {
	result := []string{}

	visit := func(path string, f fs.DirEntry, err error) error {
		if !f.IsDir() {
			result = append(result, path)
		}
		return nil
	}
	err := filepath.WalkDir(path, visit)
	if err != nil {
		return result, err
	}
	return result, nil
}

func CPUProfilingFunc(fn func(), output string) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()
	err = pprof.StartCPUProfile(f)
	if err != nil {
		return err
	}
	fn()
	pprof.StopCPUProfile()
	return nil
}

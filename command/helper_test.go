package command

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// setEnv set enviromental variables and return restore function.
func setEnv(key, val string) func() {
	preVal := os.Getenv(key)
	os.Setenv(key, val)

	return func() {
		os.Setenv(key, preVal)
	}
}

func tmpChdir(dir string) (func(), error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}

	return func() {
		os.Chdir(currentDir)
	}, nil
}

func equalFiles(dir1, dir2 string) bool {
	list1 := make(map[string]struct{})
	list2 := make(map[string]struct{})

	list, dir := list1, dir1
	f := filepath.WalkFunc(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, dir)
		list[relPath] = struct{}{}

		return nil
	})

	if err := filepath.Walk(dir1, f); err != nil {
		return false
	}

	list, dir = list2, dir2
	if err := filepath.Walk(dir2, f); err != nil {
		return false
	}

	return reflect.DeepEqual(list1, list2)
}

package command

import "os"

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

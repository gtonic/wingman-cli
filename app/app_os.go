package app

import (
	"os"
)

func MustDir() string {
	dir, err := Dir()

	if err != nil {
		panic(err)
	}

	return dir
}

func Dir() (string, error) {
	return os.Getwd()
}

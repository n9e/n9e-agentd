package util

import (
	"os"
	"path/filepath"
)

type RootDir string

func NewRootDir(root string) RootDir {
	if root == "" {
		root, _ = os.Getwd()
	}
	return RootDir(root)
}

func (d RootDir) Abs(path string, defPaths ...string) string {
	if path == "" {
		path = filepath.Join(defPaths...)
	}

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(string(d), path)
}

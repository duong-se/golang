package tools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var WORKDIR string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WORKDIR = cwd
}

func SafePath(p string) (string, error) {
	base, err := filepath.Abs(WORKDIR)
	if err != nil {
		return "", err
	}

	target, err := filepath.Abs(filepath.Join(WORKDIR, p))
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}

	// nếu có ".." là escape
	if strings.HasPrefix(rel, "..") {
		return "", errors.New("path escapes workspace")
	}

	return target, nil
}

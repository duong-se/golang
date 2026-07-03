package tools

import (
	"errors"
	"path/filepath"
	"strings"
)

var WORKDIR = "/workspace"

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

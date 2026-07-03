package tools

import (
	"os"
	"path/filepath"
	"strings"
)

func RunGlob(pattern string) string {
	base, err := filepath.Abs(WORKDIR)
	if err != nil {
		return "Error: " + err.Error()
	}

	var results []string

	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(base, path)
		if err != nil {
			return nil
		}

		matched, err := filepath.Match(pattern, rel)
		if err != nil {
			return nil
		}

		if matched {
			results = append(results, rel)
		}

		return nil
	})

	if err != nil {
		return "Error: " + err.Error()
	}

	if len(results) == 0 {
		return "(no matches)"
	}

	return strings.Join(results, "\n")
}

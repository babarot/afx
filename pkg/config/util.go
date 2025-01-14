package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/b4b4r07/afx/pkg/errors"
)

// const errors
var (
	ErrPermission = errors.New("permission denied")
)

// isExecutable returns an error if a given file is not an executable.
// https://golang.org/src/os/executable_path.go
func isExecutable(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := stat.Mode()
	if !mode.IsRegular() {
		return ErrPermission
	}
	if (mode & 0111) == 0 {
		return ErrPermission
	}
	return nil
}

func allTrue(list []bool) bool {
	if len(list) == 0 {
		return false
	}
	for _, item := range list {
		if !item {
			return false
		}
	}
	return true
}

func expandTilda(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home := ""
	switch runtime.GOOS {
	case "windows":
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}
	default:
		home = os.Getenv("HOME")
	}

	if home == "" {
		return path
	}

	return home + path[1:]
}

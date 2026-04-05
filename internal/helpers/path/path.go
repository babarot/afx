package path

import (
	"os"
	"strings"
)

// ExpandTilda replaces a leading ~ in the path with the user's home directory.
func ExpandTilda(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home := os.Getenv("HOME")
	if home == "" {
		return path
	}

	return home + path[1:]
}

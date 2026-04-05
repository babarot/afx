package path

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ExpandTilda replaces a leading ~ in the path with the user's home directory.
func ExpandTilda(path string) string {
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

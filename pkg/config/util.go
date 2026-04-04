package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"github.com/babarot/afx/pkg/errors"
)

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

// getGitAuth returns BasicAuth for git operations.
// It first tries to get token from gh auth (gh CLI), then falls back to GITHUB_TOKEN.
// Returns nil if no token is found, allowing public repository access.
func getGitAuth() *githttp.BasicAuth {
	// Try gh auth first (uses ~/.config/gh/hosts.yml or system keyring)
	token, source := auth.TokenForHost("github.com")
	if token != "" {
		log.Printf("[DEBUG] using token from gh auth (source: %s)", source)
		return &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	// Fallback to GITHUB_TOKEN environment variable
	token = os.Getenv("GITHUB_TOKEN")
	if token != "" {
		log.Printf("[DEBUG] using token from GITHUB_TOKEN environment variable")
		return &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	return nil
}

// wrapAuthError wraps an error with a hint about setting GITHUB_TOKEN
// if the error appears to be an authentication failure.
func wrapAuthError(err error, name string) error {
	if err == nil {
		return nil
	}
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "authorization") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") {
		return errors.Wrapf(err, "%s: authentication failed. Please set GITHUB_TOKEN environment variable for private repositories", name)
	}
	return err
}

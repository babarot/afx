package pkg

import (
	"os"
	"path/filepath"
)

// DataDir returns the root directory for afx data (installed packages, state).
// Priority: $AFX_DATA_DIR > $XDG_DATA_HOME/afx > ~/.afx
func DataDir() string {
	if v := os.Getenv("AFX_DATA_DIR"); v != "" {
		return v
	}
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return filepath.Join(v, "afx")
	}
	return filepath.Join(os.Getenv("HOME"), ".afx")
}

// ConfigDir returns the root directory for afx configuration files.
// Priority: $AFX_CONFIG_DIR > $XDG_CONFIG_HOME/afx > ~/.config/afx
func ConfigDir() string {
	if v := os.Getenv("AFX_CONFIG_DIR"); v != "" {
		return v
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "afx")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "afx")
}

// BinDir returns the directory where command symlinks are created.
// Priority: $AFX_COMMAND_PATH > ~/bin
func BinDir() string {
	if v := os.Getenv("AFX_COMMAND_PATH"); v != "" {
		return v
	}
	return filepath.Join(os.Getenv("HOME"), "bin")
}

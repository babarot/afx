package pkg

import (
	"runtime"
	"testing"
)

func TestDataDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to HOME handling")
	}

	t.Run("default", func(t *testing.T) {
		t.Setenv("AFX_DATA_DIR", "")
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("HOME", "/home/user")
		got := DataDir()
		if got != "/home/user/.afx" {
			t.Errorf("DataDir() = %q, want %q", got, "/home/user/.afx")
		}
	})

	t.Run("AFX_DATA_DIR", func(t *testing.T) {
		t.Setenv("AFX_DATA_DIR", "/custom/data")
		got := DataDir()
		if got != "/custom/data" {
			t.Errorf("DataDir() = %q, want %q", got, "/custom/data")
		}
	})

	t.Run("XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("AFX_DATA_DIR", "")
		t.Setenv("XDG_DATA_HOME", "/xdg/data")
		got := DataDir()
		if got != "/xdg/data/afx" {
			t.Errorf("DataDir() = %q, want %q", got, "/xdg/data/afx")
		}
	})
}

func TestConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to HOME handling")
	}

	t.Run("default", func(t *testing.T) {
		t.Setenv("AFX_CONFIG_DIR", "")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "/home/user")
		got := ConfigDir()
		if got != "/home/user/.config/afx" {
			t.Errorf("ConfigDir() = %q, want %q", got, "/home/user/.config/afx")
		}
	})

	t.Run("AFX_CONFIG_DIR", func(t *testing.T) {
		t.Setenv("AFX_CONFIG_DIR", "/custom/config")
		got := ConfigDir()
		if got != "/custom/config" {
			t.Errorf("ConfigDir() = %q, want %q", got, "/custom/config")
		}
	})

	t.Run("XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("AFX_CONFIG_DIR", "")
		t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		got := ConfigDir()
		if got != "/xdg/config/afx" {
			t.Errorf("ConfigDir() = %q, want %q", got, "/xdg/config/afx")
		}
	})
}

func TestBinDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to HOME handling")
	}

	t.Run("default", func(t *testing.T) {
		t.Setenv("AFX_COMMAND_PATH", "")
		t.Setenv("HOME", "/home/user")
		got := BinDir()
		if got != "/home/user/bin" {
			t.Errorf("BinDir() = %q, want %q", got, "/home/user/bin")
		}
	})

	t.Run("AFX_COMMAND_PATH", func(t *testing.T) {
		t.Setenv("AFX_COMMAND_PATH", "/custom/bin")
		got := BinDir()
		if got != "/custom/bin" {
			t.Errorf("BinDir() = %q, want %q", got, "/custom/bin")
		}
	})
}

package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogOutput_disabled(t *testing.T) {
	t.Setenv("AFX_LOG", "")
	t.Setenv("AFX_LOG_PATH", "")

	out, err := LogOutput()
	if err != nil {
		t.Fatalf("LogOutput() error: %v", err)
	}

	// When logging is disabled, should return ioutil.Discard
	n, err := out.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if n != 4 {
		t.Errorf("Write() n = %d, want 4", n)
	}
}

func TestLogOutput_stderr(t *testing.T) {
	t.Setenv("AFX_LOG", "DEBUG")
	t.Setenv("AFX_LOG_PATH", "")

	out, err := LogOutput()
	if err != nil {
		t.Fatalf("LogOutput() error: %v", err)
	}
	if out == nil {
		t.Fatal("LogOutput() returned nil")
	}
}

func TestLogOutput_file(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	t.Setenv("AFX_LOG", "INFO")
	t.Setenv("AFX_LOG_PATH", logFile)

	out, err := LogOutput()
	if err != nil {
		t.Fatalf("LogOutput() error: %v", err)
	}
	if out == nil {
		t.Fatal("LogOutput() returned nil")
	}

	// Verify the log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("LogOutput() did not create log file")
	}
}

func TestLogOutput_invalidPath(t *testing.T) {
	t.Setenv("AFX_LOG", "DEBUG")
	t.Setenv("AFX_LOG_PATH", "/nonexistent/dir/logfile.log")

	_, err := LogOutput()
	if err == nil {
		t.Error("LogOutput() expected error for invalid path")
	}
}

func TestPrettyPrintJsonLines(t *testing.T) {
	tests := map[string]struct {
		input    string
		contains string
	}{
		"valid json": {
			input:    `{"key":"value"}`,
			contains: `"key"`,
		},
		"non-json": {
			input:    "plain text line",
			contains: "plain text line",
		},
		"mixed": {
			input:    "header\n{\"a\":1}\nfooter",
			contains: `"a"`,
		},
		"empty": {
			input:    "",
			contains: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := prettyPrintJsonLines([]byte(tt.input))
			if !strings.Contains(got, tt.contains) {
				t.Errorf("prettyPrintJsonLines() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestPrettyPrintJsonLines_indented(t *testing.T) {
	input := `{"key":"value","nested":{"a":1}}`
	got := prettyPrintJsonLines([]byte(input))

	// Pretty-printed JSON should contain newlines from indentation
	if !strings.Contains(got, "\n") {
		t.Error("prettyPrintJsonLines() should indent JSON with newlines")
	}
}

func TestNewTransport(t *testing.T) {
	tr := NewTransport("test", nil)
	if tr == nil {
		t.Fatal("NewTransport() returned nil")
	}
	if tr.name != "test" {
		t.Errorf("NewTransport() name = %q, want %q", tr.name, "test")
	}
}

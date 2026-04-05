package printers

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestGetNewTabWriter(t *testing.T) {
	var buf bytes.Buffer
	tw := GetNewTabWriter(&buf)
	if tw == nil {
		t.Fatal("GetNewTabWriter() returned nil")
	}

	fmt.Fprintln(tw, "NAME\tVERSION\tSTATUS")
	fmt.Fprintln(tw, "enhancd\tv2.0\tinstalled")
	fmt.Fprintln(tw, "jq\tv1.7\tinstalled")
	tw.Flush()

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Verify alignment: columns should be padded
	for _, line := range lines {
		if !strings.Contains(line, "  ") {
			t.Errorf("expected padded columns, got: %q", line)
		}
	}
}

func TestTerminalSize_nonFile(t *testing.T) {
	_, _, err := TerminalSize("not a file")
	if err == nil {
		t.Error("TerminalSize() expected error for non-file argument")
	}
}

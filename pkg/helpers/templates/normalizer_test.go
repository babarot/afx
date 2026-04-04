package templates

import (
	"strings"
	"testing"
)

func TestLongDesc(t *testing.T) {
	tests := map[string]struct {
		input string
		check func(string) bool
		desc  string
	}{
		"empty": {
			input: "",
			check: func(s string) bool { return s == "" },
			desc:  "should return empty",
		},
		"simple text": {
			input: "Hello world",
			check: func(s string) bool { return strings.Contains(s, "Hello world") },
			desc:  "should contain original text",
		},
		"strips leading/trailing whitespace": {
			input: "  hello  ",
			check: func(s string) bool { return !strings.HasPrefix(s, " ") && !strings.HasSuffix(s, " ") },
			desc:  "should be trimmed",
		},
		"heredoc formatting": {
			input: `
				This is a test
				of heredoc formatting`,
			check: func(s string) bool { return strings.Contains(s, "This is a test") },
			desc:  "should contain text after heredoc processing",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := LongDesc(tt.input)
			if !tt.check(got) {
				t.Errorf("LongDesc(): %s, got %q", tt.desc, got)
			}
		})
	}
}

func TestExamples(t *testing.T) {
	tests := map[string]struct {
		input    string
		contains string
	}{
		"empty": {
			input:    "",
			contains: "",
		},
		"single line": {
			input:    "afx install",
			contains: Indentation + "afx install",
		},
		"multiline": {
			input:    "afx install\nafx update",
			contains: Indentation + "afx install",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Examples(tt.input)
			if tt.input == "" {
				if got != "" {
					t.Errorf("Examples('') = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Errorf("Examples() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestExamples_allLinesIndented(t *testing.T) {
	input := "line1\nline2\nline3"
	got := Examples(input)
	for line := range strings.SplitSeq(got, "\n") {
		if !strings.HasPrefix(line, Indentation) {
			t.Errorf("line %q should start with indentation", line)
		}
	}
}

func TestRaw(t *testing.T) {
	tests := map[string]struct {
		input string
		check func(string) bool
		desc  string
	}{
		"empty": {
			input: "",
			check: func(s string) bool { return s == "" },
			desc:  "should return empty",
		},
		"adds indentation": {
			input: "hello",
			check: func(s string) bool { return strings.Contains(s, Indentation+"hello") },
			desc:  "should have indentation",
		},
		"heredoc processing": {
			input: `
				indented text`,
			check: func(s string) bool { return strings.Contains(s, "indented text") },
			desc:  "should contain text after heredoc",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Raw(tt.input)
			if !tt.check(got) {
				t.Errorf("Raw(): %s, got %q", tt.desc, got)
			}
		})
	}
}

func TestNormalizer_trim(t *testing.T) {
	n := normalizer{"  hello  "}
	got := n.trim()
	if got.string != "hello" {
		t.Errorf("trim() = %q, want 'hello'", got.string)
	}
}

func TestNormalizer_indent(t *testing.T) {
	n := normalizer{"line1\nline2"}
	got := n.indent()
	for line := range strings.SplitSeq(got.string, "\n") {
		if !strings.HasPrefix(line, Indentation) {
			t.Errorf("indent(): line %q should start with indentation", line)
		}
	}
}

func TestNormalizer_space(t *testing.T) {
	n := normalizer{"line1\nline2"}
	got := n.space()
	for line := range strings.SplitSeq(got.string, "\n") {
		if !strings.HasPrefix(line, Indentation) {
			t.Errorf("space(): line %q should start with indentation", line)
		}
	}
}

func TestNormalizer_heredoc(t *testing.T) {
	// heredoc normalizes tab-indented text
	input := "\tline1\n\tline2"
	n := normalizer{input}
	got := n.heredoc()
	// Should produce some output containing "line1"
	if !strings.Contains(got.string, "line1") {
		t.Errorf("heredoc() should contain 'line1', got %q", got.string)
	}
}

func TestNormalizer_markdown(t *testing.T) {
	n := normalizer{"**bold** text"}
	got := n.markdown()
	// Markdown renderer should process bold
	if !strings.Contains(got.string, "bold") {
		t.Errorf("markdown() should contain 'bold', got %q", got.string)
	}
}

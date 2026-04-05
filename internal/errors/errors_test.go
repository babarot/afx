package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestErrors_Error(t *testing.T) {
	tests := map[string]struct {
		errs     Errors
		contains []string
	}{
		"single error": {
			errs:     Errors{fmt.Errorf("something failed")},
			contains: []string{"1 error occurred", "something failed"},
		},
		"multiple errors": {
			errs:     Errors{fmt.Errorf("err1"), fmt.Errorf("err2")},
			contains: []string{"2 errors occurred", "err1", "err2"},
		},
		"single nil error": {
			errs:     Errors{nil},
			contains: []string{""},
		},
		"multiline error": {
			errs:     Errors{fmt.Errorf("line1\nline2\nline3")},
			contains: []string{"1 error occurred", "line1", "\n\t  line2", "\n\t  line3"},
		},
		"multiple with nil": {
			errs:     Errors{fmt.Errorf("err1"), nil, fmt.Errorf("err2")},
			contains: []string{"2 errors occurred", "err1", "err2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.errs.Error()
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("Error() = %q, want to contain %q", got, s)
				}
			}
		})
	}
}

func TestErrors_Append(t *testing.T) {
	tests := map[string]struct {
		input   []error
		wantLen int
	}{
		"one error": {
			input:   []error{fmt.Errorf("err")},
			wantLen: 1,
		},
		"nil filtered": {
			input:   []error{nil},
			wantLen: 0,
		},
		"mixed": {
			input:   []error{fmt.Errorf("a"), nil, fmt.Errorf("b")},
			wantLen: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var e Errors
			e.Append(tt.input...)
			if len(e) != tt.wantLen {
				t.Errorf("Append() len = %d, want %d", len(e), tt.wantLen)
			}
		})
	}
}

func TestErrors_ErrorOrNil(t *testing.T) {
	tests := map[string]struct {
		errs    *Errors
		wantNil bool
	}{
		"nil receiver": {
			errs:    nil,
			wantNil: true,
		},
		"empty slice": {
			errs:    &Errors{},
			wantNil: true,
		},
		"has errors": {
			errs:    &Errors{fmt.Errorf("err")},
			wantNil: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.errs.ErrorOrNil()
			if tt.wantNil && got != nil {
				t.Errorf("ErrorOrNil() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("ErrorOrNil() = nil, want non-nil")
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := map[string]struct {
		messages []string
		wantNil  bool
		contains string
	}{
		"no messages": {
			messages: nil,
			wantNil:  true,
		},
		"one message": {
			messages: []string{"fail"},
			wantNil:  false,
			contains: "fail",
		},
		"multiple messages": {
			messages: []string{"a", "b"},
			wantNil:  false,
			contains: "2 errors occurred",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := New(tt.messages...)
			if tt.wantNil {
				if got != nil {
					t.Errorf("New() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("New() = nil, want non-nil")
			}
			if !strings.Contains(got.Error(), tt.contains) {
				t.Errorf("New().Error() = %q, want to contain %q", got.Error(), tt.contains)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	orig := fmt.Errorf("original")
	got := Wrap(orig, "context")
	if got == nil {
		t.Fatal("Wrap() = nil, want non-nil")
	}
	if !strings.Contains(got.Error(), "context") || !strings.Contains(got.Error(), "original") {
		t.Errorf("Wrap() = %q, want to contain both 'context' and 'original'", got.Error())
	}
}

func TestWrapf(t *testing.T) {
	orig := fmt.Errorf("original")
	got := Wrapf(orig, "ctx %d", 42)
	if got == nil {
		t.Fatal("Wrapf() = nil, want non-nil")
	}
	if !strings.Contains(got.Error(), "ctx 42") || !strings.Contains(got.Error(), "original") {
		t.Errorf("Wrapf() = %q, want to contain 'ctx 42' and 'original'", got.Error())
	}
}

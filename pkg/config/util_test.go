package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestAllTrue(t *testing.T) {
	tests := map[string]struct {
		input []bool
		want  bool
	}{
		"all true": {
			input: []bool{true, true, true},
			want:  true,
		},
		"one false": {
			input: []bool{true, false, true},
			want:  false,
		},
		"empty": {
			input: []bool{},
			want:  false,
		},
		"single true": {
			input: []bool{true},
			want:  true,
		},
		"single false": {
			input: []bool{false},
			want:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := allTrue(tt.input)
			if got != tt.want {
				t.Errorf("allTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapAuthError(t *testing.T) {
	tests := map[string]struct {
		err      error
		name     string
		wantNil  bool
		contains string
	}{
		"nil error": {
			err:     nil,
			name:    "pkg",
			wantNil: true,
		},
		"401 error": {
			err:      fmt.Errorf("got 401 from server"),
			name:     "mypkg",
			contains: "authentication failed",
		},
		"403 error": {
			err:      fmt.Errorf("403 forbidden"),
			name:     "mypkg",
			contains: "authentication failed",
		},
		"authentication keyword": {
			err:      fmt.Errorf("Authentication required"),
			name:     "mypkg",
			contains: "authentication failed",
		},
		"authorization keyword": {
			err:      fmt.Errorf("authorization denied"),
			name:     "mypkg",
			contains: "authentication failed",
		},
		"non-auth error": {
			err:      fmt.Errorf("connection timeout"),
			name:     "mypkg",
			contains: "connection timeout",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := wrapAuthError(tt.err, tt.name)
			if tt.wantNil {
				if got != nil {
					t.Errorf("wrapAuthError() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("wrapAuthError() = nil, want non-nil")
			}
			if !strings.Contains(got.Error(), tt.contains) {
				t.Errorf("wrapAuthError().Error() = %q, want to contain %q", got.Error(), tt.contains)
			}
		})
	}
}

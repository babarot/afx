package path

import (
	"testing"
)

func TestExpandTilda(t *testing.T) {

	tests := map[string]struct {
		input string
		home  string
		want  string
	}{
		"with tilde": {
			input: "~/foo",
			home:  "/home/user",
			want:  "/home/user/foo",
		},
		"absolute path": {
			input: "/absolute/path",
			home:  "/home/user",
			want:  "/absolute/path",
		},
		"tilde only": {
			input: "~",
			home:  "/home/user",
			want:  "/home/user",
		},
		"empty": {
			input: "",
			home:  "/home/user",
			want:  "",
		},
		"empty home": {
			input: "~/foo",
			home:  "",
			want:  "~/foo",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("HOME", tt.home)
			got := ExpandTilda(tt.input)
			if got != tt.want {
				t.Errorf("ExpandTilda(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

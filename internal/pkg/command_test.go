package pkg

import "testing"

func TestCommand_buildRequired(t *testing.T) {
	tests := map[string]struct {
		cmd  Command
		want bool
	}{
		"nil build": {
			cmd:  Command{Build: nil},
			want: false,
		},
		"empty steps": {
			cmd:  Command{Build: &Build{Steps: []string{}}},
			want: false,
		},
		"has steps": {
			cmd:  Command{Build: &Build{Steps: []string{"make"}}},
			want: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.cmd.buildRequired()
			if got != tt.want {
				t.Errorf("buildRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

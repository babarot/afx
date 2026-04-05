package manager

import (
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

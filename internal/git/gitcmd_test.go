package git

import (
	"fmt"
	"testing"
)

func TestIsAuthError(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"nil error": {
			err:  nil,
			want: false,
		},
		"401 error": {
			err:  fmt.Errorf("got 401 from server"),
			want: true,
		},
		"403 error": {
			err:  fmt.Errorf("403 forbidden"),
			want: true,
		},
		"authentication keyword": {
			err:  fmt.Errorf("Authentication required"),
			want: true,
		},
		"could not read username": {
			err:  fmt.Errorf("could not read Username"),
			want: true,
		},
		"terminal prompts disabled": {
			err:  fmt.Errorf("terminal prompts disabled"),
			want: true,
		},
		"non-auth error": {
			err:  fmt.Errorf("connection timeout"),
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsAuthError(tt.err)
			if got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

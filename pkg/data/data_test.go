package data

import (
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToEnv(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  Env
	}{
		"standard pairs": {
			input: []string{"A=1", "B=2"},
			want:  Env{"A": "1", "B": "2"},
		},
		"empty input": {
			input: []string{},
			want:  Env{},
		},
		"value with equals": {
			input: []string{"A=1=2"},
			want:  Env{"A": "1=2"},
		},
		"empty key": {
			input: []string{"=val"},
			want:  Env{},
		},
		"no equals": {
			input: []string{"NOEQUALS"},
			want:  Env{},
		},
		"empty value": {
			input: []string{"KEY="},
			want:  Env{"KEY": ""},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ToEnv(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToEnv() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type mockPkg struct {
	name string
	home string
}

func (m mockPkg) GetHome() string { return m.home }
func (m mockPkg) GetName() string { return m.name }

func TestWithPackage(t *testing.T) {
	d := &Data{}
	opt := WithPackage(mockPkg{name: "test-pkg", home: "/home/test"})
	opt(d)

	want := Package{Name: "test-pkg", Home: "/home/test"}
	if diff := cmp.Diff(want, d.Package); diff != "" {
		t.Errorf("WithPackage() mismatch (-want +got):\n%s", diff)
	}
}

func TestWithRelease(t *testing.T) {
	d := &Data{}
	opt := WithRelease(Release{Name: "v1.0.0", Tag: "v1.0.0"})
	opt(d)

	want := Release{Name: "v1.0.0", Tag: "v1.0.0"}
	if diff := cmp.Diff(want, d.Release); diff != "" {
		t.Errorf("WithRelease() mismatch (-want +got):\n%s", diff)
	}
}

func TestNew(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		d := New()
		if d.Runtime.Goos != runtime.GOOS {
			t.Errorf("Goos = %q, want %q", d.Runtime.Goos, runtime.GOOS)
		}
		if d.Runtime.Goarch != runtime.GOARCH {
			t.Errorf("Goarch = %q, want %q", d.Runtime.Goarch, runtime.GOARCH)
		}
		if d.Env == nil {
			t.Error("Env should not be nil")
		}
	})

	t.Run("with options", func(t *testing.T) {
		r := Release{Name: "myrel", Tag: "v2.0"}
		d := New(WithRelease(r))
		if diff := cmp.Diff(r, d.Release); diff != "" {
			t.Errorf("New(WithRelease) mismatch (-want +got):\n%s", diff)
		}
	})
}

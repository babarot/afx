package templates

import (
	"strings"
	"testing"

	"github.com/babarot/afx/pkg/data"
)

func testData() *data.Data {
	return &data.Data{
		Env: data.Env{"HOME": "/home/user", "GOPATH": "/go"},
		Runtime: data.Runtime{
			Goos:   "darwin",
			Goarch: "amd64",
		},
		Package: data.Package{
			Name: "test-pkg",
			Home: "/home/user/.afx/test-pkg",
		},
		Release: data.Release{
			Name: "v1.0.0",
			Tag:  "v1.0.0",
		},
	}
}

func TestTemplate_Apply(t *testing.T) {
	tmpl := New(testData())

	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple field Name": {
			input: "{{ .Name }}",
			want:  "test-pkg",
		},
		"field OS": {
			input: "{{ .OS }}",
			want:  "darwin",
		},
		"field Arch": {
			input: "{{ .Arch }}",
			want:  "amd64",
		},
		"field Home": {
			input: "{{ .Home }}",
			want:  "/home/user/.afx/test-pkg",
		},
		"field Dir": {
			input: "{{ .Dir }}",
			want:  "/home/user/.afx/test-pkg",
		},
		"env access": {
			input: `{{ index .Env "HOME" }}`,
			want:  "/home/user",
		},
		"release tag": {
			input: `{{ index .Release "Tag" }}`,
			want:  "v1.0.0",
		},
		"replace func": {
			input: `{{ replace .OS "darwin" "mac" }}`,
			want:  "mac",
		},
		"tolower func": {
			input: `{{ tolower "DARWIN" }}`,
			want:  "darwin",
		},
		"toupper func": {
			input: `{{ toupper "hello" }}`,
			want:  "HELLO",
		},
		"trim func": {
			input: `{{ trim "  hello  " }}`,
			want:  "hello",
		},
		"trimprefix func": {
			input: `{{ trimprefix "v1.0.0" "v" }}`,
			want:  "1.0.0",
		},
		"trimsuffix func": {
			input: `{{ trimsuffix "file.tar.gz" ".tar.gz" }}`,
			want:  "file",
		},
		"dir func": {
			input: `{{ dir .Home }}`,
			want:  "/home/user/.afx",
		},
		"time func does not error": {
			input: `{{ time "2006" }}`,
		},
		"invalid template": {
			input:   `{{ }`,
			wantErr: true,
		},
		"missing key": {
			input:   `{{ .NonExistent }}`,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tmpl.Apply(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("Apply() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Apply() unexpected error: %v", err)
			}
			if tt.want != "" && got != tt.want {
				t.Errorf("Apply() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplate_Replace(t *testing.T) {
	tests := map[string]struct {
		replacements map[string]string
		wantOS       string
		wantArch     string
	}{
		"match os": {
			replacements: map[string]string{"darwin": "mac"},
			wantOS:       "mac",
			wantArch:     "amd64",
		},
		"match arch": {
			replacements: map[string]string{"amd64": "x86_64"},
			wantOS:       "darwin",
			wantArch:     "x86_64",
		},
		"no match": {
			replacements: map[string]string{"linux": "lin"},
			wantOS:       "darwin",
			wantArch:     "amd64",
		},
		"empty map": {
			replacements: map[string]string{},
			wantOS:       "darwin",
			wantArch:     "amd64",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tmpl := New(testData())
			tmpl.Replace(tt.replacements)

			got, err := tmpl.Apply("{{ .OS }}")
			if err != nil {
				t.Fatalf("Apply(OS) error: %v", err)
			}
			if got != tt.wantOS {
				t.Errorf("OS = %q, want %q", got, tt.wantOS)
			}

			got, err = tmpl.Apply("{{ .Arch }}")
			if err != nil {
				t.Fatalf("Apply(Arch) error: %v", err)
			}
			if got != tt.wantArch {
				t.Errorf("Arch = %q, want %q", got, tt.wantArch)
			}
		})
	}
}

func Test_replace(t *testing.T) {
	tests := map[string]struct {
		replacements map[string]string
		original     string
		want         string
	}{
		"found": {
			replacements: map[string]string{"a": "b"},
			original:     "a",
			want:         "b",
		},
		"not found": {
			replacements: map[string]string{"a": "b"},
			original:     "c",
			want:         "c",
		},
		"empty value returns original": {
			replacements: map[string]string{"a": ""},
			original:     "a",
			want:         "a",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := replace(tt.replacements, tt.original)
			if got != tt.want {
				t.Errorf("replace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplate_Apply_combined(t *testing.T) {
	tmpl := New(testData())
	got, err := tmpl.Apply(`{{ tolower (replace .OS "darwin" "MACOS") }}`)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}
	if !strings.Contains(got, "macos") {
		t.Errorf("Apply() = %q, want 'macos'", got)
	}
}

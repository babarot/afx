package state

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func init() {
	log.SetOutput(ioutil.Discard)

	exists = func(path string) bool {
		// always returns true in testing
		return true
	}
}

func TestOpen(t *testing.T) {
	stubState(map[string]string{
		"empty.json": "{}",
		"state.json": `
{
  "resources": {
    "github.com/b4b4r07/enhancd": {
      "id": "github.com/b4b4r07/enhancd",
      "name": "b4b4r07/enhancd",
      "home": "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
      "type": "GitHub",
      "version": "",
      "paths": [
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh"
      ]
    }
  }
}
`})

	testCases := map[string]struct {
		filename string
		state    *State
	}{
		"Empty": {
			filename: "empty.json",
			state:    &State{Self: Self{Resources: map[string]Resource{}}, path: "empty.json"},
		},
		"Open": {
			filename: "state.json",
			state: &State{
				path:     "state.json",
				packages: nil,
				Self: Self{
					Resources: map[ID]Resource{
						"github.com/b4b4r07/enhancd": {
							ID:      "github.com/b4b4r07/enhancd",
							Name:    "b4b4r07/enhancd",
							Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
							Type:    "GitHub",
							Version: "",
							Paths: []string{
								"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
								"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
							},
						},
					},
				},
				Additions: nil,
				Deletions: []Resource{
					{
						ID:      "github.com/b4b4r07/enhancd",
						Name:    "b4b4r07/enhancd",
						Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						Type:    "GitHub",
						Version: "",
						Paths: []string{
							"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
							"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
						},
					},
				},
				NoChanges: nil,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, nil)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.state
			got := state
			if diff := cmp.Diff(
				want, got,
				cmp.AllowUnexported(State{}),
				cmpopts.IgnoreUnexported(State{}),
			); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestList(t *testing.T) {
	stubState(map[string]string{
		"state.json": `
{
  "resources": {
    "github.com/b4b4r07/enhancd": {
      "id": "github.com/b4b4r07/enhancd",
      "name": "b4b4r07/enhancd",
      "home": "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
      "type": "GitHub",
      "version": "",
      "paths": [
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh"
      ]
    }
  }
}
`})

	testCases := map[string]struct {
		filename  string
		resources []Resource
	}{
		"List": {
			filename: "state.json",
			resources: []Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, nil)
			if err != nil {
				t.Fatal(err)
			}
			resources, err := state.List()
			if err != nil {
				t.Fatal(err)
			}
			want := tc.resources
			got := resources
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_listChanges(t *testing.T) {
	stubState(map[string]string{
		"state.json": `
{
  "resources": {
    "github.com/release/stedolan/jq": {
      "id": "github.com/release/stedolan/jq",
      "name": "stedolan/jq",
      "home": "/Users/babarot/.afx/github.com/stedolan/jq",
      "type": "GitHub Release",
      "version": "jq-1.6",
      "paths": [
        "/Users/babarot/.afx/github.com/stedolan/jq",
        "/Users/babarot/.afx/github.com/stedolan/jq/jq",
        "/Users/babarot/bin/jq"
      ]
    }
  }
}
`})

	testCases := map[string]struct {
		filename  string
		pkgs      []Resourcer
		resources []Resource
	}{
		"Found": {
			filename: "state.json",
			pkgs: stubPackages([]Resource{
				{
					ID:      "github.com/release/stedolan/jq",
					Name:    "stedolan/jq",
					Home:    "/Users/babarot/.afx/github.com/stedolan/jq",
					Type:    "GitHub Release",
					Version: "jq-1.7",
					Paths: []string{
						"/Users/babarot/.afx/github.com/stedolan/jq",
						"/Users/babarot/.afx/github.com/stedolan/jq/jq",
						"/Users/babarot/bin/jq",
					},
				},
			}),
			resources: []Resource{
				{
					ID:      "github.com/release/stedolan/jq",
					Name:    "stedolan/jq",
					Home:    "/Users/babarot/.afx/github.com/stedolan/jq",
					Type:    "GitHub Release",
					Version: "jq-1.6",
					Paths: []string{
						"/Users/babarot/.afx/github.com/stedolan/jq",
						"/Users/babarot/.afx/github.com/stedolan/jq/jq",
						"/Users/babarot/bin/jq",
					},
				},
			},
		},
		"NotFound": {
			filename: "state.json",
			pkgs: stubPackages([]Resource{
				{
					ID:      "github.com/release/stedolan/jq",
					Name:    "stedolan/jq",
					Home:    "/Users/babarot/.afx/github.com/stedolan/jq",
					Type:    "GitHub Release",
					Version: "jq-1.6",
					Paths: []string{
						"/Users/babarot/.afx/github.com/stedolan/jq",
						"/Users/babarot/.afx/github.com/stedolan/jq/jq",
						"/Users/babarot/bin/jq",
					},
				},
			}),
			resources: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.resources
			got := state.Changes
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_listNoChanges(t *testing.T) {
	stubState(map[string]string{
		"state.json": `
{
  "resources": {
    "github.com/b4b4r07/enhancd": {
      "id": "github.com/b4b4r07/enhancd",
      "name": "b4b4r07/enhancd",
      "home": "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
      "type": "GitHub",
      "version": "",
      "paths": [
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh"
      ]
    }
  }
}
`})

	testCases := map[string]struct {
		filename  string
		pkgs      []Resourcer
		resources []Resource
	}{
		"Found": {
			filename: "state.json",
			pkgs: stubPackages([]Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			}),
			resources: []Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			},
		},
		"NotFound": {
			filename: "state.json",
			pkgs: stubPackages([]Resource{
				{
					ID:      "github.com/release/stedolan/jq",
					Name:    "stedolan/jq",
					Home:    "/Users/babarot/.afx/github.com/stedolan/jq",
					Type:    "GitHub Release",
					Version: "jq-1.7",
					Paths: []string{
						"/Users/babarot/.afx/github.com/stedolan/jq",
						"/Users/babarot/.afx/github.com/stedolan/jq/jq",
						"/Users/babarot/bin/jq",
					},
				},
			}),
			resources: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.resources
			got := state.NoChanges
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_listAdditions(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{"resources":{}}`,
	})

	testCases := map[string]struct {
		filename  string
		pkgs      []Resourcer
		resources []Resource
	}{
		"Found": {
			filename: "state.json",
			pkgs: stubPackages([]Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			}),
			resources: []Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			},
		},
		"NotFound": {
			filename:  "state.json",
			pkgs:      stubPackages([]Resource{}),
			resources: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.resources
			got := state.Additions
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_listReadditions(t *testing.T) {
	// do not test

	// it's almost same as listAdditions but the difference is
	// to check file existence
}

func Test_listDeletions(t *testing.T) {
	stubState(map[string]string{
		"state.json": `
{
  "resources": {
    "github.com/b4b4r07/enhancd": {
      "id": "github.com/b4b4r07/enhancd",
      "name": "b4b4r07/enhancd",
      "home": "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
      "type": "GitHub",
      "version": "",
      "paths": [
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
        "/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh"
      ]
    }
  }
}
`})

	testCases := map[string]struct {
		filename  string
		pkgs      []Resourcer
		resources []Resource
	}{
		"Found": {
			filename: "state.json",
			pkgs:     stubPackages([]Resource{}),
			resources: []Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			},
		},
		"NotFound": {
			filename: "state.json",
			pkgs: stubPackages([]Resource{
				{
					ID:      "github.com/release/stedolan/jq",
					Name:    "stedolan/jq",
					Home:    "/Users/babarot/.afx/github.com/stedolan/jq",
					Type:    "GitHub Release",
					Version: "jq-1.7",
					Paths: []string{
						"/Users/babarot/.afx/github.com/stedolan/jq",
						"/Users/babarot/.afx/github.com/stedolan/jq/jq",
						"/Users/babarot/bin/jq",
					},
				},
			}),
			resources: []Resource{
				{
					ID:      "github.com/b4b4r07/enhancd",
					Name:    "b4b4r07/enhancd",
					Home:    "/Users/babarot/.afx/github.com/b4b4r07/enhancd",
					Type:    "GitHub",
					Version: "",
					Paths: []string{
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd",
						"/Users/babarot/.afx/github.com/b4b4r07/enhancd/init.sh",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.resources
			got := state.Deletions
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

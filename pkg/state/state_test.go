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
}

func TestOpen(t *testing.T) {
	stubState(map[string]string{
		"empty.yaml": "{}",
		"open-test.yaml": `
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
			filename: "empty.yaml",
			state:    &State{path: "empty.yaml"},
		},
		"Open": {
			filename: "open-test.yaml",
			state: &State{
				path:     "open-test.yaml",
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
			if diff := cmp.Diff(state, tc.state, cmp.AllowUnexported(State{}), cmpopts.IgnoreUnexported(State{})); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestList(t *testing.T) {
	stubState(map[string]string{
		"list-test.yaml": `
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
			filename: "list-test.yaml",
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
			if diff := cmp.Diff(resources, tc.resources); diff != "" {
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
		filename string
		pkgs     []Resourcer
		want     []Resource
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
			want: []Resource{
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
			want: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			got := state.Changes
			if diff := cmp.Diff(got, tc.want); diff != "" {
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
		filename string
		pkgs     []Resourcer
		want     []Resource
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
			want: []Resource{
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
			want: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			got := state.NoChanges
			if diff := cmp.Diff(got, tc.want); diff != "" {
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
		filename string
		pkgs     []Resourcer
		want     []Resource
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
			want: []Resource{
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
			pkgs:     stubPackages([]Resource{}),
			want:     nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, err := Open(tc.filename, tc.pkgs)
			if err != nil {
				t.Fatal(err)
			}
			got := state.Additions
			if diff := cmp.Diff(got, tc.want); diff != "" {
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
		filename string
		pkgs     []Resourcer
		want     []Resource
	}{
		"Found": {
			filename: "state.json",
			pkgs:     stubPackages([]Resource{}),
			want: []Resource{
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
			want: []Resource{
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
			got := state.Deletions
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Compare value is mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

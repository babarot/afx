package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestState_Add(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{"resources":{}}`,
	})

	state, err := Open("state.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := Resource{
		ID:   "github.com/babarot/enhancd",
		Name: "babarot/enhancd",
		Home: "/home/.afx/github.com/babarot/enhancd",
		Type: "GitHub",
	}
	state.Add(testPackage{r: r})

	got, ok := state.Resources[r.ID]
	if !ok {
		t.Fatal("Add() did not add the resource to state")
	}
	if diff := cmp.Diff(r, got); diff != "" {
		t.Errorf("Add() resource mismatch (-want +got):\n%s", diff)
	}
}

func TestState_Remove(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{
  "resources": {
    "github.com/babarot/enhancd": {
      "id": "github.com/babarot/enhancd",
      "name": "babarot/enhancd",
      "home": "/home/.afx/github.com/babarot/enhancd",
      "type": "GitHub",
      "version": "",
      "paths": []
    }
  }
}`,
	})

	state, err := Open("state.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := Resource{
		ID:   "github.com/babarot/enhancd",
		Name: "babarot/enhancd",
	}
	state.Remove(testPackage{r: r})

	if _, ok := state.Resources[r.ID]; ok {
		t.Error("Remove() did not remove the resource from state")
	}
	if len(state.Resources) != 0 {
		t.Errorf("Remove() left %d resources, want 0", len(state.Resources))
	}
}

func TestState_Update(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{
  "resources": {
    "github.com/release/stedolan/jq": {
      "id": "github.com/release/stedolan/jq",
      "name": "stedolan/jq",
      "home": "/home/.afx/github.com/stedolan/jq",
      "type": "GitHub Release",
      "version": "jq-1.6",
      "paths": []
    }
  }
}`,
	})

	state, err := Open("state.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	updated := Resource{
		ID:      "github.com/release/stedolan/jq",
		Name:    "stedolan/jq",
		Home:    "/home/.afx/github.com/stedolan/jq",
		Type:    "GitHub Release",
		Version: "jq-1.7",
	}
	state.Update(testPackage{r: updated})

	got := state.Resources[updated.ID]
	if got.Version != "jq-1.7" {
		t.Errorf("Update() version = %q, want %q", got.Version, "jq-1.7")
	}
}

func TestState_Update_nonexistent(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{"resources":{}}`,
	})

	state, err := Open("state.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := Resource{
		ID:   "github.com/nonexistent/pkg",
		Name: "nonexistent/pkg",
	}
	state.Update(testPackage{r: r})

	if _, ok := state.Resources[r.ID]; ok {
		t.Error("Update() should not add non-existent resource")
	}
}

func TestState_Get(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{
  "resources": {
    "github.com/babarot/enhancd": {
      "id": "github.com/babarot/enhancd",
      "name": "babarot/enhancd",
      "home": "/home/.afx/github.com/babarot/enhancd",
      "type": "GitHub",
      "version": "",
      "paths": []
    }
  }
}`,
	})

	state, err := Open("state.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := state.Get("babarot/enhancd")
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got.Name != "babarot/enhancd" {
			t.Errorf("Get() name = %q, want %q", got.Name, "babarot/enhancd")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := state.Get("nonexistent/pkg")
		if err == nil {
			t.Error("Get() expected error for non-existent resource")
		}
	})
}

func TestState_New(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{"resources":{}}`,
	})

	pkgs := stubPackages([]Resource{
		{
			ID:   "github.com/babarot/enhancd",
			Name: "babarot/enhancd",
			Type: "GitHub",
		},
		{
			ID:   "github.com/release/stedolan/jq",
			Name: "stedolan/jq",
			Type: "GitHub Release",
		},
	})

	state, err := Open("state.json", pkgs)
	if err != nil {
		t.Fatal(err)
	}

	err = state.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if len(state.Resources) != 2 {
		t.Errorf("New() resources count = %d, want 2", len(state.Resources))
	}
}

func TestState_Refresh(t *testing.T) {
	t.Run("no changes updates schema", func(t *testing.T) {
		stubState(map[string]string{
			"state.json": `{"resources":{}}`,
		})

		// Build state manually to test Refresh in isolation
		r := Resource{
			ID:   "github.com/babarot/enhancd",
			Name: "babarot/enhancd",
			Home: "/home/.afx/github.com/babarot/enhancd",
			Type: "GitHub",
		}
		updated := Resource{
			ID:    "github.com/babarot/enhancd",
			Name:  "babarot/enhancd",
			Home:  "/home/.afx/github.com/babarot/enhancd",
			Type:  "GitHub",
			Paths: []string{"/home/.afx/github.com/babarot/enhancd"},
		}
		state := &State{
			Self: Self{Resources: map[ID]Resource{
				r.ID: r,
			}},
			packages: map[ID]Resource{
				updated.ID: updated,
			},
			path: "state.json",
		}

		err := state.Refresh()
		if err != nil {
			t.Fatalf("Refresh() error: %v", err)
		}

		got := state.Resources["github.com/babarot/enhancd"]
		if diff := cmp.Diff(
			[]string{"/home/.afx/github.com/babarot/enhancd"},
			got.Paths,
		); diff != "" {
			t.Errorf("Refresh() did not update paths (-want +got):\n%s", diff)
		}
	})

	t.Run("with changes returns error", func(t *testing.T) {
		stubState(map[string]string{
			"state.json": `{"resources":{}}`,
		})

		pkgs := stubPackages([]Resource{
			{
				ID:   "github.com/new/pkg",
				Name: "new/pkg",
				Type: "GitHub",
			},
		})

		state := &State{
			Self:     Self{Resources: map[ID]Resource{}},
			packages: map[ID]Resource{"github.com/new/pkg": pkgs[0].GetResource()},
		}
		state.Additions = []Resource{pkgs[0].GetResource()}

		err := state.Refresh()
		if err == nil {
			t.Error("Refresh() expected error when there are additions")
		}
	})
}

func TestKeys(t *testing.T) {
	resources := []Resource{
		{Name: "babarot/enhancd"},
		{Name: "stedolan/jq"},
		{Name: "junegunn/fzf"},
	}

	got := Keys(resources)
	want := []string{"babarot/enhancd", "stedolan/jq", "junegunn/fzf"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Keys() mismatch (-want +got):\n%s", diff)
	}
}

func TestKeys_empty(t *testing.T) {
	got := Keys(nil)
	if got != nil {
		t.Errorf("Keys(nil) = %v, want nil", got)
	}
}

func TestMap(t *testing.T) {
	resources := []Resource{
		{ID: "id1", Name: "babarot/enhancd"},
		{ID: "id2", Name: "stedolan/jq"},
	}

	got := Map(resources)
	if len(got) != 2 {
		t.Fatalf("Map() len = %d, want 2", len(got))
	}
	if got["babarot/enhancd"].ID != "id1" {
		t.Error("Map() did not map by Name")
	}
}

func TestSlice(t *testing.T) {
	m := map[ID]Resource{
		"id1": {ID: "id1", Name: "babarot/enhancd"},
		"id2": {ID: "id2", Name: "stedolan/jq"},
	}

	got := Slice(m)
	if len(got) != 2 {
		t.Fatalf("Slice() len = %d, want 2", len(got))
	}

	sortByName := cmpopts.SortSlices(func(a, b Resource) bool {
		return a.Name < b.Name
	})
	want := []Resource{
		{ID: "id1", Name: "babarot/enhancd"},
		{ID: "id2", Name: "stedolan/jq"},
	}
	if diff := cmp.Diff(want, got, sortByName); diff != "" {
		t.Errorf("Slice() mismatch (-want +got):\n%s", diff)
	}
}

func TestResource_exists(t *testing.T) {
	origExists := exists
	defer func() { exists = origExists }()

	t.Run("no paths", func(t *testing.T) {
		r := Resource{Paths: nil}
		if r.exists() {
			t.Error("exists() should return false when Paths is nil")
		}
	})

	t.Run("all exist", func(t *testing.T) {
		exists = func(string) bool { return true }
		r := Resource{Paths: []string{"/a", "/b"}}
		if !r.exists() {
			t.Error("exists() should return true when all paths exist")
		}
	})

	t.Run("one missing", func(t *testing.T) {
		exists = func(path string) bool { return path != "/b" }
		r := Resource{Paths: []string{"/a", "/b"}}
		if r.exists() {
			t.Error("exists() should return false when a path is missing")
		}
	})
}

func Test_contains(t *testing.T) {
	resources := []Resource{
		{Name: "babarot/enhancd"},
		{Name: "stedolan/jq"},
	}

	if !contains(resources, "babarot/enhancd") {
		t.Error("contains() should return true for existing name")
	}
	if contains(resources, "nonexistent") {
		t.Error("contains() should return false for non-existing name")
	}
}

func Test_add_remove_update(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		s := &State{Self: Self{Resources: map[ID]Resource{}}}
		r := Resource{ID: "id1", Name: "test/pkg"}
		add(r, s)
		if _, ok := s.Resources["id1"]; !ok {
			t.Error("add() did not add resource")
		}
	})

	t.Run("remove existing", func(t *testing.T) {
		s := &State{Self: Self{Resources: map[ID]Resource{
			"id1": {ID: "id1", Name: "test/pkg"},
		}}}
		remove(Resource{ID: "id1", Name: "test/pkg"}, s)
		if len(s.Resources) != 0 {
			t.Error("remove() did not remove resource")
		}
	})

	t.Run("remove non-existent", func(t *testing.T) {
		s := &State{Self: Self{Resources: map[ID]Resource{
			"id1": {ID: "id1", Name: "test/pkg"},
		}}}
		remove(Resource{ID: "id2", Name: "other/pkg"}, s)
		if len(s.Resources) != 1 {
			t.Error("remove() should not change state for non-existent resource")
		}
	})

	t.Run("update existing", func(t *testing.T) {
		s := &State{Self: Self{Resources: map[ID]Resource{
			"id1": {ID: "id1", Name: "test/pkg", Version: "v1"},
		}}}
		update(Resource{ID: "id1", Name: "test/pkg", Version: "v2"}, s)
		if s.Resources["id1"].Version != "v2" {
			t.Error("update() did not update resource")
		}
	})

	t.Run("update non-existent", func(t *testing.T) {
		s := &State{Self: Self{Resources: map[ID]Resource{}}}
		update(Resource{ID: "id1", Name: "test/pkg"}, s)
		if len(s.Resources) != 0 {
			t.Error("update() should not add non-existent resource")
		}
	})
}

func TestOpen_localPackageSkipped(t *testing.T) {
	stubState(map[string]string{
		"state.json": `{"resources":{}}`,
	})

	pkgs := stubPackages([]Resource{
		{
			ID:   "local/mydir",
			Name: "mydir",
			Type: "Local",
		},
	})

	state, err := Open("state.json", pkgs)
	if err != nil {
		t.Fatal(err)
	}

	if len(state.Additions) != 0 {
		t.Errorf("Open() should skip Local packages, got %d additions", len(state.Additions))
	}
}

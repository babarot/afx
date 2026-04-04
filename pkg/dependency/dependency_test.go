package dependency

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewNode(t *testing.T) {
	tests := map[string]struct {
		name string
		deps []string
		want *Node
	}{
		"no deps": {
			name: "A",
			deps: nil,
			want: &Node{Name: "A", Deps: nil},
		},
		"with deps": {
			name: "A",
			deps: []string{"B", "C"},
			want: &Node{Name: "A", Deps: []string{"B", "C"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewNode(tt.name, tt.deps...)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewNode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHas(t *testing.T) {
	tests := map[string]struct {
		graph Graph
		want  bool
	}{
		"empty graph": {
			graph: Graph{},
			want:  false,
		},
		"no deps": {
			graph: Graph{NewNode("A"), NewNode("B")},
			want:  false,
		},
		"has deps": {
			graph: Graph{NewNode("A", "B"), NewNode("B")},
			want:  true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Has(tt.graph)
			if got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func nodeNames(g Graph) []string {
	names := make([]string, len(g))
	for i, n := range g {
		names[i] = n.Name
	}
	return names
}

func TestResolve(t *testing.T) {
	tests := map[string]struct {
		graph     Graph
		wantNames []string // expected resolved order (sorted per depth level)
		wantErr   bool
	}{
		"empty graph": {
			graph:     Graph{},
			wantNames: []string{},
			wantErr:   false,
		},
		"no dependencies": {
			graph:     Graph{NewNode("A"), NewNode("B"), NewNode("C")},
			wantNames: []string{"A", "B", "C"}, // all independent, sorted
			wantErr:   false,
		},
		"linear chain": {
			// A -> B -> C
			graph:     Graph{NewNode("A", "B"), NewNode("B", "C"), NewNode("C")},
			wantNames: []string{"C", "B", "A"},
			wantErr:   false,
		},
		"diamond": {
			// C -> A, D -> A, A -> B
			graph:     Graph{NewNode("C", "A"), NewNode("D", "A"), NewNode("A", "B"), NewNode("B")},
			wantNames: []string{"B", "A", "C", "D"}, // C,D are at same depth
			wantErr:   false,
		},
		"circular dependency": {
			graph:   Graph{NewNode("A", "B"), NewNode("B", "A")},
			wantErr: true,
		},
		"self referencing": {
			graph:   Graph{NewNode("A", "A")},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Resolve(tt.graph)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Resolve() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}

			gotNames := nodeNames(got)
			// Sort names at same depth for deterministic comparison
			// mapset iteration is non-deterministic, so we sort the result
			sort.Strings(gotNames)
			sort.Strings(tt.wantNames)
			if diff := cmp.Diff(tt.wantNames, gotNames); diff != "" {
				t.Errorf("Resolve() names mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolve_order(t *testing.T) {
	// Verify that dependencies come before dependents
	// A -> B -> C means C must appear before B, B before A
	graph := Graph{NewNode("A", "B"), NewNode("B", "C"), NewNode("C")}
	resolved, err := Resolve(graph)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	pos := make(map[string]int)
	for i, n := range resolved {
		pos[n.Name] = i
	}

	if pos["C"] >= pos["B"] {
		t.Errorf("C should come before B, got C=%d B=%d", pos["C"], pos["B"])
	}
	if pos["B"] >= pos["A"] {
		t.Errorf("B should come before A, got B=%d A=%d", pos["B"], pos["A"])
	}
}

func TestGraph_String(t *testing.T) {
	tests := map[string]struct {
		graph Graph
		want  string
	}{
		"empty": {
			graph: Graph{},
			want:  "",
		},
		"with deps": {
			graph: Graph{NewNode("A", "B", "C")},
			want:  "* A -> B\n* A -> C\n",
		},
		"no deps": {
			graph: Graph{NewNode("A")},
			want:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.graph.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

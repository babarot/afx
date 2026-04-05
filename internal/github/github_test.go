package github

import (
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestGetAssetKeys(t *testing.T) {
	tests := map[string]struct {
		assets []Asset
		want   []string
	}{
		"empty": {
			assets: nil,
			want:   nil,
		},
		"single": {
			assets: []Asset{{Name: "tool-v1.0-linux.tar.gz"}},
			want:   []string{"tool-v1.0-linux.tar.gz"},
		},
		"multiple": {
			assets: []Asset{
				{Name: "a.tar.gz"},
				{Name: "b.zip"},
			},
			want: []string{"a.tar.gz", "b.zip"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := getAssetKeys(tt.assets)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getAssetKeys() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAssets_filter(t *testing.T) {
	assets := Assets{
		{Name: "tool-linux-amd64.tar.gz"},
		{Name: "tool-darwin-amd64.tar.gz"},
		{Name: "tool-windows-amd64.zip"},
	}

	tests := map[string]struct {
		fn      func(Asset) bool
		wantLen int
	}{
		"filter darwin": {
			fn:      func(a Asset) bool { return strings.Contains(a.Name, "darwin") },
			wantLen: 1,
		},
		"filter all match": {
			fn:      func(a Asset) bool { return strings.Contains(a.Name, "amd64") },
			wantLen: 3,
		},
		"filter none match": {
			fn:      func(a Asset) bool { return strings.Contains(a.Name, "arm64") },
			wantLen: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Make a copy to avoid mutation across tests
			cp := make(Assets, len(assets))
			copy(cp, assets)
			result := cp.filter(tt.fn)
			if len(*result) != tt.wantLen {
				t.Errorf("filter() len = %d, want %d", len(*result), tt.wantLen)
			}
		})
	}
}

func TestAssets_filter_shortCircuit(t *testing.T) {
	// Single asset should not be filtered (short-circuit)
	assets := Assets{{Name: "only-one.tar.gz"}}
	result := assets.filter(func(a Asset) bool { return false })
	if len(*result) != 1 {
		t.Errorf("filter() with single asset should short-circuit, got len %d", len(*result))
	}
}

func TestRelease_filterAssets_noAssets(t *testing.T) {
	r := &Release{Name: "test", Assets: Assets{}}
	_, err := r.filterAssets()
	if err == nil {
		t.Error("filterAssets() expected error for empty assets")
	}
}

func TestRelease_filterAssets_customFilter(t *testing.T) {
	r := &Release{
		Name: "test",
		Assets: Assets{
			{Name: "a.tar.gz", URL: "https://example.com/a"},
			{Name: "b.tar.gz", URL: "https://example.com/b"},
		},
		filter: func(assets Assets) *Asset {
			for _, a := range assets {
				if a.Name == "b.tar.gz" {
					return &a
				}
			}
			return nil
		},
	}

	got, err := r.filterAssets()
	if err != nil {
		t.Fatalf("filterAssets() error: %v", err)
	}
	if got.Name != "b.tar.gz" {
		t.Errorf("filterAssets() = %q, want 'b.tar.gz'", got.Name)
	}
}

func TestRelease_filterAssets_customFilterNoMatch(t *testing.T) {
	r := &Release{
		Name:   "test",
		Assets: Assets{{Name: "a.tar.gz"}},
		filter: func(assets Assets) *Asset { return nil },
	}

	_, err := r.filterAssets()
	if err == nil {
		t.Error("filterAssets() expected error when custom filter returns nil")
	}
}

func TestReleaseOptions(t *testing.T) {
	r := &Release{}

	WithOverwrite()(r)
	if !r.overwrite {
		t.Error("WithOverwrite() did not set overwrite")
	}

	WithWorkdir("/tmp/test")(r)
	if r.workdir != "/tmp/test" {
		t.Errorf("WithWorkdir() = %q, want '/tmp/test'", r.workdir)
	}

	WithVerbose()(r)
	if !r.verbose {
		t.Error("WithVerbose() did not set verbose")
	}

	called := false
	WithFilter(func(a Assets) *Asset {
		called = true
		return nil
	})(r)
	r.filter(nil)
	if !called {
		t.Error("WithFilter() did not set filter function")
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		client := NewHTTPClient()
		if client == nil {
			t.Fatal("NewHTTPClient() returned nil")
		}
	})

	t.Run("with ReplaceTripper", func(t *testing.T) {
		custom := &http.Transport{}
		client := NewHTTPClient(ReplaceTripper(custom))
		if client.Transport != custom {
			t.Error("ReplaceTripper did not set custom transport")
		}
	})
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.http == nil {
		t.Error("NewClient().http is nil")
	}
}

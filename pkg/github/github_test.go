package github

import (
	"testing"
)

func Test_filterAssets(t *testing.T) {
	testCases := map[string]struct {
		release Release
		want    Asset
	}{
		"hoge": {
			release: Release{
				Name: "owner/repo",
				Assets: Assets{
					{Name: "foo-darwin-amd64.tar.gz", URL: "http://example.com"},
					{Name: "foo-linux-amd64.tar.gz", URL: "http://example.com"},
				},
				workdir: "",
				verbose: false,
				filter:  nil,
			},
			want: Asset{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.release.filterAssets()
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
			}
		})
	}
}

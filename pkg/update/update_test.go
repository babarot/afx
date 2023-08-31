package update

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/babarot/afx/pkg/github"
	"github.com/cli/cli/v2/pkg/httpmock"
)

func TestCheckForUpdate(t *testing.T) {
	orig_GITHUB_TOKEN := os.Getenv("GITHUB_TOKEN")
	t.Cleanup(func() {
		os.Setenv("GITHUB_TOKEN", orig_GITHUB_TOKEN)
	})

	scenarios := []struct {
		Name           string
		CurrentVersion string
		LatestVersion  string
		LatestURL      string
		ExpectsResult  bool
	}{
		{
			Name:           "latest is newer",
			CurrentVersion: "v0.0.1",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "current is prerelease",
			CurrentVersion: "v1.0.0-pre.1",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "current is built from source",
			CurrentVersion: "v1.2.3-123-gdeadbeef",
			LatestVersion:  "v1.2.3",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  false,
		},
		{
			Name:           "current is built from source after a prerelease",
			CurrentVersion: "v1.2.3-rc.1-123-gdeadbeef",
			LatestVersion:  "v1.2.3",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "latest is newer than version build from source",
			CurrentVersion: "v1.2.3-123-gdeadbeef",
			LatestVersion:  "v1.2.4",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "latest is current",
			CurrentVersion: "v1.0.0",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  false,
		},
		{
			Name:           "latest is older",
			CurrentVersion: "v0.10.0-pre.1",
			LatestVersion:  "v0.9.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  false,
		},
	}

	log.SetOutput(ioutil.Discard)

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", "TOKEN")
			mock := &httpmock.Registry{}
			client := github.NewClient(github.ReplaceTripper(mock))

			mock.Register(
				httpmock.REST(http.MethodGet, "repos/OWNER/REPO/releases/latest"),
				httpmock.StringResponse(fmt.Sprintf(`{
					"tag_name": "%s",
					"html_url": "%s"
				}`, s.LatestVersion, s.LatestURL)),
			)

			rel, err := CheckForUpdate(client, tempFilePath(), "OWNER/REPO", s.CurrentVersion)
			if err != nil {
				t.Fatal(err)
			}

			if len(mock.Requests) != 1 {
				t.Fatalf("expected 1 HTTP request, got %d", len(mock.Requests))
			}
			requestPath := mock.Requests[0].URL.Path
			if requestPath != "/repos/OWNER/REPO/releases/latest" {
				t.Errorf("HTTP path: %q", requestPath)
			}

			if !s.ExpectsResult {
				if rel != nil {
					t.Fatal("expected no new release")
				}
				return
			}
			if rel == nil {
				t.Fatal("expected to report new release")
			}

			if rel.Version != s.LatestVersion {
				t.Errorf("Version: %q", rel.Version)
			}
			if rel.URL != s.LatestURL {
				t.Errorf("URL: %q", rel.URL)
			}
		})
	}
}

func tempFilePath() string {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(file.Name())
	return file.Name()
}

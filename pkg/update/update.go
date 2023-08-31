package update

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/babarot/afx/pkg/github"
	"github.com/hashicorp/go-version"
)

// refer: github.com/cli/cli/tree/<hash>/internal/update
// hash: bf83c660a1ae486d582117e0a174f8e109b64775
var gitDescribeSuffixRE = regexp.MustCompile(`\d+-\d+-g[a-f0-9]{8}$`)

// ReleaseInfo stores information about a release
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

type StateEntry struct {
	CheckedForUpdateAt time.Time   `json:"checked_for_update_at"`
	LatestRelease      ReleaseInfo `json:"latest_release"`
}

// CheckForUpdate checks whether this software has had a newer release on GitHub
func CheckForUpdate(client *github.Client, stateFilePath, repo, currentVersion string) (*ReleaseInfo, error) {
	stateEntry, _ := getStateEntry(stateFilePath)
	if stateEntry != nil && time.Since(stateEntry.CheckedForUpdateAt).Hours() < 24 {
		return nil, nil
	}

	releaseInfo, err := getLatestReleaseInfo(client, repo)
	if err != nil {
		return nil, err
	}

	err = setStateEntry(stateFilePath, time.Now(), *releaseInfo)
	if err != nil {
		return nil, err
	}

	if versionGreaterThan(releaseInfo.Version, currentVersion) {
		return releaseInfo, nil
	}

	return nil, nil
}

func getLatestReleaseInfo(client *github.Client, repo string) (*ReleaseInfo, error) {
	var latestRelease ReleaseInfo

	log.Printf("[DEBUG] call GitHub Release API to get release info")

	api := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	err := client.REST(http.MethodGet, api, nil, &latestRelease)
	if err != nil {
		return nil, err
	}

	return &latestRelease, nil
}

func getStateEntry(stateFilePath string) (*StateEntry, error) {
	content, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	var stateEntry StateEntry
	err = json.Unmarshal(content, &stateEntry)
	if err != nil {
		return nil, err
	}

	return &stateEntry, nil
}

func setStateEntry(stateFilePath string, t time.Time, r ReleaseInfo) error {
	data := StateEntry{CheckedForUpdateAt: t, LatestRelease: r}
	content, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(stateFilePath), 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(stateFilePath, content, 0600)
	return err
}

func versionGreaterThan(v, w string) bool {
	w = gitDescribeSuffixRE.ReplaceAllStringFunc(w, func(m string) string {
		idx := strings.IndexRune(m, '-')
		n, _ := strconv.Atoi(m[0:idx])
		return fmt.Sprintf("%d-pre.0", n+1)
	})

	vv, ve := version.NewVersion(v)
	vw, we := version.NewVersion(w)

	return ve == nil && we == nil && vv.GreaterThan(vw)
}

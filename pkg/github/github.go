package github

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/b4b4r07/afx/internal/diags"
	"github.com/b4b4r07/afx/pkg/logging"
	"github.com/inconshreveable/go-update"
	"github.com/mholt/archiver"
	"github.com/schollz/progressbar/v3"
)

// Release represents a GitHub release and its client
// A difference from Release is whether a client or not
type Release struct {
	Name   string
	Assets Assets

	client  *Client
	workdir string
	verbose bool
	filter  func(Assets) *Asset
}

// Asset represents GitHub release's asset.
// Basically this means one archive file attached in a release
type Asset struct {
	Name string
	URL  string
}

// ReleaseResponse is a response of github release structure
// TODO: This may be better to become same one strucure as above
type ReleaseResponse struct {
	Assets []AssetsResponse `json:"assets"`
}

type AssetsResponse struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Assets []Asset

func (as *Assets) filter(fn func(Asset) bool) *Assets {
	var assets Assets
	if len(*as) < 2 {
		// no more need to filter
		log.Printf("[DEBUG] assets.filter: finished filtering because length of assets is less than two")
		return as
	}

	for _, asset := range *as {
		if fn(asset) {
			assets = append(assets, asset)
		}
	}

	// logging if assets are changed by filter
	if len(*as) != len(assets) {
		log.Printf("[DEBUG] assets.filter: filtered: %#v", getAssetKeys(assets))
	}

	*as = assets
	return as
}

func getAssetKeys(assets []Asset) []string {
	var names []string
	for _, asset := range assets {
		names = append(names, asset.Name)
	}
	return names
}

type Option func(r *Release)

type FilterFunc func(assets Assets) *Asset

func WithWorkdir(workdir string) Option {
	return func(r *Release) {
		r.workdir = workdir
	}
}

func WithVerbose() Option {
	return func(r *Release) {
		r.verbose = true
	}
}

func WithFilter(filter func(Assets) *Asset) Option {
	return func(r *Release) {
		r.filter = filter
	}
}

func NewRelease(ctx context.Context, owner, repo, tag string, opts ...Option) (*Release, error) {
	if owner == "" || repo == "" {
		return nil, diags.New("owner and repo are required")
	}

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	switch tag {
	case "latest", "":
		releaseURL += "/latest"
	default:
		releaseURL += fmt.Sprintf("/tags/%s", tag)
	}
	log.Printf("[DEBUG] getting asset data from %s", releaseURL)

	var resp ReleaseResponse
	client := NewClient(
		ReplaceTripper(logging.NewTransport("GitHub", http.DefaultTransport)),
	)
	err := client.REST(http.MethodGet, releaseURL, nil, &resp)
	if err != nil {
		return nil, err
	}

	var assets []Asset
	for _, asset := range resp.Assets {
		assets = append(assets, Asset{
			Name: asset.Name,
			URL:  asset.BrowserDownloadURL,
		})
	}

	tmp, err := ioutil.TempDir("", repo)
	if err != nil {
		return nil, err
	}

	release := &Release{
		Name:    repo,
		Assets:  assets,
		client:  client,
		workdir: tmp,
		verbose: false,
		filter:  nil,
	}

	for _, o := range opts {
		o(release)
	}

	return release, nil
}

func (r *Release) filterAssets() (Asset, error) {
	log.Printf("[DEBUG] assets: %#v\n", getAssetKeys(r.Assets))

	if len(r.Assets) == 0 {
		return Asset{}, diags.New("no assets")
	}

	if r.filter != nil {
		log.Printf("[DEBUG] asset: filterfunc: started running")
		asset := r.filter(r.Assets)
		if asset != nil {
			log.Printf("[DEBUG] asset: filterfunc: matched in assets")
			return *asset, nil
		}
		log.Printf("[DEBUG] asset: filterfunc: not matched in assets")
		return Asset{}, diags.New("could not find assets with given name")
	}

	log.Printf("[DEBUG] asset: %s: using default assets filter", r.Name)
	assets := *r.Assets.
		filter(func(asset Asset) bool {
			expr := `.*\.sbom`
			// filter out
			return !regexp.MustCompile(expr).MatchString(asset.Name)
		}).
		filter(func(asset Asset) bool {
			expr := ".*(sha256sum|checksum).*"
			// filter out
			return !regexp.MustCompile(expr).MatchString(asset.Name)
		}).
		filter(func(asset Asset) bool {
			expr := ""
			switch runtime.GOOS {
			case "darwin":
				expr += ".*(apple|darwin|Darwin|osx|mac|macos|macOS).*"
			case "linux":
				expr += ".*(linux|hoe).*"
			}
			return regexp.MustCompile(expr).MatchString(asset.Name)
		}).
		filter(func(asset Asset) bool {
			expr := ""
			switch runtime.GOARCH {
			case "amd64":
				expr += ".*(amd64|64).*"
			case "386":
				expr += ".*(386|86).*"
			}
			return regexp.MustCompile(expr).MatchString(asset.Name)
		})

	switch len(assets) {
	case 0:
		return Asset{}, diags.New("asset not found after filtered")
	case 1:
		return assets[0], nil
	default:
		log.Printf("[WARN] %d assets found: %#v", len(assets), getAssetKeys(assets))
		log.Printf("[WARN] first one %q will be used", assets[0].Name)
		return assets[0], nil
	}
}

// Download downloads GitHub Release from given page
func (r *Release) Download(ctx context.Context) (Asset, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	asset, err := r.filterAssets()
	if err != nil {
		log.Printf("[ERROR] %s: could not find assets available on your system", r.Name)
		return asset, err
	}

	log.Printf("[DEBUG] asset: %#v", asset)

	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return asset, err
	}

	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return asset, err
	}
	defer resp.Body.Close()

	os.MkdirAll(r.workdir, os.ModePerm)
	archive := filepath.Join(r.workdir, asset.Name)

	file, err := os.Create(archive)
	if err != nil {
		return asset, diags.Wrapf(err, "%s: failed to create file", archive)
	}
	defer file.Close()

	var w io.Writer
	if r.verbose {
		w = io.MultiWriter(file, progressbar.DefaultBytes(
			resp.ContentLength,
			"Downloading",
		))
	} else {
		w = file
	}

	_, err = io.Copy(w, resp.Body)
	return asset, err
}

// Unarchive extracts downloaded asset
func (r *Release) Unarchive(asset Asset) error {
	archive := filepath.Join(r.workdir, asset.Name)

	uaIface, err := archiver.ByExtension(archive)
	if err != nil {
		// err: this will be an error of format unrecognized by filename
		// but in this case, maybe not archived file: e.g. tigrawap/slit
		//
		log.Printf("[WARN] archiver.ByExtension(): %v", err)

		// TODO: remove this logic?
		// thanks to this logic, we don't need to specify this statement to link.from
		//
		//   command:
		//     link:
		//     - from: '*jq*'
		//
		// because this logic renames a binary of 'jq-1.6' to 'jq'
		//
		target := filepath.Join(r.workdir, r.Name)
		if _, err := os.Stat(target); err != nil {
			log.Printf("[DEBUG] renamed from %s to %s", archive, target)
			os.Rename(archive, target)
			os.Chmod(target, 0755)
		}
		return nil
	}

	tar := &archiver.Tar{
		OverwriteExisting:      true,
		MkdirAll:               false,
		ImplicitTopLevelFolder: false,
		ContinueOnError:        false,
	}
	switch v := uaIface.(type) {
	case *archiver.Rar:
		v.OverwriteExisting = true
	case *archiver.Zip:
		v.OverwriteExisting = true
	case *archiver.TarBz2:
		v.Tar = tar
	case *archiver.TarGz:
		v.Tar = tar
	case *archiver.TarLz4:
		v.Tar = tar
	case *archiver.TarSz:
		v.Tar = tar
	case *archiver.TarXz:
		v.Tar = tar
	case *archiver.Gz,
		*archiver.Bz2,
		*archiver.Lz4,
		*archiver.Snappy,
		*archiver.Xz:
		// nothing to customise
	}

	u, ok := uaIface.(archiver.Unarchiver)
	if !ok {
		return diags.New("cannot type assertion with archiver.Unarchiver")
	}

	if err := u.Unarchive(archive, r.workdir); err != nil {
		return diags.Wrapf(err, "%s: failed to unarchive", r.Name)
	}

	log.Printf("[DEBUG] removed archive file: %s", archive)
	os.Remove(archive)

	return nil
}

// Install instals unarchived packages to given path
func (r *Release) Install(to string) error {
	bin := filepath.Join(r.workdir, r.Name)
	log.Printf("[DEBUG] release install: %#v", bin)

	fp, err := os.Open(bin)
	if err != nil {
		return diags.Wrap(err, "failed to open file")
	}
	defer fp.Close()

	log.Printf("[DEBUG] installing: from %s to %s", bin, to)
	return update.Apply(fp, update.Options{
		TargetPath: to,
	})
}

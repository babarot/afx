package config

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

	"golang.org/x/oauth2"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/b4b4r07/afx/pkg/data"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/logging"
	"github.com/b4b4r07/afx/pkg/state2"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/google/go-github/github"
	"github.com/mholt/archiver"
	"github.com/tidwall/gjson"
)

// GitHub represents GitHub repository
type GitHub struct {
	Name string `yaml:"name" validate:"required"`

	Owner       string `yaml:"owner" validate:"required"`
	Repo        string `yaml:"repo" validate:"required"`
	Description string `yaml:"description"`

	Branch string        `yaml:"branch"`
	Option *GitHubOption `yaml:"with"`

	Release *Release `yaml:"release"`

	Plugin  *Plugin  `yaml:"plugin"`
	Command *Command `yaml:"command" validate:"required_with=Release"` // TODO: (not required Release)

	DependsOn []string `yaml:"depends-on"`
}

type GitHubOption struct {
	Depth int `yaml:"depth"`
}

// Release represents a GitHub release structure
type Release struct {
	Name string `yaml:"name" validate:"required"`
	Tag  string `yaml:"tag"`

	// TODO: (internal change): rename Artifact to Asset
	Artifact Artifact `yaml:"asset"`
}

type Artifact struct {
	Filename     string            `yaml:"filename"`
	Replacements map[string]string `yaml:"replacements"`
}

func NewGitHub(ctx context.Context, owner, repo string) (GitHub, error) {
	r, err := getRepo(ctx, owner, repo)
	if err != nil {
		return GitHub{}, err
	}
	release, command := getRelease(ctx, owner, repo)
	return GitHub{
		Name:        repo,
		Owner:       owner,
		Repo:        repo,
		Branch:      "master",
		Description: r.GetDescription(),
		Plugin:      nil,
		Command:     command,
		Release:     release,
	}, nil
}

func githubClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func getRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	c := githubClient(ctx)
	r, _, err := c.Repositories.Get(ctx, owner, repo)
	return r, err
}

func getRelease(ctx context.Context, owner, repo string) (*Release, *Command) {
	var release *Release
	var command *Command
	c := githubClient(ctx)
	latest, _, err := c.Repositories.GetLatestRelease(
		ctx, owner, repo,
	)
	if err == nil {
		release = &Release{
			Name: repo,
			Tag:  latest.GetTagName(),
		}
		command = &Command{
			Link: []*Link{{
				From: repo,
				To:   repo,
			}},
		}
	}
	return release, command
}

// Init runs initialization step related to GitHub packages
func (c GitHub) Init() error {
	var errs errors.Errors
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Init(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Init(c))
	}
	return errs.ErrorOrNil()
}

// Clone runs git clone
func (c GitHub) Clone(ctx context.Context) error {
	writer := ioutil.Discard
	if logging.IsTrace() {
		writer = os.Stdout
	}

	var opt GitHubOption
	if c.Option != nil {
		opt = *c.Option
	}

	var r *git.Repository
	_, err := os.Stat(c.GetHome())
	switch {
	case os.IsNotExist(err):
		r, err = git.PlainCloneContext(ctx, c.GetHome(), false, &git.CloneOptions{
			URL:      fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo),
			Tags:     git.NoTags,
			Depth:    opt.Depth,
			Progress: writer,
		})
		if err != nil {
			return errors.Wrapf(err, "%s: failed to clone repository", c.GetName())
		}
	default:
		r, err = git.PlainOpen(c.GetHome())
		if err != nil {
			return errors.Wrapf(err, "%s: failed to open repository", c.GetName())
		}
	}

	w, err := r.Worktree()
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get worktree", c.GetName())
	}

	if c.Branch != "" {
		var err error
		err = r.FetchContext(ctx, &git.FetchOptions{
			RemoteName: "origin",
			RefSpecs: []config.RefSpec{
				config.RefSpec(fmt.Sprintf("+%s:%s",
					plumbing.NewBranchReferenceName(c.Branch),
					plumbing.NewBranchReferenceName(c.Branch),
				)),
			},
			Depth:    opt.Depth,
			Force:    true,
			Tags:     git.NoTags,
			Progress: writer,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return errors.Wrapf(err, "%s: failed to fetch repository", c.Branch)
		}
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + c.Branch),
			Force:  true,
		})
		if err != nil {
			return errors.Wrapf(err, "%s: failed to checkout", c.Branch)
		}
	}

	return nil
}

// Install installs from GitHub repository with git clone command
func (c GitHub) Install(ctx context.Context, status chan<- Status) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("[DEBUG] canceled")
		return nil
	default:
		// Go installing step!
	}

	switch {
	case c.Release == nil:
		err := c.Clone(ctx)
		if err != nil {
			err = errors.Wrapf(err, "%s: failed to clone repo", c.Name)
			status <- Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
	case c.Release != nil:
		err := c.InstallFromRelease(ctx)
		if err != nil {
			err = errors.Wrapf(err, "%s: failed to get from release", c.Name)
			status <- Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
	}

	var errs errors.Errors
	if c.HasPluginBlock() {
		errs.Append(c.Plugin.Install(c))
	}
	if c.HasCommandBlock() {
		errs.Append(c.Command.Install(c))
	}

	status <- Status{Name: c.GetName(), Done: true, Err: errs.ErrorOrNil() != nil}
	return errs.ErrorOrNil()
}

// Installed returns true the GitHub package is already installed
func (c GitHub) Installed() bool {
	var list []bool

	if c.HasPluginBlock() {
		list = append(list, c.Plugin.Installed(c))
	}

	if c.HasCommandBlock() {
		list = append(list, c.Command.Installed(c))
	}

	switch {
	case c.HasPluginBlock():
	case c.HasCommandBlock():
	default:
		_, err := os.Stat(c.GetHome())
		list = append(list, err == nil)
	}

	return check(list)
}

// ReleaseURL returns URL of GitHub release
func (c GitHub) ReleaseURL() string {
	tag := c.Release.Tag
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s",
		c.Owner, c.Repo, c.Release.Tag)
}

// InstallFromRelease runs install from GitHub release, from not repository
func (c GitHub) InstallFromRelease(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, c.ReleaseURL(), nil)
	if err != nil {
		return errors.Wrapf(err,
			"failed to complete the request to %v to fetch artifact list",
			c.ReleaseURL())
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN is missing")
	}
	req.Header.Set("Authorization", "token "+token)

	httpClient := http.DefaultClient
	httpClient.Transport = logging.NewTransport("GitHub", http.DefaultTransport)

	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	assets := gjson.Get(string(body), "assets")

	if !assets.Exists() {
		return errors.Detail{
			Head:    "cannot fetch the list from GitHub Releases",
			Summary: c.GetName(),
			Details: []string{
				gjson.Get(string(body), "message").String(),
				string(body),
			},
		}
	}

	filename, err := c.templateFilename()
	if err != nil {
		return errors.Wrapf(err, "failed to template filename")
	}

	release := GitHubRelease{
		Name:     c.Release.Name,
		Client:   httpClient,
		Assets:   Assets{},
		Filename: filename,
	}

	assets.ForEach(func(key, value gjson.Result) bool {
		name := value.Get("name").String()
		release.Assets = append(release.Assets, Asset{
			Name: name,
			Home: c.GetHome(),
			Path: filepath.Join(c.GetHome(), name),
			URL:  value.Get("browser_download_url").String(),
		})
		return true
	})

	if len(release.Assets) == 0 {
		return errors.Wrapf(err, "%s: failed to get releases", release.Name)
	}

	asset, err := release.Download(ctx)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to download", release.Name)
	}

	if err := release.Unarchive(asset); err != nil {
		return errors.Wrapf(err, "%s: failed to unarchive", release.Name)
	}

	return nil
}

func (c GitHub) templateFilename() (string, error) {
	filename := c.Release.Artifact.Filename
	replacements := c.Release.Artifact.Replacements

	if filename == "" {
		// no filename specified
		return "", nil
	}

	log.Printf("[DEBUG] asset: templating filename from %q", filename)

	data := data.New(
		data.WithPackage(c),
		data.WithRelease(data.Release{
			Name: c.Release.Name,
			Tag:  c.Release.Tag,
		}),
	)

	filename, err := templates.New(data).
		Replace(replacements).
		Apply(filename)

	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] asset: templated filename: -> %q", filename)
	return filename, nil
}

// GitHubRelease represents a GitHub release and its client
// A difference from Release is whether a client or not
type GitHubRelease struct {
	Client *http.Client

	Name   string
	Assets Assets

	Filename string
}

// Asset represents GitHub release's asset.
// Basically this means one archive file attached in a release
type Asset struct {
	Name string
	Home string
	Path string
	URL  string
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

// getAssetKeys just returns list of asset.Name
func getAssetKeys(assets []Asset) []string {
	var names []string
	for _, asset := range assets {
		names = append(names, asset.Name)
	}
	return names
}

func (r *GitHubRelease) GetAsset() (Asset, error) {
	log.Printf("[DEBUG] assets: %#v\n", getAssetKeys(r.Assets))

	if len(r.Assets) == 0 {
		return Asset{}, fmt.Errorf("%s: no assets found", r.Name)
	}

	if r.Filename != "" {
		log.Printf("[DEBUG] asset: found filename %q is specified in config", r.Filename)
		for _, asset := range r.Assets {
			if asset.Name == r.Filename {
				log.Printf("[DEBUG] asset: filename %q is matched with assets", r.Filename)
				return asset, nil
			}
		}
		return Asset{}, fmt.Errorf("%s: no matched in assets", r.Filename)
	}

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
		return Asset{}, errors.New("asset not found after filtered")
	case 1:
		return assets[0], nil
	default:
		log.Printf("[WARN] %d assets found: %#v", len(assets), getAssetKeys(assets))
		log.Printf("[WARN] first one %q will be used", assets[0].Name)
		return assets[0], nil
	}
}

// Download is
func (r *GitHubRelease) Download(ctx context.Context) (Asset, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	asset, err := r.GetAsset()
	if err != nil {
		return asset, err
	}

	log.Printf("[DEBUG] asset: %#v", asset)

	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return asset, err
	}

	client := new(http.Client)
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return asset, err
	}
	defer resp.Body.Close()

	os.MkdirAll(asset.Home, os.ModePerm)
	file, err := os.Create(asset.Path)
	if err != nil {
		return asset, errors.Wrapf(err, "%s: failed to create file", asset.Path)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)

	return asset, err
}

// Unarchive is
func (r *GitHubRelease) Unarchive(asset Asset) error {
	uaIface, err := archiver.ByExtension(asset.Path)
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
		target := filepath.Join(asset.Home, r.Name)
		if _, err := os.Stat(target); err != nil {
			log.Printf("[DEBUG] renamed from %s to %s", asset.Path, target)
			os.Rename(asset.Path, target)
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
		return errors.New("cannot type assertion with archiver.Unarchiver")
	}

	if err := u.Unarchive(asset.Path, asset.Home); err != nil {
		return errors.Wrapf(err, "%s: failed to unarchive", r.Name)
	}

	log.Printf("[DEBUG] removed archive file: %s", asset.Path)
	os.Remove(asset.Path)

	return nil
}

// HasPluginBlock is
func (c GitHub) HasPluginBlock() bool {
	return c.Plugin != nil
}

// HasCommandBlock is
func (c GitHub) HasCommandBlock() bool {
	return c.Command != nil
}

// HasReleaseBlock is
func (c GitHub) HasReleaseBlock() bool {
	return c.Release != nil
}

// GetPluginBlock is
func (c GitHub) GetPluginBlock() Plugin {
	if c.HasPluginBlock() {
		return *c.Plugin
	}
	return Plugin{}
}

// GetCommandBlock is
func (c GitHub) GetCommandBlock() Command {
	if c.HasCommandBlock() {
		return *c.Command
	}
	return Command{}
}

// Uninstall is
func (c GitHub) Uninstall(ctx context.Context) error {
	var errs errors.Errors

	delete := func(f string, errs *errors.Errors) {
		err := os.RemoveAll(f)
		if err != nil {
			errs.Append(err)
			return
		}
		log.Printf("[INFO] Delete %s\n", f)
	}

	if c.HasCommandBlock() {
		links, _ := c.Command.GetLink(c)
		for _, link := range links {
			delete(link.From, &errs)
			delete(link.To, &errs)
		}
	}

	if c.HasPluginBlock() {
	}

	delete(c.GetHome(), &errs)

	return errs.ErrorOrNil()
}

// GetName returns a name
func (c GitHub) GetName() string {
	return c.Name
}

// GetHome returns a path
func (c GitHub) GetHome() string {
	return filepath.Join(os.Getenv("HOME"), ".afx", "github.com", c.Owner, c.Repo)
}

func (c GitHub) GetDependsOn() []string {
	return c.DependsOn
}

func (c GitHub) GetResource() state2.Resource {
	return getResource(c)
}

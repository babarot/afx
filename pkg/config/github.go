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

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/logging"
	"github.com/google/go-github/github"
	"github.com/mholt/archiver"
	"github.com/tidwall/gjson"
)

// GitHub represents GitHub repository
type GitHub struct {
	Name string `yaml:"name"`

	Owner       string `yaml:"owner"`
	Repo        string `yaml:"repo"`
	Description string `yaml:"description"`
	Branch      string `yaml:"branch"`

	Release *Release `yaml:"release"`

	Plugin  *Plugin  `yaml:"plugin"`
	Command *Command `yaml:"command"`
}

// Release represents a GitHub release structure
type Release struct {
	Name string `yaml:"name"`
	Tag  string `yaml:"tag"`
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
			Link: []*Link{&Link{
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
	if logging.IsDebugOrHigher() {
		writer = os.Stdout
	}

	var r *git.Repository
	_, err := os.Stat(c.GetHome())
	switch {
	case os.IsNotExist(err):
		r, err = git.PlainCloneContext(ctx, c.GetHome(), false, &git.CloneOptions{
			URL:      fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo),
			Tags:     git.NoTags,
			Progress: writer,
		})
		if err != nil {
			return err
		}
	default:
		r, err = git.PlainOpen(c.GetHome())
		if err != nil {
			return err
		}
	}

	w, err := r.Worktree()
	if err != nil {
		return err
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
			Depth:    1,
			Force:    true,
			Tags:     git.NoTags,
			Progress: writer,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return errors.Wrap(err, "failed to fetch")
		}
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + c.Branch),
			Force:  true,
		})
		if err != nil {
			return err
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
			err = errors.Wrap(err, "failed to clone repo")
			status <- Status{Path: c.GetHome(), Done: true, Err: true}
			return err
		}
	case c.Release != nil:
		err := c.InstallFromRelease(ctx)
		if err != nil {
			err = errors.Wrap(err, "failed to get from release")
			status <- Status{Path: c.GetHome(), Done: true, Err: true}
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

	status <- Status{Path: c.GetHome(), Done: true, Err: errs.ErrorOrNil() != nil}
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

	release := GitHubRelease{
		Name:   c.Release.Name,
		Client: httpClient,
		Assets: []Asset{},
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
		log.Printf("[ERROR] %s is no release assets", c.Release.Name)
		return errors.New("failed to get releases")
	}

	if err := release.Download(ctx); err != nil {
		return errors.Wrapf(err, "failed to download: %q", release.Name)
	}

	if err := release.Unarchive(); err != nil {
		return errors.Wrapf(err, "failed to unarchive: %q", release.Name)
	}

	return nil
}

// GitHubRelease represents a GitHub release and its client
// A difference from Release is whether a client or not
type GitHubRelease struct {
	Client *http.Client

	Name   string
	Assets []Asset
}

// Asset represents GitHub release's asset.
// Basically this means one archive file attached in a release
type Asset struct {
	Name string
	Home string
	Path string
	URL  string
}

func (r *GitHubRelease) filter(fn func(Asset) bool) *GitHubRelease {
	var assets []Asset
	if len(r.Assets) < 2 {
		// no more need to filter
		return r
	}
	for _, asset := range r.Assets {
		if fn(asset) {
			assets = append(assets, asset)
		}
	}
	r.Assets = assets
	return r
}

// Download is
func (r *GitHubRelease) Download(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Printf("[DEBUG] assets: %#v\n", r.Assets)

	r.filter(func(asset Asset) bool {
		expr := ""
		switch runtime.GOOS {
		case "darwin":
			expr += ".*(apple|darwin|Darwin|osx|mac|macos|macOS).*"
		case "linux":
			expr += ".*(linux|hoe).*"
		}
		return regexp.MustCompile(expr).MatchString(asset.Name)
	})

	r.filter(func(asset Asset) bool {
		expr := ""
		switch runtime.GOARCH {
		case "amd64":
			expr += ".*(amd64|64).*"
		case "386":
			expr += ".*(386|86).*"
		}
		return regexp.MustCompile(expr).MatchString(asset.Name)
	})

	if len(r.Assets) == 0 {
		return fmt.Errorf("%s no assets found", r.Name)
	}

	asset := r.Assets[0]

	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return err
	}

	client := new(http.Client)
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	os.MkdirAll(asset.Home, os.ModePerm)
	file, err := os.Create(asset.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)

	return err
}

// Unarchive is
func (r *GitHubRelease) Unarchive() error {
	if len(r.Assets) == 0 {
		log.Printf("[DEBUG] no assets: %#v\n", r)
		return nil
	}
	a := r.Assets[0]

	uaIface, err := archiver.ByExtension(a.Path)
	if err != nil {
		// err: this will be an error of format unrecognized by filename
		// but in this case, maybe not archived file: e.g. tigrawap/slit
		//
		log.Printf("[ERROR] archiver.ByExtension(): %v", err)
		log.Printf("[DEBUG] %q is not an archive file so directly install", a.Name)
		target := filepath.Join(a.Home, r.Name)
		if _, err := os.Stat(target); err != nil {
			log.Printf("[DEBUG] renamed from %s to %s", a.Path, target)
			os.Rename(a.Path, target)
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
		return errors.New("not supported archive file")
	}

	if err := u.Unarchive(a.Path, a.Home); err != nil {
		log.Printf("[ERROR] failed to unarchive %s: %s\n", r.Name, err)
		return errors.Wrap(err, "archiver.Unarchive(): failed")
	}

	log.Printf("[DEBUG] removed archive file: %s\n", a.Path)
	os.Remove(a.Path)

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

package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/yaml.v2"

	"github.com/Masterminds/semver"
	"github.com/b4b4r07/afx/pkg/data"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/github"
	"github.com/b4b4r07/afx/pkg/logging"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/fatih/color"
	"github.com/go-playground/validator/v10"
)

// GitHub represents GitHub repository
type GitHub struct {
	Name string `yaml:"name" validate:"required"`

	Owner       string `yaml:"owner"       validate:"required"`
	Repo        string `yaml:"repo"        validate:"required"`
	Description string `yaml:"description"`

	Branch string        `yaml:"branch"`
	Option *GitHubOption `yaml:"with"`

	Release *GitHubRelease `yaml:"release"`

	Plugin  *Plugin   `yaml:"plugin"`
	Command *Command  `yaml:"command" validate:"required_with=Release"` // TODO: (not required Release)
	As      *GitHubAs `yaml:"as"`

	DependsOn []string `yaml:"depends-on"`
}

type GitHubAs struct {
	GHExtension *GHExtension `yaml:"gh-extension"`
}

type GHExtension struct {
	Name     string `yaml:"name" validate:"required,startswith=gh-"`
	Tag      string `yaml:"tag"`
	RenameTo string `yaml:"rename-to" validate:"gh-extension,excludesall=/"`
}

type GitHubOption struct {
	Depth int `yaml:"depth"`
}

// GitHubRelease represents a GitHub release structure
type GitHubRelease struct {
	Name string `yaml:"name" validate:"required"`
	Tag  string `yaml:"tag"`

	Asset GitHubReleaseAsset `yaml:"asset"`
}

type GitHubReleaseAsset struct {
	Filename     string            `yaml:"filename"`
	Replacements map[string]string `yaml:"replacements"`
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

	if c.IsGHExtension() {
		gh := c.As.GHExtension
		err := gh.Install(ctx, c.Owner, c.Repo, gh.GetTag())
		if err != nil {
			err = errors.Wrapf(err, "%s: failed to get from release", c.Name)
			status <- Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
	}

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

func (c GitHub) GetReleaseTag() string {
	if c.Release != nil {
		return c.Release.Tag
	}
	return "latest"
}

// InstallFromRelease runs install from GitHub release, from not repository
func (c GitHub) InstallFromRelease(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	owner, repo, tag := c.Owner, c.Repo, c.GetReleaseTag()
	log.Printf("[DEBUG] install from release: %s/%s (%s)", owner, repo, tag)

	release, err := github.NewRelease(
		ctx, owner, repo, tag,
		github.WithWorkdir(c.GetHome()),
		github.WithFilter(func(filename string) github.FilterFunc {
			if filename == "" {
				// cancel filtering
				return nil
			}
			return func(assets github.Assets) *github.Asset {
				for _, asset := range assets {
					if asset.Name == filename {
						return &asset
					}
				}
				return nil
			}
		}(c.templateFilename())),
	)
	if err != nil {
		return err
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

func (c GitHub) templateFilename() string {
	release := c.Release
	if release == nil {
		return ""
	}

	filename := release.Asset.Filename
	replacements := release.Asset.Replacements

	if filename == "" {
		// no filename specified
		return ""
	}

	log.Printf("[DEBUG] asset: templating filename from %q", filename)

	data := data.New(
		data.WithPackage(c),
		data.WithRelease(data.Release{
			Name: release.Name,
			Tag:  release.Tag,
		}),
	)

	filename, err := templates.New(data).
		Replace(replacements).
		Apply(filename)
	if err != nil {
		log.Printf("[WARN] asset: failed to template filename: %q", filename)
	}

	log.Printf("[DEBUG] asset: templated filename: -> %q", filename)
	return filename
}

func (c GitHub) HasPluginBlock() bool {
	return c.Plugin != nil
}

func (c GitHub) HasCommandBlock() bool {
	return c.Command != nil
}

func (c GitHub) HasReleaseBlock() bool {
	return c.Release != nil
}

func (c GitHub) GetPluginBlock() Plugin {
	if c.HasPluginBlock() {
		return *c.Plugin
	}
	return Plugin{}
}

func (c GitHub) GetCommandBlock() Command {
	if c.HasCommandBlock() {
		return *c.Command
	}
	return Command{}
}

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
	if c.IsGHExtension() {
		return c.As.GHExtension.GetHome()
	}
	return filepath.Join(os.Getenv("HOME"), ".afx", "github.com", c.Owner, c.Repo)
}

func (c GitHub) GetDependsOn() []string {
	return c.DependsOn
}

func (c GitHub) GetResource() state.Resource {
	return getResource(c)
}

func (c GitHub) Check(ctx context.Context, status chan<- Status) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("[DEBUG] canceled")
		return nil
	default:
		// go next
	}

	switch {
	case c.Release == nil:
		// TODO: Check git commit
		status <- Status{Name: c.GetName(), Done: true, Err: false, Message: "(github)", NoColor: true}
		return nil
	case c.Release != nil:
		report, err := c.checkUpdates(ctx)
		if err != nil {
			err = errors.Wrapf(err, "%s: failed to check release version", c.Name)
		}
		status <- Status{Name: c.GetName(), Done: true, Err: err != nil, Message: report.message}
		return err
	}

	status <- Status{Name: c.GetName(), Done: true, Err: false}
	return nil
}

type report struct {
	message string
	current *semver.Version
	next    *semver.Version
}

func (c GitHub) checkUpdates(ctx context.Context) (report, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	red := color.New(color.FgRed).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()

	tag := c.Release.Tag
	switch tag {
	case "latest", "stable", "nightly":
		return report{message: tag}, nil
	case "":
		return report{message: "(tag not set)"}, nil
	}

	release, err := github.NewRelease(
		ctx, c.Owner, c.Repo, "latest",
		github.WithWorkdir(c.GetHome()),
	)
	if err != nil {
		return report{
			message: fmt.Sprintf("%s %s", red("error!"), err),
		}, err
	}

	current, err := semver.NewVersion(tag)
	if err != nil {
		return report{}, nil
	}

	next, err := semver.NewVersion(release.Tag)
	if err != nil {
		return report{}, nil
	}

	switch current.Compare(next) {
	case -1:
		return report{
			message: fmt.Sprintf("%s v%s -> v%s",
				yellow("new!"), current, next),
		}, nil
	case 0:
		return report{message: "up-to-date"}, nil
	default:
		return report{}, errors.New("invalid version comparison")
	}
}

func ValidateGHExtension(fl validator.FieldLevel) bool {
	return fl.Field().String() == "" || strings.HasPrefix(fl.Field().String(), "gh-")
}

func (c GitHub) IsGHExtension() bool {
	return c.As != nil && c.As.GHExtension != nil && c.As.GHExtension.Name != ""
}

type ghManifest struct {
	Owner    string `yaml:"owner"`
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Tag      string `yaml:"tag"`
	IsPinned bool   `yaml:"ispinned"`
	Path     string `yaml:"path"`
}

func (gh GHExtension) GetHome() string {
	base := filepath.Join(os.Getenv("HOME"), ".local", "share", "gh", "extensions")
	var ext string
	if gh.RenameTo == "" {
		ext = filepath.Join(base, gh.Name)
	} else {
		ext = filepath.Join(base, gh.RenameTo)
	}
	return ext
}

func (gh GHExtension) GetTag() string {
	if gh.Tag != "" {
		return gh.Tag
	}
	return "latest"
}

func (gh GHExtension) Install(ctx context.Context, owner, repo, tag string) error {
	available, _ := github.HasRelease(http.DefaultClient, owner, repo, tag)
	if available {
		err := gh.InstallFromRelease(ctx, owner, repo, tag)
		if err != nil {
			return fmt.Errorf("%w: %s: failed to get gh extension", err, gh.Name)
		}
	}

	ghHome := gh.GetHome()
	// ensure to create the parent dir of each gh extension's path
	_ = os.MkdirAll(filepath.Dir(ghHome), os.ModePerm)

	// make alias
	if gh.RenameTo != "" {
		if err := os.Symlink(
			filepath.Join(ghHome, gh.Name),
			filepath.Join(ghHome, gh.RenameTo),
		); err != nil {
			return fmt.Errorf("%w: failed to symlink as alise", err)
		}
	}

	if gh.GetTag() == "latest" {
		// in case of not making manifest yaml
		return nil
	}

	return gh.makeManifest(owner)
}

func (gh GHExtension) InstallFromRelease(ctx context.Context, owner, repo, tag string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Printf("[DEBUG] install from release: %s/%s (%s)", owner, repo, tag)
	release, err := github.NewRelease(
		ctx, owner, repo, tag,
		github.WithOverwrite(),
		github.WithWorkdir(gh.GetHome()),
	)
	if err != nil {
		return err
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

func (gh GHExtension) makeManifest(owner string) error {
	// https://github.com/cli/cli/blob/c9a2d85793c4cef026d5bb941b3ac4121c81ae10/pkg/cmd/extension/manager.go#L424-L451
	manifest := ghManifest{
		Name:     gh.Name,
		Owner:    owner,
		Host:     "github.com",
		Path:     gh.GetHome(),
		Tag:      gh.GetTag(),
		IsPinned: false,
	}
	bs, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	manifestPath := filepath.Join(gh.GetHome(), "manifest.yml")
	f, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open manifest for writing: %w", err)
	}
	defer f.Close()

	_, err = f.Write(bs)
	if err != nil {
		return fmt.Errorf("failed write manifest file: %w", err)
	}
	return nil
}

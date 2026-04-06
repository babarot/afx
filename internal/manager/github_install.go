package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/babarot/afx/internal/data"
	"github.com/babarot/afx/internal/git"
	"github.com/babarot/afx/internal/github"
	"github.com/babarot/afx/internal/runner"
	"github.com/babarot/afx/internal/templates"
)

// Clone runs git clone or fetches+checkouts if already cloned.
func (c GitHub) Clone(ctx context.Context) error {
	gitCmd := git.NewRunner()

	var opt GitHubOption
	if c.Option != nil {
		opt = *c.Option
	}

	url := fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo)

	_, err := os.Stat(c.GetHome())
	switch {
	case os.IsNotExist(err):
		args := []string{"clone", "--no-tags"}
		if opt.Depth > 0 {
			args = append(args, "--depth", strconv.Itoa(opt.Depth))
		}
		if c.Branch != "" {
			args = append(args, "--branch", c.Branch)
		}
		args = append(args, url, c.GetHome())

		if _, err := gitCmd.Run(ctx, args...); err != nil {
			if git.IsAuthError(err) {
				return fmt.Errorf("%s: authentication failed. Please set GITHUB_TOKEN or configure git credentials for private repositories: %w", c.GetName(), err)
			}
			return fmt.Errorf("%s: failed to clone repository: %w", c.GetName(), err)
		}
	default:
		// Already cloned: fetch + checkout if branch is specified
		if c.Branch != "" {
			fetchArgs := []string{"fetch", "origin", c.Branch, "--no-tags", "--force"}
			if opt.Depth > 0 {
				fetchArgs = append(fetchArgs, "--depth", strconv.Itoa(opt.Depth))
			}

			if _, err := gitCmd.RunInDir(ctx, c.GetHome(), fetchArgs...); err != nil {
				return fmt.Errorf("%s: failed to fetch repository: %w", c.GetName(), err)
			}

			if _, err := gitCmd.RunInDir(ctx, c.GetHome(), "checkout", "--force", c.Branch); err != nil {
				return fmt.Errorf("%s: failed to checkout: %w", c.GetName(), err)
			}
		}
	}

	return nil
}

// Install installs from GitHub repository with git clone command
func (c GitHub) Install(ctx context.Context, status chan<- runner.Status) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("[DEBUG] canceled")
		return nil
	default:
		// Go installing step!
	}

	// gh extension: delegate entirely to gh CLI
	if c.IsGHExtension() {
		if c.GHRunner == nil {
			err := fmt.Errorf("%s: gh runner not initialized", c.Name)
			status <- runner.Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
		ext := c.As.GHExtension
		if err := ext.Install(ctx, c.GHRunner, c.Owner, c.Repo); err != nil {
			err = fmt.Errorf("%s: failed to install gh extension: %w", c.Name, err)
			status <- runner.Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
		if err := ext.createAliasSymlink(); err != nil {
			log.Printf("[WARN] %s: %v", c.Name, err)
		}
		status <- runner.Status{Name: c.GetName(), Done: true, Err: false}
		return nil
	}

	switch {
	case c.Release == nil:
		err := c.Clone(ctx)
		if err != nil {
			err = fmt.Errorf("%s: failed to clone repo: %w", c.Name, err)
			status <- runner.Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
	case c.Release != nil:
		err := c.InstallFromRelease(ctx)
		if err != nil {
			err = fmt.Errorf("%s: failed to get from release: %w", c.Name, err)
			status <- runner.Status{Name: c.GetName(), Done: true, Err: true}
			return err
		}
	}

	var errs []error

	if c.HasPluginBlock() {
		if err := c.Plugin.Install(c); err != nil {
			errs = append(errs, err)
		}
	}
	if c.HasCommandBlock() {
		if err := c.Command.Install(c); err != nil {
			errs = append(errs, err)
		}
	}

	status <- runner.Status{Name: c.GetName(), Done: true, Err: errors.Join(errs...) != nil}
	return errors.Join(errs...)
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
		return fmt.Errorf("%s: failed to download: %w", release.Name, err)
	}

	if err := release.Unarchive(asset); err != nil {
		return fmt.Errorf("%s: failed to unarchive: %w", release.Name, err)
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

func (c GitHub) Uninstall(ctx context.Context) error {
	// gh extension: delegate to gh CLI
	if c.IsGHExtension() {
		if c.GHRunner == nil {
			return fmt.Errorf("%s: gh runner not initialized", c.Name)
		}
		ext := c.As.GHExtension
		ext.removeAliasSymlink()
		return ext.Uninstall(ctx, c.GHRunner)
	}

	var errs []error

	delete := func(f string) {
		err := os.RemoveAll(f)
		if err != nil {
			errs = append(errs, err)
			return
		}
		log.Printf("[INFO] Delete %s\n", f)
	}

	if c.HasCommandBlock() {
		links, _ := c.Command.GetLink(c)
		for _, link := range links {
			delete(link.From)
			delete(link.To)
		}
	}

	delete(c.GetHome())
	return errors.Join(errs...)
}

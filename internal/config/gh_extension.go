package config

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"

	"github.com/babarot/afx/internal/errors"
	"github.com/babarot/afx/internal/github"
)

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

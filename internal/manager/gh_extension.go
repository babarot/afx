package manager

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/babarot/afx/internal/gh"
)

func ValidateGHExtension(fl validator.FieldLevel) bool {
	return fl.Field().String() == "" || strings.HasPrefix(fl.Field().String(), "gh-")
}

func (c GitHub) IsGHExtension() bool {
	return c.As != nil && c.As.GHExtension != nil && c.As.GHExtension.Name != ""
}

func (ext GHExtension) GetTag() string {
	if ext.Tag != "" {
		return ext.Tag
	}
	return "latest"
}

// GetHome returns the extension directory based on the canonical name (not rename-to).
// This path is managed by gh CLI.
func (ext GHExtension) GetHome() string {
	base := filepath.Join(os.Getenv("HOME"), ".local", "share", "gh", "extensions")
	return filepath.Join(base, ext.Name)
}

// GetAliasHome returns the symlink path when rename-to is specified.
// Returns empty string if rename-to is not set.
func (ext GHExtension) GetAliasHome() string {
	if ext.RenameTo == "" {
		return ""
	}
	base := filepath.Join(os.Getenv("HOME"), ".local", "share", "gh", "extensions")
	return filepath.Join(base, ext.RenameTo)
}

// Install installs the gh extension via the gh CLI.
func (ext GHExtension) Install(ctx context.Context, runner gh.Runner, owner, repo string) error {
	args := []string{"extension", "install", owner + "/" + repo}
	tag := ext.GetTag()
	if tag != "latest" {
		args = append(args, "--pin", tag)
	}

	log.Printf("[DEBUG] gh extension install: %s/%s (tag=%s)", owner, repo, tag)
	_, err := runner.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("gh extension install %s/%s: %w", owner, repo, err)
	}
	return nil
}

// Uninstall removes the gh extension via the gh CLI.
func (ext GHExtension) Uninstall(ctx context.Context, runner gh.Runner) error {
	log.Printf("[DEBUG] gh extension remove: %s", ext.Name)
	_, err := runner.Run(ctx, "extension", "remove", ext.Name)
	if err != nil {
		return fmt.Errorf("gh extension remove %s: %w", ext.Name, err)
	}
	return nil
}

// createAliasSymlink creates a rename-to symlink for the extension.
// It removes any existing symlink first to handle updates.
func (ext GHExtension) createAliasSymlink() error {
	if ext.RenameTo == "" {
		return nil
	}

	aliasHome := ext.GetAliasHome()

	// Remove existing symlink to handle updates (avoid EEXIST)
	if _, err := os.Lstat(aliasHome); err == nil {
		log.Printf("[DEBUG] removing existing alias symlink: %s", aliasHome)
		os.Remove(aliasHome)
	}

	target := ext.GetHome()
	log.Printf("[DEBUG] creating alias symlink: %s -> %s", aliasHome, target)
	if err := os.Symlink(target, aliasHome); err != nil {
		return fmt.Errorf("failed to create alias symlink %s -> %s: %w", aliasHome, target, err)
	}
	return nil
}

// removeAliasSymlink removes the rename-to symlink.
func (ext GHExtension) removeAliasSymlink() {
	if ext.RenameTo == "" {
		return
	}
	aliasHome := ext.GetAliasHome()
	log.Printf("[DEBUG] removing alias symlink: %s", aliasHome)
	os.Remove(aliasHome)
}

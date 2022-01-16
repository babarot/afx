package config

import (
	"context"
	"fmt"

	"github.com/mattn/go-shellwords"
)

// Installer is
type Installer interface {
	Install(context.Context, chan<- Status) error
	Uninstall(context.Context) error
	Installed() bool
}

// Loader is
type Loader interface {
	Init() error
}

// Manager is
type Manager interface {
	GetHome() string
	GetName() string
	GetType() string
	GetSlug() string
	GetURL() string

	Objects() ([]string, error)

	SetCommand(Command) Package
	SetPlugin(Plugin) Package

	HasPluginBlock() bool
	HasCommandBlock() bool
	GetPluginBlock() Plugin
	GetCommandBlock() Command
}

// Package is
type Package interface {
	Loader
	Manager
	Installer
}

// HasGitHubReleaseBlock is
func HasGitHubReleaseBlock(pkgs []Package) bool {
	for _, pkg := range pkgs {
		if pkg.Installed() {
			continue
		}
		switch pkg.GetType() {
		case "github":
			github := pkg.(*GitHub)
			if github.Release != nil {
				return true
			}
		}
	}
	return false
}

// HasSudoInCommandBuildSteps is
func HasSudoInCommandBuildSteps(pkgs []Package) bool {
	for _, pkg := range pkgs {
		if pkg.Installed() {
			continue
		}
		if !pkg.HasCommandBlock() {
			continue
		}
		command := pkg.GetCommandBlock()
		if !command.buildRequired() {
			continue
		}
		p := shellwords.NewParser()
		p.ParseEnv = true
		p.ParseBacktick = true
		for _, step := range command.Build.Steps {
			args, err := p.Parse(step)
			if err != nil {
				continue
			}
			switch args[0] {
			case "sudo":
				return true
			default:
				continue
			}
		}
	}
	return false
}

func Validate(pkgs []Package) error {
	done := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		name := pkg.GetName()
		_, already := done[name]
		if already {
			return fmt.Errorf("%s: duplicated", name)
		}
		done[name] = true
	}
	return nil
}

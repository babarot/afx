package config

import (
	"context"
	"path/filepath"

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

// Defined returns true if the package is already defined in config files
func Defined(pkgs []Package, arg Package) bool {
	for _, pkg := range pkgs {
		if pkg.GetName() == arg.GetName() {
			return true
		}
	}
	return false
}

// ConvertsFrom converts ...
func ConvertsFrom(pkgs ...Package) Config {
	var cfg Config
	for _, pkg := range pkgs {
		switch pkg.GetType() {
		case "github":
			github := pkg.(GitHub)
			cfg.GitHub = append(cfg.GitHub, &github)
		case "gist":
			gist := pkg.(Gist)
			cfg.Gist = append(cfg.Gist, &gist)
		case "http":
			http := pkg.(HTTP)
			cfg.HTTP = append(cfg.HTTP, &http)
		case "local":
			local := pkg.(Local)
			cfg.Local = append(cfg.Local, &local)
		}
	}
	return cfg
}

func ParseYAML(cfg Config) ([]Package, error) {
	var pkgs []Package
	for _, pkg := range cfg.GitHub {
		// TODO: Remove?
		if pkg.HasReleaseBlock() && !pkg.HasCommandBlock() {
			pkg.Command = &Command{
				Link: []*Link{
					&Link{From: filepath.Join("**", pkg.Release.Name)},
				},
			}
		}
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.Gist {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.Local {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.HTTP {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

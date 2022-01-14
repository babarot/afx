package config

import (
	"context"
	"log"
	"path/filepath"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/schema"
	"github.com/hashicorp/hcl/v2/gohcl"
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

// Decode is
func decode(data schema.Data) (Config, error) {
	var cfg Config

	decodedManifest, err := schema.Decode(data)
	if err != nil {
		log.Print("[ERROR] schema.Decode failed")
		return cfg, err
	}

	ctx, diags := decodedManifest.BuildContext(data.Body)

	decodeDiags := gohcl.DecodeBody(data.Body, ctx, &cfg)

	diags = append(diags, decodeDiags...)
	if diags.HasErrors() {
		log.Print("[ERROR] gohcl.DecodeBody failed")
		return cfg, errors.New(diags, data.Files)
	}

	return cfg, nil
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

// Parse is
func Parse(data schema.Data) ([]Package, error) {
	cfg, err := decode(data)
	if err != nil {
		return []Package{}, err
	}

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

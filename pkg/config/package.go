package config

import (
	"context"

	"github.com/babarot/afx/pkg/state"
	"github.com/mattn/go-shellwords"
)

// Installer is an interface related to installation of a package
type Installer interface {
	Install(context.Context, chan<- Status) error
	Uninstall(context.Context) error
	Installed() bool
	Check(context.Context, chan<- Status) error
}

// Loader is an interface related to initialize a package
type Loader interface {
	Init() error
}

// Handler is an interface of package handler
type Handler interface {
	GetHome() string
	GetName() string

	HasPluginBlock() bool
	HasCommandBlock() bool
	GetPluginBlock() Plugin
	GetCommandBlock() Command

	GetDependsOn() []string
	GetResource() state.Resource
}

// Package is an interface related to package itself
type Package interface {
	Loader
	Handler
	Installer
}

// HasGitHubReleaseBlock returns true if release block is included in one package at least
func HasGitHubReleaseBlock(pkgs []Package) bool {
	for _, pkg := range pkgs {
		github, ok := pkg.(*GitHub)
		if !ok {
			continue
		}
		if github.Release != nil {
			return true
		}
	}
	return false
}

// HasSudoInCommandBuildSteps returns true if sudo command is
// included in one build step of given package at least
func HasSudoInCommandBuildSteps(pkgs []Package) bool {
	for _, pkg := range pkgs {
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

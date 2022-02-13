package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/env"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/helpers/shell"
	"github.com/b4b4r07/afx/pkg/state"
)

type meta struct {
	Env       *env.Config
	Packages  []config.Package
	AppConfig *config.AppConfig
	State     *state.State
}

func (m *meta) init(args []string) error {
	root := filepath.Join(os.Getenv("HOME"), ".afx")
	cfgRoot := filepath.Join(os.Getenv("HOME"), ".config", "afx")
	cache := filepath.Join(root, "cache.json")

	files, err := config.WalkDir(cfgRoot)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to walk dir", cfgRoot)
	}

	app := &config.DefaultAppConfig
	for _, file := range files {
		cfg, err := config.Read(file)
		if err != nil {
			return errors.Wrapf(err, "%s: failed to read config", file)
		}
		pkgs, err := cfg.Parse()
		if err != nil {
			return errors.Wrapf(err, "%s: failed to parse config", file)
		}
		m.Packages = append(m.Packages, pkgs...)

		if cfg.AppConfig != nil {
			app = cfg.AppConfig
		}
	}

	m.AppConfig = app

	if err := config.Validate(m.Packages); err != nil {
		return errors.Wrap(err, "%s: failed to validate packages")
	}

	m.Env = env.New(cache)
	m.Env.Add(env.Variables{
		"AFX_CONFIG_PATH":  env.Variable{Value: cfgRoot},
		"AFX_LOG":          env.Variable{},
		"AFX_LOG_PATH":     env.Variable{},
		"AFX_COMMAND_PATH": env.Variable{Default: filepath.Join(os.Getenv("HOME"), "bin")},
		"AFX_SUDO_PASSWORD": env.Variable{
			Input: env.Input{
				When:    config.HasSudoInCommandBuildSteps(m.Packages),
				Message: "Please enter sudo command password",
				Help:    "Some packages build steps requires sudo command",
			},
		},
		"GITHUB_TOKEN": env.Variable{
			Input: env.Input{
				When:    config.HasGitHubReleaseBlock(m.Packages),
				Message: "Please type your GITHUB_TOKEN",
				Help:    "To fetch GitHub Releases, GitHub token is required",
			},
		},
	})

	log.Printf("[DEBUG] mkdir %s\n", root)
	os.MkdirAll(root, os.ModePerm)

	log.Printf("[DEBUG] mkdir %s\n", os.Getenv("AFX_COMMAND_PATH"))
	os.MkdirAll(os.Getenv("AFX_COMMAND_PATH"), os.ModePerm)

	s, err := state.Open(filepath.Join(root, "state.json"), m.Packages)
	if err != nil {
		return errors.Wrap(err, "faield to open state file")
	}
	m.State = s

	return nil
}

func (m *meta) Select() (config.Package, error) {
	var stdin, stdout bytes.Buffer

	cmd := shell.Shell{
		Stdin:   &stdin,
		Stdout:  &stdout,
		Stderr:  os.Stderr,
		Command: m.AppConfig.Filter.Command,
		Args:    m.AppConfig.Filter.Args,
		Env:     m.AppConfig.Filter.Env,
	}

	for _, pkg := range m.Packages {
		fmt.Fprintln(&stdin, pkg.GetName())
	}

	if err := cmd.Run(context.Background()); err != nil {
		return nil, err
	}

	search := func(name string) config.Package {
		for _, pkg := range m.Packages {
			if pkg.GetName() == name {
				return pkg
			}
		}
		return nil
	}

	for _, line := range strings.Split(stdout.String(), "\n") {
		if pkg := search(line); pkg != nil {
			return pkg, nil
		}
	}

	return nil, errors.New("pkg not found")
}

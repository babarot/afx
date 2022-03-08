package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/mattn/go-zglob"
)

// Plugin is
type Plugin struct {
	Sources        []string          `yaml:"sources" validate:"required"`
	Env            map[string]string `yaml:"env"`
	Snippet        string            `yaml:"snippet"`
	SnippetPrepare string            `yaml:"snippet-prepare"`
	If             string            `yaml:"if"`
}

func (p *Plugin) UnmarshalYAML(b []byte) error {
	type alias Plugin

	tmp := struct {
		*alias
		Sources []string `yaml:"sources"`
	}{
		alias: (*alias)(p),
	}

	if err := yaml.Unmarshal(b, &tmp); err != nil {
		return err
	}

	var sources []string
	for _, source := range tmp.Sources {
		sources = append(sources, expandTilda(os.ExpandEnv(source)))
	}
	p.Sources = sources

	return nil
}

// Installed returns true ...
func (p Plugin) Installed(pkg Package) bool {
	return len(p.GetSources(pkg)) > 0
}

// Install is
func (p Plugin) Install(pkg Package) error {
	return nil
}

func (p Plugin) GetSources(pkg Package) []string {
	var sources []string
	for _, src := range p.Sources {
		path := src
		if !filepath.IsAbs(src) {
			// basically almost all of sources are not abs path
			path = filepath.Join(pkg.GetHome(), src)
		}
		for _, src := range glob(path) {
			if _, err := os.Stat(src); errors.Is(err, os.ErrNotExist) {
				continue
			}
			sources = append(sources, src)
		}
	}
	return sources
}

// Init returns the file list which should be loaded as shell plugins
func (p Plugin) Init(pkg Package) error {
	if !pkg.Installed() {
		fmt.Printf("## package %q is not installed\n", pkg.GetName())
		return fmt.Errorf("%s: not installed", pkg.GetName())
	}

	shell := os.Getenv("AFX_SHELL")
	if shell == "" {
		shell = "bash"
	}

	if len(p.If) > 0 {
		cmd := exec.CommandContext(context.Background(), shell, "-c", p.If)
		err := cmd.Run()
		switch cmd.ProcessState.ExitCode() {
		case 0:
		default:
			log.Printf("[ERROR] %s: plugin.if exit code is not zero, so stopped to init package", pkg.GetName())
			return fmt.Errorf("%s: returned non-zero value with evaluation of `if` field: %w", pkg.GetName(), err)
		}
	}

	if s := p.SnippetPrepare; s != "" {
		fmt.Printf("%s\n", s)
	}

	sources := p.GetSources(pkg)
	if len(sources) == 0 {
		return fmt.Errorf("%s: failed to get sources", pkg.GetName())
	}

	for _, src := range sources {
		fmt.Printf("source %s\n", src)
	}

	for k, v := range p.Env {
		switch k {
		case "PATH":
			// avoid overwriting PATH
			v = fmt.Sprintf("$PATH:%s", expandTilda(v))
		default:
			// through
		}
		fmt.Printf("export %s=%q\n", k, v)
	}

	if s := p.Snippet; s != "" {
		fmt.Printf("%s\n", s)
	}

	return nil
}

func glob(path string) []string {
	var matches, sources []string
	var err error

	matches, err = filepath.Glob(path)
	if err == nil {
		sources = append(sources, matches...)
	}
	matches, err = zglob.Glob(path)
	if err == nil {
		sources = append(sources, matches...)
	}

	m := make(map[string]bool)
	unique := []string{}

	for _, source := range sources {
		if !m[source] {
			m[source] = true
			unique = append(unique, source)
		}
	}

	return unique
}

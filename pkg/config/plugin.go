package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattn/go-zglob"
)

// Plugin is
type Plugin struct {
	Sources []string          `yaml:"sources"`
	Env     map[string]string `yaml:"env,optional"`
	Snippet string            `yaml:"snippet,optional"`
}

// Installed returns true ...
func (p Plugin) Installed(pkg Package) bool {
	for _, source := range p.Sources {
		matches := glob(filepath.Join(pkg.GetHome(), source))
		if len(matches) == 0 {
			return false
		}
	}
	return true
}

// Install is
func (p Plugin) Install(pkg Package) error {
	return nil
}

// Init returns the file list which should be loaded as shell plugins
func (p Plugin) Init(pkg Package) error {
	if !pkg.Installed() {
		msg := fmt.Sprintf("package %s.%s is not installed, so skip to init",
			pkg.GetType(), pkg.GetName())
		fmt.Printf("## %s\n", msg)
		return errors.New(msg)
	}

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

	if len(sources) == 0 {
		return errors.New("no source files")
	}

	for _, src := range sources {
		fmt.Printf("source %s\n", src)
	}

	for k, v := range p.Env {
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

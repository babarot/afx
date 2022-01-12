package config

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/mattn/go-zglob"
)

// Plugin is
type Plugin struct {
	Sources   []string          `hcl:"sources"`
	Env       map[string]string `hcl:"env,optional"`
	LoadBlock *Load             `hcl:"load,block"`
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

	for _, src := range p.Sources {
		if !filepath.IsAbs(src) {
			sources := glob(filepath.Join(pkg.GetHome(), src))
			if len(sources) == 0 {
				log.Printf("[ERROR] %s: failed to get with glob/zglob\n", pkg.GetName())
				continue
			}
			for _, src := range sources {
				fmt.Printf("source %s\n", src)
			}
		}
	}

	for k, v := range p.Env {
		fmt.Printf("export %s=%q\n", k, v)
	}

	if p.LoadBlock != nil {
		for _, script := range p.LoadBlock.Scripts {
			fmt.Printf("%s\n", script)
		}
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

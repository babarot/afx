package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/b4b4r07/afx/pkg/dependency"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

// Config structure for file describing deployment. This includes the module source, inputs
// dependencies, backend etc. One config element is connected to a single deployment
type Config struct {
	GitHub []*GitHub `yaml:"github,omitempty"`
	Gist   []*Gist   `yaml:"gist,omitempty"`
	Local  []*Local  `yaml:"local,omitempty"`
	HTTP   []*HTTP   `yaml:"http,omitempty"`

	Main *Main `yaml:"main,omitempty"`
}

// Main represents configurations of this application itself
type Main struct {
	Shell     string            `yaml:"shell"`
	FilterCmd string            `yaml:"filter_command"`
	Env       map[string]string `yaml:"env"`
}

// DefaultMain is default settings of Main
// Basically this will be overridden by user config if given
var DefaultMain Main = Main{
	Shell:     "bash",
	FilterCmd: "fzf --ansi --no-preview --height=50% --reverse",
	Env:       map[string]string{},
}

// Read reads yaml file based on given path
func Read(path string) (Config, error) {
	log.Printf("[INFO] Reading config %s...", path)

	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	validate := validator.New()
	d := yaml.NewDecoder(
		bufio.NewReader(f),
		yaml.DisallowUnknownField(),
		yaml.DisallowDuplicateKey(),
		yaml.Validator(validate),
	)
	if err := d.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, err
}

func parse(cfg Config) []Package {
	var pkgs []Package

	for _, pkg := range cfg.GitHub {
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

	return pkgs
}

// Parse parses a config given via yaml files and converts it into package interface
func (c Config) Parse() ([]Package, error) {
	log.Printf("[INFO] Parsing config...")
	// TODO: divide from parse()
	return parse(c), nil
}

func visitYAML(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "%s: failed to visit", path)
		}
		switch filepath.Ext(path) {
		case ".yaml", ".yml":
			*files = append(*files, path)
		}
		return nil
	}
}

// WalkDir walks given directory path and returns full-path of all yaml files
func WalkDir(path string) ([]string, error) {
	var files []string
	fi, err := os.Stat(path)
	if err != nil {
		return files, err
	}
	if fi.IsDir() {
		return files, filepath.Walk(path, visitYAML(&files))
	}
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		files = append(files, path)
	default:
		log.Printf("[WARN] %s: found but cannot be loaded. yaml is only allowed\n", path)
	}
	return files, nil
}

func Sort(given []Package) ([]Package, error) {
	var pkgs []Package
	var graph dependency.Graph

	table := map[string]Package{}

	for _, pkg := range given {
		table[pkg.GetName()] = pkg
	}

	var errs errors.Errors
	for name, pkg := range table {
		dependencies := pkg.GetDependsOn()
		for _, dep := range pkg.GetDependsOn() {
			if _, ok := table[dep]; !ok {
				errs.Append(
					fmt.Errorf("%q: not valid package name in depends-on: %s", dep, pkg.GetName()),
				)
			}
		}
		graph = append(graph, dependency.NewNode(name, dependencies...))
	}

	if dependency.Has(graph) {
		log.Printf("[DEBUG] dependency graph is here: \n%s", graph)
	}

	resolved, err := dependency.Resolve(graph)
	if err != nil {
		return pkgs, errors.Wrap(err, "failed to resolve dependency graph")
	}

	for _, node := range resolved {
		pkgs = append(pkgs, table[node.Name])
	}

	return pkgs, errs.ErrorOrNil()
}

// Validate validates if packages are not violated some rules
func Validate(pkgs []Package) error {
	m := make(map[string]bool)
	var list []string

	for _, pkg := range pkgs {
		name := pkg.GetName()
		_, exist := m[name]
		if exist {
			list = append(list, name)
			continue
		}
		m[name] = true
	}

	if len(list) > 0 {
		return fmt.Errorf("duplicated packages: [%s]", strings.Join(list, ","))
	}

	return nil
}

func getResource(pkg Package) state.Resource {
	var paths []string

	// repository existence is also one of the path resource
	paths = append(paths, pkg.GetHome())

	if pkg.HasPluginBlock() {
		plugin := pkg.GetPluginBlock()
		paths = append(paths, plugin.GetSources(pkg)...)
	}

	if pkg.HasCommandBlock() {
		command := pkg.GetCommandBlock()
		links, _ := command.GetLink(pkg)
		for _, link := range links {
			paths = append(paths, link.From)
			paths = append(paths, link.To)
		}
	}

	var ty string
	var version string
	var id string

	switch pkg := pkg.(type) {
	case GitHub:
		ty = "GitHub"
		if pkg.HasReleaseBlock() {
			ty = "GitHub Release"
			version = pkg.Release.Tag
		}
		id = fmt.Sprintf("github.com/%s/%s", pkg.Owner, pkg.Repo)
		if pkg.HasReleaseBlock() {
			id = fmt.Sprintf("github.com/release/%s/%s", pkg.Owner, pkg.Repo)
		}
	case Gist:
		ty = "Gist"
		id = fmt.Sprintf("gist.github.com/%s/%s", pkg.Owner, pkg.ID)
	case Local:
		ty = "Local"
		id = fmt.Sprintf("local/%s", pkg.Directory)
	case HTTP:
		ty = "HTTP"
		id = pkg.URL
	default:
		ty = "Unknown"
	}

	return state.Resource{
		ID:      id,
		Name:    pkg.GetName(),
		Home:    pkg.GetHome(),
		Type:    ty,
		Version: version,
		Paths:   paths,
	}
}

func (c Config) Get(args ...string) Config {
	var part Config
	for _, arg := range args {
		for _, github := range c.GitHub {
			if github.Name == arg {
				part.GitHub = append(part.GitHub, github)
			}
		}
		for _, gist := range c.Gist {
			if gist.Name == arg {
				part.Gist = append(part.Gist, gist)
			}
		}
		for _, local := range c.Local {
			if local.Name == arg {
				part.Local = append(part.Local, local)
			}
		}
		for _, http := range c.HTTP {
			if http.Name == arg {
				part.HTTP = append(part.HTTP, http)
			}
		}
	}
	return part
}

func (c Config) Contains(args ...string) Config {
	var part Config
	for _, arg := range args {
		for _, github := range c.GitHub {
			if strings.Contains(github.Name, arg) {
				part.GitHub = append(part.GitHub, github)
			}
		}
		for _, gist := range c.Gist {
			if strings.Contains(gist.Name, arg) {
				part.Gist = append(part.Gist, gist)
			}
		}
		for _, local := range c.Local {
			if strings.Contains(local.Name, arg) {
				part.Local = append(part.Local, local)
			}
		}
		for _, http := range c.HTTP {
			if strings.Contains(http.Name, arg) {
				part.HTTP = append(part.HTTP, http)
			}
		}
	}
	return part
}

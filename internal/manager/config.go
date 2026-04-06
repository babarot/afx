package manager

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"

	"github.com/babarot/afx/internal/dependency"
)

// Config represents a parsed YAML configuration file containing package definitions.
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
	if err := validate.RegisterValidation("startswith-gh-if-not-empty", ValidateGHExtension); err != nil {
		return cfg, err
	}
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
		pkg.ParseURL()
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
			return fmt.Errorf("%w: %s: failed to visit", err, path)
		}
		switch filepath.Ext(path) {
		case ".yaml", ".yml":
			*files = append(*files, path)
		}
		return nil
	}
}

func CreateDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	} else if err != nil {
		return err
	}
	return nil
}

func resolvePath(path string) (string, bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return path, false, err
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		path, err = os.Readlink(path)
		if err != nil {
			return path, false, err
		}
		fi, err = os.Lstat(path)
		if err != nil {
			return path, false, err
		}
	}

	isDir := fi.IsDir()

	if filepath.IsAbs(path) {
		return path, isDir, nil
	}

	return path, isDir, err
}

// WalkDir walks given directory path and returns full-path of all yaml files
func WalkDir(path string) ([]string, error) {
	var files []string
	path, isDir, err := resolvePath(path)
	if err != nil {
		return files, err
	}
	if isDir {
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

	var errs []error
	for name, pkg := range table {
		dependencies := pkg.GetDependsOn()
		for _, dep := range pkg.GetDependsOn() {
			if _, ok := table[dep]; !ok {
				errs = append(errs,
					fmt.Errorf("%s: not valid package name in depends-on: %q", pkg.GetName(), dep),
				)
			}
		}
		graph = append(graph, dependency.NewNode(name, dependencies...))
	}
	if err := errors.Join(errs...); err != nil {
		return pkgs, err
	}

	if dependency.Has(graph) {
		log.Printf("[DEBUG] dependency graph is here: \n%s", graph)
	}

	resolved, err := dependency.Resolve(graph)
	if err != nil {
		return pkgs, fmt.Errorf("%w: failed to resolve dependency graph", err)
	}

	for _, node := range resolved {
		pkgs = append(pkgs, table[node.Name])
	}

	return pkgs, nil
}

// Validate validates if packages are not violated some rules
// Validate validates if packages are not violated some rules
func Validate(pkgs []Package) error {
	m := make(map[string]bool)
	var list []string
	var errs []error

	for _, pkg := range pkgs {
		name := pkg.GetName()
		if m[name] {
			list = append(list, name)
		}
		m[name] = true

		// GitHub-specific: command block is required when release block is present
		if gh, ok := pkg.(*GitHub); ok {
			if gh.Release != nil && gh.Command == nil {
				errs = append(errs, fmt.Errorf("%s: command block is required when release block is present", name))
			}
		}
	}

	if len(list) > 0 {
		errs = append(errs, fmt.Errorf("duplicated packages: [%s]", strings.Join(list, ",")))
	}

	return errors.Join(errs...)
}

func (c Config) Get(args ...string) Config {
	return c.filter(func(name, arg string) bool { return name == arg }, args...)
}

func (c Config) Contains(args ...string) Config {
	return c.filter(func(name, arg string) bool { return strings.Contains(name, arg) }, args...)
}

func (c Config) filter(match func(name, arg string) bool, args ...string) Config {
	var part Config
	for _, arg := range args {
		for _, pkg := range c.GitHub {
			if match(pkg.Name, arg) {
				part.GitHub = append(part.GitHub, pkg)
			}
		}
		for _, pkg := range c.Gist {
			if match(pkg.Name, arg) {
				part.Gist = append(part.Gist, pkg)
			}
		}
		for _, pkg := range c.Local {
			if match(pkg.Name, arg) {
				part.Local = append(part.Local, pkg)
			}
		}
		for _, pkg := range c.HTTP {
			if match(pkg.Name, arg) {
				part.HTTP = append(part.HTTP, pkg)
			}
		}
	}
	return part
}

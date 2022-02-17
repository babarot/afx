package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/b4b4r07/afx/pkg/dependency"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

// Config structure for file describing deployment. This includes the module source, inputs
// dependencies, backend etc. One config element is connected to a single deployment
type Config struct {
	GitHub []*GitHub `yaml:"github"`
	Gist   []*Gist   `yaml:"gist"`
	Local  []*Local  `yaml:"local"`
	HTTP   []*HTTP   `yaml:"http"`

	AppConfig *AppConfig `yaml:"config"`
}

// AppConfig represents configurations of this application itself
type AppConfig struct {
	Shell  string `yaml:"shell"`
	Filter Filter `yaml:"filter"`
}

// Filter represents filter command. A filter command means command-line
// fuzzy finder, e.g. fzf
type Filter struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

// DefaultAppConfig is default settings of AppConfig
// Basically this will be overridden by user config if given
var DefaultAppConfig AppConfig = AppConfig{
	Shell: "bash",
	Filter: Filter{
		Command: "fzf",
		Args:    []string{"--ansi", "--no-preview", "--height=50%", "--reverse"},
	},
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
		// TODO: Remove?
		if pkg.HasReleaseBlock() {
			msg := fmt.Sprintf(
				"[DEBUG] %s: added '**/%s' to link.from to complete missing links in github release",
				pkg.GetName(), pkg.Release.Name,
			)
			defaultLinks := []*Link{
				{From: filepath.Join("**", pkg.Release.Name)},
			}
			if pkg.HasCommandBlock() {
				links, _ := pkg.Command.GetLink(pkg)
				if len(links) == 0 {
					log.Printf(msg)
					pkg.Command.Link = defaultLinks
				}
			} else {
				log.Printf(msg)
				pkg.Command = &Command{Link: defaultLinks}
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

	return pkgs
}

// Parse parses a config given via yaml files and converts it into package interface
func (c Config) Parse() ([]Package, error) {
	log.Printf("[INFO] Parsing config...")

	var pkgs []Package

	parsed := parse(c)

	table := map[string]Package{}
	for _, pkg := range parsed {
		table[pkg.GetName()] = pkg
	}

	var errs errors.Errors
	var graph dependency.Graph

	for name, pkg := range table {
		dependencies := pkg.GetDependsOn()
		for _, dep := range dependencies {
			if !existence(parsed, dep) {
				errs.Append(
					fmt.Errorf("%q: not valid package name in depends-on: %s", dep, pkg.GetName()),
				)
			}
		}
		graph = append(graph, dependency.NewNode(name, dependencies...))
	}

	if errs.ErrorOrNil() != nil {
		return pkgs, errs.ErrorOrNil()
	}

	if dependency.Has(graph) {
		log.Printf("[DEBUG] dependency graph is here: %s", graph)
	}

	resolved, err := dependency.Resolve(graph)
	if err != nil {
		return pkgs, errors.Wrap(err, "failed to resolve dependency graph")
	}

	for _, node := range resolved {
		pkgs = append(pkgs, table[node.Name])
	}

	return pkgs, nil
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

func existence(pkgs []Package, name string) bool {
	for _, pkg := range pkgs {
		if pkg.GetName() == name {
			return true
		}
	}
	return false
}

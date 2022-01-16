package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Config structure for file describing deployment. This includes the module source, inputs
// dependencies, backend etc. One config element is connected to a single deployment
type Config struct {
	GitHub []*GitHub `yaml:"github,block"`
	Gist   []*Gist   `yaml:"gist,block"`
	Local  []*Local  `yaml:"local,block"`
	HTTP   []*HTTP   `yaml:"http,block"`
}

func Read(path string) (Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	d := yaml.NewDecoder(
		bufio.NewReader(f),
		yaml.DisallowUnknownField(),
		yaml.DisallowDuplicateKey(),
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
		if pkg.HasReleaseBlock() && !pkg.HasCommandBlock() {
			pkg.Command = &Command{
				Link: []*Link{
					{From: filepath.Join("**", pkg.Release.Name)},
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

	return pkgs
}

func (c Config) Parse() ([]Package, error) {
	return parse(c), nil
}

func visitYAML(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(path) {
		case ".yaml", ".yml":
			*files = append(*files, path)
		}
		return nil
	}
}

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

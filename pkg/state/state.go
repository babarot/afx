package state

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/b4b4r07/afx/pkg/config"
)

type Body struct {
	Resources map[string]Resource `json:"resources"`
}

type State struct {
	// records itself of state file
	body Body

	packages map[string]config.Package
	path     string

	// No record in state file
	Additions []config.Package
	// Exists but resource paths has something problem
	// so it's likely to have had problem when installing before
	Readditions []config.Package
	// Exists in state file but no in config file
	// so maybe users had deleted the package from config file
	Deletions []Resource
}

type Resource struct {
	Name  string   `json:"name"`
	Home  string   `json:"home"`
	Paths []string `json:"paths"`
}

func construct(pkgs []config.Package, s *State) {
	s.body.Resources = map[string]Resource{}
	for _, pkg := range pkgs {
		var paths []string
		if pkg.HasPluginBlock() {
			plugin := pkg.GetPluginBlock()
			for _, src := range plugin.GetSources(pkg) {
				paths = append(paths, src)
			}
		}
		if pkg.HasCommandBlock() {
			command := pkg.GetCommandBlock()
			links, err := command.GetLink(pkg)
			if err != nil {
				// no handling
			}
			for _, link := range links {
				paths = append(paths, link.From)
				paths = append(paths, link.To)
			}
		}
		s.body.Resources[pkg.GetName()] = Resource{
			Name:  pkg.GetName(),
			Home:  pkg.GetHome(),
			Paths: paths,
		}
	}
}

func Open(path string, pkgs []config.Package) (State, error) {
	s := State{path: path}
	s.packages = map[string]config.Package{}
	for _, pkg := range pkgs {
		s.packages[pkg.GetName()] = pkg
	}

	_, err := os.Stat(path)
	if err != nil {
		construct(pkgs, &s)
		return s, nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	if err := json.Unmarshal(content, &s.body); err != nil {
		return s, err
	}

	s.Additions = s.listAdditions()
	s.Readditions = s.listReadditions()
	s.Deletions = s.listDeletions()

	return s, nil
}

func (s *State) listAdditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		if _, ok := s.body.Resources[pkg.GetName()]; !ok {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

func (s *State) listReadditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		resource, ok := s.body.Resources[pkg.GetName()]
		if !ok {
			// if it's not in state file,
			// it means we need to install not reinstall
			continue
		}
		if !resource.exists() {
			pkgs = append(pkgs, pkg)
			continue
		}
	}
	return pkgs
}

func (s *State) listDeletions() []Resource {
	var resources []Resource
	for _, resource := range s.body.Resources {
		if _, ok := s.packages[resource.Name]; !ok {
			resources = append(resources, resource)
		}
	}
	return resources
}

func (s *State) Save() error {
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(s.body)
}

func (e Resource) exists() bool {
	for _, path := range e.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

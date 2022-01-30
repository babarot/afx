package state

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/b4b4r07/afx/pkg/config"
)

type Self struct {
	Resources map[string]Resource `json:"resources"`
}

type State struct {
	// State itself of state file
	Self

	packages map[string]config.Package
	path     string
	mu       sync.RWMutex

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

func (e Resource) exists() bool {
	for _, path := range e.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func add(pkg config.Package, s *State) {
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
	s.Resources[pkg.GetName()] = Resource{
		Name:  pkg.GetName(),
		Home:  pkg.GetHome(),
		Paths: paths,
	}
}

func remove(name string, s *State) {
	resources := map[string]Resource{}
	for _, resource := range s.Resources {
		if resource.Name == name {
			continue
		}
		resources[resource.Name] = resource
	}
	s.Resources = resources
}

func Open(path string, pkgs []config.Package) (*State, error) {
	s := State{path: path, mu: sync.RWMutex{}}
	s.packages = map[string]config.Package{}
	for _, pkg := range pkgs {
		s.packages[pkg.GetName()] = pkg
	}

	_, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		s.Resources = map[string]Resource{}
		for _, pkg := range pkgs {
			add(pkg, &s)
		}
	default:
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return &s, err
		}
		if err := json.Unmarshal(content, &s.Self); err != nil {
			return &s, err
		}
	}

	s.Additions = s.listAdditions()
	s.Readditions = s.listReadditions()
	s.Deletions = s.listDeletions()

	return &s, s.save()
}

func (s *State) listAdditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		if _, ok := s.Resources[pkg.GetName()]; !ok {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

func (s *State) listReadditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		resource, ok := s.Resources[pkg.GetName()]
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
	for _, resource := range s.Resources {
		if _, ok := s.packages[resource.Name]; !ok {
			resources = append(resources, resource)
		}
	}
	return resources
}

func (s *State) save() error {
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(s.Self)
}

func (s *State) Add(pkg config.Package) {
	s.mu.Lock()
	defer s.mu.Unlock()

	add(pkg, s)
	s.save()
}

func (s *State) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	remove(name, s)
	s.save()
}

package state

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
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
	// Something changes happened between config file and state file
	// Currently only version (github.release.tag) is detected as changes
	Changes []config.Package
}

type Resource struct {
	Name    string   `json:"name"`
	Home    string   `json:"home"`
	Version string   `json:"version"`
	Paths   []string `json:"paths"`
}

func (e Resource) exists() bool {
	if len(e.Paths) == 0 {
		return false
	}
	for _, path := range e.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func toResource(pkg config.Package) Resource {
	var paths []string

	if pkg.HasPluginBlock() {
		plugin := pkg.GetPluginBlock()
		paths = append(paths, plugin.GetSources(pkg)...)
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

	version := ""
	switch v := pkg.(type) {
	case *config.GitHub:
		if v.HasReleaseBlock() {
			version = v.Release.Tag
		}
	}

	log.Printf("[DEBUG] %s: add paths to state: %#v", pkg.GetName(), paths)
	return Resource{
		Name:    pkg.GetName(),
		Home:    pkg.GetHome(),
		Version: version,
		Paths:   paths,
	}
}

func add(pkg config.Package, s *State) {
	s.Resources[pkg.GetName()] = toResource(pkg)
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

func update(pkg config.Package, s *State) {
	name := pkg.GetName()
	_, ok := s.Resources[name]
	if !ok {
		// not found
		return
	}
	s.Resources[name] = toResource(pkg)
}

func (s *State) save() error {
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(s.Self)
}

func (s *State) listChanges() []config.Package {
	var pkgs []config.Package
	for _, resource := range s.Resources {
		if resource.Version == "" {
			// not target resource
			continue
		}
		pkg, ok := s.packages[resource.Name]
		if !ok {
			// something wrong happend
			continue
		}
		version := pkg.(*config.GitHub).Release.Tag
		if resource.Version != version {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
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

func Open(path string, pkgs []config.Package) (*State, error) {
	s := State{
		packages: map[string]config.Package{},
		path:     path,
		mu:       sync.RWMutex{},
	}
	for _, pkg := range pkgs {
		s.packages[pkg.GetName()] = pkg
	}

	_, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// just create empty state when state has not been created yet
		s.Resources = map[string]Resource{}
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
	s.Changes = s.listChanges()

	return &s, s.save()
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

func (s *State) Update(pkg config.Package) {
	s.mu.Lock()
	defer s.mu.Unlock()

	update(pkg, s)
	s.save()
}

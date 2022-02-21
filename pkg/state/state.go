package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/google/go-cmp/cmp"
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
	//
	NoChanges []config.Package
}

type Resource struct {
	Name    string   `json:"name"`
	Home    string   `json:"home"`
	Type    string   `json:"type"`
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

func getStateName(pkg config.Package) string {
	var name string

	switch pkg := pkg.(type) {
	case *config.GitHub:
		name = fmt.Sprintf("github.com/%s/%s", pkg.Owner, pkg.Repo)
		if pkg.HasReleaseBlock() {
			name = fmt.Sprintf("github.com/release/%s/%s", pkg.Owner, pkg.Repo)
		}
	case *config.Gist:
		name = fmt.Sprintf("gist.github.com/%s/%s", pkg.Owner, pkg.ID)
	case *config.Local:
		name = fmt.Sprintf("local/%s", pkg.Directory)
	case *config.HTTP:
		name = pkg.URL
	}

	return name
}

func toResource(pkg config.Package) Resource {
	var paths []string

	// repository existence is also one of the path resource
	paths = append(paths, pkg.GetHome())

	if pkg.HasPluginBlock() {
		plugin := pkg.GetPluginBlock()
		paths = append(paths, plugin.GetSources(pkg)...)
	}

	if pkg.HasCommandBlock() {
		command := pkg.GetCommandBlock()
		links, err := command.GetLink(pkg)
		if err != nil {
			// TODO: thinking about what to do here
			// no handling
		}
		for _, link := range links {
			paths = append(paths, link.From)
			paths = append(paths, link.To)
		}
	}

	var ty string
	var version string

	switch pkg := pkg.(type) {
	case *config.GitHub:
		ty = "GitHub"
		if pkg.HasReleaseBlock() {
			ty = "GitHub Release"
			version = pkg.Release.Tag
		}
	case *config.Gist:
		ty = "Gist"
	case *config.Local:
		ty = "Local"
	case *config.HTTP:
		ty = "HTTP"
	default:
		ty = "Unknown"
	}

	return Resource{
		Name:    getStateName(pkg),
		Home:    pkg.GetHome(),
		Type:    ty,
		Version: version,
		Paths:   paths,
	}
}

func add(pkg config.Package, s *State) {
	log.Printf("[DEBUG] %s: added to state", pkg.GetName())
	name := getStateName(pkg)
	s.Resources[name] = toResource(pkg)
}

func remove(name string, s *State) {
	resources := map[string]Resource{}
	for _, resource := range s.Resources {
		if resource.Name == name {
			continue
		}
		resources[resource.Name] = resource
	}
	log.Printf("[DEBUG] %s: removed from state", name)
	s.Resources = resources
}

func update(pkg config.Package, s *State) {
	name := getStateName(pkg)
	_, ok := s.Resources[name]
	if !ok {
		// not found
		return
	}
	log.Printf("[DEBUG] %s: updated in state", pkg.GetName())
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

func contains(pkgs []config.Package, name string) bool {
	for _, pkg := range pkgs {
		if pkg.GetName() == name {
			return true
		}
	}
	return false
}

func (s *State) listNoChanges() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		name := getStateName(pkg)
		if contains(s.listAdditions(), name) {
			continue
		}
		if contains(s.listReadditions(), name) {
			continue
		}
		if contains(s.listChanges(), name) {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func (s *State) listAdditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		name := getStateName(pkg)
		if _, ok := s.Resources[name]; !ok {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

func (s *State) listReadditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		name := getStateName(pkg)
		resource, ok := s.Resources[name]
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
		name := getStateName(pkg)
		s.packages[name] = pkg
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
	s.NoChanges = s.listNoChanges()

	// TODO: maybe better to separate to dedicated command etc?
	// this is needed to update state schema (e.g. adding new field)
	// but maybe it's danger a bit
	// so may be better to separate to dedicated command like `afx state refresh` etc
	// to run state operation explicitly
	if err := s.Refresh(); err != nil {
		log.Printf("[ERROR] there're some states or packages which needs operations: %v", err)
	}

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

func (s *State) List() ([]string, error) {
	_, err := os.Stat(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return []string{}, err
	default:
		content, err := ioutil.ReadFile(s.path)
		if err != nil {
			return []string{}, err
		}
		var state Self
		if err := json.Unmarshal(content, &state); err != nil {
			return []string{}, err
		}
		var items []string
		for k := range state.Resources {
			items = append(items, k)
		}
		return items, nil
	}
}

func (s *State) New() error {
	s.Resources = map[string]Resource{}
	for _, pkg := range s.packages {
		add(pkg, s)
	}
	return s.save()
}

func (s *State) Refresh() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	someChanges := len(s.Additions) > 0 ||
		len(s.Readditions) > 0 ||
		len(s.Changes) > 0 ||
		len(s.Deletions) > 0

	if someChanges {
		return errors.New("cannot refresh state")
	}

	done := false
	for _, pkg := range s.packages {
		name := getStateName(pkg)
		v1 := s.Resources[name]
		v2 := toResource(pkg)
		if diff := cmp.Diff(v1, v2); diff != "" {
			log.Printf("[DEBUG] refresh state to %s", diff)
			update(pkg, s)
			done = true
		}
	}

	if done {
		log.Printf("[DEBUG] refreshed state to update latest state schema")
	}

	return nil
}

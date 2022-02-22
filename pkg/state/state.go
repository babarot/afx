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

// ID is to prevent from detecting state changes unexpected by package name changing
// By using fixed string instead of package name, we can forcus on detecting the
// changes of only package contents itself.
type ID = string

type Self struct {
	Resources map[ID]Resource `json:"resources"`
}

type State struct {
	// State itself of state file
	Self

	packages map[ID]config.Package
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
	// All items recorded in state file. It means no changes between state file
	// and config file
	NoChanges []config.Package
}

type Resource struct {
	ID      ID       `json:"id"`
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

func getStateID(pkg config.Package) ID {
	var id string

	switch pkg := pkg.(type) {
	case *config.GitHub:
		id = fmt.Sprintf("github.com/%s/%s", pkg.Owner, pkg.Repo)
		if pkg.HasReleaseBlock() {
			id = fmt.Sprintf("github.com/release/%s/%s", pkg.Owner, pkg.Repo)
		}
	case *config.Gist:
		id = fmt.Sprintf("gist.github.com/%s/%s", pkg.Owner, pkg.ID)
	case *config.Local:
		id = fmt.Sprintf("local/%s", pkg.Directory)
	case *config.HTTP:
		id = pkg.URL
	}

	return id
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
		ID:      getStateID(pkg),
		Home:    pkg.GetHome(),
		Name:    pkg.GetName(),
		Type:    ty,
		Version: version,
		Paths:   paths,
	}
}

func add(pkg config.Package, s *State) {
	log.Printf("[DEBUG] %s: added to state", pkg.GetName())
	id := getStateID(pkg)
	s.Resources[id] = toResource(pkg)
}

func remove(id ID, s *State) {
	resources := map[ID]Resource{}
	for _, resource := range s.Resources {
		if resource.ID == id {
			continue
		}
		resources[resource.ID] = resource
	}
	log.Printf("[DEBUG] %s: removed from state", id)
	s.Resources = resources
}

func update(pkg config.Package, s *State) {
	id := getStateID(pkg)
	_, ok := s.Resources[id]
	if !ok {
		// not found
		return
	}
	log.Printf("[DEBUG] %s: updated in state", pkg.GetName())
	s.Resources[id] = toResource(pkg)
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
		pkg, ok := s.packages[resource.ID]
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
		if contains(s.listAdditions(), pkg.GetName()) {
			continue
		}
		if contains(s.listReadditions(), pkg.GetName()) {
			continue
		}
		if contains(s.listChanges(), pkg.GetName()) {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func (s *State) listAdditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		id := getStateID(pkg)
		if _, ok := s.Resources[id]; !ok {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

func (s *State) listReadditions() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		id := getStateID(pkg)
		resource, ok := s.Resources[id]
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
		if _, ok := s.packages[resource.ID]; !ok {
			resources = append(resources, resource)
		}
	}
	return resources
}

func Open(path string, pkgs []config.Package) (*State, error) {
	s := State{
		packages: map[ID]config.Package{},
		path:     path,
		mu:       sync.RWMutex{},
	}
	for _, pkg := range pkgs {
		id := getStateID(pkg)
		s.packages[id] = pkg
	}

	_, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// just create empty state when state has not been created yet
		s.Resources = map[ID]Resource{}
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

func (s *State) Remove(id ID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	remove(id, s)
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
		for id := range state.Resources {
			items = append(items, string(id))
		}
		return items, nil
	}
}

func (s *State) New() error {
	s.Resources = map[ID]Resource{}
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
		id := getStateID(pkg)
		v1 := s.Resources[id]
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

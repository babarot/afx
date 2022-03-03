package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"

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

	packages map[ID]Resource
	path     string
	mu       sync.RWMutex

	// No record in state file
	Additions []Resource

	// Exists in state file but no in config file
	// so maybe users had deleted the package from config file
	Deletions []Resource

	// Something changes happened between config file and state file
	// Currently only version (github.release.tag) is detected as changes
	Changes []Resource

	// All items recorded in state file. It means no changes between state file
	// and config file
	NoChanges []Resource
}

type Resourcer interface {
	GetResource() Resource
}

type Resource struct {
	ID      ID       `json:"id"`
	Name    string   `json:"name"`
	Home    string   `json:"home"`
	Type    string   `json:"type"`
	Version string   `json:"version"`
	Paths   []string `json:"paths"`
}

func (e Resource) GetResource() Resource {
	return e
}

func (e Resource) exists() bool {
	if len(e.Paths) == 0 {
		return false
	}
	for _, path := range e.Paths {
		if !exists(path) {
			return false
		}
	}
	return true
}

var exists = func(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

var ReadStateFile = func(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, pathError(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

var SaveStateFile = func(filename string) (io.Writer, error) {
	return os.Create(filename)
}

func pathError(err error) error {
	var pathError *os.PathError
	if errors.As(err, &pathError) && errors.Is(pathError.Err, syscall.ENOTDIR) {
		if p := findRegularFile(pathError.Path); p != "" {
			return fmt.Errorf("remove or rename regular file `%s` (must be a directory)", p)
		}

	}
	return err
}

func findRegularFile(p string) string {
	for {
		if s, err := os.Stat(p); err == nil && s.Mode().IsRegular() {
			return p
		}
		newPath := filepath.Dir(p)
		if newPath == p || newPath == "/" || newPath == "." {
			break
		}
		p = newPath
	}
	return ""
}

func add(r Resource, s *State) {
	log.Printf("[DEBUG] %s: added to state", r.Name)
	s.Resources[r.ID] = r
}

func remove(r Resource, s *State) {
	resources := map[ID]Resource{}
	for _, resource := range s.Resources {
		if resource.ID == r.ID {
			log.Printf("[DEBUG] %s: removed from state", r.Name)
			continue
		}
		resources[resource.ID] = resource
	}
	if len(s.Resources) == len(resources) {
		log.Printf("[WARN] %s: failed to remove from state", r.Name)
		return
	}
	s.Resources = resources
}

func update(r Resource, s *State) {
	_, ok := s.Resources[r.ID]
	if !ok {
		return
	}
	log.Printf("[DEBUG] %s: updated in state", r.Name)
	s.Resources[r.ID] = r
}

func (s *State) save() error {
	f, err := SaveStateFile(s.path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(s.Self)
}

func contains(resources []Resource, name string) bool {
	for _, resource := range resources {
		if resource.Name == name {
			return true
		}
	}
	return false
}

func (s *State) listChanges() []Resource {
	var resources []Resource
	for _, resource := range s.Resources {
		if resource.Version == "" {
			log.Printf("[TRACE] skip; version of %s is not set", resource.Name)
			continue
		}
		r, ok := s.packages[resource.ID]
		if !ok {
			log.Printf("[TRACE] skip; %s is not found in packages", resource.Name)
			continue
		}
		if resource.Version != r.Version {
			resources = append(resources, resource)
		}
	}
	return resources
}

func (s *State) listNoChanges() []Resource {
	var resources []Resource
	for _, resource := range s.packages {
		if contains(append(s.listAdditions(), s.listReadditions()...), resource.Name) {
			continue
		}
		if contains(s.listChanges(), resource.Name) {
			continue
		}
		resources = append(resources, resource)
	}
	return resources
}

func (s *State) listAdditions() []Resource {
	var resources []Resource
	for _, resource := range s.packages {
		if _, ok := s.Resources[resource.ID]; !ok {
			resources = append(resources, resource)
		}
	}
	return resources
}

func (s *State) listReadditions() []Resource {
	var resources []Resource
	for _, resource := range s.packages {
		resource, ok := s.Resources[resource.ID]
		if !ok {
			// if it's not in state file,
			// it means we need to install not reinstall
			continue
		}
		if !resource.exists() {
			resources = append(resources, resource)
			continue
		}
	}
	return resources
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

func Keys(resources []Resource) []string {
	var keys []string
	for _, resource := range resources {
		keys = append(keys, resource.Name)
	}
	return keys
}

func Open(path string, resourcers []Resourcer) (*State, error) {
	s := State{
		path:     path,
		packages: map[ID]Resource{},
		mu:       sync.RWMutex{},
	}

	for _, resourcer := range resourcers {
		resource := resourcer.GetResource()
		s.packages[resource.ID] = resource
	}

	content, err := ReadStateFile(path)
	if err != nil {
		return &s, err
	}

	if err := json.Unmarshal(content, &s.Self); err != nil {
		return &s, err
	}

	s.Additions = append(s.listAdditions(), s.listReadditions()...)
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

func (s *State) Add(resourcer Resourcer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	add(resourcer.GetResource(), s)
	s.save()
}

func (s *State) Remove(resourcer Resourcer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	remove(resourcer.GetResource(), s)
	s.save()
}

func (s *State) Update(resourcer Resourcer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	update(resourcer.GetResource(), s)
	s.save()
}

func (s *State) List() ([]Resource, error) {
	content, err := ReadStateFile(s.path)
	if err != nil {
		return []Resource{}, err
	}
	var state Self
	if err := json.Unmarshal(content, &state); err != nil {
		return []Resource{}, err
	}
	var resources []Resource
	for _, resource := range state.Resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

func (s *State) New() error {
	s.Resources = map[ID]Resource{}
	for _, resource := range s.packages {
		add(resource, s)
	}
	return s.save()
}

func (s *State) Refresh() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	someChanges := len(s.Additions) > 0 ||
		len(s.Changes) > 0 ||
		len(s.Deletions) > 0

	if someChanges {
		return errors.New("cannot refresh state")
	}

	done := false
	for _, resource := range s.packages {
		v1 := s.Resources[resource.ID]
		v2 := resource
		if diff := cmp.Diff(v1, v2); diff != "" {
			log.Printf("[DEBUG] refresh state to %s", diff)
			update(resource, s)
			done = true
		}
	}

	if done {
		log.Printf("[DEBUG] refreshed state to update latest state schema")
	}

	return nil
}

func Map(resources []Resource) map[ID]Resource {
	m := map[ID]Resource{}
	for _, resource := range resources {
		m[resource.Name] = resource
	}
	return m
}

func Slice(m map[ID]Resource) []Resource {
	var resources []Resource
	for _, resource := range m {
		resources = append(resources, resource)
	}
	return resources
}

func (s *State) Get(name string) (Resource, error) {
	for _, resource := range s.Resources {
		if resource.Name == name {
			return resource, nil
		}
	}
	return Resource{}, fmt.Errorf("%s: not found in state file", name)
}

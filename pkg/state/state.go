package state

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/b4b4r07/afx/pkg/config"
)

type State struct {
	Resources map[string]Resource `json:"resources"`

	Result Result `json:"-"`

	packages map[string]config.Package
	path     string
}

type Resource struct {
	Name  string   `json:"name"`
	Home  string   `json:"home"`
	Paths []string `json:"paths"`
}

type Result struct {
	NeedInstall   []config.Package
	NeedReinstall []config.Package
	NeedUninstall []Resource
}

func create(pkgs []config.Package, s *State) {
	s.Resources = map[string]Resource{}
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
				panic(err)
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
}

func Open(path string, pkgs []config.Package) (State, error) {
	s := State{path: path}
	s.packages = map[string]config.Package{}
	for _, pkg := range pkgs {
		s.packages[pkg.GetName()] = pkg
	}

	_, err := os.Stat(path)
	if err != nil {
		create(pkgs, &s)
		return s, nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	if err := json.Unmarshal(content, &s); err != nil {
		return s, err
	}

	s.Result.NeedInstall = s.listNeedInstall()
	s.Result.NeedReinstall = s.listNeedReinstall()
	s.Result.NeedUninstall = s.listNeedUninstall()

	return s, nil
}

func (s *State) listNeedReinstall() []config.Package {
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

func (s *State) listNeedInstall() []config.Package {
	var pkgs []config.Package
	for _, pkg := range s.packages {
		if _, ok := s.Resources[pkg.GetName()]; !ok {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

func (s *State) listNeedUninstall() []Resource {
	var resources []Resource
	for _, resource := range s.Resources {
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
	return json.NewEncoder(f).Encode(s)
}

func (e Resource) exists() bool {
	for _, path := range e.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

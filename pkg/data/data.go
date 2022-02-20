package data

import (
	"os"
	"runtime"
	"strings"
)

type Data struct {
	Env     Env
	Runtime Runtime
	Package Package
	Release Release
}

type Env map[string]string

type Runtime struct {
	Goos   string
	Goarch string
}

type PackageInterface interface {
	GetHome() string
	GetName() string
}

type Package struct {
	Name string
	Home string
}

type Release struct {
	Name string
	Tag  string
}

func New(fields ...func(*Data)) *Data {
	d := &Data{
		Package: Package{},
		Release: Release{},
		Env:     ToEnv(os.Environ()),
		Runtime: Runtime{
			Goos:   runtime.GOOS,
			Goarch: runtime.GOARCH,
		},
	}
	for _, f := range fields {
		f(d)
	}
	return d
}

func WithPackage(pkg PackageInterface) func(*Data) {
	return func(d *Data) {
		d.Package = Package{
			Home: pkg.GetHome(),
			Name: pkg.GetName(),
		}
	}
}

func WithRelease(r Release) func(*Data) {
	return func(d *Data) {
		d.Release = r
	}
}

// ToEnv converts a list of strings to an Env (aka a map[string]string).
func ToEnv(env []string) Env {
	r := Env{}
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)
		if len(p) != 2 || p[0] == "" {
			continue
		}
		r[p[0]] = p[1]
	}
	return r
}

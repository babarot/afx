package context

import (
	"os"
	"runtime"
	"strings"
)

type Context struct {
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

func New(fields ...func(*Context)) *Context {
	ctx := &Context{
		Package: Package{},
		Release: Release{},
		Env:     ToEnv(os.Environ()),
		Runtime: Runtime{
			Goos:   runtime.GOOS,
			Goarch: runtime.GOARCH,
		},
	}
	for _, f := range fields {
		f(ctx)
	}
	return ctx
}

func WithPackage(pkg PackageInterface) func(*Context) {
	return func(ctx *Context) {
		ctx.Package = Package{
			Home: pkg.GetHome(),
			Name: pkg.GetName(),
		}
	}
}

func WithRelease(r Release) func(*Context) {
	return func(ctx *Context) {
		ctx.Release = r
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

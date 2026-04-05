package manager

import (
	"fmt"

	"github.com/babarot/afx/internal/state"
)

func getResource(pkg Package) state.Resource {
	var paths []string

	// repository existence is also one of the path resource
	paths = append(paths, pkg.GetHome())

	if pkg.HasPluginBlock() {
		plugin := pkg.GetPluginBlock()
		paths = append(paths, plugin.GetSources(pkg)...)
	}

	if pkg.HasCommandBlock() {
		command := pkg.GetCommandBlock()
		links, _ := command.GetLink(pkg)
		for _, link := range links {
			paths = append(paths, link.From)
			paths = append(paths, link.To)
		}
	}

	var ty string
	var version string
	var id string

	switch pkg := pkg.(type) {
	case GitHub:
		ty = "GitHub"
		if pkg.HasReleaseBlock() {
			ty = "GitHub Release"
			version = pkg.Release.Tag
		}
		id = fmt.Sprintf("github.com/%s/%s", pkg.Owner, pkg.Repo)
		if pkg.HasReleaseBlock() {
			id = fmt.Sprintf("github.com/release/%s/%s", pkg.Owner, pkg.Repo)
		}
		if pkg.IsGHExtension() {
			ty = "GitHub (gh extension)"
			ext := pkg.As.GHExtension
			if alias := ext.GetAliasHome(); alias != "" {
				paths = append(paths, alias)
			}
		}
	case Gist:
		ty = "Gist"
		id = fmt.Sprintf("gist.github.com/%s/%s", pkg.Owner, pkg.ID)
	case Local:
		ty = "Local"
		id = fmt.Sprintf("local/%s", pkg.Directory)
	case HTTP:
		ty = "HTTP"
		id = pkg.URL
	default:
		ty = "Unknown"
	}

	return state.Resource{
		ID:      id,
		Name:    pkg.GetName(),
		Home:    pkg.GetHome(),
		Type:    ty,
		Version: version,
		Paths:   paths,
	}
}

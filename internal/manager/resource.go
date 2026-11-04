package manager

import (
	"github.com/babarot/afx/internal/state"
)

// ResourceMeta provides type-specific metadata for resource tracking.
// All concrete package types implement this interface.
type ResourceMeta interface {
	ResourceType() string
	ResourceID() string
	ResourceVersion() string
	ResourceExtraPaths() []string
}

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

	meta, ok := pkg.(ResourceMeta)
	if !ok {
		return state.Resource{
			Name:  pkg.GetName(),
			Home:  pkg.GetHome(),
			Type:  "Unknown",
			Paths: paths,
		}
	}

	paths = append(paths, meta.ResourceExtraPaths()...)

	return state.Resource{
		ID:      meta.ResourceID(),
		Name:    pkg.GetName(),
		Home:    pkg.GetHome(),
		Type:    meta.ResourceType(),
		Version: meta.ResourceVersion(),
		Paths:   paths,
	}
}

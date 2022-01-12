package schema

import (
	"fmt"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/hashicorp/hcl/v2"
)

// Data is
type Data struct {
	Body  hcl.Body
	Files map[string]*hcl.File
}

// Manifest is
type Manifest struct {
	GitHub    []*GitHub
	Gist      []*Gist
	Local     []*Local
	HTTP      []*HTTP
	Variables []*Variable
}

// manifestSchema is the schema for the top-level of a config file. We use
// the low-level HCL API for this level so we can easily deal with each
// block type separately with its own decoding logic.
var manifestSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "github",
			LabelNames: []string{"name"},
		},
		{
			Type:       "gist",
			LabelNames: []string{"name"},
		},
		{
			Type:       "local",
			LabelNames: []string{"name"},
		},
		{
			Type:       "http",
			LabelNames: []string{"name"},
		},
		{
			Type:       "function",
			LabelNames: []string{"name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
	},
}

// Decode decodes HCL files data to Manifest schema
func Decode(data Data) (*Manifest, error) {
	manifest := &Manifest{}
	content, diags := data.Body.Content(manifestSchema)

	for _, block := range content.Blocks {
		switch block.Type {

		case "variable":
			cfg, cfgDiags := decodeVariableBlock(block, false)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				manifest.Variables = append(manifest.Variables, cfg)
			}

		case "github":
			github, githubDiags := decodeGitHubSchema(block)
			diags = append(diags, githubDiags...)
			if github != nil {
				manifest.GitHub = append(manifest.GitHub, github)
			}
			diags = append(diags, checkGitHubUnique(manifest.GitHub)...)

		case "gist":
			gist, gistDiags := decodeGistSchema(block)
			diags = append(diags, gistDiags...)
			if gist != nil {
				manifest.Gist = append(manifest.Gist, gist)
			}
			diags = append(diags, checkGistUnique(manifest.Gist)...)

		case "local":
			local, localDiags := decodeLocalSchema(block)
			diags = append(diags, localDiags...)
			if local != nil {
				manifest.Local = append(manifest.Local, local)
			}
			diags = append(diags, checkLocalUnique(manifest.Local)...)

		case "http":
			http, httpDiags := decodeHTTPSchema(block)
			diags = append(diags, httpDiags...)
			if http != nil {
				manifest.HTTP = append(manifest.HTTP, http)
			}
			diags = append(diags, checkHTTPUnique(manifest.HTTP)...)

		case "function":

		default:
			// Any other block types are ones we've reserved for future use,
			// so they get a generic message.
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved block type name in resource block",
				Detail:   fmt.Sprintf("The block type name %q is reserved for use by AFX in a future version.", block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return manifest, errors.New(diags, data.Files)
}

// Plugin is
type Plugin struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Sources []string
	Env     map[string]string
	Load    *Load
}

func decodePluginBlock(block *hcl.Block) (*Plugin, hcl.Diagnostics) {
	content, config, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "sources",
				Required: true,
			},
			{
				Name: "env",
			},
			{
				Name: "load",
			},
		},
	})

	plugin := &Plugin{
		Config:    config,
		DeclRange: block.DefRange,
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "load":
			load, loadDiags := decodeLoadBlock(block)
			diags = append(diags, loadDiags...)
			if load != nil {
				plugin.Load = load
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved block type name in resource block",
				Detail:   fmt.Sprintf("The block type name %q is reserved for use by AFX in a future version.", block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return plugin, diags
}

// Command is
type Command struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Link  []*Link
	Build *Build
	Env   map[string]string
	Alias map[string]string
	Load  *Load
}

func decodeCommandBlock(block *hcl.Block) (*Command, hcl.Diagnostics) {
	content, config, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "link",
			},
			{
				Name: "build",
			},
			{
				Name: "env",
			},
			{
				Name: "alias",
			},
		},
	})

	command := &Command{
		Config:    config,
		DeclRange: block.DefRange,
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "build":
			build, buildDiags := decodeBuildBlock(block)
			diags = append(diags, buildDiags...)
			if build != nil {
				command.Build = build
			}
		case "link":
			link, linkDiags := decodeLinkBlock(block)
			diags = append(diags, linkDiags...)
			if link != nil {
				command.Link = append(command.Link, link)
			}
		case "load":
			load, loadDiags := decodeLoadBlock(block)
			diags = append(diags, loadDiags...)
			if load != nil {
				command.Load = load
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved block type name in resource block",
				Detail:   fmt.Sprintf("The block type name %q is reserved for use by AFX in a future version.", block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return command, diags
}

// Build is
type Build struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Steps []string
}

func decodeBuildBlock(block *hcl.Block) (*Build, hcl.Diagnostics) {
	_, config, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "steps",
				Required: true,
			},
		},
	})

	build := &Build{
		Config:    config,
		DeclRange: block.DefRange,
	}

	return build, diags
}

// Link is
type Link struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Steps []string
}

func decodeLinkBlock(block *hcl.Block) (*Link, hcl.Diagnostics) {
	_, config, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "from",
				Required: true,
			},
			{
				Name: "to",
				// Required: true,
			},
		},
	})

	link := &Link{
		Config:    config,
		DeclRange: block.DefRange,
	}

	return link, diags
}

// Load is
type Load struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Scripts []string
}

func decodeLoadBlock(block *hcl.Block) (*Load, hcl.Diagnostics) {
	_, config, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "scripts",
			},
		},
	})

	load := &Load{
		Config:    config,
		DeclRange: block.DefRange,
	}

	return load, diags
}

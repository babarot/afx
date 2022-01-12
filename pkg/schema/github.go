package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// GitHub is a schema for shell github
type GitHub struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Owner       string
	Repo        string
	Description string
	Path        string

	Release *Release

	Plugin  *Plugin
	Command *Command
}

var githubBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "release",
		},
		{
			Type: "plugin",
		},
		{
			Type: "command",
		},
	},
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "owner",
			Required: true,
		},
		{
			Name:     "repo",
			Required: true,
		},
		{
			Name: "description",
		},
		{
			Name: "path",
		},
	},
}

func decodeGitHubSchema(block *hcl.Block) (*GitHub, hcl.Diagnostics) {
	github := &GitHub{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}
	content, remain, diags := block.Body.PartialContent(githubBlockSchema)
	github.Config = remain

	if !hclsyntax.ValidIdentifier(github.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid output name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	// if attr, exists := content.Attributes["adderss"]; !exists {
	// }

	for _, block := range content.Blocks {
		switch block.Type {
		case "release":
			release, releaseDiags := decodeReleaseBlock(block)
			diags = append(diags, releaseDiags...)
			if release != nil {
				github.Release = release
			}
		case "plugin":
			plugin, pluginDiags := decodePluginBlock(block)
			diags = append(diags, pluginDiags...)
			if plugin != nil {
				github.Plugin = plugin
			}
		case "command":
			command, commandDiags := decodeCommandBlock(block)
			diags = append(diags, commandDiags...)
			if command != nil {
				github.Command = command
			}
		default:
			continue
		}
	}

	return github, diags
}

// Release is
type Release struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	ReleaseName string
	Tag         string
}

func decodeReleaseBlock(block *hcl.Block) (*Release, hcl.Diagnostics) {
	_, config, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: true,
			},
			{
				Name:     "tag",
				Required: true,
			},
		},
	})

	release := &Release{
		Config:    config,
		DeclRange: block.DefRange,
	}

	return release, diags
}

func checkGitHubUnique(resources []*GitHub) hcl.Diagnostics {
	encountered := map[string]*GitHub{}
	var diags hcl.Diagnostics
	for _, resource := range resources {
		if existing, exist := encountered[resource.Name]; exist {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate resource definition",
				Detail:   fmt.Sprintf("A GitHub resource named %q was already defined at %s. GitHub resource names must be unique within a policy.", existing.Name, existing.DeclRange),
				Subject:  &resource.DeclRange,
			})
		}
		encountered[resource.Name] = resource
	}
	return diags
}

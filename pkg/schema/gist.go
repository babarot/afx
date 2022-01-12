package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Gist is a schema for shell gist
type Gist struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Owner       string
	ID          string
	Description string

	Plugin  *Plugin
	Command *Command
}

var gistBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
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
			Name:     "id",
			Required: true,
		},
		{
			Name: "description",
		},
	},
}

func decodeGistSchema(block *hcl.Block) (*Gist, hcl.Diagnostics) {
	gist := &Gist{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}
	content, remain, diags := block.Body.PartialContent(gistBlockSchema)
	gist.Config = remain

	if !hclsyntax.ValidIdentifier(gist.Name) {
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
		case "plugin":
			plugin, pluginDiags := decodePluginBlock(block)
			diags = append(diags, pluginDiags...)
			if plugin != nil {
				gist.Plugin = plugin
			}
		case "command":
			command, commandDiags := decodeCommandBlock(block)
			diags = append(diags, commandDiags...)
			if command != nil {
				gist.Command = command
			}
		default:
			continue
		}
	}

	return gist, diags
}

func checkGistUnique(resources []*Gist) hcl.Diagnostics {
	encountered := map[string]*Gist{}
	var diags hcl.Diagnostics
	for _, resource := range resources {
		if existing, exist := encountered[resource.Name]; exist {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate resource definition",
				Detail:   fmt.Sprintf("A Gist resource named %q was already defined at %s. Gist resource names must be unique within a policy.", existing.Name, existing.DeclRange),
				Subject:  &resource.DeclRange,
			})
		}
		encountered[resource.Name] = resource
	}
	return diags
}

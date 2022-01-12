package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Local is a schema for shell local
type Local struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	Description string

	Plugin  *Plugin
	Command *Command
}

var localBlockSchema = &hcl.BodySchema{
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
			Name: "description",
		},
	},
}

func decodeLocalSchema(block *hcl.Block) (*Local, hcl.Diagnostics) {
	local := &Local{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}
	content, remain, diags := block.Body.PartialContent(localBlockSchema)
	local.Config = remain

	if !hclsyntax.ValidIdentifier(local.Name) {
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
				local.Plugin = plugin
			}
		case "command":
			command, commandDiags := decodeCommandBlock(block)
			diags = append(diags, commandDiags...)
			if command != nil {
				local.Command = command
			}
		default:
			continue
		}
	}

	return local, diags
}

func checkLocalUnique(resources []*Local) hcl.Diagnostics {
	encountered := map[string]*Local{}
	var diags hcl.Diagnostics
	for _, resource := range resources {
		if existing, exist := encountered[resource.Name]; exist {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate resource definition",
				Detail:   fmt.Sprintf("A Local resource named %q was already defined at %s. Local resource names must be unique within a policy.", existing.Name, existing.DeclRange),
				Subject:  &resource.DeclRange,
			})
		}
		encountered[resource.Name] = resource
	}
	return diags
}

package schema

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// HTTP is a schema for shell HTTP
type HTTP struct {
	Name string

	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range

	URL         string
	Output      string
	Description string

	Plugin  *Plugin
	Command *Command
}

var httpBlockSchema = &hcl.BodySchema{
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
			Name:     "url",
			Required: true,
		},
		{
			Name:     "output",
			Required: true,
		},
		{
			Name: "description",
		},
	},
}

func decodeHTTPSchema(block *hcl.Block) (*HTTP, hcl.Diagnostics) {
	http := &HTTP{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}
	content, remain, diags := block.Body.PartialContent(httpBlockSchema)
	http.Config = remain

	if !hclsyntax.ValidIdentifier(http.Name) {
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
				http.Plugin = plugin
			}
		case "command":
			command, commandDiags := decodeCommandBlock(block)
			diags = append(diags, commandDiags...)
			if command != nil {
				http.Command = command
			}
		default:
			continue
		}
	}

	return http, diags
}

func checkHTTPUnique(resources []*HTTP) hcl.Diagnostics {
	encountered := map[string]*HTTP{}
	var diags hcl.Diagnostics
	for _, resource := range resources {
		if existing, exist := encountered[resource.Name]; exist {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate resource definition",
				Detail:   fmt.Sprintf("A HTTP resource named %q was already defined at %s. HTTP resource names must be unique within a policy.", existing.Name, existing.DeclRange),
				Subject:  &resource.DeclRange,
			})
		}
		encountered[resource.Name] = resource
	}
	return diags
}

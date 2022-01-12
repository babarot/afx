package hcl

import (
	"fmt"

	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// MergeBodiesWithOverides merges several bodies into one. This is similar
// implementation as the one in official library, but it overwrites attributes
// that are already defined.
func MergeBodiesWithOverides(bodies []hcl2.Body) hcl2.Body {
	if len(bodies) == 0 {
		// Swap out for our singleton empty body, to reduce the number of
		// empty slices we have hanging around.
		return emptyBody
	}

	// If any of the given bodies are already merged bodies, we'll unpack
	// to flatten to a single mergedBodies, since that's conceptually simpler.
	// This also, as a side-effect, eliminates any empty bodies, since
	// empties are merged bodies with no inner bodies.
	var newLen int
	var flatten bool
	for _, body := range bodies {
		if children, merged := body.(mergedBodies); merged {
			newLen += len(children)
			flatten = true
		} else {
			newLen++
		}
	}

	if !flatten { // not just newLen == len, because we might have mergedBodies with single bodies inside
		return mergedBodies(bodies)
	}

	if newLen == 0 {
		// Don't allocate a new empty when we already have one
		return emptyBody
	}

	new := make([]hcl2.Body, 0, newLen)
	for _, body := range bodies {
		if children, merged := body.(mergedBodies); merged {
			new = append(new, children...)
		} else {
			new = append(new, body)
		}
	}
	return mergedBodies(new)
}

var emptyBody = mergedBodies([]hcl2.Body{})

// EmptyBody returns a body with no content. This body can be used as a
// placeholder when a body is required but no body content is available.
func EmptyBody() hcl2.Body {
	return emptyBody
}

type mergedBodies []hcl2.Body

// Content returns the content produced by applying the given schema to all
// of the merged bodies and merging the result.
//
// Although required attributes _are_ supported, they should be used sparingly
// with merged bodies since in this case there is no contextual information
// with which to return good diagnostics. Applications working with merged
// bodies may wish to mark all attributes as optional and then check for
// required attributes afterwards, to produce better diagnostics.
func (mb mergedBodies) Content(schema *hcl2.BodySchema) (*hcl2.BodyContent, hcl2.Diagnostics) {
	// the returned body will always be empty in this case, because mergedContent
	// will only ever call Content on the child bodies.
	content, _, diags := mb.mergedContent(schema, false)
	return content, diags
}

func (mb mergedBodies) PartialContent(schema *hcl2.BodySchema) (*hcl2.BodyContent, hcl2.Body, hcl2.Diagnostics) {
	return mb.mergedContent(schema, true)
}

func (mb mergedBodies) JustAttributes() (hcl2.Attributes, hcl2.Diagnostics) {
	attrs := make(map[string]*hcl2.Attribute)
	var diags hcl2.Diagnostics

	for _, body := range mb {
		if body == nil {
			continue
		}

		thisAttrs, thisDiags := body.JustAttributes()

		if len(thisDiags) != 0 {
			diags = append(diags, thisDiags...)
		}

		if thisAttrs != nil {
			for name, attr := range thisAttrs {
				if existing := attrs[name]; existing != nil {
					if attrMap, diags := hcl2.ExprMap(attr.Expr); diags == nil {
						existingAttrMap, diags := hcl2.ExprMap(attrs[name].Expr)
						if diags != nil {
							diags = diags.Append(&hcl2.Diagnostic{
								Severity: hcl2.DiagError,
								Summary:  "Incompatible types",
								Detail: fmt.Sprintf(
									"Argument %q has different types",
									name,
								),
								Subject: &attr.NameRange,
							})
							continue
						}

						existingAttrMap = append(existingAttrMap, attrMap...)

						items := make([]hclsyntax.ObjectConsItem, len(existingAttrMap))
						for idx, existingAttr := range existingAttrMap {
							items[idx] = hclsyntax.ObjectConsItem{
								KeyExpr:   existingAttr.Key.(hclsyntax.Expression),
								ValueExpr: existingAttr.Value.(hclsyntax.Expression),
							}
						}

						attrs[name].Expr = &hclsyntax.ObjectConsExpr{
							Items: items,
						}

						continue
					}
				}

				attrs[name] = attr
			}
		}
	}

	return attrs, diags
}

func (mb mergedBodies) MissingItemRange() hcl2.Range {
	if len(mb) == 0 {
		// Nothing useful to return here, so we'll return some garbage.
		return hcl2.Range{
			Filename: "<empty>",
		}
	}

	// arbitrarily use the first body's missing item range
	return mb[0].MissingItemRange()
}

func (mb mergedBodies) mergedContent(schema *hcl2.BodySchema, partial bool) (*hcl2.BodyContent, hcl2.Body, hcl2.Diagnostics) {
	// We need to produce a new schema with none of the attributes marked as
	// required, since _any one_ of our bodies can contribute an attribute value.
	// We'll separately check that all required attributes are present at
	// the end.
	mergedSchema := &hcl2.BodySchema{
		Blocks: schema.Blocks,
	}
	for _, attrS := range schema.Attributes {
		mergedAttrS := attrS
		mergedAttrS.Required = false
		mergedSchema.Attributes = append(mergedSchema.Attributes, mergedAttrS)
	}

	var mergedLeftovers []hcl2.Body
	content := &hcl2.BodyContent{
		Attributes: map[string]*hcl2.Attribute{},
	}

	var diags hcl2.Diagnostics
	for _, body := range mb {
		if body == nil {
			continue
		}

		var thisContent *hcl2.BodyContent
		var thisLeftovers hcl2.Body
		var thisDiags hcl2.Diagnostics

		if partial {
			thisContent, thisLeftovers, thisDiags = body.PartialContent(mergedSchema)
		} else {
			thisContent, thisDiags = body.Content(mergedSchema)
		}

		if thisLeftovers != nil {
			mergedLeftovers = append(mergedLeftovers, thisLeftovers)
		}
		if len(thisDiags) != 0 {
			diags = append(diags, thisDiags...)
		}

		if thisContent.Attributes != nil {
			for name, attr := range thisContent.Attributes {
				content.Attributes[name] = attr
			}
		}

		if len(thisContent.Blocks) != 0 {
			content.Blocks = append(content.Blocks, thisContent.Blocks...)
		}
	}

	// Finally, we check for required attributes.
	for _, attrS := range schema.Attributes {
		if !attrS.Required {
			continue
		}

		if content.Attributes[attrS.Name] == nil {
			// We don't have any context here to produce a good diagnostic,
			// which is why we warn in the Content docstring to minimize the
			// use of required attributes on merged bodies.
			diags = diags.Append(&hcl2.Diagnostic{
				Severity: hcl2.DiagError,
				Summary:  "Missing required argument",
				Detail: fmt.Sprintf(
					"The argument %q is required, but was not set.",
					attrS.Name,
				),
			})
		}
	}

	leftoverBody := MergeBodiesWithOverides(mergedLeftovers)
	return content, leftoverBody, diags
}

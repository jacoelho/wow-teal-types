package wowapi

// nativeMethodOwner separates the documentation group from the receiver that
// owns its contract. The pinned Wowless docs.yaml uiobject_method_reassignments.Region routes
// exactly these five methods to LayeredRegion; its uiobjects.yaml keeps Frame
// on common Region and gives ModelScene its own narrower SetDrawLayer contract.
func nativeMethodOwner(documentName, methodName string) string {
	if documentName == "SimpleRegionAPI" {
		switch methodName {
		case "GetDrawLayer", "SetDrawLayer", "GetVertexColor", "SetVertexColor", "SetVertexColorFromBoolean":
			return "LayeredRegion"
		}
	}
	return scriptName(documentName)
}

func (c *catalogue) applyLayeredRegionHierarchy() {
	c.definition("LayeredRegion").Bases = []string{"Region"}
	for _, owner := range []string{"TextureBase", "FontString", "VectorGraphics"} {
		definition := c.definition(owner)
		found := false
		for i, base := range definition.Bases {
			if base == "Region" {
				definition.Bases[i] = "LayeredRegion"
				found = true
			}
		}
		if !found {
			reject("review changed layered receiver hierarchy: " + owner)
		}
	}
	c.note("LayeredRegion", "route five layer/color methods from the SimpleRegionAPI documentation group to their distinct receiver; Frame retains common Region", RepresentationNote{Evidence: evidenceText("pinned wowless/data/products/wow/docs.yaml uiobject_method_reassignments.Region and uiobjects.yaml LayeredRegion/Frame/ModelScene")})
}

package wowapi

// nativeMethodOwner separates the documentation group from the receiver that
// owns its contract. The Wowless docs.yaml uiobject_method_reassignments.Region routes
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
}

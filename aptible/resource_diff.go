package aptible

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type ResourceAttrs struct {
	Required   []string
	NotAllowed []string
}

type ResourceDiff struct {
	*schema.ResourceDiff
}

// HasRequired
// Returns a true if the attribute's value is not known or if it is known to not be null.
// Designed to be used to conditionally test for required attributes in a CustomizeDiff.
func (d *ResourceDiff) HasRequired(attr string) bool {
	_, ok := d.GetOkExists(attr)
	return ok || !d.NewValueKnown(attr)
}

// HasOptional
// Returns true if the attribute's value is not known or if it is known to not be null or zero.
// Designed to be used to conditionally test if for attributes that are not allowed in a CustomizeDiff.
func (d *ResourceDiff) HasOptional(attr string) bool {
	_, ok := d.GetOk(attr)
	return ok || !d.NewValueKnown(attr)
}

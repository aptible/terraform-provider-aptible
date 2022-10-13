package aptible

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func suppressDefaultDatabaseVersion(_, old, new string, _ *schema.ResourceData) bool {
	// If the new value is empty, ignore the diff because our API will handle setting a default.
	// That value will get set whenever the Database is Read, but the config will continue to show
	// an empty value but there is not an actual change needed.
	if new == "" {
		return true
	}
	return old == new
}

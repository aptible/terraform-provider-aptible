package aptible

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceEndpointPlatformDiffSuppress(t *testing.T) {
	platformSchema := resourceEndpoint().Schema["platform"]

	tests := []struct {
		name         string
		resourceType string
		oldValue     string
		newValue     string
		want         bool
	}{
		{
			name:         "suppresses database platform drift",
			resourceType: "database",
			oldValue:     "elb",
			newValue:     "nlb",
			want:         true,
		},
		{
			name:         "does not suppress app platform drift",
			resourceType: "app",
			oldValue:     "elb",
			newValue:     "nlb",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceEndpoint().Schema, map[string]interface{}{
				"env_id":        1,
				"resource_id":   1,
				"resource_type": tt.resourceType,
			})

			got := platformSchema.DiffSuppressFunc("platform", tt.oldValue, tt.newValue, d)
			if got != tt.want {
				t.Fatalf("DiffSuppressFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

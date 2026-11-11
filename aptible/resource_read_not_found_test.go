package aptible

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestReadRemovesResourcesDeletedOutsideTerraform(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":404,"error":"not_found","message":"Not Found"}`))
	}))
	defer server.Close()
	t.Setenv("APTIBLE_API_ROOT_URL", server.URL)

	meta := &providerMetadata{
		Client: aptibleapi.NewAPIClient(aptibleapi.NewAPIConfiguration()),
		Token:  "token",
	}

	cases := []struct {
		name     string
		resource *schema.Resource
		idField  string
		read     func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	}{
		{"app", resourceApp(), "app_id", resourceAppRead},
		{"database", resourceDatabase(), "database_id", resourceDatabaseReadContext},
		{"endpoint", resourceEndpoint(), "endpoint_id", resourceEndpointRead},
		{"replica", resourceReplica(), "replica_id", resourceReplicaReadContext},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, tc.resource.Schema, map[string]interface{}{tc.idField: 1})
			d.SetId("1")

			if diags := tc.read(context.Background(), d, meta); diags.HasError() {
				t.Fatalf("expected no error, got %v", diags)
			}
			if d.Id() != "" {
				t.Fatalf("expected the resource to be removed from state, got ID %q", d.Id())
			}
		})
	}
}

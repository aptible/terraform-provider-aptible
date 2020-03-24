package aptible

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const TestEnvironmentId = 5

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"aptible": testAccProvider,
	}
}

// Ensure we're pointing at a sandbox before we run and a token is provided
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("APTIBLE_AUTH_ROOT_URL"); v == "" {
		t.Fatal("APTIBLE_AUTH_ROOT_URL must be set to a sandbox for acceptance tests")
	}
	if v := os.Getenv("APTIBLE_API_ROOT_URL"); v == "" {
		t.Fatal("APTIBLE_API_ROOT_URL must be set to a sandbox for acceptance tests")
	}
	if v := os.Getenv("APTIBLE_ACCESS_TOKEN"); v == "" {
		t.Fatal("APTIBLE_ACCESS_TOKEN must be set for acceptance tests")
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

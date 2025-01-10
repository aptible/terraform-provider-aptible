package aptible

import (
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testOrganizationId string
var testStackId int

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

const (
	AptibleStackId        = "APTIBLE_STACK_ID"
	AptibleOrganizationId = "APTIBLE_ORGANIZATION_ID"
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"aptible": testAccProvider,
	}
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"aptible": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}

	// see testAccPreCheck for more details on this being set ahead of time for dx
	testOrganizationId = os.Getenv(AptibleOrganizationId)

	stackIdStr := os.Getenv(AptibleStackId)
	stackId, _ := strconv.Atoi(stackIdStr)
	testStackId = stackId
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

	if v := os.Getenv(AptibleOrganizationId); v == "" {
		t.Fatalf("%s must be set for acceptance tests", AptibleOrganizationId)
	}

	stackId := os.Getenv(AptibleStackId)
	if stackId == "" {
		t.Fatalf("%s must be set for acceptance tests", AptibleStackId)
	}
	if _, err := strconv.Atoi(stackId); err != nil {
		t.Fatalf("%s is not a valid integer value", AptibleStackId)
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

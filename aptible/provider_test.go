package aptible

import (
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testEnvironmentId int

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

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

	i := os.Getenv("APTIBLE_ENVIRONMENT_ID")

	// Precheck confirms this will work
	id, _ := strconv.Atoi(i)
	testEnvironmentId = id
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

	id := os.Getenv("APTIBLE_ENVIRONMENT_ID")
	if id == "" {
		t.Fatal("APTIBLE_ENVIRONMENT_ID must be set for acceptance tests")
	}
	if _, err := strconv.Atoi(id); err != nil {
		t.Fatal("APTIBLE_ENVIRONMENT_ID is not a valid integer value")
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

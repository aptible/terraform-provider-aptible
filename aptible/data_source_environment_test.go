package aptible

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEnvironment_validation(t *testing.T) {
	requiredAttrs := []string{"handle"}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `data "aptible_environment" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps:             testSteps,
	})
}

func TestAccDataSourceEnvironment_basic(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEnvironment(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_environment.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_environment.test", "org_id", testOrganizationId),
					resource.TestCheckResourceAttr("aptible_environment.test", "stack_id", strconv.Itoa(testStackId)),
				),
			}, {
				ResourceName:            "aptible_environment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"org_id", "stack_id"},
			},
			{
				Config: testDataAccAptibleEnvironment(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aptible_environment.test", "handle", rHandle),
				),
			},
		},
	})
}

func testDataAccAptibleEnvironment(handle string) string {
	// also include the past environment state to prevent deletion
	return fmt.Sprintf(`
%s
data "aptible_environment" "test" {
	handle = "%s"
}`, testAccAptibleEnvironment(handle), handle)
}

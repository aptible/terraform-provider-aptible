package aptible

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/aptible/go-deploy/aptible"
)

func TestAccResourceEnvironment_validation(t *testing.T) {
	requiredAttrs := []string{"handle", "stack_id", "org_id"}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `resource "aptible_environment" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	testSteps = append(testSteps, resource.TestStep{
		PlanOnly: true,
		Config: fmt.Sprintf(`
				resource "aptible_environment" "test" {
					handle = "%s"
					org_id = "%s"
					stack_id = "%v"
				}
			`, "test", "invalid-uuid", testStackId),
		ExpectError: regexp.MustCompile(fmt.Sprintf(`expected %q to be a valid UUID, got invalid-uuid`, "org_id")),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps:             testSteps,
	})
}

func testAccCheckEnvironmentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_environment" {
			continue
		}

		envID, err := strconv.Atoi(rs.Primary.Attributes["env_id"])
		if err != nil {
			return err
		}

		environment, err := client.GetEnvironment(int64(envID))
		log.Println("Deleted? ", environment.Deleted)
		if !environment.Deleted {
			return fmt.Errorf("environment %v not removed", envID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func TestAccResourceEnvironment_basic(t *testing.T) {
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
		},
	})
}

func TestAccResourceEnvironment_basic_no_org(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEnvironmentWithoutOrg(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_environment.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_environment.test", "stack_id", strconv.Itoa(testStackId)),
				),
			}, {
				ResourceName:            "aptible_environment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"org_id", "stack_id"},
			},
		},
	})
}

func TestAccResourceEnvironment_update(t *testing.T) {
	rHandle := acctest.RandString(10)
	rUpdatedHandle := acctest.RandString(10)

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
				ResourceName:      "aptible_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			}, {
				Config: testAccAptibleEnvironment(rUpdatedHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_environment.test", "handle", rUpdatedHandle),
					resource.TestCheckResourceAttr("aptible_environment.test", "org_id", testOrganizationId),
					resource.TestCheckResourceAttr("aptible_environment.test", "stack_id", strconv.Itoa(testStackId)),
				),
			},
		},
	})
}

func testAccAptibleEnvironment(handle string) string {
	return fmt.Sprintf(`
resource "aptible_environment" "test" {
	handle = "%s"
	org_id = "%s"
	stack_id = "%v"
}`, handle, testOrganizationId, testStackId)
}

func testAccAptibleEnvironmentWithoutOrg(handle string) string {
	return fmt.Sprintf(`
resource "aptible_environment" "test" {
	handle = "%s"
	stack_id = "%v"
}`, handle, testStackId)
}

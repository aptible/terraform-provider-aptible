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
)

func TestAccResourceEnvironment_validation(t *testing.T) {
	requiredAttrs := []string{"handle", "stack_id"}
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
				handle = "test"
				org_id = "%s"
				stack_id = %d
			}
			`, "invalid-uuid", testStackId),
		ExpectError: regexp.MustCompile(fmt.Sprintf(`expected %q to be a valid UUID, got invalid-uuid`, "org_id")),
	})

	testSteps = append(testSteps, resource.TestStep{
		Config: fmt.Sprintf(`
			resource "aptible_environment" "test" {
				handle = "%s"
				org_id = "%s"
				stack_id = "%v"

				backup_retention_policy {
					daily = 3
					monthly = 2
					yearly = 1
					make_copy = true
					keep_final = false
				}

				backup_retention_policy {
					daily = 4
					monthly = 2
					yearly = 1
					make_copy = true
					keep_final = false
				}
			}
			`, acctest.RandString(10), testOrganizationId, testStackId),
		ExpectError: regexp.MustCompile("(?i)multiple backup_retention_policy"),
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps:             testSteps,
	})
}

func testAccCheckEnvironmentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
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

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.daily"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.monthly"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.yearly"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.make_copy"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.keep_final"),
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

func TestAccResourceEnvironment_no_org(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEnvironmentWithoutOrg(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_environment.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_environment.test", "org_id", testOrganizationId),
					resource.TestCheckResourceAttr("aptible_environment.test", "stack_id", strconv.Itoa(testStackId)),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.daily"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.monthly"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.yearly"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.make_copy"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.keep_final"),
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

func TestAccResourceEnvironment_backup_policy(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEnvironmentWithBackupPolicy(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_environment.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_environment.test", "org_id", testOrganizationId),
					resource.TestCheckResourceAttr("aptible_environment.test", "stack_id", strconv.Itoa(testStackId)),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.daily", "3"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.monthly", "2"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.yearly", "1"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.make_copy", "true"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.keep_final", "false"),
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

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.daily"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.monthly"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.yearly"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.make_copy"),
					resource.TestCheckResourceAttrSet("aptible_environment.test", "backup_retention_policy.0.keep_final"),
				),
			}, {
				ResourceName:      "aptible_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			}, {
				Config: testAccAptibleEnvironmentWithBackupPolicy(rUpdatedHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_environment.test", "handle", rUpdatedHandle),
					resource.TestCheckResourceAttr("aptible_environment.test", "org_id", testOrganizationId),
					resource.TestCheckResourceAttr("aptible_environment.test", "stack_id", strconv.Itoa(testStackId)),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.daily", "3"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.monthly", "2"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.yearly", "1"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.make_copy", "true"),
					resource.TestCheckResourceAttr("aptible_environment.test", "backup_retention_policy.0.keep_final", "false"),
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
	}
	`, handle, testOrganizationId, testStackId)
}

func testAccAptibleEnvironmentWithoutOrg(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		stack_id = "%v"
	}
	`, handle, testStackId)
}

func testAccAptibleEnvironmentWithBackupPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"

		backup_retention_policy {
			daily = 3
			monthly = 2
			yearly = 1
			make_copy = true
			keep_final = false
		}
	}
	`, handle, testOrganizationId, testStackId)
}

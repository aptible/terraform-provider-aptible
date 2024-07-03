package aptible

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceBackupRetentionPolicy_validation(t *testing.T) {
	requiredAttrs := []string{"env_id"}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `data "aptible_backup_retention_policy" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps:             testSteps,
	})
}

func TestAccDataSourceBackupRetentionPolicy_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataAccAptibleBackupRetentionPolicy(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "data.aptible_backup_retention_policy.test", "env_id"),
					resource.TestCheckResourceAttrSet("data.aptible_backup_retention_policy.test", "daily"),
					resource.TestCheckResourceAttrSet("data.aptible_backup_retention_policy.test", "monthly"),
					resource.TestCheckResourceAttrSet("data.aptible_backup_retention_policy.test", "yearly"),
					resource.TestCheckResourceAttrSet("data.aptible_backup_retention_policy.test", "make_copy"),
					resource.TestCheckResourceAttrSet("data.aptible_backup_retention_policy.test", "keep_final"),
				),
			},
		},
	})
}

func testDataAccAptibleBackupRetentionPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = %v
	}

	data "aptible_backup_retention_policy" "test" {
		env_id = aptible_environment.test.env_id
	}
	`, handle, testOrganizationId, testStackId)
}

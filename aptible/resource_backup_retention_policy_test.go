package aptible

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceBackupRetentionPolicy_validation(t *testing.T) {
	// Test required attributes
	requiredAttrs := []string{
		"env_id",
		"daily",
		"monthly",
		"yearly",
		"make_copy",
		"keep_final",
	}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `resource "aptible_backup_retention_policy" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	// Test minimum backup validation
	config := testAccAptibleBackupRetentionPolicy("handle", -1, true)
	testSteps = append(testSteps, resource.TestStep{
		PlanOnly:    true,
		Config:      config,
		ExpectError: regexp.MustCompile(`daily.*at least \(1\)`),
	})
	testSteps = append(testSteps, resource.TestStep{
		PlanOnly:    true,
		Config:      config,
		ExpectError: regexp.MustCompile(`monthly.*at least \(0\)`),
	})
	testSteps = append(testSteps, resource.TestStep{
		PlanOnly:    true,
		Config:      config,
		ExpectError: regexp.MustCompile(`yearly.*at least \(0\)`),
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBackupRetentionPolicyDestroy,
		Steps:             testSteps,
	})
}

// Backup retention policies aren't actually destroyed by the provider but they
// will be in the tests when the environments they're associated with are destroyed
func testAccCheckBackupRetentionPolicyDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta().(*providerMetadata)
	client := meta.Client
	ctx := meta.APIContext(context.Background())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_backup_retention_policy" {
			continue
		}

		policyId, err := strconv.Atoi(rs.Primary.Attributes["policy_id"])
		if err != nil {
			return err
		}

		_, _, err = client.BackupRetentionPoliciesAPI.
			GetBackupRetentionPolicy(ctx, int32(policyId)).
			Execute()
		if err != nil {
			if regexp.MustCompile("(?i)(not found|forbidden)").Match([]byte(err.Error())) {
				continue
			}

			return err
		}

		return fmt.Errorf("backup retention policy %v was not deleted", policyId)
	}

	return nil
}

func TestAccResourceBackupRetentionPolicy_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBackupRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleBackupRetentionPolicy(acctest.RandString(10), 5, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_backup_retention_policy.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "daily", "5"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "monthly", "4"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "yearly", "3"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "make_copy", "true"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "keep_final", "false"),
				),
			}, {
				ResourceName:      "aptible_backup_retention_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceBackupRetentionPolicy_update(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBackupRetentionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleBackupRetentionPolicy(rHandle, 2, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_backup_retention_policy.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "daily", "2"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "monthly", "1"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "yearly", "0"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "make_copy", "false"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "keep_final", "true"),
				),
			}, {
				ResourceName:      "aptible_backup_retention_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			}, {
				Config: testAccAptibleBackupRetentionPolicy(rHandle, 4, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_backup_retention_policy.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "daily", "4"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "monthly", "3"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "yearly", "2"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "make_copy", "true"),
					resource.TestCheckResourceAttr("aptible_backup_retention_policy.test", "keep_final", "false"),
				),
			},
		},
	})
}

func testAccAptibleBackupRetentionPolicy(handle string, backups int, flag bool) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = %v
	}

	resource "aptible_backup_retention_policy" "test" {
		env_id = aptible_environment.test.env_id
		daily = %v
		monthly = %v
		yearly = %v
		make_copy = %v
		keep_final = %v
	}
	`, handle, testOrganizationId, testStackId,
		backups, backups-1, backups-2,
		flag, !flag)
}

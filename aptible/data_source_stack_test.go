package aptible

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/aptible/go-deploy/aptible"
)

func TestAccStackDataSource_validation(t *testing.T) {
	requiredAttrs := []string{"name"}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `data "aptible_stack" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps:             testSteps,
	})
}

func TestAccStackDataSource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		// This guard is set because the below code should only evaluate if running in an integration test
		// setting. Typically this is honored by Terraform, but this occurs outside the context of the resource.Test
		// which must be done as values set in PreCheck are discarded and/or run in an entirely different context
		var stacks []aptible.Stack
		client, err := aptible.SetUpClient()
		if err != nil {
			t.Fatalf("Unable to generate and setup client for stacks test - %s", err.Error())
			return
		}
		stacks, err = client.GetStacks()
		if err != nil {
			t.Fatalf("Unable to retrieve stacks for test - %s", err.Error())
			return
		}
		if len(stacks) == 0 {
			t.Fatal("Unable to find stacks with a zero length")
			return
		}

		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Providers:         testAccProviders,
			ProviderFactories: testAccProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testDataAccAptibleStack(stacks[0].Name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.aptible_stack.test", "name", stacks[0].Name),
						resource.TestCheckResourceAttr("data.aptible_stack.test", "stack_id", strconv.Itoa(int(stacks[0].ID))),
						resource.TestCheckResourceAttr("data.aptible_stack.test", "org_id", stacks[0].OrganizationID),
					),
				},
			},
		})
	}
}

func testDataAccAptibleStack(name string) string {
	return fmt.Sprintf(`
data "aptible_stack" "test" {
	name = "%s"
}`,
		name)
}

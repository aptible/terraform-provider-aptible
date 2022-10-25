package aptible

import (
	"fmt"
	"github.com/aptible/go-deploy/aptible"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceStack_validation(t *testing.T) {
	requiredAttrs := []string{"name"}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `data "aptible_stack" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps:             testSteps,
	})
}

func TestAccDataSourceStack_deploy(t *testing.T) {
	var stacks []aptible.Stack

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
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

func testDataAccAptibleStack(name string) string {
	return fmt.Sprintf(`
data "aptible_stack" "test" {
	name = "%s"
}`,
		name)
}

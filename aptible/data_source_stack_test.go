package aptible

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		m := testAccProvider.Meta().(*providerMetadata)
		client := m.Client
		ctx := m.APIContext(context.Background())

		stacksResp, _, err := client.StacksAPI.ListStacks(ctx).Execute()
		if err != nil {
			t.Fatalf("Unable to retrieve stacks for test - %s", err.Error())
			return
		}
		stacks := stacksResp.Embedded.Stacks
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
					Config: testDataAccAptibleStack(stacks[0].GetName()),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.aptible_stack.test", "name", stacks[0].GetName()),
						resource.TestCheckResourceAttr("data.aptible_stack.test", "stack_id", strconv.Itoa(int(stacks[0].GetId()))),
						resource.TestCheckResourceAttr("data.aptible_stack.test", "org_id", stacks[0].GetAccountId()),
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

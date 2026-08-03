package aptible

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/aptible/aptible-api-go/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1")
	}

	diags := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags.HasError() {
		t.Fatalf("Failed to configure provider: %v", diags)
		return
	}

	m := testAccProvider.Meta().(*providerMetadata)
	ctx := context.Background()

	stacksResp, _, err := m.StacksAPI.ListStacks(ctx).Execute()
	if err != nil {
		t.Fatalf("Unable to retrieve stacks for test - %s", err.Error())
		return
	}
	stacks := stacksResp.Embedded.Stacks
	if len(stacks) == 0 {
		t.Fatal("Unable to find stacks with a zero length")
		return
	}

	stack := stacks[0]
	expectedOrgID := helpers.GetStackOrganizationID(&stack)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataAccAptibleStack(stack.GetName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aptible_stack.test", "name", stack.GetName()),
					resource.TestCheckResourceAttr("data.aptible_stack.test", "stack_id", strconv.Itoa(int(stack.GetId()))),
					resource.TestCheckResourceAttr("data.aptible_stack.test", "org_id", expectedOrgID),
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

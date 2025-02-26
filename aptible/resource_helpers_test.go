package aptible

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func checkTainted(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// This is how the terraform provided checks get the instance state of a resource
		ms := state.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		// The resource is expected to be tainted
		if !is.Tainted {
			return fmt.Errorf("%s: Resource is not tainted but was expected to be", name)
		}

		return nil
	}
}

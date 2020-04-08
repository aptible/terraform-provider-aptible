package aptible

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccResourceApp_basic(t *testing.T) {
	rHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleApp(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.example", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_app.example", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttrSet("aptible_app.example", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.example", "git_repo"),
				),
			},
		},
	})
}

func testAccAptibleApp(handle string) string {
	return fmt.Sprintf(`
resource "aptible_app" "example" {
    env_id = %d
    handle = "%v"
}
`, TestEnvironmentId, handle)
}

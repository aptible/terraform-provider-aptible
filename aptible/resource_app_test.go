package aptible

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccResourceApp_basic(t *testing.T) {
	rHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppDestroy,
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

func testAccCheckAppDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation
	time.Sleep(30 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_app" {
			continue
		}

		app_id, err := strconv.Atoi(rs.Primary.Attributes["app_id"])
		if err != nil {
			return err
		}

		deleted, err := client.GetApp(int64(app_id))
		if err != nil {
			return err
		}
		log.Println("Deleted? ", deleted)

		if !deleted {
			return fmt.Errorf("App %v not removed", app_id)
		}
	}
	return nil
}

func testAccAptibleApp(handle string) string {
	return fmt.Sprintf(`
resource "aptible_app" "example" {
    env_id = %d
    handle = "%v"
}
`, TestEnvironmentId, handle)
}

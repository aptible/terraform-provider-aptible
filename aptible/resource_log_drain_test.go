package aptible

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceLogDrain_elasticsearch(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainElastic(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "elasticsearch_database"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckLogDrainDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_log_drain" {
			continue
		}

		logDrainID, err := strconv.Atoi(rs.Primary.Attributes["log_drain_id"])
		if err != nil {
			return err
		}

		logDrain, err := client.GetLogDrain(int64(logDrainID))
		log.Println("Deleted? ", logDrain.Deleted)
		if !logDrain.Deleted {
			return fmt.Errorf("log drain %v not removed", logDrainID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAptibleLogDrainElastic(handle string) string {
	return fmt.Sprintf(`
resource "aptible_log_drain" "test" {
    env_id = %d
    database_id = %d
    handle = "%v"
    drain_type = "elasticsearch_database"
}
`, testEnvironmentId, handle)
}

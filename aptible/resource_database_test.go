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

// DONE
func TestAccResourceDatabase_basic(t *testing.T) {
	dbHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleDatabaseBasic(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_db.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_db.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_db.test", "db_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_db.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_db.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "db_id"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "connection_url"),
				),
			},
		},
	})
}

func TestAccResourceDatabase_update(t *testing.T) {
	dbHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleDatabaseBasic(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_db.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_db.test", "env_id", strconv.Itoa(TestEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_db.test", "db_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_db.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_db.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "db_id"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "connection_url"),
				),
			},
			{
				Config: testAccAptibleDatabaseUpdate(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_db.test", "container_size", "512"),
					resource.TestCheckResourceAttr("aptible_db.test", "disk_size", "20"),
				),
			},
		},
	})
}

func testAccCheckDatabaseDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation
	time.Sleep(30 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_app" {
			continue
		}

		db_id, err := strconv.Atoi(rs.Primary.Attributes["db_id"])
		if err != nil {
			return err
		}

		_, deleted, err := client.GetDatabase(int64(db_id))
		log.Println("Deleted? ", deleted)
		if !deleted {
			return fmt.Errorf("Database %v not removed", db_id)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAptibleDatabaseBasic(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_db" "test" {
    env_id = %d
	handle = "%v"
}
`, TestEnvironmentId, dbHandle)
}

func testAccAptibleDatabaseUpdate(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_db" "test" {
    env_id = %d
	handle = "%v"
	container_size = %d
	disk_size = %d
}
`, TestEnvironmentId, dbHandle, 512, 20)
}

package aptible

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

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
					resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "connection_url"),
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
					resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "connection_url"),
				),
			},
			{
				Config: testAccAptibleDatabaseUpdate(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_database.test", "container_size", "512"),
					resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "20"),
				),
			},
		},
	})
}

func TestAccResourceDatabase_expectError(t *testing.T) {
	dbHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAptibleDatabaseInvalidDBType(dbHandle),
				ExpectError: regexp.MustCompile(`expected database_type to be one of .*, got non-existent-db`),
			},
			{
				Config:      testAccAptibleDatabaseInvalidContainerSize(dbHandle),
				ExpectError: regexp.MustCompile(`expected container_size to be in the range .*, got 0`),
			},
			{
				Config:      testAccAptibleDatabaseInvalidDiskSize(dbHandle),
				ExpectError: regexp.MustCompile(`config is invalid: expected disk_size to be in the range .*, got 0`),
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

		databaseId, err := strconv.Atoi(rs.Primary.Attributes["database_id"])
		if err != nil {
			return err
		}

		database, err := client.GetDatabase(int64(databaseId))
		log.Println("Deleted? ", database.Deleted)
		if !database.Deleted {
			return fmt.Errorf("database %v not removed", databaseId)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAptibleDatabaseBasic(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
    env_id = %d
	handle = "%v"
}
`, testEnvironmentId, dbHandle)
}

func testAccAptibleDatabaseUpdate(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
    env_id = %d
	handle = "%v"
	container_size = %d
	disk_size = %d
}
`, testEnvironmentId, dbHandle, 512, 20)
}

func testAccAptibleDatabaseInvalidDBType(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
    env_id = %d
	handle = "%v"
	database_type = "%v"
}
`, testEnvironmentId, dbHandle, "non-existent-db")
}

func testAccAptibleDatabaseInvalidContainerSize(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
    env_id = %d
	handle = "%v"
	container_size = %d
}
`, testEnvironmentId, dbHandle, 0)
}

func testAccAptibleDatabaseInvalidDiskSize(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
    env_id = %d
	handle = "%v"
	disk_size = %d
}
`, testEnvironmentId, dbHandle, 0)
}

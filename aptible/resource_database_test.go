package aptible

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDatabase_basic(t *testing.T) {
	dbHandle := acctest.RandString(10)

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
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
					resource.TestMatchResourceAttr("aptible_database.test", "default_connection_url", regexp.MustCompile(`postgresql:.*\.aptible\.in:.*`)),
					resource.TestCheckResourceAttrPair("aptible_database.test", "connection_urls.0", "aptible_database.test", "default_connection_url"),
				),
			},
			{
				ResourceName:      "aptible_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDatabase_redis(t *testing.T) {
	dbHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleDatabaseRedis(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_database.test", "database_type", "redis"),
					resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
					resource.TestMatchResourceAttr("aptible_database.test", "default_connection_url", regexp.MustCompile(`redis:.*\.aptible\.in:.*`)),
					resource.TestCheckResourceAttr("aptible_database.test", "connection_urls.#", "2"),
					resource.TestCheckResourceAttrPair("aptible_database.test", "connection_urls.0", "aptible_database.test", "default_connection_url"),
					resource.TestMatchResourceAttr("aptible_database.test", "connection_urls.1", regexp.MustCompile(`rediss:.*\.aptible\.in:.*`)),
				),
			},
			{
				ResourceName:      "aptible_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDatabase_version(t *testing.T) {
	dbHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleDatabaseVersion(dbHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_database.test", "version", "9.4"),
					resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),
				),
			},
			{
				ResourceName:      "aptible_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDatabase_update(t *testing.T) {
	dbHandle := acctest.RandString(10)

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
					resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
					resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),
				),
			},
			{
				ResourceName:      "aptible_database.test",
				ImportState:       true,
				ImportStateVerify: true,
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
	dbHandle := acctest.RandString(10)

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
				ExpectError: regexp.MustCompile(`expected container_size to be one of .*, got 0`),
			},
			{
				Config:      testAccAptibleDatabaseInvalidDiskSize(dbHandle),
				ExpectError: regexp.MustCompile(`expected disk_size to be in the range .*, got 0`),
			},
		},
	})
}

func testAccCheckDatabaseDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation

	//lintignore:R018
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

func testAccAptibleDatabaseRedis(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
    env_id = %d
	handle = "%v"
  database_type = "redis"
	container_profile = "m5"
}
`, testEnvironmentId, dbHandle)
}

func testAccAptibleDatabaseVersion(dbHandle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
  env_id = %d
	handle = "%v"
	version = "9.4"
	database_type = "postgresql"
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

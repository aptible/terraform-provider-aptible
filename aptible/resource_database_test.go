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

	// Can't use an aptible_environment TF resource with databases because, when
	// the destroy is attempted, the environment will not permit deletion due to
	// the database's final backup
	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDatabaseDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleDatabaseBasic(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "3000"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
						resource.TestMatchResourceAttr("aptible_database.test", "default_connection_url", regexp.MustCompile(`postgresql://.*@db-.*`)),
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
	})
}

func TestAccResourceDatabase_redis(t *testing.T) {
	dbHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDatabaseDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleDatabaseRedis(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "redis"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "3000"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
						resource.TestMatchResourceAttr("aptible_database.test", "default_connection_url", regexp.MustCompile(`redis://.*@db-.*`)),
						resource.TestCheckResourceAttr("aptible_database.test", "connection_urls.#", "2"),
						resource.TestCheckResourceAttrPair("aptible_database.test", "connection_urls.0", "aptible_database.test", "default_connection_url"),
						resource.TestMatchResourceAttr("aptible_database.test", "connection_urls.1", regexp.MustCompile(`rediss://.*@db-.*`)),
					),
				},
				{
					ResourceName:      "aptible_database.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceDatabase_version(t *testing.T) {
	dbHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDatabaseDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleDatabaseVersion(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "version", "9.4"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "3000"),
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
	})
}

func TestAccResourceDatabase_update(t *testing.T) {
	dbHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDatabaseDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleDatabaseBasic(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "3000"),
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
					Config: testAccAptibleDatabaseUpdate(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "512"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "r5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "4000"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "20"),
					),
				},
			},
		})
	})
}

func TestAccResourceDatabase_expectError(t *testing.T) {
	dbHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDatabaseDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleDatabaseInvalidDBType(env.ID, dbHandle),
					ExpectError: regexp.MustCompile(`expected database_type to be one of .*, got non-existent-db`),
				},
				{
					Config:      testAccAptibleDatabaseInvalidContainerSize(env.ID, dbHandle),
					ExpectError: regexp.MustCompile(`expected container_size to be one of .*, got 0`),
				},
				{
					Config:      testAccAptibleDatabaseInvalidDiskSize(env.ID, dbHandle),
					ExpectError: regexp.MustCompile(`expected disk_size to be in the range .*, got 0`),
				},
			},
		})
	})
}

func TestAccResourceDatabase_scale(t *testing.T) {
	dbHandle := acctest.RandString(10)

	// Can't use an aptible_environment TF resource with databases because, when
	// the destroy is attempted, the environment will not permit deletion due to
	// the database's final backup
	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDatabaseDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleDatabaseScale(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "r5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "4000"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "12"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_image_id"),
						resource.TestMatchResourceAttr("aptible_database.test", "default_connection_url", regexp.MustCompile(`postgresql://.*@db-.*`)),
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
	})
}

func testAccCheckDatabaseDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
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

func testAccAptibleDatabaseBasic(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
	}
`, envId, dbHandle)
}

func testAccAptibleDatabaseRedis(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		database_type = "redis"
		container_profile = "m5"
	}
`, envId, dbHandle)
}

func testAccAptibleDatabaseVersion(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		version = "9.4"
		database_type = "postgresql"
	}
`, envId, dbHandle)
}

func testAccAptibleDatabaseUpdate(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		container_size = %d
		disk_size = %d
		container_profile = "r5"
		iops = 4000
	}
`, envId, dbHandle, 512, 20)
}

func testAccAptibleDatabaseInvalidDBType(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
	env_id = %d
		handle = "%v"
		database_type = "%v"
	}
`, envId, dbHandle, "non-existent-db")
}

func testAccAptibleDatabaseInvalidContainerSize(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
	env_id = %d
		handle = "%v"
		container_size = %d
	}
`, envId, dbHandle, 0)
}

func testAccAptibleDatabaseInvalidDiskSize(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
	env_id = %d
		handle = "%v"
		disk_size = %d
	}
`, envId, dbHandle, 0)
}

func testAccAptibleDatabaseScale(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		container_profile = "r5"
		iops = 4000
		disk_size = 12
	}
`, envId, dbHandle)
}

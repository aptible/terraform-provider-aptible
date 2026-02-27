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

func TestAccResourceReplica_basic(t *testing.T) {
	dbHandle := acctest.RandString(10)
	replicaHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckReplicaDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleReplicaBasic(env.ID, dbHandle, replicaHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),

						resource.TestCheckResourceAttr("aptible_replica.test", "handle", replicaHandle),
						resource.TestCheckResourceAttr("aptible_replica.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_replica.test", "iops", "3000"),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "10"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "replica_id"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "default_connection_url"),
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

func TestAccResourceReplica_withoutBackups(t *testing.T) {
	dbHandle := acctest.RandString(10)
	replicaHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckReplicaDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleReplicaWithoutBackups(env.ID, dbHandle, replicaHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttr("aptible_database.test", "enable_backups", "true"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),

						resource.TestCheckResourceAttr("aptible_replica.test", "handle", replicaHandle),
						resource.TestCheckResourceAttr("aptible_replica.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_replica.test", "iops", "3000"),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "10"),
						// TEMPORARILY DISABLED - PITR must be disabled before backups can be disabled
						// resource.TestCheckResourceAttr("aptible_replica.test", "enable_backups", "false"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "replica_id"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "default_connection_url"),
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

func TestAccResourceReplica_update(t *testing.T) {
	dbHandle := acctest.RandString(10)
	replicaHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckReplicaDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleReplicaBasic(env.ID, dbHandle, replicaHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttr("aptible_database.test", "enable_backups", "true"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),

						resource.TestCheckResourceAttr("aptible_replica.test", "handle", replicaHandle),
						resource.TestCheckResourceAttr("aptible_replica.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_replica.test", "iops", "3000"),
						resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "10"),
						resource.TestCheckResourceAttr("aptible_replica.test", "enable_backups", "true"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "replica_id"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "default_connection_url"),
					),
				},
				{
					ResourceName:      "aptible_database.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				// TEMPORARILY DISABLED - Step 3: PITR must be disabled before backups can be disabled
				// {
				// 	Config: testAccAptibleReplicaUpdate(env.ID, dbHandle, replicaHandle),
				// 	Check: resource.ComposeTestCheckFunc(
				// 		resource.TestCheckResourceAttr("aptible_replica.test", "container_profile", "r5"),
				// 		resource.TestCheckResourceAttr("aptible_replica.test", "iops", "4000"),
				// 		resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "512"),
				// 		resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "20"),
				// 		resource.TestCheckResourceAttr("aptible_replica.test", "enable_backups", "false"),
				// 	),
				// },
			},
		})
	})
}

func TestAccResourceReplica_expectError(t *testing.T) {
	replicaHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckReplicaDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleReplicaInvalidContainerSize(env.ID, replicaHandle),
					ExpectError: regexp.MustCompile(`expected container_size to be one of .*, got 0`),
				},
				{
					Config:      testAccAptibleReplicaInvalidDiskSize(env.ID, replicaHandle),
					ExpectError: regexp.MustCompile(`expected disk_size to be in the range .*, got 0`),
				},
			},
		})
	})
}

func testAccCheckReplicaDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation

	//lintignore:R018
	time.Sleep(30 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_replica" {
			continue
		}

		databaseID, err := strconv.Atoi(rs.Primary.Attributes["primary_database_id"])
		if err != nil {
			return err
		}

		replicaID, err := strconv.Atoi(rs.Primary.Attributes["replica_id"])
		if err != nil {
			return err
		}

		// Check replica is deleted first, then the primary database
		database, err := client.GetReplica(int64(replicaID))
		log.Println("Deleted? ", database.Deleted)
		if !database.Deleted {
			return fmt.Errorf("replica %v not removed", replicaID)
		}
		if err != nil {
			return err
		}

		database, err = client.GetDatabase(int64(databaseID))
		log.Println("Deleted? ", database.Deleted)
		if !database.Deleted {
			return fmt.Errorf("database %v not removed", databaseID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func TestAccResourceReplica_scale(t *testing.T) {
	dbHandle := acctest.RandString(10)
	replicaHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckReplicaDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleReplicaScale(env.ID, dbHandle, replicaHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),

						resource.TestCheckResourceAttr("aptible_replica.test", "handle", replicaHandle),
						resource.TestCheckResourceAttr("aptible_replica.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_replica.test", "iops", "4000"),
						resource.TestCheckResourceAttr("aptible_replica.test", "container_profile", "r5"),
						resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "12"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "replica_id"),
						resource.TestCheckResourceAttrSet("aptible_replica.test", "default_connection_url"),
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

func testAccAptibleReplicaBasic(envId int64, dbHandle string, replicaHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
	}

	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_database_id = aptible_database.test.database_id
	}
	`, envId, dbHandle, envId, replicaHandle)
}

func testAccAptibleReplicaWithoutBackups(envId int64, dbHandle string, replicaHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
	}

	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_database_id = aptible_database.test.database_id
		enable_backups = false
	}
	`, envId, dbHandle, envId, replicaHandle)
}

//nolint:unused // Will be re-enabled once PITR bug is fixed
func testAccAptibleReplicaUpdate(envId int64, dbHandle string, repHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
	}

	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_database_id = aptible_database.test.database_id
		container_size = %d
		disk_size = %d
		container_profile = "r5"
		iops = 4000
		enable_backups = false
	}
	`, envId, dbHandle, envId, repHandle, 512, 20)
}

func testAccAptibleReplicaInvalidContainerSize(envId int64, replicaHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_database_id = "1"
		container_size = %d
	}
	`, envId, replicaHandle, 0)
}

func testAccAptibleReplicaInvalidDiskSize(envId int64, replicaHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_database_id = "1"
		disk_size = %d
	}
	`, envId, replicaHandle, 0)
}

func testAccAptibleReplicaScale(envId int64, dbHandle string, replicaHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
	}

	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_database_id = aptible_database.test.database_id
		container_profile = "r5"
		iops = 4000
		disk_size = 12
	}
	`, envId, dbHandle, envId, replicaHandle)
}

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

func TestAccResourceReplica_basic(t *testing.T) {
	dbHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	replicaHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleReplicaBasic(dbHandle, replicaHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_db.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_db.test", "env_id", strconv.Itoa(TestEnvironmentID)),
					resource.TestCheckResourceAttr("aptible_db.test", "db_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_db.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_db.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "db_id"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "connection_url"),

					resource.TestCheckResourceAttr("aptible_replica.test", "handle", replicaHandle),
					resource.TestCheckResourceAttr("aptible_replica.test", "env_id", strconv.Itoa(TestEnvironmentID)),
					resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_replica.test", "replica_id"),
					resource.TestCheckResourceAttrSet("aptible_replica.test", "connection_url"),
				),
			},
		},
	})
}

func TestAccResourceReplica_update(t *testing.T) {
	dbHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	replicaHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleReplicaBasic(dbHandle, replicaHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_db.test", "handle", dbHandle),
					resource.TestCheckResourceAttr("aptible_db.test", "env_id", strconv.Itoa(TestEnvironmentID)),
					resource.TestCheckResourceAttr("aptible_db.test", "db_type", "postgresql"),
					resource.TestCheckResourceAttr("aptible_db.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_db.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "db_id"),
					resource.TestCheckResourceAttrSet("aptible_db.test", "connection_url"),

					resource.TestCheckResourceAttr("aptible_replica.test", "handle", replicaHandle),
					resource.TestCheckResourceAttr("aptible_replica.test", "env_id", strconv.Itoa(TestEnvironmentID)),
					resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "1024"),
					resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "10"),
					resource.TestCheckResourceAttrSet("aptible_replica.test", "replica_id"),
					resource.TestCheckResourceAttrSet("aptible_replica.test", "connection_url"),
				),
			},
			{
				Config: testAccAptibleReplicaUpdate(dbHandle, replicaHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_replica.test", "container_size", "512"),
					resource.TestCheckResourceAttr("aptible_replica.test", "disk_size", "20"),
				),
			},
		},
	})
}

func TestAccResourceReplica_expectError(t *testing.T) {
	replicaHandle := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckReplicaDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAptibleReplicaInvalidContainerSize(replicaHandle),
				ExpectError: regexp.MustCompile(`expected container_size to be in the range .*, got 0`),
			},
			{
				Config:      testAccAptibleReplicaInvalidDiskSize(replicaHandle),
				ExpectError: regexp.MustCompile(`config is invalid: expected disk_size to be in the range .*, got 0`),
			},
		},
	})
}

func testAccCheckReplicaDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation
	time.Sleep(30 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_replica" {
			continue
		}

		db_id, err := strconv.Atoi(rs.Primary.Attributes["primary_db_id"])
		if err != nil {
			return err
		}

		replica_id, err := strconv.Atoi(rs.Primary.Attributes["replica_id"])
		if err != nil {
			return err
		}

		// Check replica is deleted first, then the primary database
		_, deleted, err := client.GetReplica(int64(replica_id))
		log.Println("Deleted? ", deleted)
		if !deleted {
			return fmt.Errorf("Replica %v not removed", replica_id)
		}
		if err != nil {
			return err
		}

		_, deleted, err = client.GetDatabase(int64(db_id))
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

func testAccAptibleReplicaBasic(dbHandle string, replicaHandle string) string {
	output := fmt.Sprintf(`
resource "aptible_db" "test" {
    env_id = %d
	handle = "%v"
}

resource "aptible_replica" "test" {
	env_id = %d
	handle = "%v"
	primary_db_id = aptible_db.test.db_id
}
`, TestEnvironmentID, dbHandle, TestEnvironmentID, replicaHandle)
	log.Println("HCL generated:", output)
	return output
}

func testAccAptibleReplicaUpdate(dbHandle string, repHandle string) string {
	output := fmt.Sprintf(`
	resource "aptible_db" "test" {
		env_id = %d
		handle = "%v"
	}

	resource "aptible_replica" "test" {
		env_id = %d
		handle = "%v"
		primary_db_id = aptible_db.test.db_id
		container_size = %d
		disk_size = %d
	}
`, TestEnvironmentID, dbHandle, TestEnvironmentID, repHandle, 512, 20)
	log.Println("HCL generated:", output)
	return output
}

func testAccAptibleReplicaInvalidContainerSize(replicaHandle string) string {
	return fmt.Sprintf(`
resource "aptible_replica" "test" {
	env_id = %d
	handle = "%v"
	primary_db_id = "1"
	container_size = %d
}
`, TestEnvironmentID, replicaHandle, 0)
}

func testAccAptibleReplicaInvalidDiskSize(replicaHandle string) string {
	return fmt.Sprintf(`
resource "aptible_replica" "test" {
	env_id = %d
	handle = "%v"
	primary_db_id = "1"
	disk_size = %d
}
`, TestEnvironmentID, replicaHandle, 0)
}

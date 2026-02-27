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

func TestAccResourceDatabase_withoutBackups(t *testing.T) {
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
					Config: testAccAptibleDatabaseWithoutBackups(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_database.test", "database_type", "postgresql"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_size", "1024"),
						resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "m5"),
						resource.TestCheckResourceAttr("aptible_database.test", "iops", "3000"),
						resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "10"),
						// TEMPORARILY DISABLED - PITR must be disabled before backups can be disabled
						// resource.TestCheckResourceAttr("aptible_database.test", "enable_backups", "false"),
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
						checkConnectionUrlsInclude("aptible_database.test", []string{
							`redis://.*@db-.*`,
							`rediss://.*@db-.*`,
						}),
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
						resource.TestCheckResourceAttr("aptible_database.test", "enable_backups", "true"),
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
				// TEMPORARILY DISABLED - Step 3: PITR must be disabled before backups can be disabled
				// {
				// 	Config: testAccAptibleDatabaseUpdate(env.ID, dbHandle),
				// 	Check: resource.ComposeTestCheckFunc(
				// 		resource.TestCheckResourceAttr("aptible_database.test", "container_size", "512"),
				// 		resource.TestCheckResourceAttr("aptible_database.test", "container_profile", "r5"),
				// 		resource.TestCheckResourceAttr("aptible_database.test", "iops", "4000"),
				// 		resource.TestCheckResourceAttr("aptible_database.test", "enable_backups", "false"),
				// 		resource.TestCheckResourceAttr("aptible_database.test", "disk_size", "20"),
				// 	),
				// },
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

func testAccAptibleDatabaseWithoutBackups(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		enable_backups = false
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

//nolint:unused // Will be re-enabled once PITR bug is fixed
func testAccAptibleDatabaseUpdate(envId int64, dbHandle string) string {
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		container_size = %d
		disk_size = %d
		container_profile = "r5"
		iops = 4000
		enable_backups = false
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

func checkConnectionUrlsInclude(resourceName string, expectedPatterns []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		connectionURLs := rs.Primary.Attributes["connection_urls.#"]
		urlCount, err := strconv.Atoi(connectionURLs)
		if err != nil {
			return fmt.Errorf("error parsing connection_urls count: %v", err)
		}

		// Verify that the expected number of URLs are present
		if urlCount != len(expectedPatterns) {
			return fmt.Errorf("expected %d connection URLs, but got %d", len(expectedPatterns), urlCount)
		}

		// Track which patterns have been matched
		matchedPatterns := make(map[string]bool)

		// Loop over each connection URL and check it against the expected patterns
		for i := 0; i < urlCount; i++ {
			attrKey := fmt.Sprintf("connection_urls.%d", i)
			url := rs.Primary.Attributes[attrKey]

			matched := false
			for _, pattern := range expectedPatterns {
				if matchedPatterns[pattern] {
					continue // Skip already matched patterns
				}

				matched, err = regexp.MatchString(pattern, url)
				if err != nil {
					return fmt.Errorf("error matching URL with pattern: %v", err)
				}
				if matched {
					matchedPatterns[pattern] = true
					break
				}
			}

			if !matched {
				return fmt.Errorf("URL %s did not match any expected pattern", url)
			}
		}

		for _, pattern := range expectedPatterns {
			if !matchedPatterns[pattern] {
				return fmt.Errorf("pattern %s was not matched", pattern)
			}
		}

		return nil
	}
}

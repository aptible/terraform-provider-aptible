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

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckLogDrainDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleLogDrainElastic(env.ID, rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "elasticsearch_database"),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "logging_token", "test_token"),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "pipeline", "test_token"),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
						resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
					),
				}, {
					ResourceName:      "aptible_log_drain.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceLogDrain_syslog(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainSyslog(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "syslog_tls_tcp"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_host", "syslog.aptible.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_port", "1234"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogDrain_https(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainHttps(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "https_post"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "url", "https://test.aptible.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "false"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "false"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "false"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "true"),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogDrain_datadog(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainDatadog(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "datadog"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_username", "test_username"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "logging_token", "test_token"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "token", "test_username"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "tags", "test_token"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogDrain_sumologic(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainSumologic(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "sumologic"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "url", "https://www.sumologic.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogDrain_logdna(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainLogdna(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "logdna"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_username", "test_username"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "token", "test_username"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogDrain_papertrail(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainPapertrail(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "papertrail"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_host", "www.papertrail.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_port", "1234"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
				),
			}, {
				ResourceName:      "aptible_log_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceLogDrain_solarwinds(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainSolarwinds(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "solarwinds"),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_log_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_host", "www.solarwinds.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "logging_token", "secrettoken"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_proxies", "false"),
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
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
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

func testAccAptibleLogDrainElastic(envID int64, handle string) string {
	// Cannot use aptible_environment TF resource with a database since the final
	// DB backup will prevent the environment from being deleted
	return fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		database_type = "elasticsearch"
		container_size = 1024
		disk_size = 10
	}

	resource "aptible_log_drain" "test" {
		env_id = %d
		database_id = aptible_database.test.database_id
		handle = "%v"
		drain_type = "elasticsearch_database"
		pipeline = "test_token"
	}
`, envID, handle, envID, handle)
}

func testAccAptibleLogDrainSyslog(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		drain_type = "syslog_tls_tcp"
		drain_host = "syslog.aptible.com"
		drain_port = "1234"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleLogDrainHttps(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		url = "https://test.aptible.com"
		drain_apps = false
		drain_databases = false
		drain_ephemeral_sessions = false
		drain_proxies = true
		drain_type = "https_post"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleLogDrainDatadog(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		drain_type = "datadog"
		token = "test_username"
		tags = "test_token"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleLogDrainSumologic(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		drain_type = "sumologic"
		url = "https://www.sumologic.com"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleLogDrainLogdna(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		drain_type = "logdna"
		token = "test_username"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleLogDrainPapertrail(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		drain_type = "papertrail"
		drain_host = "www.papertrail.com"
		drain_port = "1234"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleLogDrainSolarwinds(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_log_drain" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		drain_type = "solarwinds"
		drain_host = "www.solarwinds.com"
		logging_token = "secrettoken"
	}
`, handle, testOrganizationId, testStackId, handle)
}

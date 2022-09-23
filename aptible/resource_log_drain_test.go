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
}

func TestAccResourceLogDrain_syslog(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainSyslog(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "syslog_tls_tcp"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainHttps(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "https_post"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "url", "https://test.aptible.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_apps", "false"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_databases", "true"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_ephemeral_sessions", "true"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainDatadog(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "datadog"),
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

func TestAccResourceLogDrain_fields(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLogDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleLogDrainFields(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_type", "https_post"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "url", "https://test.aptible.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_username", "test_username"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_host", "syslog.aptible.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_password", "test_password"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "drain_port", "1234"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "logging_token", "test_token"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "url", "https://test.aptible.com"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "token", "test_username"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "tags", "test_token"),
					resource.TestCheckResourceAttr("aptible_log_drain.test", "pipeline", "test_token"),
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
resource "aptible_database" "test" {
	env_id = %d
	handle = "%v"
	database_type = "postgresql"
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
`, testEnvironmentId, handle, testEnvironmentId, handle)
}

func testAccAptibleLogDrainSyslog(handle string) string {
	return fmt.Sprintf(`
resource "aptible_log_drain" "test" {
    env_id = %d
    handle = "%v"
    drain_type = "syslog_tls_tcp"
    drain_host = "syslog.aptible.com"
    drain_port = "1234"
}
`, testEnvironmentId, handle)
}

func testAccAptibleLogDrainHttps(handle string) string {
	return fmt.Sprintf(`
resource "aptible_log_drain" "test" {
    env_id = %d
    handle = "%v"
    url = "https://test.aptible.com"
    drain_apps = false
    drain_proxies = true
    drain_type = "https_post"
}
`, testEnvironmentId, handle)
}

func testAccAptibleLogDrainDatadog(handle string) string {
	return fmt.Sprintf(`
resource "aptible_log_drain" "test" {
    env_id = %d
    handle = "%v"
    drain_type = "datadog"
    token = "test_username"
    tags = "test_token"
}
`, testEnvironmentId, handle)
}

func testAccAptibleLogDrainFields(handle string) string {
	return fmt.Sprintf(`
resource "aptible_log_drain" "test" {
    env_id = %d
    handle = "%v"
    drain_type = "https_post"
    drain_username = "test_username"
    drain_host = "syslog.aptible.com"
    drain_password = "test_password"
    drain_port = "1234"
    logging_token = "test_token"
    url = "https://test.aptible.com"
    token = "test_username"
    tags = "test_token"
    pipeline = "test_token"
}
`, testEnvironmentId, handle)
}

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

func TestAccResourceMetricDrain_influxdb_database(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMetricDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleMetricDrainInfluxDBDatabase(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "drain_type", "influxdb_database"),
				),
			}, {
				ResourceName:      "aptible_metric_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceMetricDrain_influxdb(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMetricDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleMetricDrainInfluxDB(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "drain_type", "influxdb"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "url", "https://test.aptible.com:2022"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "username", "test_user"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "password", "test_password"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "database", "test_db"),
				),
			}, {
				ResourceName:      "aptible_metric_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceMetricDrain_datadog(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMetricDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleMetricDrainDatadog(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "env_id", strconv.Itoa(testEnvironmentId)),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "drain_type", "datadog"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "api_key", "test_api_key"),
				),
			}, {
				ResourceName:      "aptible_metric_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMetricDrainDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*aptible.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_metric_drain" {
			continue
		}

		metricDrainID, err := strconv.Atoi(rs.Primary.Attributes["metric_drain_id"])
		if err != nil {
			return err
		}

		metricDrain, err := client.GetMetricDrain(int64(metricDrainID))
		log.Println("Deleted? ", metricDrain.Deleted)
		if !metricDrain.Deleted {
			return fmt.Errorf("metric drain %v not removed", metricDrainID)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAptibleMetricDrainInfluxDBDatabase(handle string) string {
	return fmt.Sprintf(`
resource "aptible_database" "test" {
	env_id = %d
	handle = "%v"
	database_type = "influxdb"
	container_size = 1024
	disk_size = 10
}

resource "aptible_metric_drain" "test" {
    env_id = %d
    database_id = aptible_database.test.database_id
    handle = "%v"
    drain_type = "influxdb_database"
}
`, testEnvironmentId, handle, testEnvironmentId, handle)
}

func testAccAptibleMetricDrainInfluxDB(handle string) string {
	return fmt.Sprintf(`
resource "aptible_metric_drain" "test" {
    env_id = %d
    handle = "%v"
    drain_type = "influxdb"
		url = "https://test.aptible.com:2022"
		username = "test_user"
		password = "test_password"
		database = "test_db"
}
`, testEnvironmentId, handle)
}

func testAccAptibleMetricDrainDatadog(handle string) string {
	return fmt.Sprintf(`
resource "aptible_metric_drain" "test" {
    env_id = %d
    handle = "%v"
    drain_type = "datadog"
    api_key = "test_api_key"
		series_url = "https://test.aptible.com:2022"
}
`, testEnvironmentId, handle)
}

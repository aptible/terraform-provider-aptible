package aptible

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMetricDrain_validation(t *testing.T) {
	requiredAttrs := []string{"env_id", "handle", "drain_type"}
	var testSteps []resource.TestStep

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      `resource "aptible_metric_drain" "test" {}`,
			ExpectError: regexp.MustCompile(fmt.Sprintf("%q is required", attr)),
		})
	}

	testSteps = append(testSteps, resource.TestStep{
		PlanOnly: true,
		Config: `
			resource "aptible_metric_drain" "test" {
				env_id = 1
				handle = "test_drain"
				drain_type = "foo"
			}
		`,
		ExpectError: regexp.MustCompile(
			regexp.QuoteMeta(`expected drain_type to be one of ["influxdb_database" "influxdb" "influxdb2" "datadog"], got foo`),
		),
	})

	for _, attr := range []string{"url", "series_url"} {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly: true,
			Config: fmt.Sprintf(`
				resource "aptible_metric_drain" "test" {
					env_id = 1
					handle = "test_drain"
					drain_type = "influxdb"
					%s = "I have no scheme"
				}
			`, attr),
			ExpectError: regexp.MustCompile(fmt.Sprintf(`expected %q to have a scheme`, attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps:             testSteps,
	})
}

func TestAccResourceMetricDrain_influxdb_database_validation(t *testing.T) {
	requiredAttrs := []string{"database_id"}
	notAllowedAttrs := []string{"url", "username", "password", "database", "api_key", "series_url"}
	var testSteps []resource.TestStep

	validationConfig := `
		resource "aptible_metric_drain" "test" {
			env_id = 1
			handle = "test_drain"
			drain_type = "influxdb_database"
			url = "scheme://host"
			username = "test_user"
			password = "test_pass"
			database = "test_db"
			api_key = "test_api_key"
			series_url = "scheme://host"
		}
	`

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is required when drain_type = "influxdb_database"`, attr)),
		})
	}

	for _, attr := range notAllowedAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is not allowed when drain_type = "influxdb_database"`, attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps:             testSteps,
	})
}

func TestAccResourceMetricDrain_influxdb_database(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckMetricDrainDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleMetricDrainInfluxDBDatabase(env.ID, rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_metric_drain.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_metric_drain.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_metric_drain.test", "drain_type", "influxdb_database"),
					),
				},
				{
					ResourceName:      "aptible_metric_drain.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceMetricDrain_influxdb_validation(t *testing.T) {
	requiredAttrs := []string{"url", "username", "password", "database"}
	notAllowedAttrs := []string{"database_id"}
	var testSteps []resource.TestStep

	validationConfig := `
		resource "aptible_metric_drain" "test" {
			env_id = 1
			handle = "test_drain"
			drain_type = "influxdb"
			database_id = 1
		}
	`

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is required when drain_type = "influxdb"`, attr)),
		})
	}

	for _, attr := range notAllowedAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is not allowed when drain_type = "influxdb"`, attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps:             testSteps,
	})
}

func TestAccResourceMetricDrain_influxdb(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleMetricDrainInfluxDB(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_metric_drain.test", "env_id"),
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

func TestAccResourceMetricDrain_influxdb2_validation(t *testing.T) {
	requiredAttrs := []string{"url", "api_key", "bucket", "organization"}
	notAllowedAttrs := []string{"database_id", "username", "password"}
	var testSteps []resource.TestStep

	validationConfig := `
		resource "aptible_metric_drain" "test" {
			env_id = 1
			handle = "test_drain"
			drain_type = "influxdb2"
			database_id = 1
			username = "someone"
			password = "something"
		}
	`

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is required when drain_type = "influxdb2"`, attr)),
		})
	}

	for _, attr := range notAllowedAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is not allowed when drain_type = "influxdb2"`, attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps:             testSteps,
	})
}

func TestAccResourceMetricDrain_influxdb2(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleMetricDrainInfluxDB2(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_metric_drain.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "handle", rHandle),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "drain_type", "influxdb2"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "url", "https://test.aptible.com:2022"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "api_key", "xxx-yyy-zzz"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "bucket", "aBucket"),
					resource.TestCheckResourceAttr("aptible_metric_drain.test", "organization", "ImportantOrg"),
				),
			}, {
				ResourceName:      "aptible_metric_drain.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceMetricDrain_datadog_validation(t *testing.T) {
	requiredAttrs := []string{"api_key"}
	notAllowedAttrs := []string{"database_id"}
	var testSteps []resource.TestStep

	validationConfig := `
		resource "aptible_metric_drain" "test" {
			env_id = 1
			handle = "test_drain"
			drain_type = "datadog"
			database_id = 1
		}
	`

	for _, attr := range requiredAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is required when drain_type = "datadog"`, attr)),
		})
	}

	for _, attr := range notAllowedAttrs {
		testSteps = append(testSteps, resource.TestStep{
			PlanOnly:    true,
			Config:      validationConfig,
			ExpectError: regexp.MustCompile(fmt.Sprintf(`%q is not allowed when drain_type = "datadog"`, attr)),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps:             testSteps,
	})
}

func TestAccResourceMetricDrain_datadog(t *testing.T) {
	rHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMetricDrainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleMetricDrainDatadog(rHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_metric_drain.test", "env_id"),
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
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
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

func testAccAptibleMetricDrainInfluxDBDatabase(envId int64, handle string) string {
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
	`, envId, handle, envId, handle)
}

func testAccAptibleMetricDrainInfluxDB(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_metric_drain" "test" {
			env_id = aptible_environment.test.env_id
			handle = "%v"
			drain_type = "influxdb"
			url = "https://test.aptible.com:2022"
			username = "test_user"
			password = "test_password"
			database = "test_db"
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleMetricDrainInfluxDB2(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_metric_drain" "test" {
			env_id = aptible_environment.test.env_id
			handle = "%v"
			drain_type = "influxdb2"
			url = "https://test.aptible.com:2022"
			api_key = "xxx-yyy-zzz"
			bucket = "aBucket"
			organization = "ImportantOrg"
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleMetricDrainDatadog(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_metric_drain" "test" {
			env_id = aptible_environment.test.env_id
			handle = "%v"
			drain_type = "datadog"
			api_key = "test_api_key"
			series_url = "https://test.aptible.com:2022"
	}
	`, handle, testOrganizationId, testStackId, handle)
}

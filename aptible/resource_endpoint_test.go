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

func TestAccResourceEndpoint_customDomain(t *testing.T) {
	appHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointCustomDomain(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_endpoint.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "https"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "domain", "www.aptible-test-demo.fake"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "alb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "virtual_domain", "www.aptible-test-demo.fake"),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "external_hostname", regexp.MustCompile(`elb.*`)),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "dns_validation_record", "_acme-challenge.www.aptible-test-demo.fake"),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "dns_validation_value", regexp.MustCompile(`acme\.elb.*`)),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEndpoint_appContainerNoPort(t *testing.T) {
	appHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointAppContainerNoPort(appHandle, "20"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_endpoint.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "app"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "https"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "alb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "virtual_domain", regexp.MustCompile(`app-.*`)),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "external_hostname", regexp.MustCompile(`elb.`)),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "dns_validation_record"),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "dns_validation_value"),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEndpoint_appContainerPort(t *testing.T) {
	appHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointAppContainerPort(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_endpoint.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "app"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "https"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "container_port", "80"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "alb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "virtual_domain", regexp.MustCompile(`app-.*`)),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "external_hostname", regexp.MustCompile(`elb.*`)),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "dns_validation_record"),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "dns_validation_value"),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEndpoint_appContainerPorts(t *testing.T) {
	appHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointAppContainerPorts(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_endpoint.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "app"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "tcp"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "container_ports.0", "80"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "container_ports.1", "443"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "elb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "virtual_domain", regexp.MustCompile(`app-.*`)),
					resource.TestMatchResourceAttr("aptible_endpoint.test", "external_hostname", regexp.MustCompile(`elb.*`)),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "dns_validation_record"),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "dns_validation_value"),
					resource.TestCheckNoResourceAttr("aptible_endpoint.test", "load_balancing_algorithm_type"),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEndpoint_db(t *testing.T) {
	dbHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckEndpointDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleEndpointDatabase(env.ID, dbHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_database.test", "handle", dbHandle),
						resource.TestCheckResourceAttr("aptible_database.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttrSet("aptible_database.test", "database_id"),
						resource.TestCheckResourceAttrSet("aptible_database.test", "default_connection_url"),

						resource.TestCheckResourceAttr("aptible_endpoint.test", "env_id", strconv.Itoa(int(env.ID))),
						resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "database"),
						resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "tcp"),
						resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "false"),
						resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "elb"),
						resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
					),
				},
				{
					ResourceName:      "aptible_endpoint.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceEndpoint_updateIPWhitelist(t *testing.T) {
	appHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointAppContainerNoPort(appHandle, "21"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_app.test", "handle", appHandle),
					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
					resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),

					resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_endpoint.test", "env_id"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "resource_type", "app"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "endpoint_type", "https"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "internal", "true"),
					resource.TestCheckResourceAttr("aptible_endpoint.test", "platform", "alb"),
					resource.TestCheckResourceAttrSet("aptible_endpoint.test", "endpoint_id"),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAptibleEndpointUpdateIPWhitelist(appHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_endpoint.test", "ip_filtering.0", "1.1.1.1/32"),
				),
			},
		},
	})
}

func TestAccResourceEndpoint_provisionFailure(t *testing.T) {
	// Test that if the endpoint provision fails, subsequent applys will replace
	// the "tainted" resource
	appHandle := acctest.RandString(10)
	config := testAccAptibleEndpointBadPort(appHandle)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?i)fail.*provision.*endpoint`),
				// Check does not appear to work with ExpectError so we cannot use it to
				// verify that the resource is tainted but we can use an ImportState
				// step to verify that the resource was indeed saved to the state and
				// error + in state = tainted
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEndpoint_sharedUpgrade(t *testing.T) {
	// Test that a dedicated endpoint can be upgraded to a shared endpoint just
	// by flipping the shared flag and applying.
	appHandle := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointSetShared(appHandle, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_endpoint.test", "shared", "false"),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAptibleEndpointSetShared(appHandle, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_endpoint.test", "shared", "true"),
				),
			},
		},
	})
}

func TestAccResourceEndpoint_lbAlgorithm(t *testing.T) {
	appHandle := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAptibleEndpointLbAlgorithm(appHandle, "least_outstanding_requests"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aptible_endpoint.test", "load_balancing_algorithm_type", "least_outstanding_requests"),
				),
			},
			{
				ResourceName:      "aptible_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceEndpoint_expectError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAptibleEndpointInvalidResourceType(),
				ExpectError: regexp.MustCompile(`(?i)expected resource_type to be one of .*, got should-error`),
			},
			{
				Config:      testAccAptibleEndpointInvalidEndpointType(),
				ExpectError: regexp.MustCompile(`(?i)expected endpoint_type to be one of .*, got should-error`),
			},
			{
				Config:      testAccAptibleEndpointInvalidPlatform(),
				ExpectError: regexp.MustCompile(`(?i)expected platform to be one of .*, got should-error`),
			},
			{
				Config:      testAccAptibleEndpointInvalidDomain(),
				ExpectError: regexp.MustCompile(`(?i)managed endpoints must specify a domain`),
			},
			{
				Config:      testAccAptibleEndpointInvalidLbAlgorithm(),
				ExpectError: regexp.MustCompile(`(?i)expected load_balancing_algorithm_type to be one of .*, got should-error`),
			},
			{
				Config:      testAccAptibleEndpointInvalidContainerPort(),
				ExpectError: regexp.MustCompile(`(?i)expected container_port to be in the range \(1 \- 65535\)`),
			},
			{
				Config:      testAccAptibleEndpointInvalidContainerPortOnTcp(),
				ExpectError: regexp.MustCompile(`(?i)do not specify container port with a tls or tcp endpoint`),
			},
			{
				Config:      testAccAptibleEndpointInvalidContainerPortOnTls(),
				ExpectError: regexp.MustCompile(`(?i)do not specify container port with a tls or tcp endpoint`),
			},
			{
				Config:      testAccAptibleEndpointInvalidContainerPorts(),
				ExpectError: regexp.MustCompile(`(?i)expected container_ports.0 to be in the range \(1 \- 65535\)`),
			},
			{
				Config:      testAccAptibleEndpointInvalidContainerPortsOnHttp(),
				ExpectError: regexp.MustCompile(`(?i)do not specify container ports with https endpoint`),
			},
			{
				Config:      testAccAptibleEndpointInvalidMultipleContainerPortFields(),
				ExpectError: regexp.MustCompile(`(?i)do not specify container ports AND container port`),
			},
			{
				Config:      testAccAptibleEndpointInvalidSharedWithNoDomain(),
				ExpectError: regexp.MustCompile(`(?i)must specify a domain`),
			},
			{
				Config:      testAccAptibleEndpointInvalidSharedWithWildcardDomain(),
				ExpectError: regexp.MustCompile(`(?i)cannot use domain`),
			},
			{
				Config:      testAccAptibleEndpointInvalidLbAlgorithmWithElb(),
				ExpectError: regexp.MustCompile(`(?i)do not specify a load balancing algorithm with elb endpoint`),
			},
		},
	})
}

func testAccCheckEndpointDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
	// Allow time for deprovision operation to complete.
	// TODO: Replace this by waiting on the actual operation

	//lintignore:R018
	time.Sleep(30 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_endpoint" {
			continue
		}

		res_id, err := strconv.Atoi(rs.Primary.Attributes["resource_id"])
		if err != nil {
			return err
		}

		res_typ := rs.Primary.Attributes["resource_type"]
		if err != nil {
			return err
		}

		if res_typ == "app" {
			endpoint, err := client.GetApp(int64(res_id))
			log.Println("Deleted? ", endpoint.Deleted)
			if !endpoint.Deleted {
				return fmt.Errorf("App %v not removed", res_id)
			}

			if err != nil {
				return err
			}

		} else {
			endpoint, err := client.GetDatabase(int64(res_id))
			log.Println("Deleted? ", endpoint.Deleted)
			if !endpoint.Deleted {
				return fmt.Errorf("Database %v not removed", res_id)
			}

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func testAccAptibleEndpointCustomDomain(appHandle string) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:31"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		endpoint_type = "https"
		managed = true
		domain = "www.aptible-test-demo.fake"
		internal = true
		platform = "alb"
	}
`, appHandle, testOrganizationId, testStackId, appHandle)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointAppContainerPort(appHandle string) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/caddy-mirror:1"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		container_port = 80
		endpoint_type = "https"
		default_domain = true
		internal = true
		platform = "alb"
	}
`, appHandle, testOrganizationId, testStackId, appHandle)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointAppContainerNoPort(appHandle string, index string) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:%s"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		endpoint_type = "https"
		default_domain = true
		internal = true
		platform = "alb"
	}
`, appHandle, testOrganizationId, testStackId, appHandle, index)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointAppContainerPorts(appHandle string) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/caddy-mirror:2"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		container_ports = [80, 443]
		endpoint_type = "tcp"
		default_domain = true
		internal = true
		platform = "elb"
	}
`, appHandle, testOrganizationId, testStackId, appHandle)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointDatabase(envId int64, dbHandle string) string {
	output := fmt.Sprintf(`
	resource "aptible_database" "test" {
		env_id = %d
		handle = "%v"
		database_type = "postgresql"
		container_size = 1024
		disk_size = 10
	}

	resource "aptible_endpoint" "test" {
		env_id = %d
		resource_id = aptible_database.test.database_id
		resource_type = "database"
		endpoint_type = "tcp"
		internal = false
		platform = "elb"
	}
`, envId, dbHandle, envId)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointUpdateIPWhitelist(appHandle string) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:22"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		endpoint_type = "https"
		default_domain = true
		internal = true
		platform = "alb"
		ip_filtering = [
			"1.1.1.1/32",
		]
	}
`, appHandle, testOrganizationId, testStackId, appHandle)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointSetShared(appHandle string, shared bool) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:32"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		endpoint_type = "https"
		default_domain = true
		platform = "alb"
                shared = %t
	}
`, appHandle, testOrganizationId, testStackId, appHandle, shared)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointLbAlgorithm(appHandle string, lbAlgorithm string) string {
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:31"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		endpoint_type = "https"
		default_domain = true
		platform = "alb"
		load_balancing_algorithm_type = "%s"
	}
`, appHandle, testOrganizationId, testStackId, appHandle, lbAlgorithm)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointBadPort(appHandle string) string {
	// Use a bad port to make the provision operation fail
	output := fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:23"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
		}
	}

	resource "aptible_endpoint" "test" {
		env_id = aptible_environment.test.env_id
		resource_id = aptible_app.test.app_id
		resource_type = "app"
		process_type = "cmd"
		endpoint_type = "https"
		managed = true
		domain = "www.aptible-test-demo.fake"
		internal = true
		platform = "alb"
		container_port = 666
	}
`, appHandle, testOrganizationId, testStackId, appHandle)
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidResourceType() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "should-error"
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidEndpointType() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = true
		endpoint_type = "should-error"
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidPlatform() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = true
		platform = "should-error"
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidDomain() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		domain = ""
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidLbAlgorithm() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = true
		platform = "alb"
		load_balancing_algorithm_type = "should-error"
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidContainerPort() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_port = 99999
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidContainerPortOnTcp() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "tcp"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_port = 3000
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidContainerPortOnTls() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "tls"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_port = 3000
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidContainerPorts() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_ports = [99999]
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidContainerPortsOnHttp() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "https"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_ports = [3000]
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidMultipleContainerPortFields() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "tcp"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_port = 3000
		container_ports = [3000]
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidSharedWithNoDomain() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "https"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
		platform = "alb"
		managed = true
		container_port = 3000
                shared = true
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidSharedWithWildcardDomain() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "https"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = false
                domain = "*.example.com"
		platform = "alb"
		managed = true
		container_port = 3000
                shared = true
	}`
	log.Println("HCL generated: ", output)
	return output
}

func testAccAptibleEndpointInvalidLbAlgorithmWithElb() string {
	output := `
	resource "aptible_endpoint" "test" {
		env_id = -1
		endpoint_type = "https"
		resource_id = 1
		resource_type = "app"
		process_type = "cmd"
		default_domain = true
		platform = "elb"
		load_balancing_algorithm_type = "round_robin"
	}`
	log.Println("HCL generated: ", output)
	return output
}

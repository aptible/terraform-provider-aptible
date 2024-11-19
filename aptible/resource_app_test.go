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

func TestAccResourceApp_basic(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppBasic(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceApp_deploy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeploy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "nginx"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"force_zero_downtime": "true",
							"simple_health_check": "true",
						}),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceApp_multiple_services(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeployMultipleServices(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/terraform-multiservice-test"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceApp_updateConfig(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeploy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "nginx"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.OOPS", "mistake"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.#", "1"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"force_zero_downtime": "true",
							"simple_health_check": "true",
						}),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppUpdateConfig(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "httpd:alpine"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "nothing"),
						resource.TestCheckNoResourceAttr("aptible_app.test", "config.OOPS"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"force_zero_downtime": "false",
							"simple_health_check": "true",
						}),
					),
				},
			},
		})
	})
}

func TestAccResourceApp_scaleDown(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeploy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "nginx"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.container_count", "1"),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppScaleDown(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.container_count", "0"),
					),
				},
			},
		})
	})
}

func TestAccResourceApp_serviceSizingPolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppServiceSizingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.min_containers", "2"),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceApp_updateServiceSizingPolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppServiceSizingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.min_containers", "2"),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppUpdateServiceSizingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.autoscaling_type", "vertical"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.mem_scale_down_threshold", "0.6"),
					),
				},
			},
		})
	})
}

func TestAccResourceApp_removeServiceSizingPolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppServiceSizingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.service_sizing_policy.0.min_containers", "2"),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppWithoutServiceSizingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						// Ensure the service_sizing_policy block is no longer present
						resource.TestCheckNoResourceAttr("aptible_app.test", "service.0.service_sizing_policy"),
					),
				},
			},
		})
	})
}

func TestAccResourceApp_autoscalingTypeHorizontalMissingAttributes(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppDeployAutoscalingTypeHorizontalMissingAttributes(rHandle),
					ExpectError: regexp.MustCompile(`\w+ is required when autoscaling_type is set to 'horizontal'`),
				},
			},
		})
	})
}

func TestAccResourceApp_autoscalingTypeVerticalInvalidAttributes(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppDeployAutoscalingTypeVerticalInvalidAttributes(rHandle),
					ExpectError: regexp.MustCompile(`\w+ must not be set when autoscaling_type is set to 'vertical'`),
				},
			},
		})
	})
}

func TestAccResourceApp_invalidAutoscalingType(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppDeployInvalidAutoscalingType(rHandle),
					ExpectError: regexp.MustCompile(`expected.*autoscaling_type to be one of \[vertical horizontal\]`),
				},
			},
		})
	})
}

func testAccCheckAppDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*providerMetadata).LegacyClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aptible_app" {
			continue
		}

		appId, err := strconv.Atoi(rs.Primary.Attributes["app_id"])
		if err != nil {
			return err
		}

		app, err := client.GetApp(int64(appId))
		log.Println("Deleted? ", app.Deleted)
		if !app.Deleted {
			return fmt.Errorf("app %v not removed", appId)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAptibleAppBasic(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
			env_id = aptible_environment.test.env_id
			handle = "%v"
	}
`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeploy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
			"OOPS" = "mistake"
		}
    service {
			process_type = "cmd"
			container_profile = "m5"
			container_memory_limit = 512
			container_count = 1
			force_zero_downtime = true
			simple_health_check = true
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployMultipleServices(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/terraform-multiservice-test"
			"WHATEVER" = "something"
			"OOPS" = "mistake"
		}
		service {
			process_type = "main"
			container_profile = "m5"
			container_memory_limit = 512
			container_count = 1
		}
		service {
			process_type = "cron"
			container_profile = "r5"
			container_memory_limit = 512
			container_count = 1
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppUpdateConfig(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "httpd:alpine"
			"WHATEVER" = "nothing"
		}
        service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			simple_health_check = true
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppScaleDown(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
			"OOPS" = "mistake"
		}
    service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 0
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppServiceSizingPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
		}
		service {
			process_type           = "cmd"
			container_profile      = "m5"
			container_count        = 1
			service_sizing_policy {
				autoscaling_type  = "horizontal"
				min_containers    = 2
				max_containers	  = 4
				min_cpu_threshold = 0.1
				max_cpu_threshold = 0.9
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppWithoutServiceSizingPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
		}
		service {
			process_type           = "cmd"
			container_profile      = "m5"
			container_count        = 1
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppUpdateServiceSizingPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "nothing"
		}
		service {
			process_type           = "cmd"
			container_profile      = "m5"
			container_memory_limit = 512
			container_count        = 1
			service_sizing_policy {
				autoscaling_type = "vertical"
				mem_scale_down_threshold = 0.6
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployAutoscalingTypeHorizontalMissingAttributes(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			service_sizing_policy {
				autoscaling_type = "horizontal"
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployAutoscalingTypeVerticalInvalidAttributes(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			service_sizing_policy {
				autoscaling_type = "vertical"
				min_containers = 1
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployInvalidAutoscalingType(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		config = {
			"APTIBLE_DOCKER_IMAGE" = "nginx"
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			service_sizing_policy {
				autoscaling_type = "invalid"
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

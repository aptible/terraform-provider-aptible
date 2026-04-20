package aptible

import (
	"fmt"
	"log"
	"os"
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
					Config:             testAccAptibleAppBasic(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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
					Config: testAccAptibleAppDeploy(rHandle, "1"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:1"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"force_zero_downtime":  "true",
							"restart_free_scaling": "false",
							"simple_health_check":  "true",
						}),
					),
				},
				{
					Config:             testAccAptibleAppDeploy(rHandle, "1"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/terraform-multiservice-test"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
					),
				},
				{
					Config:             testAccAptibleAppDeployMultipleServices(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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
					Config: testAccAptibleAppDeploy(rHandle, "2"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:2"),
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
					Config:             testAccAptibleAppDeploy(rHandle, "2"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppUpdateConfig(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "httpd:alpine"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "nothing"),
						resource.TestCheckNoResourceAttr("aptible_app.test", "config.OOPS"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"force_zero_downtime": "false",
							"simple_health_check": "true",
						}),
					),
				},
				{
					Config:             testAccAptibleAppUpdateConfig(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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
					Config: testAccAptibleAppDeploy(rHandle, "3"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:3"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.container_count", "1"),
					),
				},
				{
					Config:             testAccAptibleAppDeploy(rHandle, "3"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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
				{
					Config:             testAccAptibleAppScaleDown(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

func TestAccResourceApp_autoscalingDisabledThenEnabled(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeploy(rHandle, "4"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:4"),
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
					Config:             testAccAptibleAppDeploy(rHandle, "4"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppautoscalingPolicy(rHandle, "5"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.min_containers", "2"),
					),
				},
				{
					Config:             testAccAptibleAppautoscalingPolicy(rHandle, "5"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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

func TestAccResourceApp_autoscalingPolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppautoscalingPolicy(rHandle, "6"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.min_containers", "2"),
					),
				},
				{
					Config:             testAccAptibleAppautoscalingPolicy(rHandle, "6"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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

func TestAccResourceApp_updateautoscalingPolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppautoscalingPolicy(rHandle, "7"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.min_containers", "2"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.scaling_enabled", "true"),
					),
				},
				{
					Config:             testAccAptibleAppautoscalingPolicy(rHandle, "7"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppUpdateautoscalingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.scaling_enabled", "true"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.autoscaling_type", "vertical"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.mem_scale_down_threshold", "0.6"),
					),
				},
				{
					Config:             testAccAptibleAppUpdateautoscalingPolicy(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

func TestAccResourceApp_removeautoscalingPolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppautoscalingPolicy(rHandle, "8"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckTypeSetElemNestedAttrs(
							"aptible_app.test", "service.0.autoscaling_policy.*", map[string]string{
								"autoscaling_type": "horizontal",
								"min_containers":   "2",
							},
						),
					),
				},
				{
					Config:             testAccAptibleAppautoscalingPolicy(rHandle, "8"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppWithoutautoscalingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						// Ensure the autoscaling_policy block is no longer present
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.#", "0"),
					),
				},
				{
					Config:             testAccAptibleAppWithoutautoscalingPolicy(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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

func TestAccResourceApp_autoscalingOldAndNewAttributeUsage(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppDeployAutoscalingOldAndNewPolicyAttribute(rHandle),
					ExpectError: regexp.MustCompile(`only one of autoscaling_policy or service_sizing_policy may be defined by service`),
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

func TestAccResourceApp_moreThanOnePolicy(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppDeployOnlyOnePolicy(rHandle),
					ExpectError: regexp.MustCompile(`only one autoscaling_policy is allowed per service`),
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
					ExpectError: regexp.MustCompile(`expected.*autoscaling_type to be one of \["vertical" "horizontal"\]`),
				},
			},
		})
	})
}

func TestAccResourceApp_stopTimeout(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeployStopTimeout(rHandle, "16"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:16"),
						resource.TestCheckResourceAttr("aptible_app.test", "config.WHATEVER", "something"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"force_zero_downtime": "true",
							"simple_health_check": "true",
							"stop_timeout":        "60",
						}),
					),
				},
				{
					Config:             testAccAptibleAppDeployStopTimeout(rHandle, "16"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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

func TestAccResourceApp_multipleServicesWithPartialAutoscaling(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppMultipleServicesWithPartialAutoscaling(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "service.#", "2"),
						// Check that the worker service has autoscaling policy
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"process_type": "web",
						}),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*.autoscaling_policy.*", map[string]string{
							"autoscaling_type": "vertical",
							"minimum_memory":   "512",
							"maximum_memory":   "1024",
						}),
						// Check that the web service exists but has no autoscaling policy
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"process_type": "rabbitB",
						}),
					),
				},
				{
					Config:   testAccAptibleAppMultipleServicesWithPartialAutoscaling(rHandle),
					PlanOnly: true,
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

func TestAccResourceApp_updateRestartFreeScaling(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppDeployWithRestartFreeScaling(rHandle, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:16"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"process_type":           "cmd",
							"container_count":        "1",
							"container_memory_limit": "512",
							"restart_free_scaling":   "false",
						}),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*.autoscaling_policy.*", map[string]string{
							"autoscaling_type":     "horizontal",
							"use_horizontal_scale": "false",
							"min_containers":       "1",
							"max_containers":       "3",
							"min_cpu_threshold":    "0.4",
							"max_cpu_threshold":    "0.8",
						}),
					),
				},
				{
					Config:             testAccAptibleAppDeployWithRestartFreeScaling(rHandle, false),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					Config: testAccAptibleAppDeployWithRestartFreeScaling(rHandle, true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:16"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "app_id"),
						resource.TestCheckResourceAttrSet("aptible_app.test", "git_repo"),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"process_type":           "cmd",
							"container_count":        "1",
							"container_memory_limit": "512",
							"restart_free_scaling":   "true",
						}),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*.autoscaling_policy.*", map[string]string{
							"autoscaling_type":     "horizontal",
							"use_horizontal_scale": "true",
							"min_containers":       "1",
							"max_containers":       "3",
							"min_cpu_threshold":    "0.4",
							"max_cpu_threshold":    "0.8",
						}),
					),
				},
				{
					Config:             testAccAptibleAppDeployWithRestartFreeScaling(rHandle, true),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
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

func testAccAptibleAppDeploy(handle string, index string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:%s"
		config = {
			"WHATEVER" = "something"
			"OOPS" = "mistake"
		}
    service {
			process_type = "cmd"
			container_profile = "m5"
			container_memory_limit = 512
			container_count = 1
			force_zero_downtime = true
			restart_free_scaling = false
			simple_health_check = true
		}
	}
	`, handle, testOrganizationId, testStackId, handle, index)
}

func testAccAptibleAppDeployStopTimeout(handle string, index string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:%s"
		config = {
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
			stop_timeout = 60
		}
	}
	`, handle, testOrganizationId, testStackId, handle, index)
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
		docker_image = "quay.io/aptible/terraform-multiservice-test"
		config = {
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
		docker_image = "httpd:alpine"
		config = {
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
		docker_image = "quay.io/aptible/nginx-mirror:5"
		config = {
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

func testAccAptibleAppautoscalingPolicy(handle string, index string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:%s"
		config = {
			"WHATEVER" = "something"
		}
		service {
			process_type           = "cmd"
			container_profile      = "m5"
			container_count        = 1
			autoscaling_policy {
				autoscaling_type  = "horizontal"
				min_containers    = 2
				max_containers	  = 4
				min_cpu_threshold = 0.1
				max_cpu_threshold = 0.9
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle, index)
}

func testAccAptibleAppWithoutautoscalingPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:9"
		config = {
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

func testAccAptibleAppUpdateautoscalingPolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:10"
		config = {
			"WHATEVER" = "something"
		}
		service {
			process_type           = "cmd"
			container_profile      = "m5"
			container_memory_limit = 512
			container_count        = 1
			autoscaling_policy {
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
		docker_image = "quay.io/aptible/nginx-mirror:11"
		config = {
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 1024
			container_count = 1
			autoscaling_policy {
				autoscaling_type = "horizontal"
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployAutoscalingOldAndNewPolicyAttribute(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:12"
		config = {
			"WHATEVER" = "nothing"
		}
		service {
			process_type           = "cmd"
			container_profile      = "m5"
			container_memory_limit = 512
			container_count        = 1
			autoscaling_policy {
				autoscaling_type = "vertical"
				mem_scale_down_threshold = 0.6
			}
			service_sizing_policy {
				autoscaling_type = "vertical"
				mem_scale_down_threshold = 0.6
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
		docker_image = "quay.io/aptible/nginx-mirror:13"
		config = {
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			autoscaling_policy {
				autoscaling_type = "vertical"
				min_containers = 1
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployOnlyOnePolicy(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:14"
		config = {
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			autoscaling_policy {
				autoscaling_type = "vertical"
			}
			autoscaling_policy {
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
		docker_image = "quay.io/aptible/nginx-mirror:15"
		config = {
			"WHATEVER" = "something"
		}
		service {
			process_type = "cmd"
			container_memory_limit = 512
			container_count = 1
			autoscaling_policy {
				autoscaling_type = "invalid"
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppDeployWithRestartFreeScaling(handle string, restartFreeScaling bool) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:16"
		service {
			process_type = "cmd"
			container_profile = "m5"
			container_memory_limit = 512
			container_count = 1
			restart_free_scaling = %t
			autoscaling_policy {
				autoscaling_type = "horizontal"
				use_horizontal_scale = %t
				min_containers = 1
				max_containers = 3
				min_cpu_threshold = 0.4
				max_cpu_threshold = 0.8
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle, restartFreeScaling, restartFreeScaling)
}

func testAccAptibleAppMultipleServicesWithPartialAutoscaling(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/terraform-multiservice-test:policy_order"
		service {
			process_type           = "rabbitB"
			container_count        = 1
			container_memory_limit = 1024
		}
		service {
			process_type           = "web"
			container_count        = 1
			container_memory_limit = 512
			autoscaling_policy {
				autoscaling_type = "vertical"
				percentile = 75.0
				minimum_memory = 512
				maximum_memory = 1024
				mem_scale_up_threshold = 0.9
				mem_scale_down_threshold = 0.75
				mem_cpu_ratio_r_threshold = 4.0
				mem_cpu_ratio_c_threshold = 2.0
				metric_lookback_seconds = 300
				post_scale_up_cooldown_seconds = 60
				post_scale_down_cooldown_seconds = 300
				post_release_cooldown_seconds = 60
			}
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func TestAccResourceApp_usernameWithoutPassword(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppUsernameWithoutPassword(rHandle),
					ExpectError: regexp.MustCompile(`private_registry_password is required when private_registry_username is set`),
				},
			},
		})
	})
}

func TestAccResourceApp_passwordWithoutUsername(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppPasswordWithoutUsername(rHandle),
					ExpectError: regexp.MustCompile(`private_registry_username is required when private_registry_password is set`),
				},
			},
		})
	})
}

func TestAccResourceApp_registryCredsWithoutDockerImage(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccAptibleAppRegistryCredsWithoutDockerImage(rHandle),
					ExpectError: regexp.MustCompile(`docker_image is required when private_registry_username or private_registry_password is set`),
				},
			},
		})
	})
}

func testAccAptibleAppUsernameWithoutPassword(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:latest"
		private_registry_username = "myuser"
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppPasswordWithoutUsername(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:latest"
		private_registry_password = "mypassword"
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func testAccAptibleAppRegistryCredsWithoutDockerImage(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		private_registry_username = "myuser"
		private_registry_password = "mypassword"
	}
	`, handle, testOrganizationId, testStackId, handle)
}

func TestAccResourceApp_updateAndRemovePrivateRegistry(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}

	registryUsername := os.Getenv("APTIBLE_PRIVATE_DOCKER_REPO_USERNAME")
	registryPassword := os.Getenv("APTIBLE_PRIVATE_DOCKER_REPO_PASSWORD")
	if registryUsername == "" || registryPassword == "" {
		t.Fatal("APTIBLE_PRIVATE_DOCKER_REPO_USERNAME and APTIBLE_PRIVATE_DOCKER_REPO_PASSWORD must be set for this test")
	}

	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppPrivateRegistry(rHandle, registryUsername, registryPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/test-private-repo:httpd-alpine"),
						resource.TestCheckResourceAttr("aptible_app.test", "private_registry_username", registryUsername),
						resource.TestCheckResourceAttr("aptible_app.test", "private_registry_password", registryPassword),
					),
				},
				{
					Config:             testAccAptibleAppPrivateRegistry(rHandle, registryUsername, registryPassword),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					Config: testAccAptibleAppPublicImageNoRegistry(rHandle),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "docker_image", "quay.io/aptible/nginx-mirror:17"),
						resource.TestCheckResourceAttr("aptible_app.test", "private_registry_username", ""),
						resource.TestCheckResourceAttr("aptible_app.test", "private_registry_password", ""),
					),
				},
				{
					Config:             testAccAptibleAppPublicImageNoRegistry(rHandle),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

func testAccAptibleAppPrivateRegistry(handle, username, password string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/test-private-repo:httpd-alpine"
		private_registry_username = "%s"
		private_registry_password = "%s"
		service {
			process_type           = "cmd"
			container_count        = 1
			container_memory_limit = 1024
		}
	}
	`, handle, testOrganizationId, testStackId, handle, username, password)
}

func testAccAptibleAppPublicImageNoRegistry(handle string) string {
	return fmt.Sprintf(`
	resource "aptible_environment" "test" {
		handle = "%s"
		org_id = "%s"
		stack_id = "%v"
	}

	resource "aptible_app" "test" {
		env_id = aptible_environment.test.env_id
		handle = "%v"
		docker_image = "quay.io/aptible/nginx-mirror:17"
		service {
			process_type           = "cmd"
			container_count        = 1
			container_memory_limit = 1024
		}
	}
	`, handle, testOrganizationId, testStackId, handle)
}

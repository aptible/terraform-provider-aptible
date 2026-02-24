package aptible

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/aptible/aptible-api-go/aptibleapi"
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
					Config: testAccAptibleAppDeploy(rHandle, "1"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:1"),
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
					Config: testAccAptibleAppDeploy(rHandle, "2"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:2"),
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
					Config: testAccAptibleAppDeploy(rHandle, "3"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:3"),
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
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:4"),
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
				{
					Config: testAccAptibleAppautoscalingPolicy(rHandle, "5"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.min_containers", "2"),
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
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccResourceApp_horizontalAutoscalingDoesNotCauseContainerCountDrift(t *testing.T) {
	rHandle := acctest.RandString(10)

	WithTestAccEnvironment(t, func(env aptible.Environment) {
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAptibleAppautoscalingPolicy(rHandle, "17"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.0.autoscaling_type", "horizontal"),
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.container_count", "1"),
						testAccScaleServiceOutsideTerraform("aptible_app.test", "cmd", 3),
						testAccCheckServiceContainerCount("aptible_app.test", "cmd", 3),
					),
				},
				{
					Config: testAccAptibleAppautoscalingPolicyUnrelatedUpdate(rHandle, "17"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceContainerCount("aptible_app.test", "cmd", 3),
						resource.TestCheckTypeSetElemNestedAttrs("aptible_app.test", "service.*", map[string]string{
							"process_type":    "cmd",
							"container_count": "1",
						}),
					),
				},
				{
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config:   testAccAptibleAppautoscalingPolicyUnrelatedUpdate(rHandle, "17"),
					PlanOnly: true,
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
					ResourceName:      "aptible_app.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccAptibleAppWithoutautoscalingPolicy(rHandle),
					Check: resource.ComposeTestCheckFunc(
						// Ensure the autoscaling_policy block is no longer present
						resource.TestCheckResourceAttr("aptible_app.test", "service.0.autoscaling_policy.#", "0"),
						// Ensure runtime service count is reconciled after removing horizontal autoscaling policy.
						testAccCheckServiceContainerCount("aptible_app.test", "cmd", 1),
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
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:16"),
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
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:16"),
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
					Config: testAccAptibleAppDeployWithRestartFreeScaling(rHandle, true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("aptible_environment.test", "env_id", "aptible_app.test", "env_id"),
						resource.TestCheckResourceAttr("aptible_app.test", "handle", rHandle),
						resource.TestCheckResourceAttr("aptible_app.test", "config.APTIBLE_DOCKER_IMAGE", "quay.io/aptible/nginx-mirror:16"),
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

func testAccScaleServiceOutsideTerraform(resourceName, processType string, containerCount int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		appID, err := strconv.Atoi(rs.Primary.Attributes["app_id"])
		if err != nil {
			return err
		}

		meta := testAccProvider.Meta().(*providerMetadata)
		ctx := meta.APIContext(context.Background())

		apiServicesResp, _, err := meta.Client.ServicesAPI.ListServicesForApp(ctx, int32(appID)).Execute()
		if err != nil {
			return err
		}
		service := findApiServiceByName(apiServicesResp.Embedded.Services, processType)
		if service == nil {
			return fmt.Errorf("service %s not found", processType)
		}

		payload := aptibleapi.NewCreateOperationRequest("scale")
		payload.SetContainerCount(containerCount)
		if service.ContainerMemoryLimitMb.IsSet() && service.ContainerMemoryLimitMb.Get() != nil {
			payload.SetContainerSize(*service.ContainerMemoryLimitMb.Get())
		}
		payload.SetInstanceProfile(service.InstanceClass)

		resp, _, err := meta.Client.OperationsAPI.CreateOperationForService(ctx, service.Id).CreateOperationRequest(*payload).Execute()
		if err != nil {
			return err
		}

		_, err = meta.LegacyClient.WaitForOperation(int64(resp.Id))
		return err
	}
}

func testAccCheckServiceContainerCount(resourceName, processType string, expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		appID, err := strconv.Atoi(rs.Primary.Attributes["app_id"])
		if err != nil {
			return err
		}

		meta := testAccProvider.Meta().(*providerMetadata)
		ctx := meta.APIContext(context.Background())

		apiServicesResp, _, err := meta.Client.ServicesAPI.ListServicesForApp(ctx, int32(appID)).Execute()
		if err != nil {
			return err
		}
		service := findApiServiceByName(apiServicesResp.Embedded.Services, processType)
		if service == nil {
			return fmt.Errorf("service %s not found", processType)
		}

		if service.ContainerCount != expected {
			return fmt.Errorf("expected %s container_count to be %d, got %d", processType, expected, service.ContainerCount)
		}
		return nil
	}
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:%s"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:%s"
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
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:5"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:%s"
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

func testAccAptibleAppautoscalingPolicyUnrelatedUpdate(handle string, index string) string {
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
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:%s"
			"WHATEVER" = "something-else"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:9"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:10"
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
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:11"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:12"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:13"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:14"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:15"
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/nginx-mirror:16"
		}
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
		config = {
			"APTIBLE_DOCKER_IMAGE" = "quay.io/aptible/terraform-multiservice-test:policy_order"
		}
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

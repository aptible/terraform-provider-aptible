package aptible

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/sync/errgroup"
)

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAppCreate, // POST
		Read:          resourceAppRead,   // GET
		UpdateContext: resourceAppUpdate, // PUT
		Delete:        resourceAppDelete, // DELETE
		Importer: &schema.ResourceImporter{
			State: resourceAppImport,
		},

		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"app_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"git_repo": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     resourceService(),
			},
		},
		CustomizeDiff: validateServiceSizingPolicy,
	}
}

func resourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"process_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "cmd",
			},
			"container_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"container_memory_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1024,
				ValidateFunc: validateContainerSize,
			},
			"container_profile": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "m5",
				ValidateFunc: validateContainerProfile,
			},
			"force_zero_downtime": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"simple_health_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"service_sizing_policy": {
				Type:       schema.TypeSet,
				Optional:   true,
				Elem:       resourceServiceSizingPolicy(),
				Deprecated: "Please use autoscaling_policy instead. This attribute will be removed in version 1.0",
			},
			"autoscaling_policy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     resourceServiceSizingPolicy(),
			},
		},
	}
}

func resourceServiceSizingPolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"autoscaling_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"vertical",
					"horizontal",
				}, false),
				Description: "The type of autoscaling, must be either 'vertical' or 'horizontal'.",
			},
			"metric_lookback_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1800,
				Description: "The lookback period for metrics in seconds.",
			},
			"percentile": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     99,
				Description: "The percentile threshold used for scaling.",
			},
			"post_scale_up_cooldown_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
				Description: "Cooldown period in seconds after a scale-up event.",
			},
			"post_scale_down_cooldown_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Description: "Cooldown period in seconds after a scale-down event.",
			},
			"post_release_cooldown_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
				Description: "Seconds to ignore in metrics after a release event.",
			},
			"mem_cpu_ratio_r_threshold": {
				Type:     schema.TypeFloat,
				Default:  4.0,
				Optional: true,
			},
			"mem_cpu_ratio_c_threshold": {
				Type:     schema.TypeFloat,
				Default:  2.0,
				Optional: true,
			},
			"mem_scale_up_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     0.9,
				Description: "The memory usage threshold for scaling up.",
			},
			"mem_scale_down_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     0.75,
				Description: "The memory usage threshold for scaling down.",
			},
			"minimum_memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2048,
				Description: "The minimum memory allocation in MB.",
			},
			"maximum_memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum memory allocation in MB.",
			},
			"min_cpu_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "The minimum CPU utilization threshold for scaling.",
			},
			"max_cpu_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "The maximum CPU utilization threshold for scaling.",
			},
			"min_containers": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum number of containers for scaling.",
			},
			"max_containers": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of containers for scaling.",
			},
			"scale_up_step": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of containers to add in each scale-up event.",
			},
			"scale_down_step": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of containers to remove in each scale-down event.",
			},
			"scaling_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func validateServiceSizingPolicy(ctx context.Context, d *schema.ResourceDiff, _ interface{}) error {
	services := d.Get("service").(*schema.Set).List()
	for _, service := range services {
		serviceMap := service.(map[string]interface{})

		oldPolicies := serviceMap["service_sizing_policy"].(*schema.Set).List()
		newPolicies := serviceMap["autoscaling_policy"].(*schema.Set).List()
		var policies []interface{}
		ok := false
		if len(newPolicies) > 0 {
			policies = newPolicies
			ok = true
		} else if len(oldPolicies) > 0 {
			policies = oldPolicies
			ok = true
		}

		if len(newPolicies) > 0 && len(oldPolicies) > 0 {
			return fmt.Errorf("only one of autoscaling_policy or service_sizing_policy may be defined by service. Please note service_sizing_policy is deprecated in favor of autoscaling_policy")
		}

		if ok {
			if len(policies) == 1 && policies[0] != nil {
				policyMap := policies[0].(map[string]interface{})
				autoscalingType := policyMap["autoscaling_type"].(string)
				attrsToCheck := []string{
					"min_containers",
					"max_containers",
					"min_cpu_threshold",
					"max_cpu_threshold",
				}
				if autoscalingType == "horizontal" {
					for _, attr := range attrsToCheck {
						if val, ok := policyMap[attr]; !ok || val == nil || val == 0 {
							return fmt.Errorf("%s is required when autoscaling_type is set to 'horizontal'", attr)
						}
					}
				} else if autoscalingType == "vertical" {
					for _, attr := range attrsToCheck {
						// NOTE: terraform sets numeric values to `0`, they're never nil
						val := policyMap[attr]
						// Unfortunately we *do* need separate cases for int and float64 despite the code looking identical.
						// The type system does something under the hood here and v != 0 gives wrong results if we combine the cases.
						switch v := val.(type) {
						case int:
							if v != 0 {
								return fmt.Errorf("%s must not be set when autoscaling_type is set to 'vertical'", attr)
							}
						case float64:
							if v != 0 {
								return fmt.Errorf("%s must not be set when autoscaling_type is set to 'vertical'", attr)
							}
						default:
							return fmt.Errorf("unknown issue occurred when validating %s", attr)
						}
					}
				}
			} else if len(policies) > 0 {
				return fmt.Errorf("only one autoscaling_policy is allowed per service")
			}
		}
	}
	return nil
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).LegacyClient
	envID := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)

	app, err := client.CreateApp(handle, envID)
	if err != nil {
		log.Println("There was an error when completing the request to create the app.\n[ERROR] -", err)
		return generateErrorFromClientError(err)
	}
	d.SetId(strconv.Itoa(int(app.ID)))
	_ = d.Set("app_id", app.ID)

	config := d.Get("config").(map[string]interface{})

	if len(config) != 0 {
		err := client.DeployApp(config, app.ID)
		if err != nil {
			log.Println("There was an error when completing the request to configure the app.\n[ERROR] -", err)
			return generateErrorFromClientError(err)
		}
	}

	// Services do not exist before App creation, so we need to wait until after to update settings, unlike when updating an App
	err = updateServices(context.Background(), d, meta)
	if err != nil {
		return err
	}

	// Our model prevents editing services or configurations directly. As a result, any services
	// are created as part of the deployment process and scaled to a single 1 GB container by default.
	// Unfortunately, this isn't something we can bypass without making exceptions to our API security model,
	// which I'm not prepared to do quite yet. So instead we're handling scaling after deployment, rather than
	// at the time of deployment.
	// TODO: We can check for services scaled to 1 GB/1 container before scaling.
	err = scaleServices(context.Background(), d, meta)
	if err != nil {
		return err
	}

	// Now that services exist and are scaled properly, let's see about creating any scaling policies we need
	err = updateServiceSizingPolicy(context.Background(), d, meta)
	if err != nil {
		return err
	}

	return resourceAppRead(d, meta)
}

func resourceAppImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	appID, _ := strconv.Atoi(d.Id())
	_ = d.Set("app_id", appID)
	err := resourceAppRead(d, meta)
	return []*schema.ResourceData{d}, err
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceAppRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).Client
	appID := int32(d.Get("app_id").(int))
	ctx := meta.(*providerMetadata).APIContext(context.Background())

	log.Println("Getting App with ID: " + strconv.Itoa(int(appID)))

	app, resp, err := client.AppsAPI.GetApp(ctx, appID).Execute()
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		log.Println("App with ID: " + strconv.Itoa(int(appID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}
	if err != nil {
		log.Println(err)
		return err
	}

	_ = d.Set("app_id", int(app.Id))
	_ = d.Set("git_repo", app.GitRepo)
	_ = d.Set("handle", app.Handle)
	_ = d.Set("env_id", ExtractIdFromLink(app.Links.Account.GetHref()))
	currConfId := ExtractIdFromLink(app.Links.CurrentConfiguration.GetHref())
	if currConfId != 0 {
		currConf, _, err := client.ConfigurationsAPI.GetConfiguration(ctx, currConfId).Execute()
		if err == nil {
			_ = d.Set("config", currConf.Env)
		}
	}

	var services = make([]map[string]interface{}, len(app.Embedded.Services))
	for i, s := range app.Embedded.Services {
		service := make(map[string]interface{})
		service["container_count"] = s.ContainerCount
		if s.ContainerMemoryLimitMb.IsSet() {
			service["container_memory_limit"] = *s.ContainerMemoryLimitMb.Get()
		}
		service["container_profile"] = s.InstanceClass
		service["process_type"] = s.ProcessType
		log.Printf("ZDD flags: %t, %t", s.ForceZeroDowntime, s.NaiveHealthCheck)
		service["force_zero_downtime"] = s.ForceZeroDowntime
		service["simple_health_check"] = s.NaiveHealthCheck
		// Find service_sizing_policy if any
		var policy *aptibleapi.ServiceSizingPolicy
		policy, err = getServiceSizingPolicyForService(s.Id, ctx, meta)
		if err != nil {
			log.Println(err)
			return err
		}
		if policy != nil {
			serviceSizingPolicy := make(map[string]interface{})
			serviceSizingPolicy["autoscaling_type"] = policy.Autoscaling
			serviceSizingPolicy["scaling_enabled"] = policy.GetScalingEnabled()
			serviceSizingPolicy["metric_lookback_seconds"] = policy.MetricLookbackSeconds
			serviceSizingPolicy["percentile"] = formatFloat32ToFloat64(policy.Percentile)
			serviceSizingPolicy["post_scale_up_cooldown_seconds"] = policy.PostScaleUpCooldownSeconds
			serviceSizingPolicy["post_scale_down_cooldown_seconds"] = policy.PostScaleDownCooldownSeconds
			serviceSizingPolicy["post_release_cooldown_seconds"] = policy.PostReleaseCooldownSeconds
			serviceSizingPolicy["mem_cpu_ratio_r_threshold"] = formatFloat32ToFloat64(policy.MemCpuRatioRThreshold)
			serviceSizingPolicy["mem_cpu_ratio_c_threshold"] = formatFloat32ToFloat64(policy.MemCpuRatioCThreshold)
			serviceSizingPolicy["mem_scale_up_threshold"] = formatFloat32ToFloat64(policy.MemScaleUpThreshold)
			serviceSizingPolicy["mem_scale_down_threshold"] = formatFloat32ToFloat64(policy.MemScaleDownThreshold)
			serviceSizingPolicy["minimum_memory"] = policy.MinimumMemory
			serviceSizingPolicy["maximum_memory"] = policy.GetMaximumMemory()
			serviceSizingPolicy["min_cpu_threshold"] = formatFloat32ToFloat64(policy.GetMinCpuThreshold())
			serviceSizingPolicy["max_cpu_threshold"] = formatFloat32ToFloat64(policy.GetMaxCpuThreshold())
			serviceSizingPolicy["min_containers"] = policy.GetMinContainers()
			serviceSizingPolicy["max_containers"] = policy.GetMaxContainers()
			serviceSizingPolicy["scale_up_step"] = policy.GetScaleUpStep()
			serviceSizingPolicy["scale_down_step"] = policy.GetScaleDownStep()

			service["autoscaling_policy"] = []map[string]interface{}{serviceSizingPolicy}
		}

		services[i] = service
	}
	log.Println("SETTING SERVICE")
	log.Println(services)

	_ = d.Set("service", services)

	return nil
}

func resourceAppUpdate(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).LegacyClient
	appID := int64(d.Get("app_id").(int))

	var diags diag.Diagnostics

	// Check if any updates for Service settings. If so, change before deploying
	err := updateServices(c, d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to update services settings.",
			Detail:   err.Error(),
		})
		return diags
	}

	if d.HasChange("config") {
		o, c := d.GetChange("config")
		old := o.(map[string]interface{})
		config := c.(map[string]interface{})
		// Set any old keys that are not present to an empty string.
		// The API will then clear them during normalization otherwise
		// the old values will be merged with the new
		for key := range old {
			if _, present := config[key]; !present {
				config[key] = ""
			}
		}

		err := client.DeployApp(config, appID)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to deploy the app.",
				Detail:   generateErrorFromClientError(err).Error(),
			})
			log.Println("There was an error when completing the request.\n[ERROR] -", err)
			return diags
		}
	}

	err = scaleServices(c, d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to scale services.",
			Detail:   err.Error(),
		})
		return diags
	}

	// Now that services are settled, go ahead and update scaling policy
	err = updateServiceSizingPolicy(c, d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to update autoscaling_policy.",
			Detail:   err.Error(),
		})
		return diags
	}

	handle := d.Get("handle").(string)
	if d.HasChange("handle") {
		updates := aptible.AppUpdates{
			Handle: handle,
		}
		log.Printf("[INFO] Updating handle to %s\n", handle)
		if err := client.UpdateApp(appID, updates); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to update the handle.",
				Detail:   generateErrorFromClientError(err).Error(),
			})
			return diags
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("You must restart the app to see changes. In order for the new app name (%s) to appear in log drain and metric drain destinations, you must restart the app. You can use the CLI to do this with: 'aptible restart --app=%s'.", handle, handle),
		})
		log.Printf("[WARN] In order for the new app name (%s) to appear in log drain and metric drain destinations, you must restart the app.\n", handle)
	}

	if err = resourceAppRead(d, meta); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to retrieve the updated state of the app.",
			Detail:   generateErrorFromClientError(err).Error(),
		})
	}

	return diags
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) error {
	readErr := resourceAppRead(d, meta)
	if readErr == nil {
		appID := int64(d.Get("app_id").(int))
		client := meta.(*providerMetadata).LegacyClient
		deleted, err := client.DeleteApp(appID)
		if deleted {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the app.\n[ERROR] -", err)
			return generateErrorFromClientError(err)
		}
	}
	d.SetId("")
	return nil
}

func updateServices(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).Client
	ctx = meta.(*providerMetadata).APIContext(ctx)
	appID := int32(d.Get("app_id").(int))

	// If there are no changes to services, there's no reason to update
	if !d.HasChange("service") {
		return nil
	}

	var g errgroup.Group

	// If we're changing existing services, be sure we're using the "new" service definitions and only
	// try to update ones that actually change
	log.Println("Detected change in services")
	oldService, newService := d.GetChange("service")
	services := newService.(*schema.Set).Difference(oldService.(*schema.Set)).List()

	apiServicesResp, _, err := client.ServicesAPI.ListServicesForApp(ctx, appID).Execute()
	if err != nil {
		log.Println("There was an error when loading the services \n[ERROR] -", err)
		return err
	}
	apiServices := apiServicesResp.Embedded.Services

	for _, s := range services {
		serviceData := s.(map[string]interface{})
		processType := serviceData["process_type"].(string)

		log.Printf("Looking up service %s in list: %v", processType, apiServices)
		// Find corresponding service from API response
		apiService := findApiServiceByName(apiServices, processType)
		if apiService == nil {
			log.Printf("ERROR: CANNOT FIND SERVICE %s", processType)
			return fmt.Errorf("[ERROR]Unable to find service %s", processType)
		}

		forceZeroDowntime := serviceData["force_zero_downtime"].(bool)
		naiveHealthCheck := serviceData["simple_health_check"].(bool)

		forceZeroDowntimeChanged := forceZeroDowntime != apiService.ForceZeroDowntime
		naiveHealthCheckChanged := naiveHealthCheck != apiService.NaiveHealthCheck

		if !forceZeroDowntimeChanged && !naiveHealthCheckChanged {
			log.Printf("[INFO] No relevant changes detected for service %s, skipping update.", processType)
			continue
		}

		// Clone values inside the goroutine to avoid race conditions
		svcType := processType
		svcZeroDowntime := forceZeroDowntime
		svcHealthCheck := naiveHealthCheck
		svcID := apiService.Id
		log.Printf("About to update service %s: force_zero_downtime: %t, simple_health_check: %t", svcType, svcZeroDowntime, svcHealthCheck)

		g.Go(func() error {
			payload := aptibleapi.NewUpdateServiceRequest()
			payload.SetForceZeroDowntime(svcZeroDowntime)
			payload.SetNaiveHealthCheck(svcHealthCheck)

			log.Printf("Updating service %s: force_zero_downtime: %t, simple_health_check: %t", svcType, svcZeroDowntime, svcHealthCheck)

			_, err := client.ServicesAPI.UpdateService(ctx, svcID).UpdateServiceRequest(*payload).Execute()
			if err != nil {
				return fmt.Errorf("error updating service %s: %w", svcType, err)
			}

			return nil
		})
	}

	return g.Wait()
}

func scaleServices(c context.Context, d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).Client
	legacy := meta.(*providerMetadata).LegacyClient
	appID := int32(d.Get("app_id").(int))
	ctx := meta.(*providerMetadata).APIContext(c)

	// If there are no changes to services, there's no reason to scale
	if !d.HasChange("service") {
		return nil
	}

	apiServicesResp, _, err := client.ServicesAPI.ListServicesForApp(ctx, appID).Execute()
	if err != nil {
		log.Println("There was an error when loading the services \n[ERROR] -", err)
		return err
	}
	apiServices := apiServicesResp.Embedded.Services

	var g errgroup.Group

	// If we're changing existing services, be sure we're using the "new" service definitions and only
	// try to scale ones that actually change
	log.Println("Detected change in services")
	oldService, newService := d.GetChange("service")
	services := newService.(*schema.Set).Difference(oldService.(*schema.Set)).List()

	for _, service := range services {
		serviceInterface := service.(map[string]interface{})
		// Find the corresponding old service
		var oldServiceData map[string]interface{}
		for _, oldS := range oldService.(*schema.Set).List() {
			oldService := oldS.(map[string]interface{})
			if oldService["process_type"].(string) == serviceInterface["process_type"].(string) {
				oldServiceData = oldService
				break
			}
		}
		shouldScale := false
		for key, newValue := range serviceInterface {
			if key == "force_zero_downtime" || key == "simple_health_check" || key == "autoscaling_policy" || key == "service_sizing_policy" {
				continue // Skip checking these keys. Nothing to do with manual scaling
			}

			if oldServiceData == nil || oldServiceData[key] != newValue {
				shouldScale = true
				break
			}
		}
		if !shouldScale {
			return nil
		}

		g.Go(func() error {
			memoryLimit := int32(serviceInterface["container_memory_limit"].(int))
			containerProfile := serviceInterface["container_profile"].(string)
			containerCount := int32(serviceInterface["container_count"].(int))
			processType := serviceInterface["process_type"].(string)

			log.Printf(
				"Updating %s service to count: %d, limit: %d, and container profile: %s\n",
				processType, containerCount, memoryLimit, containerProfile,
			)
			service := findApiServiceByName(apiServices, processType)
			if service == nil {
				return fmt.Errorf("there was an error when finding the service: %s", processType)
			}

			payload := aptibleapi.NewCreateOperationRequest("scale")
			payload.SetContainerCount(containerCount)
			payload.SetContainerSize(memoryLimit)
			payload.SetInstanceProfile(containerProfile)
			resp, _, err := client.OperationsAPI.CreateOperationForService(ctx, service.Id).CreateOperationRequest(*payload).Execute()
			if err != nil {
				log.Println("There was an error when scaling the service \n[ERROR] -", err)
				return err
			}

			_, err = legacy.WaitForOperation(int64(resp.Id))
			return err
		})
	}

	return g.Wait()
}

func findApiServiceByName(services []aptibleapi.Service, serviceName string) *aptibleapi.Service {
	for i := range services {
		if services[i].ProcessType == serviceName {
			return &services[i]
		}
	}
	return nil
}

func getServiceIdForAppByName(ctx context.Context, meta interface{}, appId int32, processType string) (int32, error) {
	client := meta.(*providerMetadata).Client
	ctx = meta.(*providerMetadata).APIContext(ctx)

	serviceList, _, err := client.ServicesAPI.ListServicesForApp(ctx, appId).Execute()
	if err != nil {
		return 0, fmt.Errorf("error fetching services: %w", err)
	}

	for _, service := range serviceList.Embedded.Services {
		if service.ProcessType == processType {
			return service.Id, nil
		}
	}

	return 0, fmt.Errorf("no service found for process type %s", processType)
}

func getServiceSizingPolicyForService(serviceId int32, ctx context.Context, meta interface{}) (*aptibleapi.ServiceSizingPolicy, error) {
	client := meta.(*providerMetadata).Client
	ctx = meta.(*providerMetadata).APIContext(ctx)
	resp, _, err := client.ServiceSizingPoliciesAPI.ListServiceSizingPoliciesForService(ctx, serviceId).Execute()
	if err != nil {
		return nil, err
	}
	if len(resp.Embedded.ServiceSizingPolicies) == 0 {
		return nil, nil
	}
	policy := resp.Embedded.ServiceSizingPolicies[0]
	return &policy, nil
}

func updateServiceSizingPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).Client
	ctx = meta.(*providerMetadata).APIContext(ctx)
	appID := int32(d.Get("app_id").(int))

	// If there are no changes to services, there's no reason to update
	if !d.HasChange("service") {
		return nil
	}

	services := d.Get("service").(*schema.Set).List()
	for _, serviceData := range services {
		serviceMap := serviceData.(map[string]interface{})
		serviceName := serviceMap["process_type"].(string)

		oldPolicies := serviceMap["service_sizing_policy"].(*schema.Set).List()
		newPolicies := serviceMap["autoscaling_policy"].(*schema.Set).List()
		var schemaPolicies []interface{}
		if len(newPolicies) > 0 {
			schemaPolicies = newPolicies
		} else if len(oldPolicies) > 0 {
			schemaPolicies = oldPolicies
		}

		serviceId, err := getServiceIdForAppByName(ctx, meta, appID, serviceName)
		if err != nil {
			return err
		}

		policy, err := getServiceSizingPolicyForService(serviceId, ctx, meta)
		if err != nil {
			log.Println(err)
			return err
		}

		// no policy in the schema
		if len(schemaPolicies) == 0 {
			// policy exists in deploy-api? delete
			if policy != nil {
				_, err = client.ServiceSizingPoliciesAPI.DeleteServiceSizingPolicy(ctx, serviceId).Execute()
				if err != nil {
					return fmt.Errorf("failed to delete autoscaling policy for service %s: %w", serviceName, err)
				}
			}
			return nil
		}

		// There's only ever one policy, but we have to model this as a list
		serviceSizingPolicyMap := schemaPolicies[0].(map[string]interface{})

		autoscaling := serviceSizingPolicyMap["autoscaling_type"].(string)
		delete(serviceSizingPolicyMap, "autoscaling_type")
		if autoscaling == "horizontal" {
			delete(serviceSizingPolicyMap, "maximum_memory")
		} else if autoscaling == "vertical" {
			// First, remove values without defaults that aren't used in VAS
			delete(serviceSizingPolicyMap, "min_containers")
			delete(serviceSizingPolicyMap, "max_containers")
			delete(serviceSizingPolicyMap, "min_cpu_threshold")
			delete(serviceSizingPolicyMap, "max_cpu_threshold")
			// Now ensure other values are actually set
			if serviceSizingPolicyMap["maximum_memory"] == 0 {
				delete(serviceSizingPolicyMap, "maximum_memory")
			}
		}
		// Get rid of anything marked as `0` since that is what terraform sets things not set by the user
		// Also, 0 is not a valid value for any of ServiceSizingPolicy attributes
		for key, value := range serviceSizingPolicyMap {
			switch v := value.(type) {
			case int:
				if v == 0 {
					delete(serviceSizingPolicyMap, key)
				}
			case float64:
				if v == 0 {
					delete(serviceSizingPolicyMap, key)
				}
			}
		}

		if policy == nil {
			payload := aptibleapi.NewCreateServiceSizingPolicyRequest()
			jsonData, _ := json.Marshal(serviceSizingPolicyMap)
			_ = json.Unmarshal(jsonData, &payload)
			payload.Autoscaling = &autoscaling

			_, err = client.ServiceSizingPoliciesAPI.
				CreateServiceSizingPolicy(ctx, serviceId).
				CreateServiceSizingPolicyRequest(*payload).
				Execute()
		} else {
			payload := aptibleapi.NewUpdateServiceSizingPolicyRequest()
			jsonData, _ := json.Marshal(serviceSizingPolicyMap)
			_ = json.Unmarshal(jsonData, &payload)
			payload.Autoscaling = &autoscaling
			payload.SetScalingEnabled(true)

			_, err = client.ServiceSizingPoliciesAPI.
				UpdateServiceSizingPolicy(ctx, serviceId).
				UpdateServiceSizingPolicyRequest(*payload).
				Execute()
		}

		if err != nil {
			return fmt.Errorf("failed to create autoscaling policy for service %s: %w", serviceName, err)
		}
	}
	return nil
}

package aptible

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Elem: &schema.Resource{
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
					},
				},
			},
		},
	}
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
	err = scaleServices(d, meta)
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

	err = scaleServices(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to scale services.",
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

func findServiceByName(services []aptibleapi.Service, serviceName string) *aptibleapi.Service {
	for i := range services {
		if services[i].ProcessType == serviceName {
			return &services[i]
		}
	}
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
		// Find corresponding service from API
		apiService := findServiceByName(apiServices, processType)
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

func scaleServices(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).LegacyClient
	appID := int64(d.Get("app_id").(int))

	// If there are no changes to services, there's no reason to scale
	if !d.HasChange("service") {
		return nil
	}

	var g errgroup.Group

	// If we're changing existing services, be sure we're using the "new" service definitions and only
	// try to scale ones that actually change
	log.Println("Detected change in services")
	oldService, newService := d.GetChange("service")
	services := newService.(*schema.Set).Difference(oldService.(*schema.Set)).List()

	for _, s := range services {
		// https://stackoverflow.com/a/74383278
		service := s
		serviceInterface := service.(map[string]interface{})
		g.Go(func() error {
			memoryLimit := int64(serviceInterface["container_memory_limit"].(int))
			containerProfile := serviceInterface["container_profile"].(string)
			containerCount := int64(serviceInterface["container_count"].(int))
			processType := serviceInterface["process_type"].(string)

			log.Printf(
				"Updating %s service to count: %d, limit: %d, and container profile: %s\n",
				processType, containerCount, memoryLimit, containerProfile,
			)
			service, err := client.GetServiceForAppByName(appID, processType)
			if err != nil {
				log.Println("There was an error when finding the service \n[ERROR] -", err)
				return generateErrorFromClientError(err)
			}
			err = client.ScaleService(service.ID, containerCount, memoryLimit, containerProfile)
			if err != nil {
				log.Println("There was an error when scaling the service \n[ERROR] -", err)
				return generateErrorFromClientError(err)
			}
			return nil
		})
	}

	return g.Wait()
}

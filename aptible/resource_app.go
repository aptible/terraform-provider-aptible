package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
					Type:      schema.TypeString,
					Sensitive: true,
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
							Default:      "m4",
							ValidateFunc: validateContainerProfile,
						},
					},
				},
			},
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
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
	client := meta.(*aptible.Client)
	appID := int64(d.Get("app_id").(int))

	log.Println("Getting App with ID: " + strconv.Itoa(int(appID)))

	app, err := client.GetApp(appID)
	if err != nil {
		log.Println(err)
		return err
	}
	if app.Deleted {
		d.SetId("")
		return nil
	}
	_ = d.Set("app_id", int(app.ID))
	_ = d.Set("git_repo", app.GitRepo)
	_ = d.Set("handle", app.Handle)
	_ = d.Set("env_id", app.EnvironmentID)
	_ = d.Set("config", app.Env)

	var services = make([]map[string]interface{}, len(app.Services))
	for i, s := range app.Services {
		service := make(map[string]interface{})
		service["container_count"] = s.ContainerCount
		service["container_memory_limit"] = s.ContainerMemoryLimitMb
		service["container_profile"] = s.ContainerProfile
		service["process_type"] = s.ProcessType
		services[i] = service
	}
	log.Println("SETTING SERVICE ")
	log.Println(services)

	_ = d.Set("service", services)

	return nil
}

func resourceAppUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	appID := int64(d.Get("app_id").(int))

	var diags diag.Diagnostics

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

	err := scaleServices(d, meta)
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
		client := meta.(*aptible.Client)
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

func scaleServices(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	appID := int64(d.Get("app_id").(int))

	// If there are no changes to services, there's no reason to scale
	if !d.HasChange("service") {
		return nil
	}

	// If we're changing existing services, be sure we're using the "new" service definitions and only
	// try to scale ones that actually change
	log.Println("Detected change in services")
	oldService, newService := d.GetChange("service")
	services := newService.(*schema.Set).Difference(oldService.(*schema.Set)).List()

	for _, s := range services {
		serviceInterface := s.(map[string]interface{})
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
	}

	return nil
}

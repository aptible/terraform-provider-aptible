package aptible

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"strconv"
)

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate, // POST
		Read:   resourceAppRead,   // GET
		Update: resourceAppUpdate, // PUT
		Delete: resourceAppDelete, // DELETE
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
				ForceNew: true,
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
							Required: true,
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
							ValidateFunc: validation.IntInSlice(validContainerSizes),
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
		return err
	}
	d.SetId(handle)
	_ = d.Set("app_id", app.ID)

	config := d.Get("config").(map[string]interface{})

	if len(config) != 0 {
		err := client.DeployApp(config, app.ID)
		if err != nil {
			log.Println("There was an error when completing the request to configure the app.\n[ERROR] -", err)
			return err
		}
	}

	// Our model prevents editing services or configurations directly. As a result, any services
	// are created as part of the deployment process and scaled to a single 1 GB container by default.
	// Unfortunately, this isn't something we can bypass without making exceptions to our API security model,
	// which I'm not prepared to do quite yet. So instead we're handling scaling after deployment, rather than
	// at the time of deployment.
	// TODO: We can could for services scaled to 1 GB/1 container before scaling.
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

	var services = make([]map[string]interface{}, len(app.Services), len(app.Services))
	for i, s := range app.Services {
		service := make(map[string]interface{})
		service["container_count"] = s.ContainerCount
		service["container_memory_limit"] = s.ContainerMemoryLimitMb
		service["process_type"] = s.ProcessType
		services[i] = service
	}
	log.Println("SETTING SERVICE ")
	log.Println(services)

	_ = d.Set("service", services)

	return nil
}

func resourceAppUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	appID := int64(d.Get("app_id").(int))

	if d.HasChange("config") {
		config := d.Get("config").(map[string]interface{})
		err := client.DeployApp(config, appID)
		if err != nil {
			log.Println("There was an error when completing the request.\n[ERROR] -", err)
			return err
		}
	}

	if d.HasChange("service") {
		err := scaleServices(d, meta)
		if err != nil {
			return err
		}
	}

	return resourceAppRead(d, meta)
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
			return err
		}
	}
	d.SetId("")
	return nil
}

func scaleServices(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	appID := int64(d.Get("app_id").(int))

	for _, s := range d.Get("service").(*schema.Set).List() {

		serviceInterface := s.(map[string]interface{})
		memoryLimit, _ := serviceInterface["container_memory_limit"]
		containerCount, _ := serviceInterface["container_count"]
		processType, _ := serviceInterface["process_type"]

		service, err := client.GetServiceForAppByName(appID, processType.(string))
		if err != nil {
			log.Println("There was an error when finding the service \n[ERROR] -", err)
			return err
		}
		err = client.ScaleService(service.ID, int64(containerCount.(int)), int64(memoryLimit.(int)))
		if err != nil {
			log.Println("There was an error when scaling the service \n[ERROR] -", err)
			return err
		}

	}
	return nil
}

var validContainerSizes = []int{
	512,
	1024,
	2048,
	4096,
	7168,
	15360,
	30720,
	61440,
	153600,
	245760,
}

package aptible

import (
	"log"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate, // POST
		Read:   resourceAppRead,   // GET
		Update: resourceAppUpdate, // PUT
		Delete: resourceAppDelete, // DELETE

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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	// Setting up params and client
	client := meta.(*aptible.Client)
	env_id := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)

	// Creating app
	app, err := client.CreateApp(handle, env_id)
	if err != nil {
		log.Println("There was an error when completing the request to create the app.\n[ERROR] -", err)
		return err
	}

	// Set computed attributes
	d.Set("app_id", int(*app.ID))
	d.Set("git_repo", app.GitRepo)
	d.Set("created_at", app.CreatedAt)
	d.SetId(handle)

	// Deploying app
	config := d.Get("config").(map[string]interface{})
	if len(config) != 0 {
		app_id := int64(d.Get("app_id").(int))
		err = client.DeployApp(app_id, config)
		if err != nil {
			log.Println("There was an error when completing the request to deploy the app.\n[ERROR] -", err)
			return err
		}
	}

	return resourceAppRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceAppRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	app_id := int64(d.Get("app_id").(int))
	deleted, err := client.GetApp(app_id)
	if err != nil {
		log.Println(err)
		return err
	}
	if deleted {
		d.SetId("")
		return nil
	}
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceAppUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	app_id := int64(d.Get("app_id").(int))

	// Handling config changes
	if d.HasChange("config") {
		config := d.Get("config").(map[string]interface{})
		err := client.UpdateApp(config, app_id)
		if err != nil {
			log.Println("There was an error when completing the request.\n[ERROR] -", err)
			return err
		}
	}
	return resourceAppRead(d, meta)
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) error {
	read_err := resourceAppRead(d, meta)
	if read_err == nil {
		app_id := int64(d.Get("app_id").(int))
		client := meta.(*aptible.Client)
		err := client.DestroyApp(app_id)
		if err != nil {
			log.Println("There was an error when completing the request to destroy the app.\n[ERROR] -", err)
			return err
		}
	}
	d.SetId("")
	return nil
}

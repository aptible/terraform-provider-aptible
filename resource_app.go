package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/reggregory/go-deploy/aptible"
)

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate, // POST
		Read:   resourceAppRead,   // GET
		Update: resourceAppUpdate, // PUT
		Delete: resourceAppDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"account_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"env": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"git_repo": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, m interface{}) error {
	// Setting up params and client
	client := aptible.SetUpClient()
	account_id := int64(d.Get("account_id").(int))
	handle := d.Get("handle").(string)

	// Creating app
	app, err := client.CreateApp(handle, account_id)
	if err != nil {
		AppLogger.Println("There was an error when completing the request to create the app.\n[ERROR] -", err)
		return err
	}

	// Set computed attributes
	d.Set("app_id", int(*app.ID))
	d.Set("git_repo", app.GitRepo)
	d.Set("created_at", app.CreatedAt)
	d.SetId(handle)

	// Deploying app
	env := d.Get("env").(map[string]interface{})
	app_id := int64(d.Get("app_id").(int))
	err = client.DeployApp(app_id, env)
	if err != nil {
		AppLogger.Println("There was an error when completing the request to deploy the app.\n[ERROR] -", err)
		return err
	}

	return resourceAppRead(d, m)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceAppRead(d *schema.ResourceData, m interface{}) error {
	// Setting up params and client
	client := aptible.SetUpClient()
	app_id := int64(d.Get("app_id").(int))
	deleted, err := client.GetApp(app_id)
	if err != nil {
		AppLogger.Println(err)
		return err
	}
	if deleted {
		d.SetId("")
		return nil
	}
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	// Setting up params and client
	client := aptible.SetUpClient()
	app_id := int64(d.Get("app_id").(int))

	// Handling env changes
	if d.HasChange("env") {
		env := d.Get("env").(map[string]interface{})
		err := client.UpdateEnv(env, app_id)
		if err != nil {
			AppLogger.Println("There was an error when completing the request.\n[ERROR] -", err)
			return err
		}
	}
	return resourceAppRead(d, m)
}

func resourceAppDelete(d *schema.ResourceData, m interface{}) error {
	read_err := resourceAppRead(d, m)
	if read_err == nil {
		app_id := int64(d.Get("app_id").(int))
		client := aptible.SetUpClient()
		err := client.DestroyApp(app_id)
		if err != nil {
			AppLogger.Println("There was an error when completing the request to destroy the app.\n[ERROR] -", err)
			return err
		}
	}
	d.SetId("")
	return nil
}

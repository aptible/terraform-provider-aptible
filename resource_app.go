package main

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/terraform/helper/schema"
	deploy "github.com/reggregory/go-deploy/client"
	"github.com/reggregory/go-deploy/client/operations"
	"github.com/reggregory/go-deploy/helpers"
	"github.com/reggregory/go-deploy/models"
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
	client, bearerTokenAuth := helpers.SetUpClient()
	account_id := int64(d.Get("account_id").(int))
	handle := d.Get("handle").(string)

	// Creating app
	appreq := models.AppRequest3{Handle: &handle}
	params := operations.NewPostAccountsAccountIDAppsParams().WithAccountID(account_id).WithAppRequest(&appreq)
	resp, err := client.Operations.PostAccountsAccountIDApps(params, bearerTokenAuth)
	if err != nil {
		AppLogger.Println("There was an error when completing the request to create the app.\n[ERROR] -", resp)
		return err
	}
	AppLogger.Println("This is the response.\n[INFO] -", resp)
	app := resp.Payload
	d.Set("app_id", int(*app.ID))
	d.Set("git_repo", app.GitRepo)
	d.Set("created_at", app.CreatedAt)
	d.SetId(handle)

	// Deploying app
	env := d.Get("env")
	req_type := "deploy"
	app_req := models.AppRequest21{Type: &req_type, Env: env, ContainerCount: 1, ContainerSize: 1024}
	app_id := int64(d.Get("app_id").(int))
	app_params := operations.NewPostAppsAppIDOperationsParams().WithAppID(app_id).WithAppRequest(&app_req)
	app_resp, err := client.Operations.PostAppsAppIDOperations(app_params, bearerTokenAuth)
	if err != nil {
		AppLogger.Println("There was an error when completing the request to deploy the app.\n[ERROR] -", app_resp)
		return err
	}

	AppLogger.Println("This is the response.\n[INFO] -", app_resp)
	return resourceAppRead(d, m)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceAppRead(d *schema.ResourceData, m interface{}) error {
	// Setting up params and client
	client, bearerTokenAuth := helpers.SetUpClient()

	app_id := int64(d.Get("app_id").(int))
	params := operations.NewGetAppsIDParams().WithID(app_id)
	resp, err := client.Operations.GetAppsID(params, bearerTokenAuth)
	if err != nil {
		err_struct := err.(*operations.GetAppsIDDefault)
		switch err_struct.Code() {
		case 404:
			d.SetId("")
			return nil
		case 401:
			AppLogger.Println("Make sure you have the correct auth token.")
			return err
		default:
			AppLogger.Println(fmt.Sprintf("There was an error when completing the request to get the app with handle: %s.\n[ERROR] - %s", d.Get("handle").(string), err))
			return err
		}
	}
	AppLogger.Println("This is the response.\n[INFO] -", resp)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	// Setting up params and client
	client, bearerTokenAuth := helpers.SetUpClient()
	app_id := int64(d.Get("app_id").(int))

	// Handling env changes
	if d.HasChange("env") {
		env := d.Get("env").(map[string]interface{})
		err := updateEnv(env, client, app_id, bearerTokenAuth)
		if err != nil {
			return err
		}
	}
	return resourceAppRead(d, m)
}

func resourceAppDelete(d *schema.ResourceData, m interface{}) error {
	read_err := resourceAppRead(d, m)
	if read_err == nil {
		app_id := int64(d.Get("app_id").(int))
		client, bearerTokenAuth := helpers.SetUpClient()

		req_type := "deprovision"
		app_req := models.AppRequest21{Type: &req_type}
		app_params := operations.NewPostAppsAppIDOperationsParams().WithAppID(app_id).WithAppRequest(&app_req)
		app_resp, err := client.Operations.PostAppsAppIDOperations(app_params, bearerTokenAuth)
		if err != nil {
			AppLogger.Println("There was an error when completing the request to destroy the app.\n[ERROR] -", app_resp)
			return err
		}
	}
	d.SetId("")
	return nil
}

// Updates the `env` based on changes made in the config file
func updateEnv(env map[string]interface{}, client *deploy.DeployAPIV1, app_id int64, bearerTokenAuth runtime.ClientAuthInfoWriter) error {
	app_req := models.AppRequest21{}
	if _, ok := env["APTIBLE_DOCKER_IMAGE"]; ok {
		// Deploying app
		req_type := "deploy"
		app_req = models.AppRequest21{Type: &req_type, Env: env, ContainerCount: 1, ContainerSize: 1024}
	} else {
		// Configuring app
		req_type := "configure"
		app_req = models.AppRequest21{Type: &req_type, Env: env}
	}

	app_params := operations.NewPostAppsAppIDOperationsParams().WithAppID(app_id).WithAppRequest(&app_req)
	app_resp, err := client.Operations.PostAppsAppIDOperations(app_params, bearerTokenAuth)
	if err != nil {
		AppLogger.Println("There was an error when completing the request.\n[ERROR] -", app_resp)
		return err
	}
	return nil
}

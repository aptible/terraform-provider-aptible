package main

import (
	"os"
	"strconv"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform/helper/schema"
	deploy "github.com/reggregory/go-deploy/client"
	"github.com/reggregory/go-deploy/client/operations"
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
				Type:     schema.TypeString,
				Required: true,
			},
			"handle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"env": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
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
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, m interface{}) error {
	// Setting up params and client
	account_id_str := d.Get("account_id").(string)
	account_id, err := strconv.ParseInt(account_id_str, 10, 64) // WithAccountID takes in an int64
	handle := d.Get("handle").(string)

	rt := httptransport.New(
		"api-rachel.aptible-sandbox.com",
		deploy.DefaultBasePath,
		deploy.DefaultSchemes)
	rt.Consumers["application/hal+json"] = runtime.JSONConsumer()
	rt.Producers["application/hal+json"] = runtime.JSONProducer()
	client := deploy.New(rt, strfmt.Default)

	var token = os.Getenv("APTIBLE_ACCESS_TOKEN")
	bearerTokenAuth := httptransport.BearerToken(token)

	// Creating app
	appreq := models.AppRequest3{Handle: &handle}
	params := operations.NewPostAccountsAccountIDAppsParams().WithAccountID(account_id).WithAppRequest(&appreq)
	resp, err := client.Operations.PostAccountsAccountIDApps(params, bearerTokenAuth)
	if err != nil {
		CreateLogger.Println("There was an error when completing the request to create the app.\n[ERROR] -", resp)
		return err
	}
	CreateLogger.Println("This is the response.\n[INFO] -", resp)
	app := resp.Payload
	app_id := strconv.Itoa(int(*app.ID))
	d.Set("app_id", app_id)
	d.Set("git_repo", app.GitRepo)
	d.Set("created_at", app.CreatedAt)
	d.SetId(handle)

	// Deploying app
	req_type := "deploy"
	env := d.Get("env")
	app_req := models.AppRequest21{Type: &req_type, Env: env, ContainerCount: 1, ContainerSize: 1024}
	app_id_str := d.Get("app_id").(string)
	id, err := strconv.ParseInt(app_id_str, 10, 64) // WithAppID takes in an int64
	app_params := operations.NewPostAppsAppIDOperationsParams().WithAppID(id).WithAppRequest(&app_req)
	app_resp, err := client.Operations.PostAppsAppIDOperations(app_params, bearerTokenAuth)
	if err != nil {
		CreateLogger.Println("There was an error when completing the request to deploy the app.\n[ERROR] -", app_resp)
		return err
	}

	CreateLogger.Println("This is the response.\n[INFO] -", app_resp)
	return resourceAppRead(d, m)
}

func resourceAppRead(d *schema.ResourceData, m interface{}) error {
	app_id_str := d.Get("app_id").(string)
	app_id, err := strconv.ParseInt(app_id_str, 10, 64) // WithID takes in an int64

	rt := httptransport.New(
		"api-rachel.aptible-sandbox.com",
		deploy.DefaultBasePath,
		deploy.DefaultSchemes)
	rt.Consumers["application/hal+json"] = runtime.JSONConsumer()
	rt.Producers["application/hal+json"] = runtime.JSONProducer()
	client := deploy.New(rt, strfmt.Default)

	var token = os.Getenv("APTIBLE_ACCESS_TOKEN")
	bearerTokenAuth := httptransport.BearerToken(token)

	params := operations.NewGetAppsIDParams().WithID(app_id)
	resp, err := client.Operations.GetAppsID(params, bearerTokenAuth)
	if err != nil {
		CreateLogger.Println("There was an error when completing the request to get the app.\n[ERROR] -", err)
		CreateLogger.Println("The app id was: ", app_id)
		return err
	}
	CreateLogger.Println("This is the response.\n[INFO] -", resp)
	return nil
}

func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceAppRead(d, m)
}

func resourceAppDelete(d *schema.ResourceData, m interface{}) error {
	read_err := resourceAppRead(d, m)
	// TODO: it could error for many reasons: bad token, 404, etc. so this isn't accurate
	// if the app exists
	if read_err == nil {
		app_id_str := d.Get("app_id").(string)
		app_id, err := strconv.ParseInt(app_id_str, 10, 64) // WithID takes in an int64

		rt := httptransport.New(
			"api-rachel.aptible-sandbox.com",
			deploy.DefaultBasePath,
			deploy.DefaultSchemes)
		rt.Consumers["application/hal+json"] = runtime.JSONConsumer()
		rt.Producers["application/hal+json"] = runtime.JSONProducer()
		client := deploy.New(rt, strfmt.Default)

		var token = os.Getenv("APTIBLE_ACCESS_TOKEN")
		bearerTokenAuth := httptransport.BearerToken(token)

		req_type := "deprovision"
		app_req := models.AppRequest21{Type: &req_type}
		app_params := operations.NewPostAppsAppIDOperationsParams().WithAppID(app_id).WithAppRequest(&app_req)
		app_resp, err := client.Operations.PostAppsAppIDOperations(app_params, bearerTokenAuth)
		if err != nil {
			CreateLogger.Println("There was an error when completing the request to deploy the app.\n[ERROR] -", app_resp)
			return err
		}
	}

	d.SetId("")
	return nil
}

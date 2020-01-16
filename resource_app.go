package main

import (
	"os"
	"strconv"

	deploy "github.com/aptible/go-deploy/client"
	"github.com/aptible/go-deploy/client/operations"
	"github.com/aptible/go-deploy/models"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform/helper/schema"
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
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"git_repo": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, m interface{}) error {
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

	appreq := models.AppRequest3{Handle: &handle}
	params := operations.NewPostAccountsAccountIDAppsParams().WithAccountID(account_id).WithAppRequest(&appreq)
	resp, err := client.Operations.PostAccountsAccountIDApps(params, bearerTokenAuth)
	if err != nil {
		CreateLogger.Println("There was an error when completing the request.\n[ERROR] -", resp)
		return err
	}
	CreateLogger.Println("This is the response.\n[INFO] -", resp)

	d.SetId(handle)
	return resourceAppRead(d, m)
}

func resourceAppRead(d *schema.ResourceData, m interface{}) error {
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

	page := int64(1)
	params := operations.NewGetAccountsAccountIDAppsParams().WithAccountID(account_id).WithPage(&page)
	resp, err := client.Operations.GetAccountsAccountIDApps(params, bearerTokenAuth)
	if err != nil {
		CreateLogger.Println("There was an error when completing the request.\n[ERROR] -", resp)
		return err
	}

	for _, app := range resp.Payload.Embedded.Apps {
		if app.Handle == handle {
			d.Set("app_id", app.ID)
			d.Set("git_repo", app.GitRepo)
			d.Set("created_at", app.CreatedAt)
			break
		}
	}
	return nil
}

func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceAppRead(d, m)
}

func resourceAppDelete(d *schema.ResourceData, m interface{}) error {
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return nil
}

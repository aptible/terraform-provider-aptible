package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/reggregory/go-deploy/aptible"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate, // POST
		Read:   resourceEndpointRead,   // GET
		Update: resourceEndpointUpdate, // PUT
		Delete: resourceEndpointDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"account_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"service_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"certificate": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEndpointCreate(d *schema.ResourceData, m interface{}) error {
	client := aptible.SetUpClient()
	app_id := int64(d.Get("app_id").(int))
	payload, err := client.CreateEndpoint(app_id)
	if err != nil {
		return err
	}
	d.Set("endpoint_id", *payload.ID)
	d.Set("hostname", *payload.VirtualDomain)
	d.SetId(*payload.VirtualDomain)
	return resourceEndpointRead(d, m)
}

func resourceEndpointRead(d *schema.ResourceData, m interface{}) error {
	client := aptible.SetUpClient()
	endpoint_id := int64(d.Get("endpoint_id").(int))
	_, deleted, err := client.GetEndpoint(endpoint_id)
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

func resourceEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceEndpointRead(d, m)
}

func resourceEndpointDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}

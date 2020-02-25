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
				ForceNew: true,
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"service_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"internal": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"container_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"ip_filtering": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"platform": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
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
		AppLogger.Println("There was an error when completing the request to create the endpoint.\n[ERROR] -", err)
		return err
	}

	d.Set("hostname", *payload.VirtualDomain)
	d.Set("endpoint_id", int(*payload.ID))
	d.SetId(*payload.VirtualDomain)
	return resourceEndpointRead(d, m)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceEndpointRead(d *schema.ResourceData, m interface{}) error {
	client := aptible.SetUpClient()
	endpoint_id := int64(d.Get("endpoint_id").(int))
	payload, deleted, err := client.GetEndpoint(endpoint_id)
	if err != nil {
		AppLogger.Println(err)
		return err
	}
	if deleted {
		d.SetId("")
		return nil
	}

	d.Set("ip_filtering", payload.IPWhitelist)
	d.Set("platform", *payload.Platform)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	client := aptible.SetUpClient()
	endpoint_id := int64(d.Get("endpoint_id").(int))
	updates := map[string]interface{}{}

	if d.HasChange("container_port") {
		container_port := int64(d.Get("container_port").(int))
		updates["container_port"] = container_port
	}

	if d.HasChange("ip_filtering") {
		ip_filtering := d.Get("internal").([]string)
		updates["ip_filtering"] = ip_filtering
	}

	if d.HasChange("platform") {
		platform := d.Get("platform").(string)
		updates["platform"] = platform
	}

	err := client.UpdateEndpoint(endpoint_id, updates)
	if err != nil {
		AppLogger.Println("There was an error when completing the request.\n[ERROR] -", err)
		return err
	}

	return resourceEndpointRead(d, m)
}

func resourceEndpointDelete(d *schema.ResourceData, m interface{}) error {
	client := aptible.SetUpClient()
	endpoint_id := int64(d.Get("endpoint_id").(int))
	err := client.DeleteEndpoint(endpoint_id)
	if err != nil {
		AppLogger.Println(err)
		return err
	}

	d.SetId("")
	return nil
}

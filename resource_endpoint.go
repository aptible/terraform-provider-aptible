package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aptible/go-deploy/aptible"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate, // POST
		Read:   resourceEndpointRead,   // GET
		Update: resourceEndpointUpdate, // PUT
		Delete: resourceEndpointDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"env_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true, // TODO: Make optional once we support database endpoints.
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "HTTPS",
				ForceNew: true,
			},
			// v2, for now there's only one service per app
			"service_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			// v2, for now Default = true
			"certificate": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"internal": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"container_port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  80,
			},
			"ip_filtering": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"platform": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "alb",
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
	client := m.(*aptible.Client)
	app_id := int64(d.Get("app_id").(int))

	if_slice := d.Get("ip_filtering").([]interface{})
	ip_whitelist, _ := aptible.MakeStringSlice(if_slice)

	t := d.Get("type").(string)
	t, err := aptible.GetEndpointType(t)
	if err != nil {
		AppLogger.Println(err)
		return err
	}

	attrs := aptible.CreateAttrs{
		Type:          &t,
		Default:       true,
		Internal:      d.Get("internal").(bool),
		ContainerPort: int64(d.Get("container_port").(int)),
		IPWhitelist:   ip_whitelist,
		Platform:      d.Get("platform").(string),
	}

	payload, err := client.CreateEndpoint(app_id, attrs)
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
	client := m.(*aptible.Client)
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

	if payload.ContainerPort != nil {
		d.Set("container_port", *payload.ContainerPort)
	}
	d.Set("ip_filtering", payload.IPWhitelist)
	d.Set("platform", *payload.Platform)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aptible.Client)
	endpoint_id := int64(d.Get("endpoint_id").(int))
	if_slice := d.Get("ip_filtering").([]interface{})
	ip_whitelist, _ := aptible.MakeStringSlice(if_slice)

	updates := aptible.Updates{
		ContainerPort: int64(d.Get("container_port").(int)),
		IPWhitelist:   ip_whitelist,
		Platform:      d.Get("platform").(string),
	}

	err := client.UpdateEndpoint(endpoint_id, updates)
	if err != nil {
		AppLogger.Println("There was an error when completing the request.\n[ERROR] -", err)
		return err
	}

	return resourceEndpointRead(d, m)
}

func resourceEndpointDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aptible.Client)
	endpoint_id := int64(d.Get("endpoint_id").(int))
	err := client.DeleteEndpoint(endpoint_id)
	if err != nil {
		AppLogger.Println(err)
		return err
	}

	d.SetId("")
	return nil
}

package aptible

import (
	"log"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate, // POST
		Read:   resourceEndpointRead,   // GET
		Update: resourceEndpointUpdate, // PUT
		Delete: resourceEndpointDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validResourceTypes, false),
				ForceNew:     true,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validEndpointTypes, false),
				Default:      "https",
				ForceNew:     true,
			},
			// v2, for now there's only one service per app
			"service_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			// v2, for now Default = true
			"certificate": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"container_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
				Default:      80,
			},
			"ip_filtering": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"platform": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validPlatforms, false),
				Default:      "alb",
			},
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	resource_id := int64(d.Get("resource_id").(int))
	resource_type := d.Get("resource_type").(string)

	if_slice := d.Get("ip_filtering").([]interface{})
	ip_whitelist, _ := aptible.MakeStringSlice(if_slice)

	t := d.Get("endpoint_type").(string)
	t, err := aptible.GetEndpointType(t)
	if err != nil {
		log.Println(err)
		return err
	}

	attrs := aptible.EndpointCreateAttrs{
		ResourceType:  resource_type,
		Type:          &t,
		Internal:      d.Get("internal").(bool),
		ContainerPort: int64(d.Get("container_port").(int)),
		IPWhitelist:   ip_whitelist,
		Platform:      d.Get("platform").(string),
	}
	if resource_type == "app" {
		attrs.Default = true
	} else {
		attrs.Default = false
	}

	ep, err := client.CreateEndpoint(resource_id, attrs)
	if err != nil {
		log.Println("There was an error when completing the request to create the endpoint.\n[ERROR] -", err)
		return err
	}

	d.SetId(ep.Hostname)
	d.Set("hostname", ep.Hostname)
	d.Set("endpoint_id", int(ep.ID))

	return resourceEndpointRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpoint_id := int64(d.Get("endpoint_id").(int))
	resource_type := d.Get("resource_type").(string)
	ep, deleted, err := client.GetEndpoint(endpoint_id, resource_type)
	if err != nil {
		log.Println(err)
		return err
	}
	if deleted {
		d.SetId("")
		return nil
	}

	if resource_type == "app" {
		d.Set("container_port", ep.ContainerPort)
		d.Set("platform", ep.Platform)
	}
	d.Set("ip_filtering", ep.IPWhitelist)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpoint_id := int64(d.Get("endpoint_id").(int))
	if_slice := d.Get("ip_filtering").([]interface{})
	ip_whitelist, _ := aptible.MakeStringSlice(if_slice)

	updates := aptible.EndpointUpdates{
		ContainerPort: int64(d.Get("container_port").(int)),
		IPWhitelist:   ip_whitelist,
		Platform:      d.Get("platform").(string),
	}

	err := client.UpdateEndpoint(endpoint_id, updates)
	if err != nil {
		log.Println("There was an error when completing the request.\n[ERROR] -", err)
		return err
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpoint_id := int64(d.Get("endpoint_id").(int))
	err := client.DeleteEndpoint(endpoint_id)
	if err != nil {
		log.Println(err)
		return err
	}

	d.SetId("")
	return nil
}

var validResourceTypes = []string{
	"app",
	"database",
}

var validEndpointTypes = []string{
	"https",
	"tls",
	"tcp",
}

var validPlatforms = []string{
	"alb",
	"elb",
}

package aptible

import (
	"fmt"
	"log"
	"strconv"

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
			"service_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validEndpointTypes, false),
				Default:      "https",
				ForceNew:     true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"default_domain": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
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

	serviceName := d.Get("service_name").(string)
	resourceID := int64(d.Get("resource_id").(int))
	service, err := client.GetServiceForAppByName(resourceID, serviceName)
	if err != nil {
		log.Println(err)
		return err
	}

	interfaceSlice := d.Get("ip_filtering").([]interface{})
	ipWhitelist, _ := aptible.MakeStringSlice(interfaceSlice)

	humanReadableEndpointType := d.Get("endpoint_type").(string)
	endpointType, err := aptible.GetEndpointType(humanReadableEndpointType)
	if err != nil {
		log.Println(err)
		return err
	}

	defaultDomain := d.Get("default_domain").(bool)
	managed := d.Get("managed").(bool)
	domain := d.Get("domain").(string)

	if defaultDomain && managed {
		return fmt.Errorf("do not specify Managed HTTPS if using the Default Domain")
	}
	if defaultDomain && domain != "" {
		return fmt.Errorf("cannot specify domain when using Default Domain")
	}

	if service.ResourceType == "database" && defaultDomain {
		return fmt.Errorf("cannot use Default Domain on Databases")
	}
	if service.ResourceType == "database" && domain != "" {
		return fmt.Errorf("cannot specify domain on Databases")
	}

	attrs := aptible.EndpointCreateAttrs{
		Type:          &endpointType,
		Internal:      d.Get("internal").(bool),
		ContainerPort: int64(d.Get("container_port").(int)),
		IPWhitelist:   ipWhitelist,
		Platform:      d.Get("platform").(string),
		Default:       defaultDomain,
		Acme:          managed,
	}
	if domain != "" {
		attrs.UserDomain = domain
	}

	endpoint, err := client.CreateEndpoint(service, attrs)
	if err != nil {
		log.Println("There was an error when completing the request to create the endpoint.\n[ERROR] -", err)
		return err
	}

	_ = d.Set("endpoint_id", endpoint.ID)
	_ = d.Set("service_id", service.ID)

	d.SetId(endpoint.ExternalHost)

	return resourceEndpointRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpointID := int64(d.Get("endpoint_id").(int))

	log.Println("getting Endpoint with ID: " + strconv.Itoa(int(endpointID)))

	endpoint, err := client.GetEndpoint(endpointID)
	if err != nil {
		log.Println(err)
		return err
	}
	if endpoint.Deleted {
		d.SetId("")
		log.Println("Endpoint with ID: " + strconv.Itoa(int(endpointID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	serviceID := int64(d.Get("service_id").(int))
	service, err := client.GetService(serviceID)
	if err != nil {
		log.Println(err)
		return err
	}

	if service.ResourceType == "app" {
		_ = d.Set("container_port", endpoint.ContainerPort)
		_ = d.Set("platform", endpoint.Platform)
	}
	_ = d.Set("ip_filtering", endpoint.IPWhitelist)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpointID := int64(d.Get("endpoint_id").(int))
	interfaceSlice := d.Get("ip_filtering").([]interface{})
	ipWhitelist, _ := aptible.MakeStringSlice(interfaceSlice)

	updates := aptible.EndpointUpdates{
		ContainerPort: int64(d.Get("container_port").(int)),
		IPWhitelist:   ipWhitelist,
		Platform:      d.Get("platform").(string),
	}

	err := client.UpdateEndpoint(endpointID, updates)
	if err != nil {
		log.Println("There was an error when completing the request.\n[ERROR] -", err)
		return err
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpointID := int64(d.Get("endpoint_id").(int))
	err := client.DeleteEndpoint(endpointID)
	if err != nil {
		log.Println(err)
		return err
	}

	d.SetId("")
	return nil
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

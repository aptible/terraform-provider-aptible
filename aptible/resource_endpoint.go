package aptible

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate, // POST
		Read:   resourceEndpointRead,   // GET
		Update: resourceEndpointUpdate, // PUT
		Delete: resourceEndpointDelete, // DELETE
		Importer: &schema.ResourceImporter{
			State: resourceEndpointImport,
		},

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
				ValidateFunc: validation.StringInSlice(validResourceTypes, false),
				Required:     true,
				ForceNew:     true,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validEndpointTypes, false),
				Default:      "https",
				ForceNew:     true,
			},
			"process_type": {
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
				ForceNew: true,
			},
			"container_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
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
			"endpoint_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"virtual_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_validation_record": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_validation_value": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	service := aptible.Service{}
	var err error

	processType := d.Get("process_type").(string)
	resourceID := int64(d.Get("resource_id").(int))
	resourceType := d.Get("resource_type").(string)
	interfaceSlice := d.Get("ip_filtering").([]interface{})
	ipWhitelist, _ := aptible.MakeStringSlice(interfaceSlice)
	defaultDomain := d.Get("default_domain").(bool)
	managed := d.Get("managed").(bool)
	domain := d.Get("domain").(string)

	if defaultDomain && managed {
		return fmt.Errorf("do not specify Managed HTTPS if using the Default Domain")
	}
	if managed && domain == "" {
		return fmt.Errorf("managed endpoints must specify a domain")
	}
	if defaultDomain && domain != "" {
		return fmt.Errorf("cannot specify domain when using Default Domain")
	}

	if resourceType == "app" {
		service, err = client.GetServiceForAppByName(resourceID, processType)
		if err != nil {
			log.Println(err)
			return generateErrorFromClientError(err)
		}
	} else {
		database, err := client.GetDatabase(resourceID)
		if err != nil {
			log.Println(err)
			return generateErrorFromClientError(err)
		}
		service = database.Service
	}

	if service.ResourceType == "database" && defaultDomain {
		return fmt.Errorf("cannot use Default Domain on Databases")
	}
	if service.ResourceType == "database" && domain != "" {
		return fmt.Errorf("cannot specify domain on Databases")
	}

	humanReadableEndpointType := d.Get("endpoint_type").(string)
	endpointType, err := aptible.GetEndpointType(humanReadableEndpointType)
	if err != nil {
		log.Println(err)
		return err
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
		return generateErrorFromClientError(err)
	}

	_ = d.Set("endpoint_id", endpoint.ID)
	d.SetId(strconv.Itoa(int(endpoint.ID)))

	return resourceEndpointRead(d, meta)
}

func resourceEndpointImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	endpointID, _ := strconv.Atoi(d.Id())
	_ = d.Set("endpoint_id", endpointID)
	err := resourceEndpointRead(d, meta)
	return []*schema.ResourceData{d}, err
}

func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpointID := int64(d.Get("endpoint_id").(int))

	endpoint, err := client.GetEndpoint(endpointID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}
	if endpoint.Deleted {
		d.SetId("")
		log.Println("Endpoint with ID: " + strconv.Itoa(int(endpointID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	service := endpoint.Service
	if service.ResourceType == "app" {
		_ = d.Set("container_port", endpoint.ContainerPort)
		_ = d.Set("platform", endpoint.Platform)
		_ = d.Set("process_type", endpoint.Service.ProcessType)
	}

	_ = d.Set("resource_type", endpoint.Service.ResourceType)
	_ = d.Set("ip_filtering", endpoint.IPWhitelist)
	_ = d.Set("env_id", endpoint.Service.EnvironmentID)
	_ = d.Set("resource_id", endpoint.Service.ResourceID)
	_ = d.Set("endpoint_type", endpoint.Type)
	_ = d.Set("default_domain", endpoint.Default)
	_ = d.Set("managed", endpoint.Acme)
	_ = d.Set("domain", endpoint.UserDomain)
	_ = d.Set("virtual_domain", endpoint.VirtualDomain)
	_ = d.Set("internal", endpoint.Internal)
	_ = d.Set("ip_filtering", endpoint.IPWhitelist)
	_ = d.Set("platform", endpoint.Platform)
	_ = d.Set("endpoint_id", endpoint.ID)
	_ = d.Set("external_hostname", endpoint.ExternalHost)

	for _, c := range endpoint.AcmeChallenges {
		if c.Method != "dns01" {
			continue
		}
		_ = d.Set("dns_validation_record", c.From)
		_ = d.Set("dns_validation_value", c.To)
		break
	}

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
		return generateErrorFromClientError(err)
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	endpointID := int64(d.Get("endpoint_id").(int))
	err := client.DeleteEndpoint(endpointID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
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

var validResourceTypes = []string{
	"app",
	"database",
}

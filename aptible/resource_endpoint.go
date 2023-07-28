package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create:        resourceEndpointCreate, // POST
		Read:          resourceEndpointRead,   // GET
		Update:        resourceEndpointUpdate, // PUT
		Delete:        resourceEndpointDelete, // DELETE
		CustomizeDiff: resourceEndpointValidate,
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
			"container_ports": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeInt,
					ValidateFunc: validation.IntBetween(1, 65535),
				},
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

func resourceEndpointValidate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	d := ResourceDiff{ResourceDiff: diff}
	interfaceContainerPortsSlice := d.Get("container_ports").([]interface{})
	containerPorts, _ := aptible.MakeInt64Slice(interfaceContainerPortsSlice)
	containerPort, _ := (d.Get("container_port").(int))
	endpointType := d.Get("endpoint_type").(string)
	var err error

	// container_port and container ports are mutually exclusive fields
	if containerPort != 0 && len(containerPorts) != 0 {
		err = multierror.Append(err, fmt.Errorf("do not specify container ports AND container port (see terraform docs)"))
	}

	if containerPort != 0 && (endpointType == "tcp" || endpointType == "tls") {
		err = multierror.Append(err, fmt.Errorf("do not specify container port with a tls or tcp endpoint"))
	}

	// container ports can only be used with tls/tcp
	if len(containerPorts) != 0 && (endpointType == "https") {
		err = multierror.Append(err, fmt.Errorf("do not specify container ports with https endpoint (see terraform docs)"))
	}

	return err
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
	interfaceContainerPortsSlice := d.Get("container_ports").([]interface{})
	containerPorts, _ := aptible.MakeInt64Slice(interfaceContainerPortsSlice)
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
		Type:           &endpointType,
		Internal:       d.Get("internal").(bool),
		ContainerPort:  int64(d.Get("container_port").(int)),
		ContainerPorts: containerPorts,
		IPWhitelist:    ipWhitelist,
		Platform:       d.Get("platform").(string),
		Default:        defaultDomain,
		Acme:           managed,
	}
	if len(containerPorts) > 0 {
		// zero values for arrays give non-deterministic responses from backend
		attrs.ContainerPorts = containerPorts
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
		_ = d.Set("container_ports", endpoint.ContainerPorts)
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
	interfaceContainerPortsSlice := d.Get("container_ports").([]interface{})
	containerPorts, _ := aptible.MakeInt64Slice(interfaceContainerPortsSlice)

	updates := aptible.EndpointUpdates{
		ContainerPort:  int64(d.Get("container_port").(int)),
		ContainerPorts: containerPorts,
		IPWhitelist:    ipWhitelist,
		Platform:       d.Get("platform").(string),
	}
	if len(containerPorts) > 0 {
		// zero values for arrays give non-deterministic responses from backend
		updates.ContainerPorts = containerPorts
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

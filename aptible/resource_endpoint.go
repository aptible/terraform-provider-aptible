package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/go-deploy/aptible"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEndpointCreate, // POST
		ReadContext:   resourceEndpointRead,   // GET
		UpdateContext: resourceEndpointUpdate, // PUT
		DeleteContext: resourceEndpointDelete, // DELETE
		CustomizeDiff: resourceEndpointValidate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImport,
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
	if len(containerPorts) != 0 && (endpointType == "https" || endpointType == "grpc") {
		err = multierror.Append(err, fmt.Errorf("do not specify container ports with %s endpoint (see terraform docs)", endpointType))
	}

	return err
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	legacy := m.LegacyClient
	client := m.Client
	ctx = m.APIContext(ctx)
	service := aptible.Service{}
	diags := diag.Diagnostics{}

	processType := d.Get("process_type").(string)
	resourceID := int64(d.Get("resource_id").(int))
	resourceType := d.Get("resource_type").(string)
	ipWhitelist, err := makeStringSlice(d.Get("ip_filtering").([]interface{}))
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   fmt.Sprintf("Failed to parse ip_filtering: %s", err.Error()),
		})
	}
	containerPorts, err := makeInt32Slice(d.Get("container_ports").([]interface{}))
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   fmt.Sprintf("Failed to parse container_ports: %s", err.Error()),
		})
	}
	defaultDomain := d.Get("default_domain").(bool)
	managed := d.Get("managed").(bool)
	domain := d.Get("domain").(string)

	if defaultDomain && managed {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Do not specify Managed HTTPS if using the Default Domain",
		})
	}
	if managed && domain == "" {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Managed endpoints must specify a domain",
		})
	}
	if defaultDomain && domain != "" {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Cannot specify domain when using Default Domain",
		})
	}

	if resourceType == "app" {
		service, err = legacy.GetServiceForAppByName(resourceID, processType)
		if err != nil {
			log.Println(err)
			return generateDiagnosticsFromClientError(err)
		}
	} else {
		database, err := legacy.GetDatabase(resourceID)
		if err != nil {
			log.Println(err)
			return generateDiagnosticsFromClientError(err)
		}
		service = database.Service
	}

	if service.ResourceType == "database" && defaultDomain {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Cannot use Default Domain on Databases",
		})
	}
	if service.ResourceType == "database" && domain != "" {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Cannot specify domain on Databases",
		})
	}

	humanReadableEndpointType := d.Get("endpoint_type").(string)
	endpointType, err := aptible.GetEndpointType(humanReadableEndpointType)
	if err != nil {
		log.Println(err)
		return generateDiagnosticsFromClientError(err)
	}

	internal := d.Get("internal").(bool)
	port := int32(d.Get("container_port").(int))
	platform := d.Get("platform").(string)

	attrs := aptibleapi.CreateVhostRequest{
		Type:           endpointType,
		Internal:       &internal,
		ContainerPort:  &port,
		ContainerPorts: containerPorts,
		IpWhitelist:    ipWhitelist,
		Platform:       &platform,
		Default:        &defaultDomain,
		Acme:           &managed,
	}
	if domain != "" {
		attrs.UserDomain = &domain
	}

	endpoint, _, err := client.VhostsAPI.
		CreateVhost(ctx, int32(service.ID)).
		CreateVhostRequest(attrs).
		Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error creating endpoint",
			Detail:   err.Error(),
		})
	}

	payload := aptibleapi.NewCreateOperationRequest("Provision")
	operation, _, err := client.OperationsAPI.
		CreateOperationForVhost(ctx, endpoint.Id).
		CreateOperationRequest(*payload).
		Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error creating endpoint",
			Detail:   err.Error(),
		})
	}

	_ = d.Set("endpoint_id", endpoint.Id)
	d.SetId(strconv.Itoa(int(endpoint.Id)))

	_, err = legacy.WaitForOperation(int64(operation.Id))
	if err != nil {
		// Do not return here so that the read method can hydrate the state
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error provisioning new endpoint %d", endpoint.Id),
			Detail:   err.Error(),
		})
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	endpointID, _ := strconv.Atoi(d.Id())
	_ = d.Set("endpoint_id", endpointID)
	diags := resourceEndpointRead(ctx, d, meta)
	return []*schema.ResourceData{d}, diagnosticsToError(diags)
}

func resourceEndpointRead(_ context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*providerMetadata).LegacyClient
	endpointID := int64(d.Get("endpoint_id").(int))

	endpoint, err := client.GetEndpoint(endpointID)
	if err != nil {
		log.Println(err)
		return generateDiagnosticsFromClientError(err)
	}
	if endpoint.Deleted {
		d.SetId("")
		log.Printf("Endpoint with ID: %d was deleted outside of Terraform. Removing it from Terraform state.", endpointID)
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
func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).LegacyClient
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

	err := client.UpdateEndpoint(endpointID, updates)
	if err != nil {
		log.Println("There was an error when completing the request.\n[ERROR] -", err)
		return generateDiagnosticsFromClientError(err)
	}

	return resourceEndpointRead(ctx, d, meta)
}

func resourceEndpointDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).LegacyClient
	endpointID := int64(d.Get("endpoint_id").(int))
	err := client.DeleteEndpoint(endpointID)
	if err != nil {
		log.Println(err)
		return generateDiagnosticsFromClientError(err)
	}

	d.SetId("")
	return nil
}

var validEndpointTypes = []string{
	"https",
	"tls",
	"tcp",
	"grpc",
}

var validPlatforms = []string{
	"alb",
	"elb",
}

var validResourceTypes = []string{
	"app",
	"database",
}

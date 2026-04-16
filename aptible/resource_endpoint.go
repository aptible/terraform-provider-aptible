package aptible

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"load_balancing_algorithm_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validLbAlgorithms, false),
			},
			"force_ssl": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"true", "false", ""}, false),
			},
			"maintenance_page_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"idle_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateZeroOrIntBetween(30, 2400),
			},
			"release_healthcheck_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateZeroOrIntBetween(1, 900),
			},
			"strict_health_checks": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"true", "false", ""}, false),
			},
			"show_elb_healthchecks": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"true", "false", ""}, false),
			},
			"ssl_protocols_override": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_ciphers_override": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disable_weak_cipher_suites": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"true", "false", ""}, false),
			},
		},
	}
}

// endpointCategory returns a canonical category string from the user-facing
// endpoint_type and platform values, used to drive settings validation.
// Returns "" when type/platform are not yet determined (plan not yet complete).
func endpointCategory(endpointType, platform string) string {
	switch endpointType {
	case "https":
		switch platform {
		case "alb":
			return "alb"
		case "elb":
			return "elb"
		}
	case "tls":
		return "tls"
	case "grpc":
		return "grpc"
	case "tcp":
		return "tcp"
	}
	return ""
}

// endpointSettingCategories maps each setting attribute to the endpoint
// categories that support it.
var endpointSettingCategories = map[string][]string{
	"force_ssl":                   {"alb", "elb"},
	"maintenance_page_url":        {"alb", "elb"},
	"idle_timeout":                {"alb", "elb", "tls", "grpc", "tcp"},
	"release_healthcheck_timeout": {"alb", "elb"},
	"strict_health_checks":        {"alb", "elb"},
	"show_elb_healthchecks":       {"alb", "elb"},
	"ssl_protocols_override":      {"alb", "elb", "tls", "grpc"},
	"ssl_ciphers_override":        {"elb", "tls", "grpc"},
	"disable_weak_cipher_suites":  {"elb", "tls", "grpc"},
}

func resourceEndpointValidate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	d := ResourceDiff{ResourceDiff: diff}
	interfaceContainerPortsSlice := d.Get("container_ports").([]interface{})
	containerPorts, _ := aptible.MakeInt64Slice(interfaceContainerPortsSlice)
	containerPort, _ := (d.Get("container_port").(int))
	endpointType := d.Get("endpoint_type").(string)
	platform := d.Get("platform").(string)
	lbAlgorithmType := d.Get("load_balancing_algorithm_type").(string)
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

	// load balancing algorithm can only be used with ALBs
	if (lbAlgorithmType != "") && (platform != "alb") {
		err = multierror.Append(err, fmt.Errorf("do not specify a load balancing algorithm with %s endpoint", platform))
	}

	// Validate setting attributes against endpoint category
	if category := endpointCategory(endpointType, platform); category != "" {
		for attr, validCategories := range endpointSettingCategories {
			if !isEndpointSettingSet(d.ResourceDiff, attr) {
				continue
			}
			valid := false
			for _, c := range validCategories {
				if c == category {
					valid = true
					break
				}
			}
			if !valid {
				err = multierror.Append(err, fmt.Errorf("%s is not supported for %s endpoints", attr, category))
			}
		}
	}

	return err
}

// isEndpointSettingSet returns true when a setting attribute has a non-zero
// value — i.e. the user actually specified it.
func isEndpointSettingSet(d *schema.ResourceDiff, attr string) bool {
	v := d.Get(attr)
	switch val := v.(type) {
	case int:
		return val != 0
	case string:
		return val != ""
	}
	return false
}

// validateZeroOrIntBetween returns a ValidateFunc that allows 0 (meaning
// "unset") as well as any integer in [min, max].
func validateZeroOrIntBetween(min, max int) schema.SchemaValidateFunc {
	return func(val interface{}, key string) (warns []string, errs []error) {
		if val.(int) == 0 {
			return
		}
		return validation.IntBetween(min, max)(val, key)
	}
}

// buildEndpointSettingsMap constructs the settings map for Create from the
// individual typed attributes.  Only non-zero values are included.
func buildEndpointSettingsMap(d *schema.ResourceData) map[string]string {
	m := map[string]string{}
	if v := d.Get("force_ssl").(string); v == "true" {
		m["FORCE_SSL"] = "true"
	}
	if v := d.Get("maintenance_page_url").(string); v != "" {
		m["MAINTENANCE_PAGE_URL"] = v
	}
	if v := d.Get("idle_timeout").(int); v != 0 {
		m["IDLE_TIMEOUT"] = strconv.Itoa(v)
	}
	if v := d.Get("release_healthcheck_timeout").(int); v != 0 {
		m["RELEASE_HEALTHCHECK_TIMEOUT"] = strconv.Itoa(v)
	}
	if v := d.Get("strict_health_checks").(string); v == "true" {
		m["STRICT_HEALTH_CHECKS"] = "true"
	}
	if v := d.Get("show_elb_healthchecks").(string); v == "true" {
		m["SHOW_ELB_HEALTHCHECKS"] = "true"
	}
	if v := d.Get("ssl_protocols_override").(string); v != "" {
		m["SSL_PROTOCOLS_OVERRIDE"] = v
	}
	if v := d.Get("ssl_ciphers_override").(string); v != "" {
		m["SSL_CIPHERS_OVERRIDE"] = v
	}
	if v := d.Get("disable_weak_cipher_suites").(string); v == "true" {
		m["DISABLE_WEAK_CIPHER_SUITES"] = "true"
	}
	return m
}

// applyEndpointSettingsToState reads individual settings from the API response
// map and sets each typed attribute in Terraform state.
func applyEndpointSettingsToState(d *schema.ResourceData, settings map[string]interface{}) {
	forceSslStr, _ := settings["FORCE_SSL"].(string)
	_ = d.Set("force_ssl", forceSslStr)
	maintenancePageURL, _ := settings["MAINTENANCE_PAGE_URL"].(string)
	_ = d.Set("maintenance_page_url", maintenancePageURL)
	idleTimeoutStr, _ := settings["IDLE_TIMEOUT"].(string)
	idleTimeout, _ := strconv.Atoi(idleTimeoutStr)
	_ = d.Set("idle_timeout", idleTimeout)
	rhtStr, _ := settings["RELEASE_HEALTHCHECK_TIMEOUT"].(string)
	rht, _ := strconv.Atoi(rhtStr)
	_ = d.Set("release_healthcheck_timeout", rht)
	strictStr, _ := settings["STRICT_HEALTH_CHECKS"].(string)
	_ = d.Set("strict_health_checks", strictStr)
	showStr, _ := settings["SHOW_ELB_HEALTHCHECKS"].(string)
	_ = d.Set("show_elb_healthchecks", showStr)
	sslProto, _ := settings["SSL_PROTOCOLS_OVERRIDE"].(string)
	_ = d.Set("ssl_protocols_override", sslProto)
	sslCiphers, _ := settings["SSL_CIPHERS_OVERRIDE"].(string)
	_ = d.Set("ssl_ciphers_override", sslCiphers)
	disableStr, _ := settings["DISABLE_WEAK_CIPHER_SUITES"].(string)
	_ = d.Set("disable_weak_cipher_suites", disableStr)
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
	shared := d.Get("shared").(bool)
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
	if shared && !defaultDomain && domain == "" {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Shared endpoints must specify a domain",
		})
	}
	if shared && strings.ContainsAny(domain, "*?") {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Validation Error",
			Detail:   "Shared endpoints cannot use domains containing '*' or '?'",
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

	attrs := aptibleapi.NewCreateVhostRequest(endpointType)
	attrs.SetInternal(d.Get("internal").(bool))
	attrs.SetIpWhitelist(ipWhitelist)
	attrs.SetPlatform(d.Get("platform").(string))
	attrs.SetDefault(defaultDomain)
	attrs.SetAcme(managed)
	attrs.SetShared(shared)
	lbAlgorithmType := d.Get("load_balancing_algorithm_type").(string)
	if lbAlgorithmType != "" {
		attrs.SetLoadBalancingAlgorithmType(lbAlgorithmType)
	}

	containerPort := int32(d.Get("container_port").(int))

	if endpointType == "tcp" {
		attrs.SetContainerPorts(containerPorts)
	} else if containerPort != 0 {
		attrs.SetContainerPort(containerPort)
	}

	if domain != "" {
		attrs.SetUserDomain(domain)
	}

	endpoint, _, err := client.VhostsAPI.
		CreateVhost(ctx, int32(service.ID)).
		CreateVhostRequest(*attrs).
		Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create endpoint",
			Detail:   err.Error(),
		})
	}

	payload := aptibleapi.NewCreateOperationRequest("provision")

	if settingsMap := buildEndpointSettingsMap(d); len(settingsMap) > 0 {
		payload.SetSettings(settingsMap)
	}

	operation, _, err := client.OperationsAPI.
		CreateOperationForVhost(ctx, endpoint.Id).
		CreateOperationRequest(*payload).
		Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create endpoint",
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
			Summary:  fmt.Sprintf("Failed to provision new endpoint %d", endpoint.Id),
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

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	m := meta.(*providerMetadata)
	client := m.Client
	ctx = m.APIContext(ctx)
	diags = diag.Diagnostics{}
	endpointID := int32(d.Get("endpoint_id").(int))

	endpoint, resp, err := client.VhostsAPI.GetVhost(ctx, endpointID).Execute()
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		log.Printf("Endpoint with ID: %d was deleted outside of Terraform. Removing it from Terraform state.", endpointID)
		return nil
	}
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to fetch endpoint with ID %d", endpointID),
			Detail:   err.Error(),
		})
	}

	serviceID := ExtractIdFromLink(endpoint.Links.Service.GetHref())

	service, _, err := client.ServicesAPI.GetServiceWithOperationStatus(ctx, serviceID).Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to fetch service %d for endpoint %d", serviceID, endpointID),
			Detail:   err.Error(),
		})
	}

	endpointType, err := aptible.GetHumanReadableEndpointType(endpoint.GetType())
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get endpoint type",
			Detail:   err.Error(),
		})
	}

	var resourceType string
	var resourceID int32

	// If there's an app link, it's an app service
	if service.Links.App != nil {
		resourceType = "app"
		resourceID = ExtractIdFromLink(service.Links.App.GetHref())
		_ = d.Set("container_port", endpoint.GetContainerPort())
		_ = d.Set("container_ports", endpoint.GetContainerPorts())
		_ = d.Set("platform", endpoint.GetPlatform())
		_ = d.Set("process_type", service.GetProcessType())
	} else {
		resourceType = "database"
		resourceID = ExtractIdFromLink(service.Links.Database.GetHref())
	}

	_ = d.Set("endpoint_id", endpoint.GetId())
	_ = d.Set("endpoint_type", endpointType)
	_ = d.Set("resource_type", resourceType)
	_ = d.Set("ip_filtering", endpoint.GetIpWhitelist())
	_ = d.Set("env_id", ExtractIdFromLink(service.Links.Account.GetHref()))
	_ = d.Set("resource_id", resourceID)
	_ = d.Set("default_domain", endpoint.GetDefault())
	_ = d.Set("managed", endpoint.GetAcme())
	_ = d.Set("domain", endpoint.GetUserDomain())
	_ = d.Set("virtual_domain", endpoint.GetVirtualDomain())
	_ = d.Set("internal", endpoint.GetInternal())
	_ = d.Set("ip_filtering", endpoint.GetIpWhitelist())
	_ = d.Set("platform", endpoint.GetPlatform())
	_ = d.Set("external_hostname", endpoint.GetExternalHost())
	_ = d.Set("shared", endpoint.GetShared())
	_ = d.Set("load_balancing_algorithm_type", endpoint.GetLoadBalancingAlgorithmType())

	for _, c := range endpoint.GetAcmeConfiguration().Challenges {
		// Skip non-DNS challenges
		if c.GetMethod() != "dns01" {
			continue
		}

		// Skip if we're missing critical info
		if c.From == nil || c.To == nil {
			continue
		}

		var toName *string
		for _, to := range c.To {
			if to.GetLegacy() {
				// Skip deprecated challenges
				continue
			}
			toName = to.Name

			if toName != nil {
				// We only need the first valid entry
				break
			}
		}

		if toName == nil {
			return
		}

		_ = d.Set("dns_validation_record", c.From.GetName())
		_ = d.Set("dns_validation_value", toName)
		break
	}

	currentSettingLink, hasCurrentSetting := endpoint.Links.GetCurrentSettingOk()
	if hasCurrentSetting {
		currentSettingID := ExtractIdFromLink(currentSettingLink.GetHref())
		if currentSettingID != 0 {
			currentSetting, _, err := client.SettingsAPI.GetSetting(ctx, currentSettingID).Execute()
			if err == nil {
				applyEndpointSettingsToState(d, currentSetting.GetSettings())
			}
		}
	}

	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.Client
	legacy := m.LegacyClient
	ctx = m.APIContext(ctx)
	diags := diag.Diagnostics{}

	endpointID := int32(d.Get("endpoint_id").(int))

	needsDeploy := false
	settingsMap := map[string]string{}
	attrs := aptibleapi.NewUpdateVhostRequest()

	if d.HasChange("ip_filtering") {
		needsDeploy = true
		ipWhitelist, err := makeStringSlice(d.Get("ip_filtering").([]interface{}))
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Validation Error",
				Detail:   fmt.Sprintf("Failed to parse ip_filtering: %s", err.Error()),
			})
		}
		attrs.SetIpWhitelist(ipWhitelist)
	}

	if d.HasChange("container_port") {
		needsDeploy = true
		attrs.SetContainerPort(int32(d.Get("container_port").(int)))
	}

	if d.HasChange("container_ports") {
		needsDeploy = true
		containerPorts, err := makeInt32Slice(d.Get("container_ports").([]interface{}))
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Validation Error",
				Detail:   fmt.Sprintf("Failed to parse container_ports: %s", err.Error()),
			})
		}
		attrs.SetContainerPorts(containerPorts)
	}

	if d.HasChange("shared") {
		needsDeploy = true
		attrs.SetShared(d.Get("shared").(bool))
	}

	if d.HasChange("load_balancing_algorithm_type") {
		needsDeploy = true
		lbAlgorithmType := d.Get("load_balancing_algorithm_type").(string)
		if lbAlgorithmType != "" {
			attrs.SetLoadBalancingAlgorithmType(lbAlgorithmType)
		}
	}

	_, err := client.VhostsAPI.
		UpdateVhost(ctx, endpointID).
		UpdateVhostRequest(*attrs).
		Execute()
	if err != nil {
		log.Printf("error: %v", err)
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update endpoint",
			Detail:   err.Error(),
		})
	}

	// Bool-like string settings: "true" → "true", "false" or "" → "" to clear
	for attr, apiKey := range map[string]string{
		"force_ssl":                  "FORCE_SSL",
		"strict_health_checks":       "STRICT_HEALTH_CHECKS",
		"show_elb_healthchecks":      "SHOW_ELB_HEALTHCHECKS",
		"disable_weak_cipher_suites": "DISABLE_WEAK_CIPHER_SUITES",
	} {
		if d.HasChange(attr) {
			needsDeploy = true
			if d.Get(attr).(string) == "true" {
				settingsMap[apiKey] = "true"
			} else {
				settingsMap[apiKey] = ""
			}
		}
	}

	// Int settings: non-zero → stringified value, changed to zero → "" to clear
	for attr, apiKey := range map[string]string{
		"idle_timeout":                "IDLE_TIMEOUT",
		"release_healthcheck_timeout": "RELEASE_HEALTHCHECK_TIMEOUT",
	} {
		if d.HasChange(attr) {
			needsDeploy = true
			if v := d.Get(attr).(int); v != 0 {
				settingsMap[apiKey] = strconv.Itoa(v)
			} else {
				settingsMap[apiKey] = ""
			}
		}
	}

	// String settings: non-empty → value, changed to empty → "" to clear
	for attr, apiKey := range map[string]string{
		"maintenance_page_url":   "MAINTENANCE_PAGE_URL",
		"ssl_protocols_override": "SSL_PROTOCOLS_OVERRIDE",
		"ssl_ciphers_override":   "SSL_CIPHERS_OVERRIDE",
	} {
		if d.HasChange(attr) {
			needsDeploy = true
			settingsMap[apiKey] = d.Get(attr).(string)
		}
	}

	if needsDeploy {
		payload := aptibleapi.NewCreateOperationRequest("provision")
		if len(settingsMap) > 0 {
			payload.SetSettings(settingsMap)
		}
		operation, _, err := client.OperationsAPI.
			CreateOperationForVhost(ctx, endpointID).
			CreateOperationRequest(*payload).
			Execute()
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create operation",
				Detail:   err.Error(),
			})
		}

		_, err = legacy.WaitForOperation(int64(operation.Id))
		if err != nil {
			// Do not return here so that the read method can hydrate the state
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to update endpoint %d", endpointID),
				Detail:   err.Error(),
			})
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
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

var validLbAlgorithms = []string{
	"round_robin",
	"least_outstanding_requests",
	"weighted_random",
}

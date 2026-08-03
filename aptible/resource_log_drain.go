package aptible

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLogDrain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLogDrainCreateContext,
		ReadContext:   resourceLogDrainReadContext,
		DeleteContext: resourceLogDrainDeleteContext,
		Importer: &schema.ResourceImporter{
			State: resourceLogDrainImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"log_drain_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"drain_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"database_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"drain_username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"drain_host": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"drain_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"drain_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"logging_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"drain_apps": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"drain_databases": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"drain_ephemeral_sessions": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"drain_proxies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			// aliases
			"token": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"drain_username"},
			},
			"pipeline": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"logging_token"},
			},
			"tags": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"logging_token"},
			},
		},
	}
}

func resourceLogDrainCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.APIClient

	handle := d.Get("handle").(string)
	accountID := int32(d.Get("env_id").(int))
	drainType := d.Get("drain_type").(string)

	drainHost := d.Get("drain_host").(string)
	drainPort := int32(d.Get("drain_port").(int))
	drainUsername := d.Get("drain_username").(string)
	drainPassword := d.Get("drain_password").(string)
	loggingToken := d.Get("logging_token").(string)
	url := d.Get("url").(string)
	databaseID := int32(d.Get("database_id").(int))
	drainApps := d.Get("drain_apps").(bool)
	drainDatabases := d.Get("drain_databases").(bool)
	drainEphemeralSessions := d.Get("drain_ephemeral_sessions").(bool)
	drainProxies := d.Get("drain_proxies").(bool)

	// alias support
	if drainType == "elasticsearch_database" && loggingToken == "" {
		loggingToken = d.Get("pipeline").(string)
	}
	if drainType == "datadog" || drainType == "logdna" {
		if drainUsername == "" {
			drainUsername = d.Get("token").(string)
		}
		if loggingToken == "" {
			loggingToken = d.Get("tags").(string)
		}
	}

	req := aptibleapi.CreateLogDrainRequest{
		Handle:                 handle,
		DrainType:              drainType,
		DrainApps:              &drainApps,
		DrainDatabases:         &drainDatabases,
		DrainEphemeralSessions: &drainEphemeralSessions,
		DrainProxies:           &drainProxies,
	}
	if drainHost != "" {
		req.DrainHost = &drainHost
	}
	if drainPort != 0 {
		req.DrainPort = &drainPort
	}
	if drainUsername != "" {
		req.DrainUsername = &drainUsername
	}
	if drainPassword != "" {
		req.DrainPassword = &drainPassword
	}
	if loggingToken != "" {
		req.LoggingToken = &loggingToken
	}
	if url != "" {
		req.Url = &url
	}
	if databaseID != 0 {
		req.DatabaseId = &databaseID
	}

	logDrain, _, err := client.LogDrainsAPI.CreateLogDrain(ctx, accountID).
		CreateLogDrainRequest(req).Execute()
	if err != nil {
		log.Println("There was an error when completing the request to create the log drain.\n[ERROR] -", err)
		return diag.FromErr(err)
	}

	// Provision the log drain
	op, _, err := client.OperationsAPI.CreateOperationForLogDrain(ctx, logDrain.Id).
		CreateOperationRequest(aptibleapi.CreateOperationRequest{Type: "provision"}).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = m.WaitForOperation(ctx, op.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(int(logDrain.Id)))
	_ = d.Set("log_drain_id", int(logDrain.Id))

	return resourceLogDrainReadContext(ctx, d, meta)
}

func resourceLogDrainReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.APIClient

	logDrainID := int32(d.Get("log_drain_id").(int))
	log.Println("Getting log drain with ID: " + strconv.Itoa(int(logDrainID)))

	logDrain, resp, err := client.LogDrainsAPI.GetLogDrain(ctx, logDrainID).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		log.Println(err)
		return diag.FromErr(err)
	}

	_ = d.Set("log_drain_id", int(logDrain.Id))
	_ = d.Set("handle", logDrain.Handle)
	_ = d.Set("drain_type", logDrain.DrainType)
	_ = d.Set("url", logDrain.GetUrl())
	_ = d.Set("logging_token", logDrain.GetLoggingToken())
	_ = d.Set("drain_port", int(logDrain.DrainPort))
	_ = d.Set("drain_username", logDrain.GetDrainUsername())
	_ = d.Set("drain_password", logDrain.GetDrainPassword())
	_ = d.Set("drain_host", logDrain.DrainHost)
	_ = d.Set("drain_proxies", logDrain.DrainProxies)
	_ = d.Set("drain_ephemeral_sessions", logDrain.DrainEphemeralSessions)
	_ = d.Set("drain_databases", logDrain.DrainDatabases)
	_ = d.Set("drain_apps", logDrain.DrainApps)

	if logDrain.Links != nil {
		if logDrain.Links.Account != nil && logDrain.Links.Account.Href != nil {
			_ = d.Set("env_id", int(ExtractIdFromLink(*logDrain.Links.Account.Href)))
		}
		if logDrain.Links.Database != nil && logDrain.Links.Database.Href != nil {
			_ = d.Set("database_id", int(ExtractIdFromLink(*logDrain.Links.Database.Href)))
		}
	}

	// alias support
	if logDrain.DrainType == "elasticsearch_database" {
		_ = d.Set("pipeline", logDrain.GetLoggingToken())
	}
	if logDrain.DrainType == "datadog" || logDrain.DrainType == "logdna" {
		_ = d.Set("token", logDrain.GetDrainUsername())
		_ = d.Set("tags", logDrain.GetLoggingToken())
	}

	return nil
}

func resourceLogDrainDeleteContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)

	if diags := resourceLogDrainReadContext(ctx, d, meta); !diags.HasError() {
		logDrainID := int32(d.Get("log_drain_id").(int))
		deleted, err := m.DeleteLogDrain(ctx, logDrainID)
		if deleted {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the log drain.\n[ERROR] -", err)
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}

func resourceLogDrainImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logDrainID, _ := strconv.Atoi(d.Id())
	_ = d.Set("log_drain_id", logDrainID)
	err := diagnosticsToError(resourceLogDrainReadContext(context.Background(), d, meta))
	return []*schema.ResourceData{d}, err
}

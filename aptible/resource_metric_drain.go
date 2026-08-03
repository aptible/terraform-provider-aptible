package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMetricDrain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMetricDrainCreate,
		ReadContext:   resourceMetricDrainRead,
		DeleteContext: resourceMetricDrainDelete,
		CustomizeDiff: resourceMetricDrainValidate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMetricDrainImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"metric_drain_id": {
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(validMetricDrainTypes, false),
			},
			"database_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateURL,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"database": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"api_key": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"series_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateURL,
			},
		},
	}
}

var validMetricDrainTypes = []string{"influxdb_database", "influxdb", "influxdb2", "datadog"}

var metricDrainAttrs = map[string]ResourceAttrs{
	"influxdb_database": {
		Required:   []string{"database_id"},
		NotAllowed: []string{"url", "username", "password", "database", "api_key", "series_url"},
	},
	"influxdb": {
		Required:   []string{"url", "username", "password", "database"},
		NotAllowed: []string{"database_id"},
	},
	"influxdb2": {
		Required:   []string{"url", "api_key", "bucket", "organization"},
		NotAllowed: []string{"database_id", "username", "password"},
	},
	"datadog": {
		Required:   []string{"api_key"},
		NotAllowed: []string{"database_id"},
	},
}

func resourceMetricDrainValidate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	d := ResourceDiff{ResourceDiff: diff}
	drainType := d.Get("drain_type").(string)
	var err error

	allowedAttrs, ok := metricDrainAttrs[drainType]
	if !ok {
		return fmt.Errorf("error during validation: drain_type %q not found", drainType)
	}

	for _, attr := range allowedAttrs.Required {
		if !d.HasRequired(attr) {
			err = multierror.Append(err, fmt.Errorf("%q is required when drain_type = %q", attr, drainType))
		}
	}
	for _, attr := range allowedAttrs.NotAllowed {
		if d.HasOptional(attr) {
			err = multierror.Append(err, fmt.Errorf("%q is not allowed when drain_type = %q", attr, drainType))
		}
	}

	return err
}

func resourceMetricDrainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.APIClient

	handle := d.Get("handle").(string)
	accountID := int32(d.Get("env_id").(int))
	drainType := d.Get("drain_type").(string)

	req := aptibleapi.CreateMetricDrainRequest{
		Handle:    handle,
		DrainType: drainType,
	}

	if databaseID := int32(d.Get("database_id").(int)); databaseID != 0 {
		req.DatabaseId = &databaseID
	}

	// influxdb_database drains don't have a DrainConfiguration
	if drainType != "influxdb_database" {
		url := d.Get("url").(string)
		username := d.Get("username").(string)
		password := d.Get("password").(string)
		database := d.Get("database").(string)
		apiKey := d.Get("api_key").(string)
		seriesURL := d.Get("series_url").(string)
		bucket := d.Get("bucket").(string)
		org := d.Get("organization").(string)

		config := &aptibleapi.CreateMetricDrainRequestDrainConfiguration{}
		if url != "" {
			config.Address = &url
		}
		if username != "" {
			config.Username = &username
		}
		if password != "" {
			config.Password = &password
		}
		if database != "" {
			config.Database = &database
		}
		if apiKey != "" {
			config.ApiKey = &apiKey
			config.AuthToken = &apiKey
		}
		if seriesURL != "" {
			config.SeriesUrl = &seriesURL
		}
		if bucket != "" {
			config.Bucket = &bucket
		}
		if org != "" {
			config.Org = &org
		}
		req.DrainConfiguration = config
	}

	metricDrain, _, err := client.MetricDrainsAPI.CreateMetricDrain(ctx, accountID).
		CreateMetricDrainRequest(req).Execute()
	if err != nil {
		log.Println("There was an error when completing the request to create the metric drain.\n[ERROR] -", err)
		return diag.FromErr(err)
	}

	// Provision the metric drain
	op, _, err := client.OperationsAPI.CreateOperationForMetricDrain(ctx, metricDrain.Id).
		CreateOperationRequest(aptibleapi.CreateOperationRequest{Type: "provision"}).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = m.WaitForOperation(ctx, op.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(int(metricDrain.Id)))
	_ = d.Set("metric_drain_id", int(metricDrain.Id))

	return resourceMetricDrainRead(ctx, d, meta)
}

func resourceMetricDrainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.APIClient

	metricDrainID := int32(d.Get("metric_drain_id").(int))
	log.Println("Getting metric drain with ID: " + strconv.Itoa(int(metricDrainID)))

	metricDrain, resp, err := client.MetricDrainsAPI.GetMetricDrain(ctx, metricDrainID).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		log.Println(err)
		return diag.FromErr(err)
	}

	_ = d.Set("metric_drain_id", int(metricDrain.Id))
	_ = d.Set("handle", metricDrain.Handle)
	_ = d.Set("drain_type", metricDrain.DrainType)

	if metricDrain.Links != nil {
		if metricDrain.Links.Account != nil && metricDrain.Links.Account.Href != nil {
			_ = d.Set("env_id", int(ExtractIdFromLink(*metricDrain.Links.Account.Href)))
		}
		if metricDrain.Links.Database != nil && metricDrain.Links.Database.Href != nil {
			_ = d.Set("database_id", int(ExtractIdFromLink(*metricDrain.Links.Database.Href)))
		}
	}

	if metricDrain.DrainConfiguration != nil {
		config := metricDrain.DrainConfiguration
		_ = d.Set("url", config.GetAddress())
		_ = d.Set("username", config.GetUsername())
		_ = d.Set("password", config.GetPassword())
		_ = d.Set("database", config.GetDatabase())
		if apiKey := config.GetApiKey(); apiKey != "" {
			_ = d.Set("api_key", apiKey)
		}
		_ = d.Set("series_url", config.GetSeriesUrl())

		if config.AdditionalProperties != nil {
			if authToken, ok := config.AdditionalProperties["authToken"].(string); ok && authToken != "" {
				_ = d.Set("api_key", authToken)
			}
			if bucket, ok := config.AdditionalProperties["bucket"].(string); ok {
				_ = d.Set("bucket", bucket)
			}
			if org, ok := config.AdditionalProperties["org"].(string); ok {
				_ = d.Set("organization", org)
			}
		}
	}

	return nil
}

func resourceMetricDrainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)

	readDiags := resourceMetricDrainRead(ctx, d, meta)
	if !readDiags.HasError() {
		metricDrainID := int32(d.Get("metric_drain_id").(int))
		deleted, err := m.DeleteMetricDrain(ctx, metricDrainID)
		if deleted {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the metric drain.\n[ERROR] -", err)
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}

func resourceMetricDrainImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	metricDrainID, _ := strconv.Atoi(d.Id())
	_ = d.Set("metric_drain_id", metricDrainID)
	if err := diagnosticsToError(resourceMetricDrainRead(ctx, d, meta)); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

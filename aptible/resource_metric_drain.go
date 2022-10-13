package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/go-openapi/strfmt"
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

var validMetricDrainTypes = []string{"influxdb_database", "influxdb", "datadog"}

var metricDrainAttrs = map[string]ResourceAttrs{
	"influxdb_database": {
		Required:   []string{"database_id"},
		NotAllowed: []string{"url", "username", "password", "database", "api_key", "series_url"},
	},
	"influxdb": {
		Required:   []string{"url", "username", "password", "database"},
		NotAllowed: []string{"database_id"},
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
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)
	accountID := int64(d.Get("env_id").(int))
	data := &aptible.MetricDrainCreateAttrs{
		DrainType:  d.Get("drain_type").(string),
		DatabaseID: int64(d.Get("database_id").(int)),
		URL:        strfmt.URI(d.Get("url").(string)),
		Username:   d.Get("username").(string),
		Password:   d.Get("password").(string),
		Database:   d.Get("database").(string),
		APIKey:     d.Get("api_key").(string),
		SeriesURL:  strfmt.URI(d.Get("series_url").(string)),
	}

	metricDrain, err := client.CreateMetricDrain(handle, accountID, data)
	if err != nil {
		log.Println("There was an error when completing the request to create the metric drain.\n[ERROR] -", err)
		return generateDiagnosticsFromClientError(err)
	}
	d.SetId(strconv.Itoa(int(metricDrain.ID)))
	_ = d.Set("metric_drain_id", metricDrain.ID)

	return resourceMetricDrainRead(ctx, d, meta)
}

func resourceMetricDrainRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	metricDrainID := int64(d.Get("metric_drain_id").(int))

	log.Println("Getting metric drain with ID: " + strconv.Itoa(int(metricDrainID)))

	metricDrain, err := client.GetMetricDrain(metricDrainID)
	if err != nil {
		log.Println(err)
		return generateDiagnosticsFromClientError(err)
	}
	if metricDrain.Deleted {
		d.SetId("")
		return nil
	}
	_ = d.Set("metric_drain_id", int(metricDrain.ID))
	_ = d.Set("env_id", metricDrain.AccountID)
	_ = d.Set("handle", metricDrain.Handle)
	_ = d.Set("drain_type", metricDrain.DrainType)
	_ = d.Set("database_id", metricDrain.DatabaseID)
	_ = d.Set("url", metricDrain.URL)
	_ = d.Set("username", metricDrain.Username)
	_ = d.Set("password", metricDrain.Password)
	_ = d.Set("database", metricDrain.Database)
	_ = d.Set("api_key", metricDrain.APIKey)
	_ = d.Set("series_url", metricDrain.SeriesURL)

	return nil
}

func resourceMetricDrainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	readDiags := resourceMetricDrainRead(ctx, d, meta)
	if !readDiags.HasError() {
		metricDrainID := int64(d.Get("metric_drain_id").(int))
		client := meta.(*aptible.Client)
		deleted, err := client.DeleteMetricDrain(metricDrainID)
		if deleted {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the metric drain.\n[ERROR] -", err)
			return generateDiagnosticsFromClientError(err)
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

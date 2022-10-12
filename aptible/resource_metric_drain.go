package aptible

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMetricDrain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMetricDrainCreate,
		Read:          resourceMetricDrainRead,
		Delete:        resourceMetricDrainDelete,
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
				ValidateFunc: validation.StringInSlice([]string{"influxdb_database", "influxdb", "datadog"}, false),
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
				ValidateFunc: ValidateURL,
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
				ValidateFunc: ValidateURL,
			},
		},
	}
}

var resourceMetricDrainAttrs = map[string]ResourceAttrs{
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

	allowedAttrs := resourceMetricDrainAttrs[drainType]

	for _, attr := range allowedAttrs.Required {
		if !d.HasRequired(attr) {
			err = multierror.Append(err, fmt.Errorf("%s is required when drain_type = \"%s\"", attr, drainType))
		}
	}
	for _, attr := range allowedAttrs.NotAllowed {
		if d.HasOptional(attr) {
			err = multierror.Append(err, fmt.Errorf("%s is not allowed when drain_type = \"%s\"", attr, drainType))
		}
	}

	return err
}

func resourceMetricDrainCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if err := resourceMetricDrainRead(d, meta); err != nil {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Error(),
		}}
	}

	return diag.Diagnostics{}
}

func resourceMetricDrainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	metricDrainID := int64(d.Get("metric_drain_id").(int))

	log.Println("Getting metric drain with ID: " + strconv.Itoa(int(metricDrainID)))

	metricDrain, err := client.GetMetricDrain(metricDrainID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
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

func resourceMetricDrainDelete(d *schema.ResourceData, meta interface{}) error {
	readErr := resourceMetricDrainRead(d, meta)
	if readErr == nil {
		metricDrainID := int64(d.Get("metric_drain_id").(int))
		client := meta.(*aptible.Client)
		deleted, err := client.DeleteMetricDrain(metricDrainID)
		if deleted {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the metric drain.\n[ERROR] -", err)
			return generateErrorFromClientError(err)
		}
	}
	d.SetId("")
	return nil
}

func resourceMetricDrainImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	metricDrainID, _ := strconv.Atoi(d.Id())
	_ = d.Set("metric_drain_id", metricDrainID)
	err := resourceMetricDrainRead(d, meta)
	return []*schema.ResourceData{d}, err
}

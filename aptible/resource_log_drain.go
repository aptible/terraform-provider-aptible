package aptible

import (
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLogDrain() *schema.Resource {
	return &schema.Resource{
		Create: resourceLogDrainCreate,
		Read:   resourceLogDrainRead,
		Delete: resourceLogDrainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceLogDrainImport,
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
				Computed:  true, // The API generates a password if one isn't provided
				ForceNew:  true,
				Sensitive: true,
			},
			"drain_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"logging_token": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"pipeline": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLogDrainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)
	accountID := int64(d.Get("env_id").(int))
	drainType := d.Get("drain_type").(string)
	data := &aptible.LogDrainCreateAttrs{
		DrainType:              &drainType,
		URL:                    strfmt.URI(d.Get("url").(string)),
		LoggingToken:           d.Get("logging_token").(string),
		DrainPort:              int64(d.Get("drain_port").(int)),
		DrainPassword:          d.Get("drain_password").(string),
		DrainHost:              strfmt.URI(d.Get("drain_host").(string)),
		DrainUsername:          d.Get("drain_username").(string),
		DatabaseID:             int64(d.Get("database_id").(int)),
		DrainProxies:           d.Get("drain_proxies").(bool),
		DrainEphemeralSessions: d.Get("drain_ephemeral_sessions").(bool),
		DrainDatabases:         d.Get("drain_databases").(bool),
		DrainApps:              d.Get("drain_apps").(bool),
	}

	// alias support
	if drainType == "elasticsearch_database" && data.LoggingToken == "" {
		data.LoggingToken = d.Get("pipeline").(string)
	}

	if drainType == "datadog" || drainType == "logdna" {
		if data.DrainUsername == "" {
			data.DrainUsername = d.Get("token").(string)
		}

		if data.LoggingToken == "" {
			data.LoggingToken = d.Get("tags").(string)
		}
	}

	logDrain, err := client.CreateLogDrain(handle, accountID, data)
	if err != nil {
		log.Println("There was an error when completing the request to create the log drain.\n[ERROR] -", err)
		return generateErrorFromClientError(err)
	}
	d.SetId(strconv.Itoa(int(logDrain.ID)))
	_ = d.Set("log_drain_id", logDrain.ID)
	// The API generates a password if one isn't provided so we need to set it after creation
	_ = d.Set("drain_password", logDrain.DrainPassword)
	// alias support
	_ = d.Set("drain_username", logDrain.DrainUsername)
	_ = d.Set("logging_token", logDrain.LoggingToken)

	return resourceLogDrainRead(d, meta)
}

func resourceLogDrainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	logDrainID := int64(d.Get("log_drain_id").(int))

	log.Println("Getting log drain with ID: " + strconv.Itoa(int(logDrainID)))

	logDrain, err := client.GetLogDrain(logDrainID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}
	if logDrain.Deleted {
		d.SetId("")
		return nil
	}
	_ = d.Set("log_drain_id", int(logDrain.ID))
	_ = d.Set("handle", logDrain.Handle)
	_ = d.Set("drain_type", logDrain.DrainType)
	_ = d.Set("url", logDrain.URL)
	_ = d.Set("logging_token", logDrain.LoggingToken)
	_ = d.Set("drain_port", logDrain.DrainPort)
	_ = d.Set("drain_username", logDrain.DrainUsername)
	_ = d.Set("drain_password", logDrain.DrainPassword)
	_ = d.Set("drain_host", logDrain.DrainHost)
	_ = d.Set("drain_proxies", logDrain.DrainProxies)
	_ = d.Set("drain_ephemeral_sessions", logDrain.DrainEphemeralSessions)
	_ = d.Set("drain_databases", logDrain.DrainDatabases)
	_ = d.Set("drain_apps", logDrain.DrainApps)
	_ = d.Set("env_id", logDrain.AccountID)
	_ = d.Set("database_id", logDrain.DatabaseID)
	_ = d.Set("token", logDrain.DrainUsername)
	_ = d.Set("tags", logDrain.LoggingToken)
	_ = d.Set("pipeline", logDrain.LoggingToken)

	return nil
}

func resourceLogDrainDelete(d *schema.ResourceData, meta interface{}) error {
	readErr := resourceLogDrainRead(d, meta)
	if readErr == nil {
		logDrainID := int64(d.Get("log_drain_id").(int))
		client := meta.(*aptible.Client)
		deleted, err := client.DeleteLogDrain(logDrainID)
		if deleted {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the app.\n[ERROR] -", err)
			return generateErrorFromClientError(err)
		}
	}
	d.SetId("")
	return nil
}

func resourceLogDrainImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logDrainID, _ := strconv.Atoi(d.Id())
	_ = d.Set("log_drain_id", logDrainID)
	err := resourceLogDrainRead(d, meta)
	return []*schema.ResourceData{d}, err
}

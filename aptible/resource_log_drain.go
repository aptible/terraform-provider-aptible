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
			"drain_host": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"drain_password": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"drain_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"logging_token": {
				Type:     schema.TypeString,
				Optional: true,
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
		DatabaseID:             int64(d.Get("database_id").(int)),
		DrainProxies:           d.Get("drain_proxies").(bool),
		DrainEphemeralSessions: d.Get("drain_ephemeral_sessions").(bool),
		DrainDatabases:         d.Get("drain_databases").(bool),
		DrainApps:              d.Get("drain_apps").(bool),
	}

	logDrain, err := client.CreateLogDrain(handle, accountID, data)
	if err != nil {
		log.Println("There was an error when completing the request to create the log drain.\n[ERROR] -", err)
		return err
	}
	d.SetId(strconv.Itoa(int(logDrain.ID)))
	_ = d.Set("log_drain_id", logDrain.ID)

	return resourceLogDrainRead(d, meta)
}

func resourceLogDrainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	logDrainID := int64(d.Get("log_drain_id").(int))

	log.Println("Getting log drain with ID: " + strconv.Itoa(int(logDrainID)))

	logDrain, err := client.GetLogDrain(logDrainID)
	if err != nil {
		log.Println(err)
		return err
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
	_ = d.Set("drain_host", logDrain.DrainHost)
	_ = d.Set("drain_proxies", logDrain.DrainProxies)
	_ = d.Set("drain_ephemeral_sessions", logDrain.DrainEphemeralSessions)
	_ = d.Set("drain_databases", logDrain.DrainDatabases)
	_ = d.Set("drain_apps", logDrain.DrainApps)
	_ = d.Set("account_id", logDrain.AccountID)
	_ = d.Set("database_id", logDrain.DatabaseID)

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
			return err
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

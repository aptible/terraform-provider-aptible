package aptible

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: resourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"env_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)
	id, err := client.GetEnvironmentIDFromHandle(handle)
	if err != nil {
		return err
	}

	_ = d.Set("env_id", id)
	d.SetId("handle")
	return nil
}

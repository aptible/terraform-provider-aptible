package aptible

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/aptible/go-deploy/aptible"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: resourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"handle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"env_id": &schema.Schema{
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

	d.Set("env_id", id)
	d.SetId("handle")
	return nil
}

package aptible

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEnvironmentRead,
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

func dataSourceEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)
	id, err := client.GetEnvironmentIDFromHandle(handle)
	if err != nil {
		return generateErrorFromClientError(err)
	}

	_ = d.Set("env_id", id)
	d.SetId("handle")
	return nil
}

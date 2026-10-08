package aptible

import (
	"context"

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
	m := meta.(*client)
	client := m.APIClient
	ctx := context.Background()

	handle := d.Get("handle").(string)
	account, _, err := client.AccountsAPI.GetAccountByHandle(ctx).Handle(handle).Execute()
	if err != nil {
		return err
	}

	_ = d.Set("env_id", int(account.Id))
	d.SetId("handle")
	return nil
}

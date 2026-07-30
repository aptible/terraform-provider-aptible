package aptible

import (
	"context"

	"github.com/aptible/aptible-api-go/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stack_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceStackRead(d *schema.ResourceData, meta interface{}) error {
	m := meta.(*providerMetadata)
	client := m.Client
	ctx := m.APIContext(context.Background())

	name := d.Get("name").(string)
	stack, err := helpers.GetStackByName(ctx, client, name)
	if err != nil {
		return err
	}

	_ = d.Set("stack_id", int(stack.Id))
	_ = d.Set("org_id", helpers.GetOrgIDFromStackLinks(stack))
	d.SetId(name)
	return nil
}

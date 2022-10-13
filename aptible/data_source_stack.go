package aptible

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackRead,
		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceStackRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	stack_id := d.Get("stack_id").(int64)
	stack, err := client.GetStack(stack_id)
	if err != nil {
		return generateErrorFromClientError(err)
	}

	_ = d.Set("org_id": stack.OrganizationID)
	_ = d.Set("name": stack.Name)
	return nil
}

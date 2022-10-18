package aptible

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEnvironmentImport,
		},
		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"org_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"stack_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)
	orgID := d.Get("org_id").(string)
	stackID := int64(d.Get("stack_id").(int))

	data := aptible.EnvironmentCreateAttrs{
		Handle: handle,
	}

	environment, err := client.CreateEnvironment(orgID, stackID, data)
	if err != nil {
		log.Println("There was an error when completing the request to create the environment.\n[ERROR] -", err)
		return generateDiagnosticsFromClientError(err)
	}

	d.SetId(strconv.Itoa(int(environment.ID)))
	_ = d.Set("env_id", environment.ID)

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	envID := int64(d.Get("env_id").(int))

	log.Println("Getting environment with ID: " + strconv.Itoa(int(envID)))

	environment, err := client.GetEnvironment(envID)
	if err != nil {
		log.Println(err)
		return generateDiagnosticsFromClientError(err)
	}
	if environment.Deleted {
		d.SetId("")
		return nil
	}

	_ = d.Set("handle", environment.Handle)
	_ = d.Set("env_id", int(environment.ID))
	_ = d.Set("stack_id", environment.StackID)
	_ = d.Set("org_id", environment.OrganizationID)
	return nil
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)
	envId := int64(d.Get("env_id").(int))
	environmentUpdates := aptible.EnvironmentUpdates{
		Handle: handle,
	}

	if err := client.UpdateEnvironment(envId, environmentUpdates); err != nil {
		log.Println("There was an error when completing the request to update the environment.\n[ERROR] -", err)
		return generateDiagnosticsFromClientError(err)
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	readDiags := resourceEnvironmentRead(ctx, d, meta)
	if !readDiags.HasError() {
		envID := int64(d.Get("env_id").(int))
		client := meta.(*aptible.Client)
		err := client.DeleteEnvironment(envID)
		if err == nil {
			d.SetId("")
			return nil
		}
		if err != nil {
			log.Println("There was an error when completing the request to destroy the environment.\n[ERROR] -", err)
			return generateDiagnosticsFromClientError(err)
		}
	}
	d.SetId("")
	return nil
}

func resourceEnvironmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	envID, _ := strconv.Atoi(d.Id())
	_ = d.Set("env_id", envID)
	if err := diagnosticsToError(resourceEnvironmentRead(ctx, d, meta)); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

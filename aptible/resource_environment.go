package aptible

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Computed:     true,
				Optional:     true,
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
	client := meta.(*providerMetadata).LegacyClient
	handle := d.Get("handle").(string)
	stackID := int64(d.Get("stack_id").(int))

	// there are a few  scenarios where org_id can be fully gathered from the data provided above
	// 1. If it is provided explicitly
	// 2. It is using a dedicated stack, and so when creating an environment, use the stack id to get stack
	//    and check that stack's org id and use it
	// 3. The user only belongs to one organization, so fall back to that. If they belong to multiple organizations
	// 	  they will have to specify it explicitly at this point (if these points no longer work for inference.)
	orgID := d.Get("org_id").(string) // scenario #1 outlined above
	if orgID == "" {
		stack, err := client.GetStack(stackID) // scenario #2 outlined above
		if err != nil {
			log.Println("There was an error trying to retrieve the stack with the stack id provided to determine"+
				"an organization id.\n[ERROR] - ", err)
			return generateDiagnosticsFromClientError(err)
		}
		orgID = stack.OrganizationID
		if orgID == "" {
			org, err := client.GetOrganization() // scenario #3 outlined above
			if err != nil {
				log.Println("There was an error trying to retrieve an organization id (org_id). You can "+
					"either specify it explicitly or review the error message to attempt to fix the issue. "+
					"\n[ERROR] - ", err)
				return generateDiagnosticsFromClientError(err)
			}
			orgID = org.ID
		}
	}

	if orgID == "" {
		errorMessage := "[ERROR] - Unable to infer organization ID from stack or user. You may have to specify it explicitly"
		log.Println(errorMessage)
		return generateDiagnosticsFromClientError(errors.New(errorMessage))
	}

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
	client := meta.(*providerMetadata).LegacyClient
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
	client := meta.(*providerMetadata).LegacyClient
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
		client := meta.(*providerMetadata).LegacyClient
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

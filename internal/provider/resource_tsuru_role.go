package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruRoleCreate,
		ReadContext:   resourceTsuruRoleRead,
		UpdateContext: resourceTsuruRoleUpdate,
		DeleteContext: resourceTsuruRoleDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name in Tsuru.",
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Context type for the role (e.g., 'app' or 'team').",
			},
			"permissions": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of permissions for this role.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Role description.",
			},
		},
	}
}

func resourceTsuruRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("name").(string))
	return resourceTsuruRoleRead(ctx, d, m)
}

func resourceTsuruRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTsuruRoleRead(ctx, d, m)
}

func resourceTsuruRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

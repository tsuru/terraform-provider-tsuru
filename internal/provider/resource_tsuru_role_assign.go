package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruRoleAssign() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruRoleAssignCreate,
		ReadContext:   resourceTsuruRoleAssignRead,
		DeleteContext: resourceTsuruRoleAssignDelete,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User email to assign the role.",
			},
			"role_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the role being assigned.",
			},
			"context_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of context (e.g., 'app' or 'team').",
			},
			"context_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value for the context (e.g., app name or team name).",
			},
		},
	}
}

func resourceTsuruRoleAssignCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("email").(string) + ":" + d.Get("role_name").(string))
	return resourceTsuruRoleAssignRead(ctx, d, m)
}

func resourceTsuruRoleAssignRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruRoleAssignDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

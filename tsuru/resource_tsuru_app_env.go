// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	PublicVariable  = "public"
	PrivateVariable = "private"
)

func resourceTsuruApplicationEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Environment",
		CreateContext: resourceTsuruApplicationEnvironmentCreate,
		ReadContext:   resourceTsuruApplicationEnvironmentRead,
		UpdateContext: resourceTsuruApplicationEnvironmentUpdate,
		DeleteContext: resourceTsuruApplicationEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"environment_variable": {
				Type:        schema.TypeList,
				Description: "Environment variables",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Variable name",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "Variable value",
							Optional:    true,
						},
						"sensitive_value": {
							Type:        schema.TypeString,
							Description: "Sensitive variable value",
							Sensitive:   true,
							Optional:    true,
						},
					},
				},
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "restart app after applying",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceTsuruApplicationEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruApplicationEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return nil
}

func resourceTsuruApplicationEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTsuruApplicationEnvironmentRead(ctx, d, meta)
}

func resourceTsuruApplicationEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

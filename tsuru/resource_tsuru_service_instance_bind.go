// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruServiceInstanceBind() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Service Instance Bind",
		CreateContext: resourceTsuruServiceInstanceBindCreate,
		ReadContext:   resourceTsuruServiceInstanceBindRead,
		UpdateContext: resourceTsuruServiceInstanceBindUpdate,
		DeleteContext: resourceTsuruServiceInstanceBindDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service kind",
			},
			"service_instance": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service instance",
			},
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
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

func resourceTsuruServiceInstanceBindCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruServiceInstanceBindRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return nil
}

func resourceTsuruServiceInstanceBindUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTsuruServiceInstanceBindRead(ctx, d, meta)
}

func resourceTsuruServiceInstanceBindDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

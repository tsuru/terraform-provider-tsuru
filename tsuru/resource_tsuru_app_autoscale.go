// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruApplicationAutoscale() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Autoscale",
		CreateContext: resourceTsuruApplicationAutoscaleCreate,
		ReadContext:   resourceTsuruApplicationAutoscaleRead,
		UpdateContext: resourceTsuruApplicationAutoscaleUpdate,
		DeleteContext: resourceTsuruApplicationAutoscaleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"process": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service instance",
			},
			"min_units": {
				Type:        schema.TypeInt,
				Description: "minimum number of units",
				Required:    true,
				Default:     true,
			},
			"max_units": {
				Type:        schema.TypeInt,
				Description: "maximum number of units",
				Required:    true,
				Default:     true,
			},
			"cpu_average": {
				Type:        schema.TypeInt,
				Description: "cpu average",
				Required:    true,
				Default:     true,
			},
		},
	}
}

func resourceTsuruApplicationAutoscaleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruApplicationAutoscaleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return nil
}

func resourceTsuruApplicationAutoscaleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTsuruApplicationAutoscaleRead(ctx, d, meta)
}

func resourceTsuruApplicationAutoscaleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

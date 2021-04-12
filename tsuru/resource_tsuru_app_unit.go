// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruApplicationUnits() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Units",
		CreateContext: resourceTsuruApplicationUnitsCreate,
		ReadContext:   resourceTsuruApplicationUnitsRead,
		UpdateContext: resourceTsuruApplicationUnitsUpdate,
		DeleteContext: resourceTsuruApplicationUnitsDelete,
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
				Description: "Process name",
				Required:    true,
			},
			"units_count": {
				Type:        schema.TypeInt,
				Description: "Units count",
				Required:    true,
			},
		},
	}
}

func resourceTsuruApplicationUnitsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruApplicationUnitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return nil
}

func resourceTsuruApplicationUnitsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTsuruApplicationUnitsRead(ctx, d, meta)
}

func resourceTsuruApplicationUnitsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruVolume() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Service Volume",
		CreateContext: resourceTsuruVolumeCreate,
		ReadContext:   resourceTsuruVolumeRead,
		UpdateContext: resourceTsuruVolumeUpdate,
		DeleteContext: resourceTsuruVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Volume name",
			},
			"plan": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Team owner of this volume",
			},
			"pool": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Volume Pool",
			},
			"options": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Volume additional options",
			},
		},
	}
}

func resourceTsuruVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return nil
}

func resourceTsuruVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTsuruVolumeRead(ctx, d, meta)
}

func resourceTsuruVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

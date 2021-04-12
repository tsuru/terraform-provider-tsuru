// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruVolumeBind() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Service Volume Bind",
		CreateContext: resourceTsuruVolumeBindCreate,
		ReadContext:   resourceTsuruVolumeBindRead,
		UpdateContext: resourceTsuruVolumeBindUpdate,
		DeleteContext: resourceTsuruVolumeBindDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"volume_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service kind",
			},
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"mount_point": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service instance",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Description: "restart app after applying",
				Optional:    true,
				Default:     true,
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

func resourceTsuruVolumeBindCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruVolumeBindRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	return nil
}

func resourceTsuruVolumeBindUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceTsuruVolumeBindRead(ctx, d, meta)
}

func resourceTsuruVolumeBindDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

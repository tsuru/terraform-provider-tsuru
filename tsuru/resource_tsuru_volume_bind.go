// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
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
				Default:     false,
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
	provider := meta.(*tsuruProvider)

	name := d.Get("volume_name").(string)

	bindData := tsuru_client.VolumeBindData{
		App:        d.Get("app").(string),
		Mountpoint: d.Get("mount_point").(string),
		Readonly:   false,
		Norestart:  false,
	}

	if roi, ok := d.GetOk("read_only"); ok {
		ro := roi.(bool)
		if ro {
			bindData.Readonly = true
		}
	}

	if ri, ok := d.GetOk("restart_on_update"); ok {
		r := ri.(bool)
		if !r {
			bindData.Norestart = true
		}
	}

	provider.TsuruClient.VolumeApi.VolumeBind(ctx, name, bindData)

	return resourceTsuruVolumeBindRead(ctx, d, meta)
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

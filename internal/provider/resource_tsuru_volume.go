// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruVolume() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Service Volume",
		CreateContext: resourceTsuruVolumeCreate,
		ReadContext:   resourceTsuruVolumeRead,
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
				ForceNew: true,
				Required: true,
			},
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Team owner of this volume",
			},
			"pool": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Volume Pool",
			},
			"options": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				ForceNew:    true,
				Description: "Volume additional options",
			},
		},
	}
}

func resourceTsuruVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	volume := tsuru_client.Volume{
		Name:      d.Get("name").(string),
		TeamOwner: d.Get("owner").(string),
		Pool:      d.Get("pool").(string),
		Plan: tsuru_client.VolumePlan{
			Name: d.Get("plan").(string),
		},
	}

	options := map[string]string{}
	if o, ok := d.GetOk("options"); ok {
		opts := o.(map[string]interface{})
		for key, value := range opts {
			options[key] = value.(string)
		}
	}

	if len(options) > 0 {
		volume.Opts = options
	}

	_, err := provider.TsuruClient.VolumeApi.VolumeCreate(ctx, volume)
	if err != nil {
		return diag.Errorf("Unable to create volume: %v", err)
	}

	d.SetId(volume.Name)

	return resourceTsuruVolumeRead(ctx, d, meta)
}

func resourceTsuruVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()

	volume, _, err := provider.TsuruClient.VolumeApi.VolumeGet(ctx, name)
	if err != nil {
		return diag.Errorf("Unable to read volume: %v", err)
	}

	d.Set("name", volume.Name)
	d.Set("plan", volume.Plan.Name)
	d.Set("pool", volume.Pool)
	d.Set("owner", volume.TeamOwner)
	d.Set("options", volume.Opts)

	return nil
}

func resourceTsuruVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("name").(string)

	_, err := provider.TsuruClient.VolumeApi.VolumeDelete(ctx, name)
	if err != nil {
		return diag.Errorf("Unable to delete volume: %v", err)
	}

	return nil
}

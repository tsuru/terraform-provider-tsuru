// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruVolumeBind() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Service Volume Bind",
		CreateContext: resourceTsuruVolumeBindCreate,
		ReadContext:   resourceTsuruVolumeBindRead,
		DeleteContext: resourceTsuruVolumeBindDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"volume": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service kind",
			},
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
				ForceNew:    true,
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
				ForceNew:    true,
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "restart app after applying",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTsuruVolumeBindCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("volume").(string)

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

	ri := d.Get("restart_on_update").(bool)
	if !ri {
		bindData.Norestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := provider.TsuruClient.VolumeApi.VolumeBind(ctx, name, bindData)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
		}

		d.SetId(createID([]string{bindData.App, name, bindData.Mountpoint}))
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTsuruVolumeBindRead(ctx, d, meta)
}

func resourceTsuruVolumeBindRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 3)
	if err != nil {
		return diag.FromErr(err)
	}
	app := parts[0]
	name := parts[1]
	mountPath := parts[2]

	volume, resp, err := provider.TsuruClient.VolumeApi.VolumeGet(ctx, name)
	if err != nil {
		return diag.Errorf("unable to read volume info: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to read volume info, error code: %d", resp.StatusCode)
	}

	for _, bind := range volume.Binds {
		if bind.Id.App != app && bind.Id.Mountpoint != mountPath {
			continue
		}
		d.Set("volume", name)
		d.Set("app", app)
		d.Set("mount_point", bind.Id.Mountpoint)
		d.Set("read_only", bind.Readonly)
		return nil
	}

	return diag.Errorf("unable to find bind for volume %s to %s", volume.Name, app)
}

func resourceTsuruVolumeBindDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("volume").(string)

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

	ri := d.Get("restart_on_update").(bool)
	if !ri {
		bindData.Norestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := provider.TsuruClient.VolumeApi.VolumeUnbind(ctx, name, bindData)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

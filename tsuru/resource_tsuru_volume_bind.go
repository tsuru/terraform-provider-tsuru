// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"fmt"
	"io/ioutil"
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

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := provider.TsuruClient.VolumeApi.VolumeBind(ctx, name, bindData)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode != http.StatusOK {
			if isLocked(string(body)) {
				return resource.RetryableError(errors.Errorf("App locked"))
			}
			return resource.NonRetryableError(errors.Errorf("unable to bind volume, error code: %d", resp.StatusCode))
		}

		d.SetId(fmt.Sprintf("%s-%s", bindData.App, bindData.Mountpoint))
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTsuruVolumeBindRead(ctx, d, meta)
}

func resourceTsuruVolumeBindRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("volume_name").(string)
	app := d.Get("app").(string)
	mountPath := d.Get("mount_path").(string)

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
		d.Set("volume_name", name)
		d.Set("app", app)
		d.Set("mount_path", bind.Id.Mountpoint)
		d.Set("read_only", bind.Readonly)
		return nil
	}

	return diag.Errorf("unable to find bind for volume %s to %s", volume.Name, app)
}

func resourceTsuruVolumeBindDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		resp, err := provider.TsuruClient.VolumeApi.VolumeUnbind(ctx, name, bindData)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if resp.StatusCode != http.StatusOK {
			if isLocked(string(body)) {
				return resource.RetryableError(errors.Errorf("App locked"))
			}
			return resource.NonRetryableError(errors.Errorf("unable to unbind volume, error code: %d", resp.StatusCode))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

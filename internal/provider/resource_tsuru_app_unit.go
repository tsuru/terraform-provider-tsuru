// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
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
			"version": {
				Type:        schema.TypeInt,
				Description: "Process name",
				Optional:    true,
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
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	process := d.Get("process").(string)
	units := d.Get("units_count").(int)

	var version *int
	if _v, ok := d.GetOk("version"); ok {
		v := _v.(int)
		version = &v
	}

	baseID := []string{app, process}

	curUnits, err := countUnits(ctx, provider, app, process, version)
	if err != nil {
		return diag.Errorf("Unable to read app %s: %v", app, err)
	}

	delta := units - curUnits
	if delta < 0 {
		return diag.Errorf("App has more running units for process %s than defined, update your tf file", process)
	} else if delta > 0 {

		deltaRequest := tsuru_client.UnitsDelta{
			Units:   strconv.Itoa(delta),
			Process: process,
		}

		if version != nil {
			vStr := strconv.Itoa(*version)
			baseID = append(baseID, vStr)
			deltaRequest.Version = vStr
		}

		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			_, err = provider.TsuruClient.AppApi.UnitsAdd(ctx, app, deltaRequest)
			if err != nil {
				var apiError tsuru_client.GenericOpenAPIError
				if errors.As(err, &apiError) {
					if isRetryableError(apiError.Body()) {
						return resource.RetryableError(err)
					}
				}
				return resource.NonRetryableError(errors.Errorf("unable to add units to %s %s: %v", app, process, err))
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(createID(baseID))

	return resourceTsuruApplicationUnitsRead(ctx, d, meta)
}

func resourceTsuruApplicationUnitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 3)
	if err != nil {
		parts, err = IDtoParts(d.Id(), 2)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	appName := parts[0]
	process := parts[1]
	var version *int
	if len(parts) == 3 {
		v, _ := strconv.Atoi(parts[2])
		version = &v
	}

	units, err := countUnits(ctx, provider, appName, process, version)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("process", process)
	d.Set("units_count", units)

	return nil
}

func resourceTsuruApplicationUnitsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	process := d.Get("process").(string)
	units := d.Get("units_count").(int)

	var version *int
	if _v, ok := d.GetOk("version"); ok {
		v := _v.(int)
		version = &v
	}

	curUnits, err := countUnits(ctx, provider, app, process, version)
	if err != nil {
		return diag.Errorf("Unable to read app %s: %v", app, err)
	}

	delta := units - curUnits
	deltaRequest := tsuru_client.UnitsDelta{
		Process: process,
	}

	if version != nil {
		vStr := strconv.Itoa(*version)
		deltaRequest.Version = vStr
	}

	if delta < 0 {
		deltaRequest.Units = strconv.Itoa(-delta)
		_, err = provider.TsuruClient.AppApi.UnitsRemove(ctx, app, deltaRequest)
		if err != nil {
			return diag.Errorf("unable to remove units from %s %s: %v", app, process, err)
		}

	} else if delta > 0 {
		deltaRequest.Units = strconv.Itoa(delta)
		_, err = provider.TsuruClient.AppApi.UnitsAdd(ctx, app, deltaRequest)
		if err != nil {
			return diag.Errorf("unable to add units to %s %s: %v", app, process, err)
		}
	}

	return resourceTsuruApplicationUnitsRead(ctx, d, meta)
}

func resourceTsuruApplicationUnitsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	process := d.Get("process").(string)
	units := d.Get("units_count").(int)

	var version *int
	if _v, ok := d.GetOk("version"); ok {
		v := _v.(int)
		version = &v
	}

	deltaRequest := tsuru_client.UnitsDelta{
		Units:   strconv.Itoa(units),
		Process: process,
	}

	if version != nil {
		vStr := strconv.Itoa(*version)
		deltaRequest.Version = vStr
	}

	_, err := provider.TsuruClient.AppApi.UnitsRemove(ctx, app, deltaRequest)
	if err != nil {
		return diag.Errorf("unable to remove units from %s %s: %v", app, process, err)
	}

	return nil
}

func countUnits(ctx context.Context, provider *tsuruProvider, appName, process string, version *int) (int, error) {
	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, appName)
	if err != nil {
		if isNotFoundError(err) {
			return 0, nil
		}
		return 0, errors.Errorf("unable to read app %s: %v", app.Name, err)
	}

	units := 0
	for _, u := range app.Units {
		if u.Processname != process {
			continue
		}
		if version != nil && int(u.Version) == *version {
			units++
			continue
		}
		units++
	}
	return units, nil
}

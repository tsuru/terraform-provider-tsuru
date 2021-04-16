// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

var (
	cpuRegex = regexp.MustCompile(`^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$`)
)

func resourceTsuruApplicationAutoscale() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Autoscale",
		CreateContext: resourceTsuruApplicationAutoscaleCreate,
		ReadContext:   resourceTsuruApplicationAutoscaleRead,
		DeleteContext: resourceTsuruApplicationAutoscaleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
				ForceNew:    true,
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
				ForceNew:    true,
				DefaultFunc: func() (interface{}, error) {
					return 1, nil
				},
			},
			"max_units": {
				Type:        schema.TypeInt,
				Description: "maximum number of units",
				Required:    true,
				ForceNew:    true,
			},
			"cpu_average": {
				Type:        schema.TypeString,
				Description: "cpu average",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(i interface{}, k string) (s []string, es []error) {
					v, ok := i.(string)
					if !ok {
						es = append(es, fmt.Errorf("expected type of %s to be string", k))
						return
					}

					if !cpuRegex.MatchString(v) {
						es = append(es, fmt.Errorf("%s value %s must match regexp ^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$", k, v))
						return
					}
					return
				},
			},
		},
	}
}

func resourceTsuruApplicationAutoscaleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)

	appInfo, _, err := provider.TsuruClient.AppApi.AppGet(ctx, app)
	if err != nil {
		return diag.Errorf("Unable to read app %s: %v", app, err)
	}

	if appInfo.Deploys == 0 {
		return nil
	}

	process := d.Get("process").(string)
	minUnits := d.Get("min_units").(int)
	maxUnits := d.Get("max_units").(int)
	if minUnits < 0 {
		minUnits = 1
	}
	if maxUnits < 0 {
		maxUnits = 1
	}
	if minUnits > maxUnits {
		minUnits = maxUnits
	}
	cpu := d.Get("cpu_average").(string)

	autoscale := tsuru_client.AutoScaleSpec{
		Process:    process,
		MinUnits:   int32(minUnits),
		MaxUnits:   int32(maxUnits),
		AverageCPU: cpu,
	}

	resp, err := provider.TsuruClient.AppApi.AutoScaleAdd(ctx, app, autoscale)
	if err != nil {
		return diag.Errorf("Unable to create autoscale %s %s: %v", app, process, err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("Unable to create autoscale: error code %d", resp.StatusCode)
	}

	d.SetId(fmt.Sprintf("%s-%s", app, process))

	return resourceTsuruApplicationAutoscaleRead(ctx, d, meta)
}

func resourceTsuruApplicationAutoscaleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	process := d.Get("process").(string)

	autoscales, _, err := provider.TsuruClient.AppApi.AutoScaleInfo(ctx, app)
	if err != nil {
		return diag.Errorf("Unable to read autoscale %s %s: %v", app, process, err)
	}

	for _, autoscale := range autoscales {
		if autoscale.Process != process {
			continue
		}

		d.Set("app", app)
		d.Set("process", autoscale.Process)
		d.Set("min_units", autoscale.MinUnits)
		d.Set("max_units", autoscale.MaxUnits)
		d.Set("cpu_average", autoscale.AverageCPU)
		return nil
	}

	return diag.Errorf("Unable to read autoscale %s %p: process not found", app, process)
}

func resourceTsuruApplicationAutoscaleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	process := d.Get("process").(string)

	resp, err := provider.TsuruClient.AppApi.AutoScaleRemove(ctx, app, process)
	if err != nil {
		return diag.Errorf("Unable to remove autoscale %s %s: %v", app, process, err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("Unable to remove autoscale: error code %d", resp.StatusCode)
	}

	return nil
}

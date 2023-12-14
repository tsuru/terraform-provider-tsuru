// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruApplicationAutoscale() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Autoscale",
		CreateContext: resourceTsuruApplicationAutoscaleSet,
		ReadContext:   resourceTsuruApplicationAutoscaleRead,
		UpdateContext: resourceTsuruApplicationAutoscaleSet,
		DeleteContext: resourceTsuruApplicationAutoscaleDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
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
				Description: "Application process",
			},
			"min_units": {
				Type:        schema.TypeInt,
				Description: "minimum number of units",
				Required:    true,
				DefaultFunc: func() (interface{}, error) {
					return 1, nil
				},
			},
			"max_units": {
				Type:        schema.TypeInt,
				Description: "maximum number of units",
				Required:    true,
			},
			"cpu_average": {
				Type:         schema.TypeString,
				Description:  "CPU average, for example: 20%, mean that we trigger autoscale when the average of CPU Usage of units is 20%.",
				Optional:     true,
				AtLeastOneOf: []string{"cpu_average", "schedule"},
			},
			"schedule": {
				Type:        schema.TypeList,
				Description: "List of schedules that determine scheduled up/downscales",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_replicas": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Optional:     true,
				AtLeastOneOf: []string{"cpu_average", "schedule"},
			},
		},
	}
}

func resourceTsuruApplicationAutoscaleSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)

	appInfo, _, err := provider.TsuruClient.AppApi.AppGet(ctx, app)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to read app %s: %v", app, err)
	}

	if appInfo.Deploys == 0 {
		return diag.Errorf("We can not set a autoscale without a first deploy")
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

	autoscale := tsuru_client.AutoScaleSpec{
		Process:  process,
		MinUnits: int32(minUnits),
		MaxUnits: int32(maxUnits),
	}

	if cpu, ok := d.GetOk("cpu_average"); ok {
		autoscale.AverageCPU = cpu.(string)
	}

	if m, ok := d.GetOk("schedule"); ok {
		schedules := schedulesFromResourceData(m)
		if schedules != nil {
			autoscale.Schedules = schedules
		}
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = provider.TsuruClient.AppApi.AutoScaleAdd(ctx, app, autoscale)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(errors.Errorf("Unable to create autoscale %s %s: %v", app, process, err))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createID([]string{app, process}))

	return resourceTsuruApplicationAutoscaleRead(ctx, d, meta)
}

func resourceTsuruApplicationAutoscaleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	app := parts[0]
	process := parts[1]

	currentCPUAverage := d.Get("cpu_average").(string)
	isMilli := strings.HasSuffix(currentCPUAverage, "m")
	isPercentage := strings.HasSuffix(currentCPUAverage, "%")

	retryCount := 0
	maxRetries := 5
	// autoscale info reflects near realtime
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		retryCount++
		autoscales, _, err := provider.TsuruClient.AppApi.AutoScaleInfo(ctx, app)
		if err != nil {
			return resource.NonRetryableError(errors.Wrapf(err, "unable to read autoscale %s %s", app, process))
		}

		for _, autoscale := range autoscales {
			if autoscale.Process != process {
				continue
			}

			d.Set("app", app)
			d.Set("process", autoscale.Process)
			d.Set("min_units", autoscale.MinUnits)
			d.Set("max_units", autoscale.MaxUnits)
			if isPercentage {
				d.Set("cpu_average", milliToPercentage(autoscale.AverageCPU)+"%")
			} else if isMilli {
				d.Set("cpu_average", autoscale.AverageCPU)
			} else if strings.HasSuffix(autoscale.AverageCPU, "m") {
				d.Set("cpu_average", milliToPercentage(autoscale.AverageCPU))
			} else if currentCPUAverage != "" {
				d.Set("cpu_average", autoscale.AverageCPU)
			}

			d.Set("schedule", flattenSchedules(autoscale.Schedules))

			return nil
		}

		if retryCount >= maxRetries {
			return resource.NonRetryableError(&MaxRetriesError{Message: fmt.Sprintf("Unable to read autoscale for %s::%s (after %d retries)", app, process, maxRetries)})
		}

		log.Print("[INFO] no autoscales found, trying again")
		return resource.RetryableError(fmt.Errorf("unable to read autoscale for %s::%s: process not found", app, process))
	})

	if err != nil {
		var mrErr *MaxRetriesError
		if errors.As(err, &mrErr) {
			d.SetId("")
			return diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  mrErr.Message,
					Detail:   fmt.Sprintf("%s. Removing it from state.", mrErr.Message),
				},
			}
		}
		return diag.FromErr(err)
	}

	return nil
}

func resourceTsuruApplicationAutoscaleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	process := d.Get("process").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := provider.TsuruClient.AppApi.AutoScaleRemove(ctx, app, process)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(errors.Errorf("Unable to remove autoscale %s %s: %v", app, process, err))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func milliToPercentage(milli string) string {
	if milli == "" {
		return ""
	}

	milliInt, err := strconv.ParseFloat(strings.TrimRight(milli, "m"), 64)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%.2g", milliInt/10)
}

func schedulesFromResourceData(meta interface{}) []tsuru_client.AutoScaleSchedule {
	scheduleMeta := meta.([]interface{})

	if len(scheduleMeta) == 0 {
		return nil
	}

	schedules := []tsuru_client.AutoScaleSchedule{}

	for _, iface := range scheduleMeta {
		schedule := tsuru_client.AutoScaleSchedule{}
		sm := iface.(map[string]interface{})

		if v, ok := sm["min_replicas"]; ok {
			minReplicas := v.(int)
			schedule.MinReplicas = int32(minReplicas)
		}

		if v, ok := sm["start"]; ok {
			schedule.Start = v.(string)
		}

		if v, ok := sm["end"]; ok {
			schedule.End = v.(string)
		}

		if v, ok := sm["timezone"]; ok {
			schedule.Timezone = v.(string)
		}

		schedules = append(schedules, schedule)
	}

	return schedules
}

func flattenSchedules(schedules []tsuru_client.AutoScaleSchedule) []interface{} {
	result := []interface{}{}

	for _, schedule := range schedules {
		result = append(result, map[string]interface{}{
			"min_replicas": schedule.MinReplicas,
			"start":        schedule.Start,
			"end":          schedule.End,
			"timezone":     schedule.Timezone,
		})
	}

	return result
}

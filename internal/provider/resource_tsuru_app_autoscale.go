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
	"k8s.io/utils/ptr"

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
				AtLeastOneOf: []string{"cpu_average", "schedule", "prometheus"},
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
				AtLeastOneOf: []string{"cpu_average", "schedule", "prometheus"},
			},
			"prometheus": {
				Type:        schema.TypeList,
				Description: "List of Prometheus autoscale rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the Prometheus autoscale rule",
						},
						"threshold": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "Threshold value to trigger the autoscaler",
						},
						"query": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Prometheus query to be used in the autoscale rule",
						},
						"custom_address": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom Prometheus URL. If not specified, it will use the default Prometheus from the app's pool",
						},
						"prometheus_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Custom Prometheus URL. If not specified, it will use the default Prometheus from the app's pool",
						},
					},
				},
				Optional:     true,
				AtLeastOneOf: []string{"cpu_average", "schedule", "prometheus"},
			},
			"scale_down": {
				Type:        schema.TypeList,
				Description: "Behavior of the auto scale down",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"units": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Number of units to scale down",
						},
						"percentage": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Percentage of units to scale down",
						},
						"stabilization_window": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Stabilization window in seconds",
						},
					},
				},
				Optional:     true,
				AtLeastOneOf: []string{"cpu_average", "schedule", "prometheus"},
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
		Behavior: tsuru_client.AutoScaleSpecBehavior{
			ScaleDown: tsuru_client.AutoScaleSpecBehaviorScaleDown{},
		},
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

	if p, ok := d.GetOk("prometheus"); ok {
		prometheus := prometheusFromResourceData(p)
		if prometheus != nil {
			autoscale.Prometheus = prometheus
		}
	}
	if m, ok := d.GetOk("scale_down"); ok {
		scaleDown := scaleDownFromResourceData(m)
		autoscale.Behavior.ScaleDown = scaleDown
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

	_, proposed := d.GetChange("scale_down")
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
			d.Set("prometheus", flattenPrometheus(autoscale.Prometheus, d))
			d.Set("scale_down", flattenScaleDown(autoscale.Behavior.ScaleDown, proposed))
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

func prometheusFromResourceData(meta interface{}) []tsuru_client.AutoScalePrometheus {
	prometheusMeta := meta.([]interface{})

	if len(prometheusMeta) == 0 {
		return nil
	}

	prometheus := []tsuru_client.AutoScalePrometheus{}

	for _, iface := range prometheusMeta {
		prom := tsuru_client.AutoScalePrometheus{}
		pm := iface.(map[string]interface{})

		if v, ok := pm["name"]; ok {
			prom.Name = v.(string)
		}

		if v, ok := pm["threshold"]; ok {
			prom.Threshold = v.(float64)
		}

		if v, ok := pm["query"]; ok {
			prom.Query = v.(string)
		}

		if v, ok := pm["custom_address"]; ok {
			prom.PrometheusAddress = v.(string)
		}

		prometheus = append(prometheus, prom)
	}

	return prometheus
}

func scaleDownFromResourceData(meta interface{}) tsuru_client.AutoScaleSpecBehaviorScaleDown {
	scaleDownMeta := meta.([]interface{})
	if len(scaleDownMeta) == 0 {
		return tsuru_client.AutoScaleSpecBehaviorScaleDown{}
	}
	scaleDown := tsuru_client.AutoScaleSpecBehaviorScaleDown{}
	for _, iFace := range scaleDownMeta {
		sd := iFace.(map[string]interface{})
		if v, ok := sd["percentage"]; ok {
			if val, ok := v.(int); ok {
				scaleDown.PercentagePolicyValue = ptr.To(int32(val))
			}
		}
		if v, ok := sd["units"]; ok {
			if val, ok := v.(int); ok {
				scaleDown.UnitsPolicyValue = ptr.To(int32(val))
			}
		}
		if v, ok := sd["stabilization_window"]; ok {
			if val, ok := v.(int); ok {
				scaleDown.StabilizationWindow = ptr.To(int32(val))
			}
		}
	}
	return scaleDown
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

func flattenPrometheus(prometheus []tsuru_client.AutoScalePrometheus, d *schema.ResourceData) []interface{} {
	result := []interface{}{}

	for i, prom := range prometheus {
		customAddressStr := fmt.Sprintf("prometheus.%d.custom_address", i)
		customAddress := d.Get(customAddressStr).(string)
		result = append(result, map[string]interface{}{
			"name":               prom.Name,
			"threshold":          prom.Threshold,
			"query":              prom.Query,
			"custom_address":     customAddress,
			"prometheus_address": prom.PrometheusAddress,
		})
	}

	return result
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"k8s.io/apimachinery/pkg/api/resource"
)

func resourceTsuruApplication() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application",
		CreateContext: resourceTsuruApplicationCreate,
		UpdateContext: resourceTsuruApplicationUpdate,
		ReadContext:   resourceTsuruApplicationRead,
		DeleteContext: resourceTsuruApplicationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceTsuruApplicationImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Application description",
				Optional:    true,
			},
			"platform": {
				Type:        schema.TypeString,
				Description: "Platform",
				Required:    true,
			},
			"plan": {
				Type:        schema.TypeString,
				Description: "Plan",
				Required:    true,
			},
			"custom_cpu_burst": {
				Type:        schema.TypeFloat,
				Description: "CPU burst factory override",
				Optional:    true,
			},
			"custom_cpu": {
				Type:             schema.TypeString,
				Description:      "CPU limit override, such as `500m` or `0.5`. Replaces the CPU defined by the plan",
				Optional:         true,
				ValidateFunc:     validateCPUQuantity,
				DiffSuppressFunc: suppressEquivalentCPU,
			},
			"custom_memory": {
				Type:             schema.TypeString,
				Description:      "Memory limit override, such as `512Mi` or `1Gi`. Replaces the memory defined by the plan",
				Optional:         true,
				ValidateFunc:     validateMemoryQuantity,
				DiffSuppressFunc: suppressEquivalentMemory,
			},
			"team_owner": {
				Type:        schema.TypeString,
				Description: "Application owner",
				Required:    true,
			},
			"cluster": {
				Type:        schema.TypeString,
				Description: "The name of cluster",
				Computed:    true,
			},
			"pool": {
				Type:        schema.TypeString,
				Description: "The name of pool",
				Required:    true,
			},
			"tags": {
				Type:        schema.TypeList,
				Description: "Tags",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": metadataSchema(),
			"process": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"plan": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"metadata": metadataSchema(),
					},
				},
			},

			"default_router": {
				Type:        schema.TypeString,
				Description: "Default router at creation of app",
				Optional:    true,
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "Restart app after applying changes",
				Optional:    true,
			},

			"internal_address": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"process": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"router": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"addresses": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"options": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func metadataSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"labels": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"annotations": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func resourceTsuruApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	platform := d.Get("platform").(string)
	if err := validPlatform(ctx, provider, platform); err != nil {
		return diag.FromErr(err)
	}

	pool := d.Get("pool").(string)
	if err := validPool(ctx, provider, pool); err != nil {
		return diag.FromErr(err)
	}

	plan := d.Get("plan").(string)
	if err := validPlan(ctx, provider, plan); err != nil {
		return diag.FromErr(err)
	}

	tags := []string{}
	for _, item := range d.Get("tags").([]interface{}) {
		tags = append(tags, item.(string))
	}

	defaultRouter := ""
	if i, ok := d.GetOk("default_router"); ok {
		defaultRouter = i.(string)
	}

	app := tsuru_client.InputApp{
		Name:      d.Get("name").(string),
		Platform:  platform,
		Pool:      pool,
		Plan:      plan,
		TeamOwner: d.Get("team_owner").(string),
		Router:    defaultRouter,
		Tags:      tags,
	}

	if m, ok := d.GetOk("metadata"); ok {
		metadata := metadataFromResourceData(m)
		if metadata != nil {
			app.Metadata = *metadata
		}
	}

	if m, ok := d.GetOk("process"); ok {
		processes := processesFromResourceData(m)
		if processes != nil {
			err := validateProcessesOrder(processes)
			if err != nil {
				return diag.FromErr(err)
			}
			app.Processes = processes
		}
	}

	if desc, ok := d.GetOk("description"); ok {
		app.Description = desc.(string)
	}

	_, _, err := provider.TsuruClient.AppApi.AppCreate(ctx, app)
	if err != nil {
		return diag.Errorf("unable to create app %s: %v", app.Name, err)
	}

	d.SetId(app.Name)

	override, err := planOverrideFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if override != (tsuru_client.PlanOverride{}) {
		resp, err := provider.TsuruClient.AppApi.AppUpdate(ctx, app.Name, tsuru_client.UpdateApp{
			Planoverride: override,
			NoRestart:    true,
		})
		if err != nil {
			return diag.Errorf("unable to set plan override on app %s: %v", app.Name, err)
		}
		defer resp.Body.Close()
		logTsuruStream(resp.Body)
	}

	return resourceTsuruApplicationRead(ctx, d, meta)
}

func resourceTsuruApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)
	platform := d.Get("platform").(string)
	if err := validPlatform(ctx, provider, platform); err != nil {
		return diag.FromErr(err)
	}

	pool := d.Get("pool").(string)
	if err := validPool(ctx, provider, pool); err != nil {
		return diag.FromErr(err)
	}

	plan := d.Get("plan").(string)
	if err := validPlan(ctx, provider, plan); err != nil {
		return diag.FromErr(err)
	}

	tags := []string{}
	for _, item := range d.Get("tags").([]interface{}) {
		tags = append(tags, item.(string))
	}

	app := tsuru_client.UpdateApp{
		Platform:  platform,
		Pool:      pool,
		Plan:      plan,
		TeamOwner: d.Get("team_owner").(string),
		Tags:      tags,
	}

	override, err := planOverrideFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}
	app.Planoverride = override

	if d.HasChange("custom_cpu") && app.Planoverride.Cpumilli == nil {
		unset := 0
		app.Planoverride.Cpumilli = &unset
	}

	if d.HasChange("custom_memory") && app.Planoverride.Memory == nil {
		var unset int64
		app.Planoverride.Memory = &unset
	}

	if d.HasChange("metadata") {
		old, new := d.GetChange("metadata")
		oldMetadata := metadataFromResourceData(old)
		if oldMetadata == nil {
			oldMetadata = &tsuru_client.Metadata{}
		}
		newMetadata := metadataFromResourceData(new)
		if newMetadata == nil {
			newMetadata = &tsuru_client.Metadata{}
		}

		app.Metadata = tsuru_client.Metadata{
			Annotations: markRemovedMetadataItemAsDeleted(oldMetadata.Annotations, newMetadata.Annotations),
			Labels:      markRemovedMetadataItemAsDeleted(oldMetadata.Labels, newMetadata.Labels),
		}
	}

	if d.HasChange("process") {
		old, new := d.GetChange("process")
		oldProcesses := processesFromResourceData(old)
		if oldProcesses == nil {
			oldProcesses = []tsuru_client.AppProcess{}
		}
		newProcesses := processesFromResourceData(new)
		if newProcesses == nil {
			newProcesses = []tsuru_client.AppProcess{}
		}
		err := validateProcessesOrder(newProcesses)
		if err != nil {
			return diag.FromErr(err)
		}

		app.Processes = markRemovedProcessAsDefaultPlan(oldProcesses, newProcesses)
	}

	if desc, ok := d.GetOk("description"); ok {
		app.Description = desc.(string)
	}

	restart := true
	if r, ok := d.GetOk("restart_on_update"); ok {
		restart = r.(bool)
	}

	if !restart {
		app.NoRestart = true
	}

	resp, err := provider.TsuruClient.AppApi.AppUpdate(ctx, name, app)
	if err != nil {
		return diag.Errorf("unable to update app %s: %v", name, err)
	}

	defer resp.Body.Close()
	logTsuruStream(resp.Body)

	return resourceTsuruApplicationRead(ctx, d, meta)
}

func planOverrideFromResourceData(d *schema.ResourceData) (tsuru_client.PlanOverride, error) {
	override := tsuru_client.PlanOverride{}

	if cpuBurst, ok := d.GetOk("custom_cpu_burst"); ok {
		cpuBurstValue := cpuBurst.(float64)
		override.CpuBurst = &cpuBurstValue
	}

	if cpu, ok := d.GetOk("custom_cpu"); ok {
		cpuMilli, err := parseCPUMilli(cpu.(string))
		if err != nil {
			return override, err
		}
		override.Cpumilli = &cpuMilli
	}

	if memory, ok := d.GetOk("custom_memory"); ok {
		memoryBytes, err := parseMemoryQuantity(memory.(string))
		if err != nil {
			return override, fmt.Errorf("invalid custom_memory %q: %w", memory, err)
		}
		override.Memory = &memoryBytes
	}

	return override, nil
}

func parseCPUMilli(cpu string) (int, error) {
	quantity, err := resource.ParseQuantity(cpu)
	if err != nil {
		return 0, fmt.Errorf("invalid custom_cpu %q: %w", cpu, err)
	}
	return int(quantity.MilliValue()), nil
}

func validateCPUQuantity(v interface{}, k string) ([]string, []error) {
	if _, err := parseCPUMilli(v.(string)); err != nil {
		return nil, []error{err}
	}
	return nil, nil
}

func validateMemoryQuantity(v interface{}, k string) ([]string, []error) {
	if _, err := parseMemoryQuantity(v.(string)); err != nil {
		return nil, []error{fmt.Errorf("invalid %s %q: %w", k, v, err)}
	}
	return nil, nil
}

func suppressEquivalentCPU(k, old, new string, d *schema.ResourceData) bool {
	oldMilli, oldErr := parseCPUMilli(old)
	newMilli, newErr := parseCPUMilli(new)
	return oldErr == nil && newErr == nil && oldMilli == newMilli
}

func suppressEquivalentMemory(k, old, new string, d *schema.ResourceData) bool {
	oldBytes, oldErr := parseMemoryQuantity(old)
	newBytes, newErr := parseMemoryQuantity(new)
	return oldErr == nil && newErr == nil && oldBytes == newBytes
}

func validateProcessesOrder(processes []tsuru_client.AppProcess) error {
	for i := 1; i < len(processes); i++ {
		if processes[i-1].Name > processes[i].Name {
			return errors.New("please, sort app processes alphabetically")
		}
	}
	return nil
}

func resourceTsuruApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Id()

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, name)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read app %s: %v", app.Name, err)
	}

	d.Set("name", name)
	d.Set("platform", app.Platform)
	d.Set("pool", app.Pool)
	d.Set("plan", app.Plan.Name)
	d.Set("team_owner", app.TeamOwner)
	d.Set("cluster", app.Cluster)

	if app.Plan.Override.CpuBurst != nil {
		d.Set("custom_cpu_burst", app.Plan.Override.CpuBurst)
	}

	if app.Plan.Override.Cpumilli != nil {
		d.Set("custom_cpu", cpuMillisToString(int32(*app.Plan.Override.Cpumilli)))
	} else {
		d.Set("custom_cpu", "")
	}

	if app.Plan.Override.Memory != nil {
		d.Set("custom_memory", memoryBytesToString(*app.Plan.Override.Memory))
	} else {
		d.Set("custom_memory", "")
	}

	if app.Description != "" {
		d.Set("description", app.Description)
	}

	d.Set("tags", app.Tags)

	d.Set("metadata", flattenMetadata(app.Metadata))
	d.Set("internal_address", flattenInternalAddresses(app.InternalAddresses))
	d.Set("router", flattenRouters(app.Routers))
	d.Set("process", flattenProcesses(app.Processes))

	return nil
}

func resourceTsuruApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	_, err := provider.TsuruClient.AppApi.AppDelete(ctx, name)
	if err != nil {
		return diag.Errorf("unable to delete app %s: %v", name, err)
	}

	return nil
}

func resourceTsuruApplicationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	provider := meta.(*tsuruProvider)

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("name", app.Name)
	d.SetId(app.Name)

	return []*schema.ResourceData{d}, nil
}

func processesFromResourceData(meta interface{}) []tsuru_client.AppProcess {
	m := meta.([]interface{})

	if len(m) == 0 {
		return nil
	}

	processes := []tsuru_client.AppProcess{}

	for _, iface := range m {
		process := tsuru_client.AppProcess{}
		m := iface.(map[string]interface{})

		if v, ok := m["name"]; ok {
			process.Name = v.(string)
		}

		if v, ok := m["plan"]; ok {
			process.Plan = v.(string)
		}

		if process.Plan == "" {
			process.Plan = "$default"
		}

		if v, ok := m["metadata"]; ok {
			metadata := metadataFromResourceData(v)

			if metadata != nil {
				process.Metadata = *metadata
			}
		}

		processes = append(processes, process)
	}

	return processes
}

func metadataFromResourceData(meta interface{}) *tsuru_client.Metadata {
	m := meta.([]interface{})
	if len(m) == 0 || m[0] == nil {
		return nil
	}

	cm := tsuru_client.Metadata{}

	metadataMap := m[0].(map[string]interface{})
	if v, ok := metadataMap["labels"]; ok && len(v.(map[string]interface{})) > 0 {
		l := v.(map[string]interface{})

		for key, value := range l {
			cm.Labels = append(cm.Labels, tsuru_client.MetadataItem{Name: key, Value: value.(string)})
		}
	}

	if v, ok := metadataMap["annotations"]; ok && len(v.(map[string]interface{})) > 0 {
		l := v.(map[string]interface{})
		for key, value := range l {
			cm.Annotations = append(cm.Annotations, tsuru_client.MetadataItem{Name: key, Value: value.(string)})
		}
	}

	if len(cm.Labels) == 0 && len(cm.Annotations) == 0 {
		return nil
	}

	return &cm
}

func validPlatform(ctx context.Context, provider *tsuruProvider, platform string) error {
	platforms, _, err := provider.TsuruClient.PlatformApi.PlatformList(ctx)
	if err != nil {
		return err
	}

	platformParts := strings.SplitN(platform, ":", 2)
	availablePlatforms := []string{}
	for _, p := range platforms {
		if p.Disabled {
			continue
		}

		availablePlatforms = append(availablePlatforms, p.Name)
		if p.Name == platformParts[0] {
			return nil
		}
	}
	plaformList := strings.Join(availablePlatforms, ",")

	return fmt.Errorf("invalid platform: %s available platforms are [%s]", platform, plaformList)
}

func validPool(ctx context.Context, provider *tsuruProvider, pool string) error {
	pools, _, err := provider.TsuruClient.PoolApi.PoolList(ctx)
	if err != nil {
		return err
	}

	availablePools := []string{}
	for _, p := range pools {
		availablePools = append(availablePools, p.Name)
		if p.Name == pool {
			return nil
		}
	}
	poolList := strings.Join(availablePools, ",")

	return fmt.Errorf("invalid pool: %s available pools are [%s]", pool, poolList)
}

func validPlan(ctx context.Context, provider *tsuruProvider, plan string) error {
	plans, _, err := provider.TsuruClient.PlanApi.PlanList(ctx)
	if err != nil {
		return err
	}

	availablePlans := []string{}
	for _, p := range plans {
		availablePlans = append(availablePlans, p.Name)
		if p.Name == plan {
			return nil
		}
	}
	plansList := strings.Join(availablePlans, ",")

	return fmt.Errorf("invalid plan: %s available plans are [%s]", plan, plansList)
}

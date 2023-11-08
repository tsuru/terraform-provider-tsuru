// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
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

	if m, ok := d.GetOk("metadata"); ok {
		metadata := metadataFromResourceData(m)
		if metadata != nil {
			app.Metadata = *metadata
		}
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

	return errors.Errorf("invalid platform: %s available platforms are [%s]", platform, plaformList)
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

	return errors.Errorf("invalid pool: %s available pools are [%s]", pool, poolList)
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

	return errors.Errorf("invalid plan: %s available plans are [%s]", plan, plansList)
}

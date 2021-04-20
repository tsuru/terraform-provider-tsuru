// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"strings"

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
		Importer: &schema.ResourceImporter{
			StateContext: resourceTsuruApplicationImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
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
			"metadata": {
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
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "Restart app after applying changes",
				Optional:    true,
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

	app := tsuru_client.InputApp{
		Name:      d.Get("name").(string),
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

	_, err := provider.TsuruClient.AppApi.AppUpdate(ctx, name, app)
	if err != nil {
		return diag.Errorf("unable to update app %s: %v", name, err)
	}

	return nil
}

func resourceTsuruApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, name)
	if err != nil {
		return diag.Errorf("unable to create app %s: %v", app.Name, err)
	}

	d.Set("name", name)
	d.Set("platform", app.Platform)
	d.Set("pool", app.Pool)
	d.Set("plan", app.Plan.Name)
	d.Set("team_owner", app.TeamOwner)

	if app.Description != "" {
		d.Set("description", app.Description)
	}

	if len(app.Tags) > 0 {
		tags := []interface{}{}
		for _, item := range app.Tags {
			tags = append(tags, item)
		}
		d.Set("tags", tags)
	}

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

	availablePlatforms := []string{}
	for _, p := range platforms {
		if p.Disabled {
			continue
		}

		availablePlatforms = append(availablePlatforms, p.Name)
		if p.Name == platform {
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

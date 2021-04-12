// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"net/http"

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
			StateContext: schema.ImportStatePassthroughContext,
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

	if desc, ok := d.GetOk("description"); ok {
		app.Description = desc.(string)
	}

	_, resp, err := provider.TsuruClient.AppApi.AppCreate(ctx, app)
	if err != nil {
		return diag.Errorf("unable to create app %s: %v", app.Name, err)
	}

	if resp.StatusCode != http.StatusCreated {
		return diag.Errorf("unable to create app %s: error code %d", app.Name, resp.StatusCode)
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

	restart := true
	if r, ok := d.GetOk("restart_on_update"); ok {
		restart = r.(bool)
	}

	if !restart {
		app.NoRestart = true
	}

	if desc, ok := d.GetOk("description"); ok {
		app.Description = desc.(string)
	}

	resp, err := provider.TsuruClient.AppApi.AppUpdate(ctx, name, app)
	if err != nil {
		return diag.Errorf("unable to update app %s: %v", name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to update app %s: error code %d", name, resp.StatusCode)
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

	resp, err := provider.TsuruClient.AppApi.AppDelete(ctx, name)
	if err != nil {
		return diag.Errorf("unable to delete app %s: %v", name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to delete app %s: error code %d", name, resp.StatusCode)
	}

	return nil
}

func validPlatform(ctx context.Context, provider *tsuruProvider, platform string) error {
	platforms, _, err := provider.TsuruClient.PlatformApi.PlatformList(ctx)
	if err != nil {
		return err
	}

	for _, p := range platforms {
		if p.Name == platform && !p.Disabled {
			return nil
		}
	}
	return errors.Errorf("invalid platform: %s", platform)
}

func validPool(ctx context.Context, provider *tsuruProvider, pool string) error {
	pools, _, err := provider.TsuruClient.PoolApi.PoolList(ctx)
	if err != nil {
		return err
	}

	for _, p := range pools {
		if p.Name == pool {
			return nil
		}
	}
	return errors.Errorf("invalid pool: %s", pool)
}

func validPlan(ctx context.Context, provider *tsuruProvider, plan string) error {
	plans, _, err := provider.TsuruClient.PlanApi.PlanList(ctx)
	if err != nil {
		return err
	}

	for _, p := range plans {
		if p.Name == plan {
			return nil
		}
	}
	return errors.Errorf("invalid plan: %s", plan)
}

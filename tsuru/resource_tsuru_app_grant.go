// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruApplicationGrant() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Access Grant",
		CreateContext: resourceTsuruApplicationGrantCreate,
		UpdateContext: resourceTsuruApplicationGrantCreate,
		ReadContext:   resourceTsuruApplicationGrantRead,
		DeleteContext: resourceTsuruApplicationGrantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"team": {
				Type:        schema.TypeString,
				Description: "Teams to grant access to the app",
				Required:    true,
			},
		},
	}
}

func resourceTsuruApplicationGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	team := d.Get("team").(string)

	resp, err := provider.TsuruClient.AppApi.AppTeamGrant(ctx, app, team)
	if err != nil {
		return diag.Errorf("unable to add team grant: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to add team grant, error code: %d", resp.StatusCode)
	}

	d.SetId(app)

	return resourceTsuruApplicationGrantRead(ctx, d, meta)
}

func resourceTsuruApplicationGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)
	team := d.Get("team").(string)

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, appName)
	if err != nil {
		return diag.Errorf("unable to get app %s: %v", appName, err)
	}

	for _, t := range app.Teams {
		if t == team {
			d.Set("team", t)
		}
	}

	return nil
}

func resourceTsuruApplicationGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	team := d.Get("team").(string)

	resp, err := provider.TsuruClient.AppApi.AppTeamRevoke(ctx, app, team)
	if err != nil {
		return diag.Errorf("unable to revoke team grant: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to revoke team grant, error code: %d", resp.StatusCode)
	}

	return nil
}

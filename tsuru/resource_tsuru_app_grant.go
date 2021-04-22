// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruApplicationGrant() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Access Grant",
		CreateContext: resourceTsuruApplicationGrantCreate,
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
				ForceNew:    true,
			},
			"team": {
				Type:        schema.TypeString,
				Description: "Teams to grant access to the app",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTsuruApplicationGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	team := d.Get("team").(string)

	_, err := provider.TsuruClient.AppApi.AppTeamGrant(ctx, app, team)
	if err != nil {
		return diag.Errorf("unable to add team grant: %v", err)
	}

	d.SetId(createID([]string{app, team}))

	return resourceTsuruApplicationGrantRead(ctx, d, meta)
}

func resourceTsuruApplicationGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	appName := parts[0]
	team := parts[1]

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

	_, err := provider.TsuruClient.AppApi.AppTeamRevoke(ctx, app, team)
	if err != nil {
		return diag.Errorf("unable to revoke team grant: %v", err)
	}

	return nil
}

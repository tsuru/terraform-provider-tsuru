// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruTeam() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a tsuru team. Changing `name` renames the team in place: tsuru updates the apps, services, service instances, volumes and pools that reference it.",
		CreateContext: resourceTsuruTeamCreate,
		ReadContext:   resourceTsuruTeamRead,
		UpdateContext: resourceTsuruTeamUpdate,
		DeleteContext: resourceTsuruTeamDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Team name",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the team",
			},
		},
	}
}

func resourceTsuruTeamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	_, err := provider.TsuruClient.TeamApi.TeamCreate(ctx, tsuru.TeamCreateArgs{
		Name: name,
		Tags: teamTags(d),
	})
	if err != nil {
		return diag.Errorf("unable to create team %q: %v", name, err)
	}
	d.SetId(name)

	return resourceTsuruTeamRead(ctx, d, meta)
}

func resourceTsuruTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Id()

	// GET /teams/{name} also expands units, routers and metrics of every app owned
	// by the team, so the lighter team list is used to refresh the state.
	teams, resp, err := provider.TsuruClient.TeamApi.TeamsList(ctx)
	if resp != nil && resp.StatusCode == http.StatusNoContent {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("unable to read team %q: %v", name, err)
	}

	for _, team := range teams {
		if team.Name != name {
			continue
		}
		d.Set("name", team.Name)
		d.Set("tags", team.Tags)
		return nil
	}

	d.SetId("")
	return nil
}

func resourceTsuruTeamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	args := tsuru.TeamUpdateArgs{
		Tags: teamTags(d),
	}
	if d.HasChange("name") {
		args.Newname = d.Get("name").(string)
	}

	_, err := provider.TsuruClient.TeamApi.TeamUpdate(ctx, d.Id(), args)
	if err != nil {
		return diag.Errorf("unable to update team %q: %v", d.Id(), err)
	}
	if args.Newname != "" {
		d.SetId(args.Newname)
	}

	return resourceTsuruTeamRead(ctx, d, meta)
}

func resourceTsuruTeamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	_, err := provider.TsuruClient.TeamApi.TeamDelete(ctx, d.Id())
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return diag.Errorf("unable to delete team %q: %v", d.Id(), err)
	}

	return nil
}

func teamTags(d *schema.ResourceData) []string {
	tags := []string{}
	for _, item := range d.Get("tags").([]interface{}) {
		tags = append(tags, item.(string))
	}
	return tags
}

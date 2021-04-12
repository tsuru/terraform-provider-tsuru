// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

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
			"teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Teams to grant access to the app",
				Optional:    true,
			},
		},
	}
}

func resourceTsuruApplicationGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)

	errs := []diag.Diagnostic{}
	for _, team := range d.Get("teams").([]interface{}) {
		resp, err := provider.TsuruClient.AppApi.AppTeamGrant(ctx, app, team.(string))
		if err != nil {
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to add team grant: %v", err),
			}
			errs = append(errs, diagnostic)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to add team grant, error code: %d", resp.StatusCode),
			}
			errs = append(errs, diagnostic)
			continue
		}
	}

	if len(errs) > 0 {
		return errs
	}

	d.SetId(app)

	return resourceTsuruApplicationGrantRead(ctx, d, meta)
}

func resourceTsuruApplicationGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)

	teams := []string{}
	for _, team := range d.Get("teams").([]interface{}) {
		teams = append(teams, team.(string))
	}

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, appName)
	if err != nil {
		return diag.Errorf("unable to get app %s: %v", appName, err)
	}

	if reflect.DeepEqual(teams, app.Teams) {
		d.Set("teams", app.Teams)
	}

	return nil
}

func resourceTsuruApplicationGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)

	errs := []diag.Diagnostic{}
	for _, team := range d.Get("teams").([]interface{}) {
		resp, err := provider.TsuruClient.AppApi.AppTeamRevoke(ctx, app, team.(string))
		if err != nil {
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to revoke team grant: %v", err),
			}
			errs = append(errs, diagnostic)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to revoke team grant, error code: %d", resp.StatusCode),
			}
			errs = append(errs, diagnostic)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

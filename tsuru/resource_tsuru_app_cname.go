// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruApplicationCName() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application CName",
		CreateContext: resourceTsuruApplicationCNameCreate,
		UpdateContext: resourceTsuruApplicationCNameCreate,
		ReadContext:   resourceTsuruApplicationCNameRead,
		DeleteContext: resourceTsuruApplicationCNameDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"hostname": {
				Type:        schema.TypeString,
				Description: "Application description",
				Optional:    true,
			},
		},
	}
}

func resourceTsuruApplicationCNameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	hostname := d.Get("hostname").(string)
	cname := tsuru_client.AppCName{
		Cname: []string{hostname},
	}

	resp, err := provider.TsuruClient.AppApi.AppCnameAdd(ctx, app, cname)
	if err != nil {
		return diag.Errorf("unable to add cname: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to add cname, error code: %d", resp.StatusCode)
	}

	d.SetId(hostname)

	return resourceTsuruApplicationCNameRead(ctx, d, meta)
}

func resourceTsuruApplicationCNameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)
	hostname := d.Get("hostname").(string)

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, appName)
	if err != nil {
		return diag.Errorf("unable to get app %s: %v", appName, err)
	}

	for _, name := range app.Cname {
		if hostname == name {
			d.Set("hostname", name)
			break
		}
	}

	return nil
}

func resourceTsuruApplicationCNameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	hostname := d.Get("hostname").(string)
	cname := tsuru_client.AppCName{
		Cname: []string{hostname},
	}

	resp, err := provider.TsuruClient.AppApi.AppCnameDelete(ctx, app, cname)
	if err != nil {
		return diag.Errorf("unable to delete cname: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to delete cname, error code: %d", resp.StatusCode)
	}

	return nil
}

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

func resourceTsuruApplicationRouter() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Router",
		CreateContext: resourceTsuruApplicationRouterCreate,
		ReadContext:   resourceTsuruApplicationRouterRead,
		UpdateContext: resourceTsuruApplicationRouterUpdate,
		DeleteContext: resourceTsuruApplicationRouterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Router name",
				Required:    true,
			},
			"options": {
				Type:        schema.TypeMap,
				Description: "Application description",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTsuruApplicationRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)
	name := d.Get("name").(string)

	if err := validRouter(ctx, provider, name); err != nil {
		return diag.Errorf("unable to create router: %v", err)
	}

	options := map[string]interface{}{}
	for key, value := range d.Get("options").(map[string]interface{}) {
		options[key] = value.(string)
	}

	router := tsuru_client.AppRouter{
		Name: name,
		Opts: options,
	}

	resp, err := provider.TsuruClient.AppApi.AppRouterAdd(ctx, appName, router)
	if err != nil {
		return diag.Errorf("unable to create router: %v", err)
	}

	switch resp.StatusCode {
	case http.StatusConflict:
		fallthrough
	case http.StatusOK:
		d.SetId(name)
	default:
		return diag.Errorf("unable to create router, error code: %d", resp.StatusCode)
	}

	return resourceTsuruApplicationRouterRead(ctx, d, meta)
}

func resourceTsuruApplicationRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)
	name := d.Get("name").(string)

	routers, _, err := provider.TsuruClient.AppApi.AppRouterList(ctx, appName)
	if err != nil {
		return diag.Errorf("unable to get app %s: %v", appName, err)
	}

	for _, router := range routers {
		if name != router.Name {
			continue
		}
		d.Set("name", name)
		d.Set("options", router.Opts)
		return nil
	}

	return nil
}

func resourceTsuruApplicationRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)
	name := d.Get("name").(string)

	options := map[string]interface{}{}
	for key, value := range d.Get("options").(map[string]interface{}) {
		options[key] = value.(string)
	}

	router := tsuru_client.AppRouter{
		Name: name,
		Opts: options,
	}

	resp, err := provider.TsuruClient.AppApi.AppRouterUpdate(ctx, appName, name, router)
	if err != nil {
		return diag.Errorf("unable to update router: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to update router, error code: %d", resp.StatusCode)
	}

	return resourceTsuruApplicationRouterRead(ctx, d, meta)
}

func resourceTsuruApplicationRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	appName := d.Get("app").(string)
	name := d.Get("name").(string)

	resp, err := provider.TsuruClient.AppApi.AppRouterDelete(ctx, appName, name)
	if err != nil {
		return diag.Errorf("unable to delete router: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("unable to delete router, error code: %d", resp.StatusCode)
	}

	return nil
}

func validRouter(ctx context.Context, provider *tsuruProvider, router string) error {
	routers, _, err := provider.TsuruClient.RouterApi.RouterList(ctx)
	if err != nil {
		return err
	}

	for _, r := range routers {
		if r.Name == router {
			return nil
		}
	}
	return errors.Errorf("invalid router: %s", router)
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},
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

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := provider.TsuruClient.AppApi.AppRouterAdd(ctx, appName, router)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(errors.Errorf("unable to create router: %v", err))
			}
		}
		d.SetId(createID([]string{appName, name}))
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTsuruApplicationRouterRead(ctx, d, meta)
}

func resourceTsuruApplicationRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	appName := parts[0]
	name := parts[1]

	routers, _, err := provider.TsuruClient.AppApi.AppRouterList(ctx, appName)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
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

	return diag.Errorf("unable to find router %s on app %s", name, appName)
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

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := provider.TsuruClient.AppApi.AppRouterUpdate(ctx, appName, name, router)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
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

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruServiceInstanceBind() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Service Instance Bind",
		CreateContext: resourceTsuruServiceInstanceBindCreate,
		ReadContext:   resourceTsuruServiceInstanceBindRead,
		DeleteContext: resourceTsuruServiceInstanceBindDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service kind",
			},
			"service_instance": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service instance",
			},
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				ForceNew:    true,
				Required:    true,
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "restart app after applying",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTsuruServiceInstanceBindCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	service := d.Get("service_name").(string)
	instance := d.Get("service_instance").(string)
	app := d.Get("app").(string)

	noRestart := false
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		noRestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := provider.TsuruClient.ServiceApi.ServiceInstanceBind(ctx, service, instance, app, tsuru_client.ServiceInstanceBind{
			NoRestart: noRestart,
		})
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
		}
		defer resp.Body.Close()
		logTsuruStream(resp.Body)
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createID([]string{service, instance, app}))

	return resourceTsuruServiceInstanceBindRead(ctx, d, meta)
}

func resourceTsuruServiceInstanceBindRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 3)
	if err != nil {
		return diag.FromErr(err)
	}
	service := parts[0]
	instanceName := parts[1]
	app := parts[2]

	instance, _, err := provider.TsuruClient.ServiceApi.InstanceGet(ctx, service, instanceName)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read bind %s %s: %v", service, instanceName, err)
	}

	for _, a := range instance.Apps {
		if app == a {
			d.Set("app", a)
			d.Set("service_name", service)
			d.Set("service_instance", instanceName)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceTsuruServiceInstanceBindDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	service := d.Get("service_name").(string)
	instance := d.Get("service_instance").(string)
	app := d.Get("app").(string)

	noRestart := false
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		noRestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := provider.TsuruClient.ServiceApi.ServiceInstanceUnbind(ctx, service, instance, app, tsuru_client.ServiceInstanceUnbind{
			NoRestart: noRestart,
		})
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

	d.SetId("")
	return nil
}

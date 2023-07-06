// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"log"
	"net/http"
	"time"

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
		UpdateContext: resourceTsuruServiceInstanceBindUpdate,
		ReadContext:   resourceTsuruServiceInstanceBindRead,
		DeleteContext: resourceTsuruServiceInstanceBindDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
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
				Type:         schema.TypeString,
				Description:  "Application name",
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"job", "app"},
			},
			"job": {
				Type:         schema.TypeString,
				Description:  "Application name",
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"job", "app"},
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "restart app after applying (default = true)",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceTsuruServiceInstanceBindCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	service := d.Get("service_name").(string)
	instance := d.Get("service_instance").(string)

	noRestart := false
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		noRestart = true
	}

	var idToSet, app_name, job_name string
	if app, ok := d.GetOk("app"); ok {
		app_name = app.(string)
		idToSet = createID([]string{service, instance, app_name})
	}
	if job, ok := d.GetOk("job"); ok {
		job_name = job.(string)
		idToSet = createID([]string{service, instance, "tsuru-job", job_name})
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var resp *http.Response
		var err error
		if app_name != "" {
			resp, err = provider.TsuruClient.ServiceApi.ServiceInstanceBind(ctx, service, instance, app_name, tsuru_client.ServiceInstanceBind{
				NoRestart: noRestart,
			})
		} else {
			resp, err = provider.TsuruClient.ServiceApi.JobServiceInstanceBind(ctx, service, instance, job_name, tsuru_client.JobServiceInstanceBind{})
		}
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

	d.SetId(idToSet)

	return resourceTsuruServiceInstanceBindRead(ctx, d, meta)
}

func resourceTsuruServiceInstanceBindUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("[INFO] Update bind is a no-op")
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
	name := parts[2]
	if name == "tsuru-job" {
		name = parts[3]
	}

	instance, _, err := provider.TsuruClient.ServiceApi.InstanceGet(ctx, service, instanceName)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read bind %s %s: %v", service, instanceName, err)
	}

	for _, a := range instance.Apps {
		if name == a {
			d.Set("app", a)
			d.Set("service_name", service)
			d.Set("service_instance", instanceName)
			return nil
		}
	}

	for _, j := range instance.Jobs {
		if name == j {
			d.Set("job", j)
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

	var app, job string
	if app, ok := d.GetOk("app"); ok {
		app = app.(string)
	}
	if job, ok := d.GetOk("job"); ok {
		job = job.(string)
	}

	noRestart := false
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		noRestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var err error
		if app != "" {
			_, err = provider.TsuruClient.ServiceApi.ServiceInstanceUnbind(ctx, service, instance, app, tsuru_client.ServiceInstanceUnbind{
				NoRestart: noRestart,
			})
		} else {
			_, err = provider.TsuruClient.ServiceApi.JobServiceInstanceUnbind(ctx, service, instance, job, tsuru_client.JobServiceInstanceUnbind{})
		}
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

	return nil
}

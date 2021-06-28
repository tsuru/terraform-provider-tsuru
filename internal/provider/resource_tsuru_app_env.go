// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

const (
	PublicVariable  = "public"
	PrivateVariable = "private"
)

func resourceTsuruApplicationEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Environment",
		CreateContext: resourceTsuruApplicationEnvironmentCreate,
		ReadContext:   resourceTsuruApplicationEnvironmentRead,
		UpdateContext: resourceTsuruApplicationEnvironmentUpdate,
		DeleteContext: resourceTsuruApplicationEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
			},
			"environment_variable": {
				Type:        schema.TypeSet,
				Description: "Environment variables",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Variable name",
							Required:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "Variable value",
							Optional:    true,
						},
						"sensitive_value": {
							Type:        schema.TypeString,
							Description: "Sensitive variable value",
							Sensitive:   true,
							Optional:    true,
						},
					},
				},
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Description: "restart app after applying",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceTsuruApplicationEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)

	envs := envsFromResource(d.Get("environment_variable"))
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		envs.Norestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(envs.Envs) == 0 {
			return resource.NonRetryableError(errors.Errorf("No environment variables to create"))
		}
		_, err := provider.TsuruClient.AppApi.EnvSet(ctx, app, *envs)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(app)

	return resourceTsuruApplicationEnvironmentRead(ctx, d, meta)
}

func resourceTsuruApplicationEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	app := d.Id()

	envs, _, err := provider.TsuruClient.AppApi.EnvGet(ctx, app, nil)
	if err != nil {
		return diag.Errorf("unable to read envs for app %s: %v", app, err)
	}

	var envVars []map[string]interface{}
	for _, env := range envs {
		if isReservedEnv(env.Name) {
			continue
		}
		if env.Public {
			envVars = append(envVars, map[string]interface{}{
				"name":  env.Name,
				"value": env.Value,
			})
		} else {
			envVars = append(envVars, map[string]interface{}{
				"name":            env.Name,
				"sensitive_value": env.Value,
			})
		}
	}
	d.Set("environment_variable", envVars)

	return nil
}

func resourceTsuruApplicationEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	envs := envsFromResource(d.Get("environment_variable"))

	ri := d.Get("restart_on_update").(bool)
	if !ri {
		envs.Norestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		if len(envs.Envs) == 0 {
			return resource.NonRetryableError(errors.Errorf("No environment variables to update"))
		}
		_, err := provider.TsuruClient.AppApi.EnvSet(ctx, app, *envs)
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

	return resourceTsuruApplicationEnvironmentRead(ctx, d, meta)
}

func resourceTsuruApplicationEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	envSet := envsFromResource(d.Get("environment_variable"))

	envs := []string{}
	for _, e := range envSet.Envs {
		envs = append(envs, e.Name)
	}

	noRestart := false
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		noRestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := provider.TsuruClient.AppApi.EnvUnset(ctx, app, envs, noRestart)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func envsFromResource(envvars interface{}) *tsuru_client.EnvSetData {
	evs := envvars.(*schema.Set)

	envs := &tsuru_client.EnvSetData{
		Envs:        []tsuru_client.Env{},
		ManagedBy:   "terraform",
		PruneUnused: true,
	}

	for _, raw := range evs.List() {
		e := raw.(map[string]interface{})
		env := tsuru_client.Env{
			Name: e["name"].(string),
		}
		if v, ok := e["value"]; ok && v != "" {
			env.Value = v.(string)
		} else if v, ok := e["sensitive_value"]; ok && v != "" {
			env.Value = v.(string)
			env.Private = true
		}
		envs.Envs = append(envs.Envs, env)
	}

	return envs
}

func isReservedEnv(variable string) bool {
	return strings.HasPrefix(variable, "TSURU_")
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
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
			"environment_variables": {
				Description: "Environment variables",
				Required:    false,
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"private_environment_variables": {
				Description: "Environment variables",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, Sensitive: true},
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

	envs := &tsuru_client.EnvSetData{
		Envs:        []tsuru_client.Env{},
		ManagedBy:   "terraform",
		PruneUnused: true,
	}
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("environment_variables"), false)...)
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("private_environment_variables"), true)...)

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
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read envs for app %s: %v", app, err)
	}

	envs = filterUnmanagedTerraformEnvs(envs)

	envVars := map[string]string{}
	sensitiveEnvVars := map[string]string{}

	for _, env := range envs {
		if env.Public {
			envVars[env.Name] = env.Value
		} else {
			sensitiveEnvVars[env.Name] = env.Value
		}
	}

	d.Set("environment_variables", envVars)
	d.Set("private_environment_variables", sensitiveEnvVars)

	return nil
}

func resourceTsuruApplicationEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	envs := &tsuru_client.EnvSetData{
		Envs:        []tsuru_client.Env{},
		ManagedBy:   "terraform",
		PruneUnused: true,
	}
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("environment_variables"), false)...)
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("private_environment_variables"), true)...)

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

	noRestart := false
	ri := d.Get("restart_on_update").(bool)
	if !ri {
		noRestart = true
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := provider.TsuruClient.AppApi.EnvSet(ctx, app, tsuru_client.EnvSetData{
			Envs:      []tsuru.Env{},
			ManagedBy: "terraform",
			Norestart: noRestart,
		})
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

func envsFromResource(envvars interface{}, private bool) []tsuru_client.Env {
	m := envvars.(map[string]interface{})

	envs := []tsuru_client.Env{}

	for key, raw := range m {
		env := tsuru_client.Env{
			Name:    key,
			Value:   raw.(string),
			Private: private,
		}
		envs = append(envs, env)
	}

	return envs
}

func filterUnmanagedTerraformEnvs(envs []tsuru.EnvVar) []tsuru.EnvVar {
	n := 0
	for _, env := range envs {
		if isReservedEnv(env.Name) {
			continue
		}
		if env.ManagedBy != "terraform" {
			continue
		}
		envs[n] = env
		n++
	}
	envs = envs[:n]
	return envs
}

func isReservedEnv(variable string) bool {
	for _, internalEnv := range internalEnvs() {
		if internalEnv == variable {
			return true
		}
	}
	return false
}

func internalEnvs() []string {
	return []string{
		"TSURU_HOST", "TSURU_APPNAME", "TSURU_APP_TOKEN", "TSURU_SERVICE", "TSURU_APPDIR", "TSURU_SERVICES",
	}
}

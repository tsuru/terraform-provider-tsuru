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

func resourceTsuruJobEnvironment() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Job Environment",
		CreateContext: resourceTsuruJobEnvironmentCreate,
		ReadContext:   resourceTsuruJobEnvironmentRead,
		UpdateContext: resourceTsuruJobEnvironmentUpdate,
		DeleteContext: resourceTsuruJobEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"job": {
				Type:        schema.TypeString,
				Description: "Job name",
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
		},
	}
}

func resourceTsuruJobEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	job := d.Get("job").(string)

	envs := &tsuru_client.EnvSetData{
		Envs:        []tsuru_client.Env{},
		ManagedBy:   "terraform",
		PruneUnused: true,
	}
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("environment_variables"), false)...)
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("private_environment_variables"), true)...)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(envs.Envs) == 0 {
			return resource.NonRetryableError(errors.Errorf("No environment variables to create"))
		}
		resp, err := provider.TsuruClient.JobApi.JobEnvSet(ctx, job, *envs)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}

		defer resp.Body.Close()
		logTsuruStream(resp.Body)

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(job)

	return resourceTsuruJobEnvironmentRead(ctx, d, meta)
}

func resourceTsuruJobEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	job := d.Id()

	envs, _, err := provider.TsuruClient.JobApi.JobEnvGet(ctx, job, nil)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read envs for job %s: %v", job, err)
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

func resourceTsuruJobEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	job := d.Get("job").(string)
	envs := &tsuru_client.EnvSetData{
		Envs:        []tsuru_client.Env{},
		ManagedBy:   "terraform",
		PruneUnused: true,
	}
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("environment_variables"), false)...)
	envs.Envs = append(envs.Envs, envsFromResource(d.Get("private_environment_variables"), true)...)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		if len(envs.Envs) == 0 {
			return resource.NonRetryableError(errors.Errorf("No environment variables to update"))
		}
		_, err := provider.TsuruClient.JobApi.JobEnvSet(ctx, job, *envs)
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

	return resourceTsuruJobEnvironmentRead(ctx, d, meta)
}

func resourceTsuruJobEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	job := d.Get("job").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := provider.TsuruClient.JobApi.JobEnvSet(ctx, job, tsuru_client.EnvSetData{
			Envs:      []tsuru.Env{},
			ManagedBy: "terraform",
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

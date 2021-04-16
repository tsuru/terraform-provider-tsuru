// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
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

	privateEnvs, publicEnvs := envsFromResource(d.Get("environment_variable"))
	var envsPtr *tsuru_client.EnvSetData
	if len(privateEnvs.Envs) > 0 && len(publicEnvs.Envs) > 0 {
		// Don't restart when private and public have variables
		privateEnvs.Norestart = true
		envsPtr = publicEnvs
	} else if len(publicEnvs.Envs) == 0 {
		envsPtr = privateEnvs
	} else if len(privateEnvs.Envs) == 0 {
		envsPtr = publicEnvs
	}

	if ri, ok := d.GetOk("restart_on_update"); ok {
		r := ri.(bool)
		if !r {
			envsPtr.Norestart = true
		}
	}

	provider.Log.Infof("create public: %#v private: %#v", publicEnvs, privateEnvs)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(privateEnvs.Envs) > 0 {
			provider.TsuruClient.AppApi.EnvSet(ctx, app, *privateEnvs)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(publicEnvs.Envs) > 0 {
			provider.TsuruClient.AppApi.EnvSet(ctx, app, *publicEnvs)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s-envs", app))

	return resourceTsuruApplicationEnvironmentRead(ctx, d, meta)
}

func resourceTsuruApplicationEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	app := d.Get("app").(string)

	envs, _, err := provider.TsuruClient.AppApi.EnvGet(ctx, app, nil)
	if err != nil {
		return diag.Errorf("unable to read envs for app %s: %v", app, err)
	}

	prefix := "environment_variable"
	for i, env := range envs {
		if isReservedEnv(env.Name) {
			continue
		}
		d.Set(fmt.Sprintf("%s.%d.name", prefix, i), env.Name)
		if env.Public {
			d.Set(fmt.Sprintf("%s.%d.value", prefix, i), env.Value)
		} else {
			d.Set(fmt.Sprintf("%s.%d.sensitive_value", prefix, i), env.Value)
		}
	}

	return nil
}

func resourceTsuruApplicationEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)

	privateEnvs, publicEnvs := envsFromResource(d.Get("environment_variable"))
	curPrivate, curPublic, err := getRemoteEnvironment(ctx, provider.TsuruClient, app)
	if err != nil {
		return diag.Errorf("unable to get current environment variables for app %s: %v", app, err)
	}
	privateUpdate, privateRemove := diffEnvSetData(privateEnvs.Envs, curPrivate)
	publicUpdate, publicRemove := diffEnvSetData(publicEnvs.Envs, curPublic)

	provider.Log.Infof("update publicRemove: %#v privateRemove: %#v", publicRemove, privateRemove)

	err = removeEnvVars(ctx, provider.TsuruClient, d.Timeout(schema.TimeoutUpdate), app, privateRemove)
	if err != nil {
		return diag.Errorf("unable to remove private envs: %v", err)
	}
	err = removeEnvVars(ctx, provider.TsuruClient, d.Timeout(schema.TimeoutUpdate), app, publicRemove)
	if err != nil {
		return diag.Errorf("unable to remove public envs: %v", err)
	}

	privateEnvs.Envs = privateUpdate
	publicEnvs.Envs = publicUpdate
	var envsPtr *tsuru_client.EnvSetData
	if len(privateEnvs.Envs) > 0 && len(publicEnvs.Envs) > 0 {
		// Don't restart when private and public have variables
		privateEnvs.Norestart = true
		envsPtr = publicEnvs
	} else if len(publicEnvs.Envs) == 0 {
		envsPtr = privateEnvs
	} else if len(privateEnvs.Envs) == 0 {
		envsPtr = publicEnvs
	}

	if ri, ok := d.GetOk("restart_on_update"); ok {
		r := ri.(bool)
		if !r {
			envsPtr.Norestart = true
		}
	}

	provider.Log.Infof("update public: %#v private: %#v", publicEnvs, privateEnvs)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(privateEnvs.Envs) > 0 {
			provider.TsuruClient.AppApi.EnvSet(ctx, app, *privateEnvs)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(publicEnvs.Envs) > 0 {
			provider.TsuruClient.AppApi.EnvSet(ctx, app, *publicEnvs)
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

	privateEnvs, publicEnvs := envsFromResource(d.Get("environment_variable"))
	var envsPtr *tsuru_client.EnvSetData
	if len(privateEnvs.Envs) > 0 && len(publicEnvs.Envs) > 0 {
		// Don't restart when private and public have variables
		privateEnvs.Norestart = true
		envsPtr = publicEnvs
	} else if len(publicEnvs.Envs) == 0 {
		envsPtr = privateEnvs
	} else if len(privateEnvs.Envs) == 0 {
		envsPtr = publicEnvs
	}

	if ri, ok := d.GetOk("restart_on_update"); ok {
		r := ri.(bool)
		if !r {
			envsPtr.Norestart = true
		}
	}

	provider.Log.Infof("delete public: %#v private: %#v", publicEnvs, privateEnvs)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if len(privateEnvs.Envs) > 0 {
			varNames := getVarNames(privateEnvs.Envs)
			provider.TsuruClient.AppApi.EnvUnset(ctx, app, varNames, privateEnvs.Norestart)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if len(publicEnvs.Envs) > 0 {
			varNames := getVarNames(privateEnvs.Envs)
			provider.TsuruClient.AppApi.EnvUnset(ctx, app, varNames, privateEnvs.Norestart)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func removeEnvVars(ctx context.Context, client *tsuru_client.APIClient, timeout time.Duration, app string, envs []tsuru_client.Env) error {
	return resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		if len(envs) > 0 {
			varNames := getVarNames(envs)
			resp, err := client.AppApi.EnvUnset(ctx, app, varNames, true)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			if resp.StatusCode != http.StatusOK {
				if isLocked(string(body)) {
					return resource.RetryableError(errors.Errorf("App locked"))
				}
				return resource.NonRetryableError(errors.Errorf("unable to add cname, error code: %d", resp.StatusCode))
			}
		}
		return nil
	})
}

func getRemoteEnvironment(ctx context.Context, client *tsuru_client.APIClient, app string) (private, public []tsuru_client.Env, err error) {
	envs, _, err := client.AppApi.EnvGet(ctx, app, nil)
	if err != nil {
		return nil, nil, err
	}

	for _, env := range envs {
		if isReservedEnv(env.Name) {
			continue
		}
		if env.Public {
			public = append(public, env)
		} else {
			private = append(private, env)
		}
	}

	return
}

func diffEnvSetData(newEnvs, oldEnvs []tsuru_client.Env) (update, remove []tsuru_client.Env) {
	update = []tsuru_client.Env{}
	remove = []tsuru_client.Env{}

	for _, newEnv := range newEnvs {
		if curEnv, ok := envInSet(newEnv.Name, oldEnvs); ok {
			if curEnv.Value != newEnv.Value {
				update = append(update, newEnv)
			}
		} else {
			update = append(update, newEnv)
		}
	}

	for _, oldEnv := range oldEnvs {
		if _, ok := envInSet(oldEnv.Name, newEnvs); !ok {
			remove = append(remove, oldEnv)
		}
	}

	return
}

func envInSet(envName string, envs []tsuru_client.Env) (*tsuru_client.Env, bool) {
	for _, e := range envs {
		if e.Name == envName {
			return &e, true
		}
	}
	return nil, false
}

func getVarNames(envs []tsuru_client.Env) []string {
	names := []string{}
	for _, e := range envs {
		names = append(names, e.Name)
	}
	return names
}

func envsFromResource(envvars interface{}) (*tsuru_client.EnvSetData, *tsuru_client.EnvSetData) {
	evs := envvars.(*schema.Set)

	private := &tsuru_client.EnvSetData{
		Envs:    []tsuru_client.Env{},
		Private: true,
	}

	public := &tsuru_client.EnvSetData{
		Envs: []tsuru_client.Env{},
	}

	for _, raw := range evs.List() {
		e := raw.(map[string]interface{})
		env := tsuru_client.Env{
			Name: e["name"].(string),
		}
		if v, ok := e["value"]; ok && v != "" {
			env.Value = v.(string)
			public.Envs = append(public.Envs, env)
			continue
		}
		if v, ok := e["sensitive_value"]; ok && v != "" {
			env.Value = v.(string)
			private.Envs = append(private.Envs, env)
			continue
		}
	}

	return private, public
}

func isReservedEnv(variable string) bool {
	reserved := map[string]bool{
		"TSURU_APPDIR":    true,
		"TSURU_APPNAME":   true,
		"TSURU_APP_TOKEN": true,
		"TSURU_SERVICES":  true,
	}
	_, ok := reserved[variable]
	return ok
}

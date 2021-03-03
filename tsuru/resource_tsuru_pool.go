// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruPoolCreate,
		ReadContext:   resourceTsuruPoolRead,
		UpdateContext: resourceTsuruPoolUpdate,
		DeleteContext: resourceTsuruPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Unique name of pool",
				Required:    true,
			},
			"tsuru_provisioner": {
				Type:        schema.TypeString,
				Description: "Provisioner of pool",
				Default:     "kubernetes",
				ForceNew:    true, // there is no way to update this field
				Optional:    true,
			},
			"public": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"default": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Key/value to store additional config",
			},
		},
	}
}

func resourceTsuruPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	labels := make(map[string]string)
	for key, value := range d.Get("labels").(map[string]interface{}) {
		labels[key] = value.(string)
	}

	_, err := provider.TsuruClient.PoolApi.PoolCreate(ctx, tsuru.PoolCreateData{
		Name:        name,
		Provisioner: d.Get("tsuru_provisioner").(string),
		Public:      d.Get("public").(bool),
		Default:     d.Get("default").(bool),
		Labels:      labels,
	})

	if err != nil {
		return diag.Errorf("Could not create tsuru pool: %q, err: %s", name, err.Error())
	}
	d.SetId(name)
	return nil
}

func resourceTsuruPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	pool, _, err := provider.TsuruClient.PoolApi.PoolGet(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Could not read tsuru pool: %q, err: %s", d.Id(), err.Error())
	}
	d.Set("name", pool.Name)
	d.Set("tsuru_provisioner", pool.Provisioner)
	d.Set("default", pool.Default)
	d.Set("public", pool.Public)
	d.Set("labels", pool.Labels)

	return nil
}

func resourceTsuruPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	if d.HasChange("name") {
		return diag.Errorf("Could not change property \"name\"")
	}

	_, err := provider.TsuruClient.PoolApi.PoolUpdate(ctx, d.Id(), tsuru.PoolUpdateData{
		Default: d.Get("default").(bool),
		Public:  d.Get("public").(bool),
	})
	if err != nil {
		return diag.Errorf("Could not update tsuru pool: %q, err: %s", d.Id(), err.Error())
	}
	return nil
}

func resourceTsuruPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	_, err := provider.TsuruClient.PoolApi.PoolDelete(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Could not delete tsuru pool: %q, err: %s", d.Id(), err.Error())
	}
	return nil
}

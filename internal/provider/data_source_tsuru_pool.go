// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTsuruPool() *schema.Resource {
	return &schema.Resource{
		Description: "Reads a tsuru pool, including the values allowed by its constraints.",
		ReadContext: dataSourceTsuruPoolRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Pool name",
			},
			"tsuru_provisioner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Provisioner of pool",
			},
			"default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is the default pool",
			},
			"public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the pool is available to every team",
			},
			"teams": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Teams allowed to use the pool",
			},
			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Labels of the pool",
			},
			"allowed": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Values allowed by the pool constraints, one entry per constraint field",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Constraint field, like team, router, service, plan, volume-plan or cert-issuer",
						},
						"values": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Values allowed for the field",
						},
					},
				},
			},
		},
	}
}

func dataSourceTsuruPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	pool, _, err := provider.TsuruClient.PoolApi.PoolGet(ctx, name)
	if err != nil {
		return diag.Errorf("unable to read pool %q: %v", name, err)
	}
	d.SetId(name)

	d.Set("tsuru_provisioner", pool.Provisioner)
	d.Set("default", pool.Default)
	d.Set("public", pool.Public)
	d.Set("teams", pool.Teams)
	d.Set("labels", pool.Labels)
	d.Set("allowed", flattenPoolAllowed(pool.Allowed))

	return nil
}

func flattenPoolAllowed(allowed map[string][]string) []interface{} {
	fields := make([]string, 0, len(allowed))
	for field := range allowed {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	result := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		result = append(result, map[string]interface{}{
			"field":  field,
			"values": allowed[field],
		})
	}
	return result
}

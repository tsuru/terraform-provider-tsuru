// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruPoolConstraint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruPoolConstraintSet,
		ReadContext:   resourceTsuruPoolConstraintRead,
		UpdateContext: resourceTsuruPoolConstraintSet,
		DeleteContext: resourceTsuruPoolConstraintDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"pool_expr": {
				Type:        schema.TypeString,
				Description: "The name of pools, allow glob match style",
				Required:    true,
				ForceNew:    true, // there is no way to update this field
			},
			"field": {
				Type:        schema.TypeString,
				Description: "field of constraint",
				ForceNew:    true, // there is no way to update this field
				Required:    true,
			},
			"values": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"blacklist": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func resourceTsuruPoolConstraintSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	id := d.Get("pool_expr").(string) + "/" + d.Get("field").(string)

	values := []string{}

	for _, item := range d.Get("values").([]interface{}) {
		values = append(values, item.(string))
	}

	constraint := tsuru.PoolConstraintSet{
		PoolExpr:  d.Get("pool_expr").(string),
		Field:     d.Get("field").(string),
		Values:    values,
		Blacklist: d.Get("blacklist").(bool),
	}

	err := tsuruRetry(ctx, d, func() error {
		_, internalErr := provider.TsuruClient.PoolApi.ConstraintSet(ctx, constraint)
		return internalErr
	})

	if err != nil {
		return diag.Errorf("Could not set tsuru pool pool constraint: %q, err: %s", id, err.Error())
	}
	d.SetId(id)

	return resourceTsuruPoolConstraintRead(ctx, d, meta)

}

func resourceTsuruPoolConstraintRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	parts := strings.SplitN(d.Id(), "/", 2)

	poolExpr := parts[0]
	field := parts[1]

	constraints, _, err := provider.TsuruClient.PoolApi.ConstraintList(ctx)

	if err != nil {
		return diag.Errorf("Could not list tsuru pool pool constraints, err: %s", err.Error())
	}

	for _, constraint := range constraints {
		if constraint.PoolExpr != poolExpr || constraint.Field != field {
			continue
		}

		values := []interface{}{}
		for _, value := range constraint.Values {
			values = append(values, value)
		}

		d.Set("pool_expr", constraint.PoolExpr)
		d.Set("field", constraint.Field)
		d.Set("values", values)
		d.Set("blacklist", constraint.Blacklist)

		return nil
	}

	d.SetId("")
	return nil
}

func resourceTsuruPoolConstraintDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	id := d.Get("pool_expr").(string) + "/" + d.Get("field").(string)

	err := tsuruRetry(ctx, d, func() error {
		_, internalErr := provider.TsuruClient.PoolApi.ConstraintSet(ctx, tsuru.PoolConstraintSet{
			PoolExpr: d.Get("pool_expr").(string),
			Field:    d.Get("field").(string),
			Values:   []string{},
		})

		return internalErr
	})

	if err != nil {
		return diag.Errorf("Could not set tsuru pool empty pool constraints: %q, err: %s", id, err.Error())
	}

	return nil
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"time"

	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruServiceCreate,
		ReadContext:   resourceTsuruServiceRead,
		UpdateContext: resourceTsuruServiceUpdate,
		DeleteContext: resourceTsuruServiceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service name",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username for service authentication",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for service authentication",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service endpoint URL",
			},
			"multi_cluster": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the service supports multi-cluster",
			},
			"team": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Team owner of this service",
			},
		},
	}
}

func resourceTsuruServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("name").(string)

	opts := &tsuru.ServiceCreateOpts{
		Id:       optional.NewString(name),
		Endpoint: optional.NewString(d.Get("endpoint").(string)),
		Team:     optional.NewString(d.Get("team").(string)),
	}

	if v, ok := d.GetOk("username"); ok {
		opts.Username = optional.NewString(v.(string))
	}

	if v, ok := d.GetOk("password"); ok {
		opts.Password = optional.NewString(v.(string))
	}

	if d.Get("multi_cluster").(bool) {
		opts.MultiCluster = optional.NewString("true")
	}

	_, err := provider.TsuruClient.ServiceApi.ServiceCreate(ctx, opts)
	if err != nil {
		return diag.Errorf("Could not create tsuru service %q, err: %s", name, err.Error())
	}

	d.SetId(name)

	return resourceTsuruServiceRead(ctx, d, meta)
}

func resourceTsuruServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()

	services, _, err := provider.TsuruClient.ServiceApi.ServicesList(ctx)
	if err != nil {
		return diag.Errorf("Could not list tsuru services, err: %s", err.Error())
	}

	for _, svc := range services {
		if svc.Service == name {
			d.Set("name", name)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceTsuruServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()

	opts := &tsuru.ServiceUpdateOpts{
		Id:       optional.NewString(name),
		Endpoint: optional.NewString(d.Get("endpoint").(string)),
		Team:     optional.NewString(d.Get("team").(string)),
	}

	if v, ok := d.GetOk("username"); ok {
		opts.Username = optional.NewString(v.(string))
	}

	if v, ok := d.GetOk("password"); ok {
		opts.Password = optional.NewString(v.(string))
	}

	if d.Get("multi_cluster").(bool) {
		opts.MultiCluster = optional.NewString("true")
	} else {
		opts.MultiCluster = optional.NewString("false")
	}

	_, err := provider.TsuruClient.ServiceApi.ServiceUpdate(ctx, name, opts)
	if err != nil {
		return diag.Errorf("Could not update tsuru service %q, err: %s", name, err.Error())
	}

	return resourceTsuruServiceRead(ctx, d, meta)
}

func resourceTsuruServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()

	_, err := provider.TsuruClient.ServiceApi.ServiceDelete(ctx, name)
	if err != nil {
		return diag.Errorf("Could not delete tsuru service %q, err: %s", name, err.Error())
	}

	return nil
}

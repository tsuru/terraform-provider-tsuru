// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruApplicationDeploy() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application",
		CreateContext: resourceTsuruApplicationDeployCreate,
		ReadContext:   resourceTsuruApplicationDeployRead,
		DeleteContext: resourceTsuruApplicationDeployDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
				ForceNew:    true,
			},
			"image_url": {
				Type:        schema.TypeString,
				Description: "Application description",
				Required:    true,
				ForceNew:    true,
			},
			"message": {
				Type:        schema.TypeString,
				Description: "A message describing this deploy",
				Optional:    true,
				ForceNew:    true,
			},
			"new_version": {
				Type:        schema.TypeBool,
				Description: "Creates a new version for the current deployment while preserving existing versions",
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
			"override_old_versions": {
				Type:        schema.TypeBool,
				Description: "Force replace all deployed versions by this new deploy",
				Default:     true,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTsuruApplicationDeployCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	appDeployOptions := tsuru_client.AppDeployOptions{}

	if v, ok := d.GetOk("image_url"); ok {
		appDeployOptions.Image = v.(string)
	}
	if v, ok := d.GetOk("message"); ok {
		appDeployOptions.Message = v.(string)
	}
	if v, ok := d.GetOk("new_version"); ok {
		appDeployOptions.NewVersion = v.(bool)
	}
	if v, ok := d.GetOk("override_old_versions"); ok {
		appDeployOptions.OverrideVersions = v.(bool)
	}

	_, err := provider.TsuruClient.AppApi.AppDeploy(ctx, app, appDeployOptions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(app)
	return resourceTsuruApplicationDeployRead(ctx, d, meta)
}

func resourceTsuruApplicationDeployRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceTsuruApplicationDeployDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

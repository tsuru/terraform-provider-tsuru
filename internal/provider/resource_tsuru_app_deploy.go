// Copyright 2023 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"bufio"
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruApplicationDeploy() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Application Deploy",
		CreateContext: resourceTsuruApplicationDeployCreate,
		ReadContext:   resourceTsuruApplicationDeployRead,
		DeleteContext: resourceTsuruApplicationDeployDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
				ForceNew:    true,
			},
			"image": {
				Type:        schema.TypeString,
				Description: "Docker Image",
				Required:    true,
				ForceNew:    true,
			},

			"new_version": {
				Type:        schema.TypeBool,
				Description: "Creates a new version for the current deployment while preserving existing versions",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},

			"override_old_versions": {
				Type:        schema.TypeBool,
				Description: "Force replace all deployed versions by this new deploy",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},

			"wait": {
				Type:        schema.TypeBool,
				Description: "Wait for the rollout of deploy",
				Optional:    true,
				ForceNew:    true,
				Default:     true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"output_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTsuruApplicationDeployCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	app := d.Get("app").(string)
	image := d.Get("image").(string)
	newVersion := d.Get("new_version").(bool)
	overrideOldVersions := d.Get("override_old_versions").(bool)
	wait := d.Get("wait").(bool)

	resp, err := provider.TsuruClient.AppApi.AppDeploy(ctx, app, tsuru.AppDeployOptions{
		Image:            image,
		Message:          "deploy via terraform",
		NewVersion:       newVersion,
		OverrideVersions: overrideOldVersions,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	eventID := resp.Header.Get("X-Tsuru-Eventid")
	d.SetId(eventID)

	if wait {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			log.Println("[DEBUG]", scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal("[ERROR]", err)
		}
	}

	return resourceTsuruApplicationDeployRead(ctx, d, meta)
}

func resourceTsuruApplicationDeployRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//provider := meta.(*tsuruProvider)
	// TODO read event and update status and output_image
	return nil
}

func resourceTsuruApplicationDeployDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("[DEBUG] delete a deploy is a no-op by terraform")
	return nil
}

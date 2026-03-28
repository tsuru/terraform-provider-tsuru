// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruPlatform() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Platform",
		CreateContext: resourceTsuruPlatformCreate,
		ReadContext:   resourceTsuruPlatformRead,
		UpdateContext: resourceTsuruPlatformUpdate,
		DeleteContext: resourceTsuruPlatformDelete,
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
				Description: "Platform name",
				Required:    true,
				ForceNew:    true,
			},
			"dockerfile": {
				Type:        schema.TypeString,
				Description: "Platform dockerfile URL",
				Optional:    true,
			},
			"dockerfile_content": {
				Type:        schema.TypeString,
				Description: "Platform dockerfile content",
				Optional:    true,
			},
			"disabled": {
				Type:        schema.TypeBool,
				Description: "Platform disabled status",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceTsuruPlatformCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Implement platform creation
	// provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)
	
	// For now, just set the ID - we'll implement the actual API calls later
	d.SetId(name)
	
	return resourceTsuruPlatformRead(ctx, d, meta)
}

func resourceTsuruPlatformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Implement platform reading
	// provider := meta.(*tsuruProvider)
	// name := d.Id()
	
	return nil
}

func resourceTsuruPlatformUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Implement platform update
	// provider := meta.(*tsuruProvider)
	
	return resourceTsuruPlatformRead(ctx, d, meta)
}

func resourceTsuruPlatformDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Implement platform deletion
	// provider := meta.(*tsuruProvider)
	// name := d.Id()
	
	return nil
}
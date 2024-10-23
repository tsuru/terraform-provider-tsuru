// Copyright 2024 tsuru authors. All rights reserved.
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
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruToken() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Token",
		CreateContext: resourceTsuruTokenCreate,
		ReadContext:   resourceTsuruTokenRead,
		UpdateContext: resourceTsuruTokenUpdate,
		DeleteContext: resourceTsuruTokenDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"team": {
				Type:        schema.TypeString,
				Description: "The team name responsible for this token",
				Required:    true,
			},
			"token_id": {
				Type:        schema.TypeString,
				Description: "Token name, must be a unique identifier, if empty it will be generated automatically",
				Optional:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Token description",
				Optional:    true,
			},
			"expires": {
				Type:        schema.TypeString,
				Description: "Token expiration with suffix (s for seconds, m for minutos, h for hours, ...) 0 or unset means it never expires",
				Optional:    true,
				Default:     "0s",
			},
			"regenerate_on_update": {
				Type:        schema.TypeBool,
				Description: "Setting regenerate will change de value of the token, invalidating the previous value",
				Optional:    true,
				Default:     false,
			},
			"token": {
				Type:        schema.TypeString,
				Description: "Tsuru token",
				Computed:    true,
				Sensitive:   true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Token creation date",
				Computed:    true,
			},
			"expires_at": {
				Type:        schema.TypeString,
				Description: "Token expiration date",
				Computed:    true,
			},
			"last_access": {
				Type:        schema.TypeString,
				Description: "Token last access date",
				Computed:    true,
			},
			"creator_email": {
				Type:        schema.TypeString,
				Description: "Token creator email",
				Computed:    true,
			},
			"roles": {
				Type:        schema.TypeList,
				Description: "Tsuru token roles",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"context_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceTsuruTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	teamToken := tsuru_client.TeamTokenCreateArgs{
		Team: d.Get("team").(string),
	}

	if tokenId, ok := d.GetOk("token_id"); ok {
		teamToken.TokenId = tokenId.(string)
	}

	if desc, ok := d.GetOk("description"); ok {
		teamToken.Description = desc.(string)
	}

	if expires, ok := d.GetOk("expires"); ok {
		duration, err := time.ParseDuration(expires.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		teamToken.ExpiresIn = int64(duration.Seconds())
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		token, _, err := provider.TsuruClient.AuthApi.TeamTokenCreate(ctx, teamToken)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}
		d.SetId(token.TokenId)
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTsuruTokenRead(ctx, d, meta)
}

func resourceTsuruTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	tokenId := d.Id()

	teamToken, _, err := provider.TsuruClient.AuthApi.TeamTokenInfo(ctx, tokenId)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read token %s: %v", tokenId, err)
	}

	d.Set("token", teamToken.Token)
	d.Set("created_at", formatDate(teamToken.CreatedAt))
	d.Set("expires_at", formatDate(teamToken.ExpiresAt))
	d.Set("last_access", formatDate(teamToken.LastAccess))
	d.Set("creator_email", teamToken.CreatorEmail)
	d.Set("team", teamToken.Team)
	d.Set("description", teamToken.Description)
	d.Set("roles", flattenRoles(teamToken.Roles))

	return nil
}

func resourceTsuruTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	tokenId := d.Id()

	teamToken := tsuru_client.TeamTokenUpdateArgs{}

	if regenerate, ok := d.GetOk("regenerate_on_update"); ok {
		teamToken.Regenerate = regenerate.(bool)
	}

	if desc, ok := d.GetOk("description"); ok {
		teamToken.Description = desc.(string)
	}

	if expires, ok := d.GetOk("expires"); ok {
		duration, err := time.ParseDuration(expires.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		teamToken.ExpiresIn = int64(duration.Seconds())
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, _, err := provider.TsuruClient.AuthApi.TeamTokenUpdate(ctx, tokenId, teamToken)
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
		return diag.Errorf("unable to update token %s: %v", tokenId, err)
	}

	return resourceTsuruTokenRead(ctx, d, meta)
}

func resourceTsuruTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	tokenId := d.Id()

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := provider.TsuruClient.AuthApi.TeamTokenDelete(ctx, tokenId)
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
		return diag.Errorf("unable to delete token %s: %v", tokenId, err)
	}

	return nil
}

func flattenRoles(roles []tsuru_client.RoleInstance) []interface{} {
	result := []interface{}{}

	for _, role := range roles {
		result = append(result, map[string]interface{}{
			"name":          role.Name,
			"context_value": role.Contextvalue,
		})
	}

	return result
}

func formatDate(date time.Time) string {
	if date.IsZero() {
		return "-"
	}
	return date.In(time.Local).Format(time.RFC822)
}

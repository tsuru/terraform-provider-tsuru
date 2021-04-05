// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruWebhookCreate,
		ReadContext:   resourceTsuruWebhookRead,
		DeleteContext: resourceTsuruWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_owner": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"event_filter": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_types": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							ForceNew: true,
						},
						"target_values": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							ForceNew: true,
						},
						"kind_types": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							ForceNew: true,
						},
						"kind_names": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							ForceNew: true,
						},
						"error_only": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						"success_only": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
					},
				},
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"proxy_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"method": {
				Type:     schema.TypeString,
				Default:  "POST",
				Optional: true,
				ForceNew: true,
			},
			"headers": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
			},
			"body": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"insecure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceTsuruWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	webhook := tsuru.Webhook{
		Name:        name,
		Description: d.Get("description").(string),
		TeamOwner:   d.Get("team_owner").(string),
		Url:         d.Get("url").(string),
		Method:      d.Get("method").(string),
		Headers:     map[string][]string{},
		Insecure:    d.Get("insecure").(bool),
	}

	if v, ok := d.GetOk("proxy_url"); ok {
		webhook.ProxyUrl = v.(string)
	}

	if v, ok := d.GetOk("body"); ok {
		webhook.Body = v.(string)
	}

	for key, value := range d.Get("headers").(map[string]interface{}) {
		webhook.Headers[key] = []string{
			value.(string),
		}
	}

	if eventFilters, ok := d.Get("event_filter").([]interface{}); ok && len(eventFilters) > 0 {
		eventFilter := eventFilters[0].(map[string]interface{})

		if value, ok := parseStringSlice(eventFilter["kind_names"]); ok {
			webhook.EventFilter.KindNames = value
		}

		if value, ok := parseStringSlice(eventFilter["kind_types"]); ok {
			webhook.EventFilter.KindTypes = value
		}

		if value, ok := parseStringSlice(eventFilter["target_types"]); ok {
			webhook.EventFilter.TargetTypes = value
		}

		if value, ok := parseStringSlice(eventFilter["target_values"]); ok {
			webhook.EventFilter.TargetValues = value
		}

		if reflect.DeepEqual(eventFilter["error_only"], true) {
			webhook.EventFilter.ErrorOnly = true
		}

		if reflect.DeepEqual(eventFilter["success_only"], true) {
			webhook.EventFilter.SuccessOnly = true
		}
	}

	_, err := provider.TsuruClient.EventApi.WebhookCreate(ctx, webhook)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceTsuruWebhookRead(ctx, d, meta)
}

func parseStringSlice(i interface{}) ([]string, bool) {
	arrayInterface, ok := i.([]interface{})
	if !ok {
		return nil, false
	}

	result := []string{}

	for _, v := range arrayInterface {
		result = append(result, v.(string))
	}

	return result, true
}

func resourceTsuruWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	provider := meta.(*tsuruProvider)
	webhook, _, err := provider.TsuruClient.EventApi.WebhookGet(ctx, id)

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", webhook.Name)
	d.Set("description", webhook.Description)
	d.Set("team_owner", webhook.TeamOwner)
	d.Set("event_filter", flattenEventFilters(webhook.EventFilter))
	d.Set("url", webhook.Url)
	d.Set("method", webhook.Method)

	headers := map[string]string{}
	for key, values := range webhook.Headers {
		headers[key] = values[0]
	}

	d.Set("headers", headers)
	d.Set("proxy_url", webhook.ProxyUrl)
	d.Set("body", webhook.Body)
	d.Set("insecure", webhook.Insecure)

	return nil
}

func resourceTsuruWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	provider := meta.(*tsuruProvider)

	_, err := provider.TsuruClient.EventApi.WebhookDelete(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenEventFilters(filter tsuru.WebhookEventFilter) []interface{} {
	result := map[string]interface{}{
		"target_types":  filter.TargetTypes,
		"target_values": filter.TargetValues,
		"kind_names":    filter.KindNames,
		"kind_types":    filter.KindTypes,
		"error_only":    filter.ErrorOnly,
		"success_only":  filter.SuccessOnly,
	}

	return []interface{}{result}
}

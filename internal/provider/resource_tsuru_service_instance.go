// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"

	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruServiceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruServiceInstanceCreate,
		ReadContext:   resourceTsuruServiceInstanceRead,
		UpdateContext: resourceTsuruServiceInstanceUpdate,
		DeleteContext: resourceTsuruServiceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Instance name",
			},
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of service kind",
			},
			"plan": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Team owner of this instance",
			},
			"pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service Pool",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Human readable description for instance",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Custom tags for instance",
			},
			"parameters": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Service instance addicional parameters",
			},
			"unbind_on_delete": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Unbind service instance from apps on delete",
			},
		},
	}
}

func resourceTsuruServiceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("name").(string)
	serviceName := d.Get("service_name").(string)
	plan := d.Get("plan").(string)
	owner := d.Get("owner").(string)
	pool := d.Get("pool").(string)

	instance := tsuru.ServiceInstance{
		Name:        name,
		ServiceName: serviceName,
		PlanName:    plan,
		TeamOwner:   owner,
		Pool:        pool,
	}

	if description, ok := d.GetOk("description"); ok {
		instance.Description = description.(string)
	}

	if tags, ok := d.GetOk("tags"); ok {
		instance.Tags = parseTags(tags)
	}

	if parameters, ok := d.GetOk("parameters"); ok {
		instance.Parameters = parseParameters(parameters)
	}

	_, err := provider.TsuruClient.ServiceApi.InstanceCreate(ctx, serviceName, instance)

	if err != nil {
		return diag.Errorf("Could not create tsuru router, err : %s", err.Error())
	}

	d.SetId(createID([]string{serviceName, name}))

	return nil
}

func resourceTsuruServiceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts, err := IDtoParts(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	serviceName := parts[0]
	name := parts[1]

	instance, _, err := provider.TsuruClient.ServiceApi.InstanceGet(ctx, serviceName, name)
	if err != nil {
		return diag.Errorf("Could not read tsuru service (%s) instance (%s), err : %s", serviceName, name, err.Error())
	}

	d.Set("name", name)
	d.Set("service_name", serviceName)
	d.Set("plan", instance.Planname)
	d.Set("owner", instance.Teamowner)
	d.Set("pool", instance.Pool)

	if instance.Description != "" {
		d.Set("description", instance.Description)
	}

	if len(instance.Tags) > 0 {
		d.Set("tags", instance.Tags)
	}

	if len(instance.Parameters) > 0 {
		d.Set("parameters", instance.Parameters)
	}
	return nil

}

func resourceTsuruServiceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)
	serviceName := d.Get("service_name").(string)

	instanceData := tsuru.ServiceInstanceUpdateData{
		Teamowner: d.Get("owner").(string),
		Plan:      d.Get("plan").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		instanceData.Description = description.(string)
	}

	if tags, ok := d.GetOk("tags"); ok {
		instanceData.Tags = parseTags(tags)
	}

	if parameters, ok := d.GetOk("parameters"); ok {
		instanceData.Parameters = parseParameters(parameters)
	}

	updateOpts := &tsuru.InstanceUpdateOpts{
		ServiceInstanceUpdateData: optional.NewInterface(instanceData),
	}

	_, err := provider.TsuruClient.ServiceApi.InstanceUpdate(ctx, serviceName, name, updateOpts)
	if err != nil {
		return diag.Errorf("Could not update tsuru service instance: %q, err: %s", d.Id(), err.Error())
	}
	return nil
}

func resourceTsuruServiceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)
	serviceName := d.Get("service_name").(string)
	unbind := d.Get("unbind_on_delete").(bool)

	_, err := provider.TsuruClient.ServiceApi.InstanceDelete(ctx, serviceName, name, unbind)
	if err != nil {
		return diag.Errorf("Could not delete tsuru service instance, err: %s", err.Error())
	}

	return nil
}

func parseTags(data interface{}) []string {
	values := []string{}

	for _, item := range data.([]interface{}) {
		values = append(values, item.(string))
	}

	return values
}

func parseParameters(data interface{}) map[string]string {
	values := map[string]string{}

	for key, value := range data.(map[string]interface{}) {
		values[key] = value.(string)
	}

	return values
}

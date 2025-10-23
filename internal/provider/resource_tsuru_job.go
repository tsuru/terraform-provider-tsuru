// Copyright 2021 tsuru authors. All rights reserved.
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
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruJob() *schema.Resource {
	return &schema.Resource{
		Description:   "Tsuru Job",
		CreateContext: resourceTsuruJobCreate,
		UpdateContext: resourceTsuruJobUpdate,
		ReadContext:   resourceTsuruJobRead,
		DeleteContext: resourceTsuruJobDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceTsuruJobImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Job name",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Job description",
				Optional:    true,
			},
			"plan": {
				Type:        schema.TypeString,
				Description: "Plan",
				Required:    true,
			},
			"team_owner": {
				Type:        schema.TypeString,
				Description: "Job owner",
				Required:    true,
			},
			"pool": {
				Type:        schema.TypeString,
				Description: "The name of pool",
				Required:    true,
			},
			"cluster": {
				Type:        schema.TypeString,
				Description: "The name of cluster",
				Computed:    true,
			},
			"tags": {
				Type:        schema.TypeList,
				Description: "Tags",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"schedule": {
				Type:        schema.TypeString,
				Description: "Cron-like schedule for when the job should be triggered (keep empty for manual jobs)",
				Optional:    true,
			},

			"container": {
				Type:     schema.TypeList,
				MinItems: 0,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"command": {
							Type:        schema.TypeList,
							Description: "Command",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"active_deadline_seconds": {
				Type:        schema.TypeInt,
				Description: "Time a Job can run before its terminated. Defaults is 3600",
				Optional:    true,
			},

			"concurrency_policy": {
				Type:        schema.TypeString,
				Description: "Concurrency policy",
				Optional:    true,
			},
		},
	}
}

func resourceTsuruJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	job, err := inputJobFromResourceData(ctx, d, provider)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = provider.TsuruClient.JobApi.CreateJob(ctx, job)
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
		return diag.Errorf("unable to create job %s: %v", job.Name, err)
	}

	d.SetId(job.Name)

	return resourceTsuruJobRead(ctx, d, meta)
}

func resourceTsuruJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	job, err := inputJobFromResourceData(ctx, d, provider)
	if err != nil {
		return diag.FromErr(err)
	}
	jobName := d.Id()

	if d.HasChange("metadata") {
		old, new := d.GetChange("metadata")
		oldMetadata := metadataFromResourceData(old)
		if oldMetadata == nil {
			oldMetadata = &tsuru_client.Metadata{}
		}
		newMetadata := metadataFromResourceData(new)
		if newMetadata == nil {
			newMetadata = &tsuru_client.Metadata{}
		}

		job.Metadata = tsuru_client.Metadata{
			Annotations: markRemovedMetadataItemAsDeleted(oldMetadata.Annotations, newMetadata.Annotations),
			Labels:      markRemovedMetadataItemAsDeleted(oldMetadata.Labels, newMetadata.Labels),
		}
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := provider.TsuruClient.JobApi.UpdateJob(ctx, jobName, job)
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}

		defer resp.Body.Close()
		logTsuruStream(resp.Body)

		return nil
	})
	if err != nil {
		return diag.Errorf("unable to update job %s: %v", jobName, err)
	}

	return resourceTsuruJobRead(ctx, d, meta)
}

func resourceTsuruJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Id()

	job, _, err := provider.TsuruClient.JobApi.GetJob(ctx, name)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to read job %s: %v", name, err)
	}

	d.Set("name", name)
	d.Set("pool", job.Job.Pool)
	d.Set("cluster", job.Cluster)

	d.Set("plan", job.Job.Plan.Name)
	d.Set("team_owner", job.Job.TeamOwner)

	d.Set("container", flattenJobContainer(job.Job.Spec.Container))

	if job.Job.Spec.Manual {
		d.Set("schedule", "")
	} else {
		d.Set("schedule", job.Job.Spec.Schedule)
	}

	if job.Job.Description != "" {
		d.Set("description", job.Job.Description)
	}

	for key, value := range flattenJobSpec(job.Job.Spec) {
		d.Set(key, value)
	}

	d.Set("metadata", flattenMetadata(job.Job.Metadata))

	return nil
}

func resourceTsuruJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Id()

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := provider.TsuruClient.JobApi.DeleteJob(ctx, name)
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
		return diag.Errorf("unable to delete job %s: %v", name, err)
	}

	return nil
}

func resourceTsuruJobImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	provider := meta.(*tsuruProvider)

	job, _, err := provider.TsuruClient.JobApi.GetJob(ctx, d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("name", job.Job.Name)
	d.SetId(job.Job.Name)

	return []*schema.ResourceData{d}, nil
}

func inputJobFromResourceData(ctx context.Context, d *schema.ResourceData, provider *tsuruProvider) (tsuru_client.InputJob, error) {
	pool := d.Get("pool").(string)
	if err := validPool(ctx, provider, pool); err != nil {
		return tsuru_client.InputJob{}, err
	}

	plan := d.Get("plan").(string)
	if err := validPlan(ctx, provider, plan); err != nil {
		return tsuru_client.InputJob{}, err
	}

	tags := []string{}
	for _, item := range d.Get("tags").([]interface{}) {
		tags = append(tags, item.(string))
	}

	var container tsuru_client.InputJobContainer

	if m, ok := d.GetOk("container"); ok {
		container = jobContainerFromResourceData(m)
	}

	job := tsuru_client.InputJob{
		Name:      d.Get("name").(string),
		Pool:      pool,
		Plan:      plan,
		TeamOwner: d.Get("team_owner").(string),
		Tags:      tags,
		Container: container,
	}

	if m, ok := d.GetOk("metadata"); ok {
		metadata := metadataFromResourceData(m)
		if metadata != nil {
			job.Metadata = *metadata
		}
	}

	if desc, ok := d.GetOk("description"); ok {
		job.Description = desc.(string)
	}

	if schedule, ok := d.GetOk("schedule"); ok {
		job.Schedule = schedule.(string)
	} else {
		job.Manual = true
	}
	if concurrencyPolicyInterface, ok := d.GetOk("concurrency_policy"); ok {
		concurrencyPolicy := concurrencyPolicyInterface.(string)
		job.ConcurrencyPolicy = &concurrencyPolicy
	}

	if activeDeadLineSecondsInterface, ok := d.GetOk("active_deadline_seconds"); ok {
		activeDeadLineSeconds := int64(activeDeadLineSecondsInterface.(int))
		job.ActiveDeadlineSeconds = &activeDeadLineSeconds
	}

	return job, nil
}

func jobContainerFromResourceData(meta interface{}) tsuru_client.InputJobContainer {
	container := tsuru_client.InputJobContainer{}

	m := meta.([]interface{})
	if len(m) == 0 || m[0] == nil {
		return container
	}

	containerMap := m[0].(map[string]interface{})
	if v, ok := containerMap["image"]; ok {
		container.Image = v.(string)
	}

	if v, ok := containerMap["command"]; ok && len(v.([]interface{})) > 0 {
		container.Command = []string{}
		for _, value := range v.([]interface{}) {
			container.Command = append(container.Command, value.(string))
		}
	}

	return container
}

func flattenJobContainer(container tsuru.InputJobContainer) []interface{} {
	if container.Image == "" && len(container.Command) == 0 {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"image":   container.Image,
		"command": container.Command,
	}

	return []interface{}{m}
}

func flattenJobSpec(spec tsuru.JobSpec) map[string]any {
	m := map[string]any{}

	if spec.ConcurrencyPolicy != nil {
		m["concurrency_policy"] = spec.ConcurrencyPolicy
	}

	if spec.ActiveDeadlineSeconds != nil {
		m["active_deadline_seconds"] = *spec.ActiveDeadlineSeconds
	}

	return m
}

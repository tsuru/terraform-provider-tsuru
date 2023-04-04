// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Description: "When trigger the job",
				Optional:    true,
			},

			"container": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:     schema.TypeString,
							Required: true,
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

			"spec": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"completions": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"parallelism": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"active_deadline_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"backoff_limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
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

	_, err = provider.TsuruClient.JobApi.CreateJob(ctx, job)
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

	resp, err := provider.TsuruClient.JobApi.UpdateJob(ctx, jobName, job)
	if err != nil {
		return diag.Errorf("unable to update job %s: %v", jobName, err)
	}

	defer resp.Body.Close()
	logTsuruStream(resp.Body)

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
	d.Set("plan", job.Job.Plan.Name)
	d.Set("team_owner", job.Job.TeamOwner)

	if job.Job.Description != "" {
		d.Set("description", job.Job.Description)
	}

	annotations := map[string]interface{}{}
	if len(job.Job.Metadata.Annotations) > 0 {
		for _, annotation := range job.Job.Metadata.Annotations {
			annotations[annotation.Name] = annotation.Value
		}
	}

	labels := map[string]interface{}{}
	if len(job.Job.Metadata.Labels) > 0 {
		for _, label := range job.Job.Metadata.Labels {
			labels[label.Name] = label.Value
		}
	}

	if len(annotations) > 0 || len(labels) > 0 {
		d.Set("metadata", []map[string]interface{}{{
			"annotations": annotations,
			"labels":      labels,
		}})
	}

	return nil
}

func resourceTsuruJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)

	_, err := provider.TsuruClient.JobApi.DeleteJob(ctx, name)
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
	platform := d.Get("platform").(string)
	if err := validPlatform(ctx, provider, platform); err != nil {
		return tsuru_client.InputJob{}, err
	}

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

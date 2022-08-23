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
)

func resourceTsuruClusterPool() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource used to assign pools into clusters",
		CreateContext: resourceTsuruClusterPoolSet,
		ReadContext:   resourceTsuruClusterPoolRead,
		DeleteContext: resourceTsuruClusterPoolUnset,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"cluster": {
				Type:        schema.TypeString,
				Description: "The name of defined cluster",
				Required:    true,
				ForceNew:    true, // there is no way to update this field
			},
			"pool": {
				Type:        schema.TypeString,
				Description: "The name of defined pool",
				Required:    true,
				ForceNew:    true, // there is no way to update this field
			},
		},
	}
}

func resourceTsuruClusterPoolSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	clusterName, poolName := getClusterAndPoolFromResource(d)

	cluster, _, err := provider.TsuruClient.ClusterApi.ClusterInfo(ctx, clusterName)
	if err != nil {
		return diag.Errorf("Could not read tsuru cluster: %q, err: %s", clusterName, err.Error())
	}
	found := false
	for _, foundPool := range cluster.Pools {
		if foundPool == poolName {
			found = true
		}
	}
	if !found {
		cluster.Pools = append(cluster.Pools, poolName)
	}

	_, err = provider.TsuruClient.ClusterApi.ClusterUpdate(ctx, clusterName, cluster)
	if err != nil {
		return diag.Errorf("Could not update tsuru cluster: %q, err: %s", clusterName, err.Error())
	}

	d.SetId(clusterName + "/" + poolName)

	return resourceTsuruClusterPoolRead(ctx, d, meta)
}

func resourceTsuruClusterPoolUnset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	clusterName, poolName := getClusterAndPoolFromResource(d)

	cluster, _, err := provider.TsuruClient.ClusterApi.ClusterInfo(ctx, clusterName)
	if err != nil {
		return diag.Errorf("Could not read tsuru cluster: %q, err: %s", clusterName, err.Error())
	}

	cluster.Pools = removeItemFromSlice(cluster.Pools, poolName)

	_, err = provider.TsuruClient.ClusterApi.ClusterUpdate(ctx, clusterName, cluster)
	if err != nil {
		return diag.Errorf("Could not update tsuru cluster: %q, err: %s", clusterName, err.Error())
	}

	return nil
}

func resourceTsuruClusterPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	parts := strings.SplitN(d.Id(), "/", 2)

	clusterName := parts[0]
	poolName := parts[1]

	cluster, _, err := provider.TsuruClient.ClusterApi.ClusterInfo(ctx, clusterName)

	if isNotFoundError(err) {
		setClusterPoolNotFound(d)
		return nil
	}

	if err != nil {
		return diag.Errorf("Could not read tsuru cluster: %q, err: %s", clusterName, err.Error())
	}

	for _, foundPool := range cluster.Pools {
		if foundPool == poolName {
			d.Set("cluster", clusterName)
			d.Set("pool", poolName)
			return nil
		}
	}
	setClusterPoolNotFound(d)
	return nil
}

func getClusterAndPoolFromResource(d *schema.ResourceData) (cluster, pool string) {
	id := d.Id()
	cluster = d.Get("cluster").(string)
	pool = d.Get("pool").(string)
	if id != "" {
		parts := strings.SplitN(d.Id(), "/", 2)
		cluster = parts[0]
		pool = parts[1]
	}
	return
}

func setClusterPoolNotFound(d *schema.ResourceData) {
	d.Set("cluster", "")
	d.Set("pool", "")
}

func removeItemFromSlice(s []string, item string) []string {
	n := 0
	for _, foundItem := range s {
		if foundItem != item {
			s[n] = foundItem
			n++
		}
	}
	return s[:n]
}

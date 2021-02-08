package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruClusterCreate,
		ReadContext:   resourceTsuruClusterRead,
		UpdateContext: resourceTsuruClusterUpdate,
		DeleteContext: resourceTsuruClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique name of cluster",
			},
			"addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of addresses to tsuru API connect in",
			},
			"tsuru_provisioner": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "kubernetes",
				Description: "Provisioner of cluster",
			},
			"ca_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "CA certificate to connect in cluster, must be enconded in base64",
			},
			"client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client key to connect in cluster, must be enconded in base64",
			},
			"custom_data": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Key/value to store additional config",
			},
			"client_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client certificate to connect in cluster, must be enconded in base64",
			},
			"default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether true, the cluster is the default for all pools",
			},
			"initial_pools": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Name of initial pools, required when is no default cluster",
				Optional:    true,
			},
			"pools": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Name of pools that belongs to the cluster",
			},
		},
	}
}

func resourceTsuruClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	initialPools := []string{}

	for _, item := range d.Get("initial_pools").([]interface{}) {
		initialPools = append(initialPools, item.(string))
	}

	cluster := clusterFromResourceData(d)
	cluster.Pools = initialPools

	_, err := provider.TsuruClient.ClusterApi.ClusterCreate(ctx, cluster)

	if err != nil {
		return diag.Errorf("Could not create tsuru cluster, err : %s", err.Error())
	}

	d.SetId(cluster.Name)

	return resourceTsuruClusterRead(ctx, d, meta)
}

func resourceTsuruClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	clusterName := d.Id()

	cluster, _, err := provider.TsuruClient.ClusterApi.ClusterInfo(ctx, clusterName)

	if err != nil {
		return diag.Errorf("Could not read tsuru cluster, err : %s", err.Error())
	}

	d.Set("addresses", cluster.Addresses)
	d.Set("name", cluster.Name)
	d.Set("tsuru_provisioner", cluster.Provisioner)
	d.Set("ca_cert", string(cluster.Cacert))
	d.Set("client_key", string(cluster.Clientkey))
	d.Set("client_cert", string(cluster.Clientcert))
	d.Set("pools", cluster.Pools)

	return nil
}

func resourceTsuruClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	cluster := clusterFromResourceData(d)

	existentCluster, _, err := provider.TsuruClient.ClusterApi.ClusterInfo(ctx, d.Id())

	if err != nil {
		return diag.Errorf("Could not read tsuru cluster, err : %s", err.Error())
	}
	cluster.Pools = existentCluster.Pools

	_, err = provider.TsuruClient.ClusterApi.ClusterUpdate(ctx, d.Id(), cluster)

	if err != nil {
		return diag.Errorf("Could not update tsuru cluster: %q, err: %s", d.Id(), err.Error())
	}
	return nil
}

func resourceTsuruClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	_, err := provider.TsuruClient.ClusterApi.ClusterDelete(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Could not delete tsuru cluster, err: %s", err.Error())
	}
	return nil
}

func clusterFromResourceData(d *schema.ResourceData) tsuru.Cluster {
	addresses := []string{}
	customData := make(map[string]string)

	for _, item := range d.Get("addresses").([]interface{}) {
		addresses = append(addresses, item.(string))
	}

	for key, value := range d.Get("custom_data").(map[string]interface{}) {
		customData[key] = value.(string)
	}

	return tsuru.Cluster{
		Name:        d.Get("name").(string),
		Addresses:   addresses,
		Provisioner: d.Get("tsuru_provisioner").(string),
		Cacert:      []byte(d.Get("ca_cert").(string)),
		Clientcert:  []byte(d.Get("client_cert").(string)),
		Clientkey:   []byte(d.Get("client_key").(string)),
		CustomData:  customData,
		Default:     d.Get("default").(bool),
	}
}

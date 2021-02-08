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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"tsuru_provisioner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "kubernetes",
			},
			"ca_cert": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"client_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_data": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"client_cert": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceTsuruClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	addresses := []string{}
	customData := make(map[string]string)

	for _, item := range d.Get("addresses").([]interface{}) {
		addresses = append(addresses, item.(string))
	}

	for key, value := range d.Get("custom_data").(map[string]interface{}) {
		customData[key] = value.(string)
	}

	name := d.Get("name").(string)

	cluster := tsuru.Cluster{
		Name:        name,
		Addresses:   addresses,
		Provisioner: d.Get("tsuru_provisioner").(string),
		Cacert:      []byte(d.Get("ca_cert").(string)),
		Clientcert:  []byte(d.Get("client_cert").(string)),
		Clientkey:   []byte(d.Get("client_key").(string)),
		Pools:       []string{},
		CustomData:  customData,
		Default:     d.Get("default").(bool),
	}

	_, err := provider.TsuruClient.ClusterApi.ClusterCreate(ctx, cluster)

	if err != nil {
		return diag.Errorf("Could not create tsuru cluster, err : %s", err.Error())
	}

	d.SetId(name)

	return nil
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

	return nil
}

func resourceTsuruClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	_, err := provider.TsuruClient.ClusterApi.ClusterUpdate(ctx, d.Id(), tsuru.Cluster{
		Default:     d.Get("default").(bool),
		Provisioner: d.Get("tsuru_provisioner").(string),
	})

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

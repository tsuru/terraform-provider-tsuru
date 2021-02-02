package tsuru

import (
	"context"
	"strings"

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
			"nome": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"addresses": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"tsuru_tsuru_provisioner": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ca_certificate": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"client_key": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"custom_data": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"client_certificate": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"default": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceTsuruClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	addresses := []string{}
	CaCertificate := []byte{}
	ClientCertificate := []byte{}
	ClientKey := []byte{}
	CustomData := make(map[string]string)

	for _, item := range d.Get("addresses").(*schema.Set).List() {
		addresses = append(addresses, item.(string))
	}

	for _, item := range d.Get("ca_certificate").(*schema.Set).List() {
		CaCertificate = append(CaCertificate, item.(byte))
	}

	for c, item := range d.Get("custom_data").(map[string]interface{}) {
		CustomData[c] = item.(string)
	}

	for _, item := range d.Get("client_certificate").(*schema.Set).List() {
		ClientCertificate = append(ClientCertificate, item.(byte))
	}

	for _, item := range d.Get("client_key").(*schema.Set).List() {
		ClientKey = append(ClientKey, item.(byte))
	}
	_, err := provider.TsuruClient.ClusterApi.ClusterCreate(ctx, tsuru.Cluster{
		Name:        d.Get("nome").(string),
		Addresses:   addresses,
		Provisioner: d.Get("tsuru_provisioner").(string),
		Cacert:      CaCertificate,
		CustomData:  CustomData,
		Clientcert:  ClientCertificate,
		Clientkey:   ClientKey,
	})
	if err != nil {
		return diag.Errorf("Could not create tsuru cluster, err : %s", err.Error())
	}
	return nil
}

func resourceTsuruClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	parts := strings.SplitN(d.Id(), "/", 2)

	ClusterName := parts[0]

	cluster, _, err := provider.TsuruClient.ClusterApi.ClusterInfo(ctx, ClusterName)

	if err != nil {
		return diag.Errorf("Could not read tsuru cluster, err : %s", err.Error())
	}

	d.Set("addresses", cluster.Addresses)
	d.Set("name", cluster.Name)
	d.Set("tsuru_provisioner", cluster.Provisioner)
	d.Set("ca_certificate", cluster.Cacert)
	d.Set("client_key", cluster.Clientkey)
	d.Set("client_certificate", cluster.Clientcert)

	return nil
}

func resourceTsuruClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	if d.HasChange("name") {
		return diag.Errorf("teste")
	}

	_, err := provider.TsuruClient.ClusterApi.ClusterUpdate(ctx, d.Id(), tsuru.Cluster{
		Default: d.Get("default").(bool),
	})

	if err != nil {
		return diag.Errorf("Could not update tsuru cluster: %q, err: %s", d.Id(), err.Error())
	}
	return nil
}

func resourceTsuruClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	//id = d.Get("name").(string) + "/" + d.Get("tsuru_provisioner").(string)

	_, err := provider.TsuruClient.ClusterApi.ClusterDelete(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Could not delete tsuru cluster, err: %s", err.Error())
	}
	return nil
}

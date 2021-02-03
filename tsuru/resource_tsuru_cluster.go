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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"addresses": {
				Type:    schema.TypeString,
				Default: "testcarlos",
				//Required: true,
				Optional: true,
				ForceNew: true,
			},
			"tsuru_provisioner": {
				Type: schema.TypeString,
				//Required: true,
				Optional: true,
				ForceNew: true,
				Default:  "kubernetes",
			},
			"ca_certificate": {
				Type:    schema.TypeString,
				Default: "testcarlos",
				//Required: true,
				Optional: true,
				ForceNew: true,
			},
			"client_key": {
				Type:    schema.TypeString,
				Default: "testcarlos",
				//Required: true,
				Optional: true,
				ForceNew: true,
			},
			"custom_data": {
				Type: schema.TypeString,
				//Required: false,
				Optional: true,
				ForceNew: true,
			},
			"client_certificate": {
				Type:    schema.TypeString,
				Default: "testcarlos",
				//Required: true,
				Optional: true,
				ForceNew: true,
			},
			"default": {
				Type: schema.TypeBool,
				//Required: true,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceTsuruClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	//addresses := []string{}
	//CaCertificate := []byte{}
	//ClientCertificate := []byte{}
	//ClientKey := []byte{}
	CustomData := make(map[string]string)

	//for _, item := range d.Get("addresses").(*schema.Set).List() {
	//	addresses = append(addresses, item.(string))
	//}

	//for _, item := range d.Get("ca_certificate").(*schema.Set).List() {
	//	CaCertificate = append(CaCertificate, item.(byte))
	//}

	//for c, item := range d.Get("custom_data").(map[string]interface{}) {
	//	CustomData[c] = item.(string)
	//}

	//for _, item := range d.Get("client_certificate").(*schema.Set).List() {
	//	ClientCertificate = append(ClientCertificate, item.(byte))
	//}

	//for _, item := range d.Get("client_key").(*schema.Set).List() {
	//	ClientKey = append(ClientKey, item.(byte))
	//}

	cluster := tsuru.Cluster{
		Name:        d.Get("name").(string),
		Addresses:   []string{d.Get("addresses").(string)},
		Provisioner: d.Get("tsuru_provisioner").(string),
		Cacert:      []byte(d.Get("ca_certificate").(string)),
		Clientcert:  []byte(d.Get("client_certificate").(string)),
		Clientkey:   []byte(d.Get("client_key").(string)),
		Pools:       []string{},
		CustomData:  CustomData,
		//CreateData:  map[string]string{},
		Default: false,
	}

	_, err := provider.TsuruClient.ClusterApi.ClusterCreate(ctx, cluster)

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

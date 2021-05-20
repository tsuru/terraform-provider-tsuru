package tsuru

import (
	"context"
	"reflect"

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
			"local": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether true, the cluster is auth by the local credentials of kubernetes",
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
			"kube_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster": kubeConfigClusterSchema(),
						"user":    kubeConfigUserSchema(),
					},
				},
			},
			"http_proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client HTTP proxy",
			},
		},
	}
}

func kubeConfigUserSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"auth_provider": kubeConfigAuthProviderSchema(),
				"exec":          kubeConfigAuthExecSchema(),
				"client_certificate_data": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"client_key_data": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"token": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"username": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"password": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
			},
		},
	}
}

func kubeConfigAuthProviderSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"config": {
					Type:     schema.TypeMap,
					Optional: true,
				},
			},
		},
	}
}

func kubeConfigAuthExecSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"api_version": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"command": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"args": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"env": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func kubeConfigClusterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"server": {
					Type:     schema.TypeString,
					Required: true,
				},
				"tls_server_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"insecure_skip_tls_verify": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"certificate_authority_data": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
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
	d.Set("default", cluster.Default)
	d.Set("local", cluster.Local)
	d.Set("pools", cluster.Pools)
	d.Set("http_proxy", cluster.HttpProxy)
	d.Set("kube_config", flattenKubeConfig(cluster.KubeConfig))

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

	var kubeConfig tsuru.ClusterKubeConfig

	if data, ok := d.GetOk("kube_config"); ok {
		dataArray := data.([]interface{})
		kubeConfigRaw := dataArray[0].(map[string]interface{})

		dataArray = kubeConfigRaw["cluster"].([]interface{})
		serverRaw, _ := dataArray[0].(map[string]interface{})

		dataArray = kubeConfigRaw["user"].([]interface{})
		userRaw, _ := dataArray[0].(map[string]interface{})

		dataArray = userRaw["auth_provider"].([]interface{})
		authProviderRaw, _ := dataArray[0].(map[string]interface{})

		kubeConfig.Cluster.Server = serverRaw["server"].(string)
		kubeConfig.Cluster.CertificateAuthorityData = serverRaw["certificate_authority_data"].(string)
		kubeConfig.Cluster.TlsServerName = serverRaw["tls_server_name"].(string)
		kubeConfig.Cluster.InsecureSkipTlsVerify = serverRaw["insecure_skip_tls_verify"].(bool)

		kubeConfig.User.AuthProvider.Name = authProviderRaw["name"].(string)
		kubeConfig.User.AuthProvider.Config = map[string]string{}

		configRaw, _ := authProviderRaw["config"].(map[string]interface{})

		for key, item := range configRaw {
			kubeConfig.User.AuthProvider.Config[key] = item.(string)
		}

		kubeConfig.User.ClientCertificateData = userRaw["client_certificate_data"].(string)
		kubeConfig.User.ClientKeyData = userRaw["client_key_data"].(string)
		kubeConfig.User.Token = userRaw["token"].(string)
		kubeConfig.User.Username = userRaw["username"].(string)
		kubeConfig.User.Password = userRaw["password"].(string)

		dataArray = userRaw["exec"].([]interface{})
		execRaw, _ := dataArray[0].(map[string]interface{})

		kubeConfig.User.Exec.ApiVersion = execRaw["api_version"].(string)
		kubeConfig.User.Exec.Command = execRaw["command"].(string)
		kubeConfig.User.Exec.Args = []string{}
		kubeConfig.User.Exec.Env = []tsuru.ClusterKubeConfigUserExecEnv{}

		for _, arg := range execRaw["args"].([]interface{}) {
			kubeConfig.User.Exec.Args = append(kubeConfig.User.Exec.Args, arg.(string))
		}

		for _, arg := range execRaw["env"].([]interface{}) {

			m := arg.(map[string]interface{})
			kubeConfig.User.Exec.Env = append(kubeConfig.User.Exec.Env, tsuru.ClusterKubeConfigUserExecEnv{
				Name:  m["name"].(string),
				Value: m["value"].(string),
			})
		}

	}

	clusterDefault := false
	if value, ok := d.GetOk("default"); ok {
		clusterDefault = value.(bool)
	}

	clusterLocal := false
	if value, ok := d.GetOk("local"); ok {
		clusterLocal = value.(bool)
	}

	return tsuru.Cluster{
		Name:        d.Get("name").(string),
		Addresses:   addresses,
		Provisioner: d.Get("tsuru_provisioner").(string),
		Cacert:      []byte(d.Get("ca_cert").(string)),
		Clientcert:  []byte(d.Get("client_cert").(string)),
		Clientkey:   []byte(d.Get("client_key").(string)),
		CustomData:  customData,
		Default:     clusterDefault,
		Local:       clusterLocal,
		KubeConfig:  kubeConfig,
		HttpProxy:   d.Get("http_proxy").(string),
	}
}

func flattenKubeConfig(kubeconfig tsuru.ClusterKubeConfig) []interface{} {
	if reflect.DeepEqual(kubeconfig, tsuru.ClusterKubeConfig{}) {
		return []interface{}{}
	}

	result := map[string]interface{}{
		"cluster": flattenKubeConfigCluster(kubeconfig.Cluster),
		"user":    flattenKubeConfigUser(kubeconfig.User),
	}
	return []interface{}{result}
}

func flattenKubeConfigCluster(cluster tsuru.ClusterKubeConfigCluster) []interface{} {
	result := map[string]interface{}{
		"server":                     cluster.Server,
		"certificate_authority_data": cluster.CertificateAuthorityData,
		"tls_server_name":            cluster.TlsServerName,
		"insecure_skip_tls_verify":   cluster.InsecureSkipTlsVerify,
	}
	return []interface{}{result}
}

func flattenKubeConfigUser(user tsuru.ClusterKubeConfigUser) []interface{} {
	result := map[string]interface{}{
		"client_certificate_data": user.ClientCertificateData,
		"client_key_data":         user.ClientKeyData,
		"username":                user.Username,
		"password":                user.Password,
		"token":                   user.Token,
		"auth_provider":           flattenAuthProvider(user.AuthProvider),
		"exec":                    flattenExec(user.Exec),
	}
	return []interface{}{result}
}

func flattenAuthProvider(authProvider tsuru.ClusterKubeConfigUserAuthprovider) []interface{} {
	result := map[string]interface{}{
		"name":   authProvider.Name,
		"config": authProvider.Config,
	}
	return []interface{}{result}
}

func flattenExec(exec tsuru.ClusterKubeConfigUserExec) []interface{} {
	result := map[string]interface{}{
		"args":        exec.Args,
		"api_version": exec.ApiVersion,
		"command":     exec.Command,
		"env":         flattenEnvs(exec.Env),
	}
	return []interface{}{result}
}

func flattenEnvs(envs []tsuru.ClusterKubeConfigUserExecEnv) []interface{} {
	result := []interface{}{}

	for _, item := range envs {
		result = append(result, map[string]string{
			"name":  item.Name,
			"value": item.Value,
		})
	}

	return result
}

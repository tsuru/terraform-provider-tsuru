package provider

import (
	"context"
	"time"

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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
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
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
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
				"interactive_mode": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "Never",
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
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
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
	d.Set("custom_data", cluster.CustomData)

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

	return resourceTsuruClusterRead(ctx, d, meta)
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

	kubeConfig := kubeConfigFromResourceData(d.Get("kube_config"))

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

func kubeConfigFromResourceData(data interface{}) *tsuru.ClusterKubeConfig {
	dataArray := data.([]interface{})
	if len(dataArray) == 0 {
		return nil
	}
	kubeConfig := tsuru.ClusterKubeConfig{}

	kubeConfigRaw := dataArray[0].(map[string]interface{})

	kubeConfig.Cluster = kubeConfigClusterFromResourceData(kubeConfigRaw["cluster"])
	kubeConfig.User = kubeConfigUserFromResourceData(kubeConfigRaw["user"])

	return &kubeConfig
}

func kubeConfigClusterFromResourceData(data interface{}) tsuru.ClusterKubeConfigCluster {
	cluster := tsuru.ClusterKubeConfigCluster{}
	dataArray := data.([]interface{})

	if len(dataArray) == 0 {
		return cluster
	}

	clusterRaw, _ := dataArray[0].(map[string]interface{})

	cluster.Server = clusterRaw["server"].(string)
	cluster.CertificateAuthorityData = clusterRaw["certificate_authority_data"].(string)
	cluster.TlsServerName = clusterRaw["tls_server_name"].(string)
	cluster.InsecureSkipTlsVerify = clusterRaw["insecure_skip_tls_verify"].(bool)

	return cluster
}

func kubeConfigUserFromResourceData(data interface{}) tsuru.ClusterKubeConfigUser {
	user := tsuru.ClusterKubeConfigUser{}

	dataArray := data.([]interface{})

	if len(dataArray) == 0 {
		return user
	}

	userRaw, _ := dataArray[0].(map[string]interface{})

	user.ClientCertificateData = userRaw["client_certificate_data"].(string)
	user.ClientKeyData = userRaw["client_key_data"].(string)
	user.Token = userRaw["token"].(string)
	user.Username = userRaw["username"].(string)
	user.Password = userRaw["password"].(string)

	user.AuthProvider = authProviderFromResourceData(userRaw["auth_provider"])
	user.Exec = execFromResourceData(userRaw["exec"])

	return user
}

func authProviderFromResourceData(data interface{}) *tsuru.ClusterKubeConfigUserAuthprovider {
	authProvider := &tsuru.ClusterKubeConfigUserAuthprovider{}

	dataArray := data.([]interface{})
	if len(dataArray) == 0 {
		return nil
	}

	authProviderRaw, _ := dataArray[0].(map[string]interface{})

	authProvider.Name = authProviderRaw["name"].(string)

	configRaw, _ := authProviderRaw["config"].(map[string]interface{})
	if len(configRaw) > 0 {
		authProvider.Config = map[string]string{}

		for key, item := range configRaw {
			authProvider.Config[key] = item.(string)
		}
	}

	return authProvider
}

func execFromResourceData(data interface{}) *tsuru.ClusterKubeConfigUserExec {

	dataArray := data.([]interface{})
	if len(dataArray) == 0 {
		return nil
	}
	exec := &tsuru.ClusterKubeConfigUserExec{}

	execRaw, _ := dataArray[0].(map[string]interface{})

	exec.ApiVersion, _ = execRaw["api_version"].(string)
	exec.Command, _ = execRaw["command"].(string)
	exec.InteractiveMode = execRaw["interactive_mode"].(string)

	argsRaw, _ := execRaw["args"].([]interface{})
	for _, arg := range argsRaw {
		exec.Args = append(exec.Args, arg.(string))
	}

	envRaw, _ := execRaw["env"].([]interface{})
	for _, arg := range envRaw {
		m := arg.(map[string]interface{})
		exec.Env = append(exec.Env, tsuru.ClusterKubeConfigUserExecEnv{
			Name:  m["name"].(string),
			Value: m["value"].(string),
		})
	}
	return exec
}

func flattenKubeConfig(kubeconfig *tsuru.ClusterKubeConfig) []interface{} {
	if kubeconfig == nil {
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
	}
	if user.AuthProvider != nil {
		result["auth_provider"] = flattenAuthProvider(user.AuthProvider)
	}
	if user.Exec != nil {
		result["exec"] = flattenExec(user.Exec)
	}

	return []interface{}{result}
}

func flattenAuthProvider(authProvider *tsuru.ClusterKubeConfigUserAuthprovider) []interface{} {
	result := map[string]interface{}{
		"name":   authProvider.Name,
		"config": authProvider.Config,
	}

	return []interface{}{result}
}

func flattenExec(exec *tsuru.ClusterKubeConfigUserExec) []interface{} {
	result := map[string]interface{}{
		"args":             exec.Args,
		"api_version":      exec.ApiVersion,
		"command":          exec.Command,
		"interactive_mode": exec.InteractiveMode,
		"env":              flattenEnvs(exec.Env),
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

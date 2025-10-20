package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruServiceBroker() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruServiceBrokerCreate,
		ReadContext:   resourceTsuruServiceBrokerRead,
		UpdateContext: resourceTsuruServiceBrokerUpdate,
		DeleteContext: resourceTsuruServiceBrokerDelete,
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
				Description: "Unique name for the service broker",
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				Type:         schema.TypeString,
				Description:  "URL endpoint of the service broker",
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"config": {
				Type:        schema.TypeList,
				Description: "Configuration for the service broker",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"insecure": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"context": {
							Type:        schema.TypeMap,
							Description: "Context information passed to the broker",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"cache_expiration_seconds": {
							Type:         schema.TypeInt,
							Description:  "Cache expiration duration in seconds",
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"auth_config": {
							Type:        schema.TypeList,
							Description: "Authentication configuration for the broker",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"basic_auth_config": {
										Type:        schema.TypeList,
										Description: "Basic authentication configuration",
										Optional:    true,
										MaxItems:    1,
										ConflictsWith: []string{
											"config.0.auth_config.0.bearer_config",
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:        schema.TypeString,
													Description: "Username for basic authentication",
													Required:    true,
												},
												"password": {
													Type:        schema.TypeString,
													Description: "Password for basic authentication",
													Required:    true,
													Sensitive:   true,
												},
											},
										},
									},
									"bearer_config": {
										Type:        schema.TypeList,
										Description: "Bearer token authentication configuration",
										Optional:    true,
										MaxItems:    1,
										ConflictsWith: []string{
											"config.0.auth_config.0.basic_auth_config",
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"token": {
													Type:        schema.TypeString,
													Description: "Bearer token for authentication",
													Required:    true,
													Sensitive:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceTsuruServiceBrokerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("name").(string)
	url := d.Get("url").(string)

	broker := tsuru.ServiceBroker{
		Name: name,
		URL:  url,
	}

	if v, ok := d.GetOk("config"); ok {
		broker.Config = *expandServiceBrokerConfig(v.([]interface{}))
	}

	_, err := provider.TsuruClient.ServiceApi.ServiceBrokerCreate(ctx, broker)

	if err != nil {
		return diag.Errorf("Could not create tsuru service broker, err : %s", err.Error())
	}

	d.SetId(name)

	return resourceTsuruServiceBrokerRead(ctx, d, meta)
}

func resourceTsuruServiceBrokerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	brokerId := d.Id()

	brokers, _, err := provider.TsuruClient.ServiceApi.ServiceBrokerList(ctx)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("could not read tsuru service broker: %v", err)
	}

	broker := new(tsuru.ServiceBroker)
	for _, b := range brokers.Brokers {
		if b.Name == brokerId {
			broker = &b
			break
		}
	}

	if broker == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", broker.Name)
	d.Set("url", broker.URL)

	if &broker.Config != nil {
		if err := d.Set("config", flattenServiceBrokerConfig(&broker.Config)); err != nil {
			return diag.Errorf("could not set config: %v", err)
		}
	}

	return nil
}

func expandServiceBrokerConfig(config []interface{}) *tsuru.ServiceBrokerConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	cfg := config[0].(map[string]interface{})
	brokerConfig := &tsuru.ServiceBrokerConfig{}

	if v, ok := cfg["insecure"].(bool); ok {
		brokerConfig.Insecure = v
	}

	if v, ok := cfg["context"].(map[string]string); ok && len(v) > 0 {
		context := make(map[string]string)
		for key, val := range v {
			context[key] = val
		}
		brokerConfig.Context = context
	}

	if v, ok := cfg["cache_expiration_seconds"].(int); ok && v > 0 {
		expiration := int32(v)
		brokerConfig.CacheExpirationSeconds = expiration
	}

	if authConfigList, ok := cfg["auth_config"].([]interface{}); ok && len(authConfigList) > 0 {
		authCfg := authConfigList[0].(map[string]interface{})
		brokerConfig.AuthConfig = *expandAuthConfig(authCfg)
	}

	return brokerConfig
}

func expandAuthConfig(inputAuthConfig map[string]interface{}) *tsuru.ServiceBrokerConfigAuthConfig {
	authConfig := &tsuru.ServiceBrokerConfigAuthConfig{}

	if basicAuthMap, ok := inputAuthConfig["basic_auth_config"].([]interface{}); ok && len(basicAuthMap) > 0 {
		basicAuth := basicAuthMap[0].(map[string]interface{})
		authConfig.BasicAuthConfig = tsuru.ServiceBrokerConfigAuthConfigBasicAuthConfig{
			Username: basicAuth["username"].(string),
			Password: basicAuth["password"].(string),
		}
	}

	if bearerMap, ok := inputAuthConfig["bearer_config"].([]interface{}); ok && len(bearerMap) > 0 {
		bearer := bearerMap[0].(map[string]interface{})
		authConfig.BearerConfig = tsuru.ServiceBrokerConfigAuthConfigBearerConfig{
			Token: bearer["token"].(string),
		}
	}

	return authConfig
}

func flattenServiceBrokerConfig(config *tsuru.ServiceBrokerConfig) []interface{} {
	if config == nil {
		return nil
	}

	cfg := make(map[string]interface{})

	cfg["insecure"] = config.Insecure

	if config.Context != nil && len(config.Context) > 0 {
		contextMap := make(map[string]interface{})
		for key, val := range config.Context {
			contextMap[key] = fmt.Sprintf("%v", val)
		}
		cfg["context"] = contextMap
	}

	if &config.CacheExpirationSeconds != nil {
		cfg["cache_expiration_seconds"] = int(config.CacheExpirationSeconds)
	}

	if &config.AuthConfig != nil {
		cfg["auth_config"] = flattenAuthConfig(&config.AuthConfig)
	}

	return []interface{}{cfg}
}

func flattenAuthConfig(authConfig *tsuru.ServiceBrokerConfigAuthConfig) []interface{} {
	if authConfig == nil {
		return nil
	}

	auth := make(map[string]interface{})

	if &authConfig.BasicAuthConfig != nil {
		basicAuth := map[string]interface{}{
			"username": authConfig.BasicAuthConfig.Username,
			"password": authConfig.BasicAuthConfig.Password,
		}
		auth["basic_auth_config"] = []interface{}{basicAuth}
	}

	if &authConfig.BearerConfig != nil {
		bearer := map[string]interface{}{
			"token": authConfig.BearerConfig.Token,
		}
		auth["bearer_config"] = []interface{}{bearer}
	}

	return []interface{}{auth}
}

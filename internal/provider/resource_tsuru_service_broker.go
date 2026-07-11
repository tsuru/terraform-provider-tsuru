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

	if config, ok := d.GetOk("config"); ok {
		broker.Config = *expandServiceBrokerConfig(config.([]interface{}))
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
		return diag.Errorf("could not read tsuru service broker: %v", err)
	}

	broker := new(tsuru.ServiceBroker)
	for _, b := range brokers.Brokers {
		if b.Name == brokerId {
			broker = &b
			break
		}
	}

	if broker == nil || broker.Name == "" {
		d.SetId("")
		return nil
	}

	d.Set("name", broker.Name)
	d.Set("url", broker.URL)

	if err := d.Set("config", flattenServiceBrokerConfig(&broker.Config)); err != nil {
		return diag.Errorf("could not set config: %v", err)
	}

	return nil
}

func resourceTsuruServiceBrokerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	_, err := provider.TsuruClient.ServiceApi.ServiceBrokerUpdate(ctx, name, broker)
	if err != nil {
		return diag.Errorf("Could not update tsuru service broker, err: %s", err.Error())
	}

	return resourceTsuruServiceBrokerRead(ctx, d, meta)
}

func resourceTsuruServiceBrokerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()

	_, err := provider.TsuruClient.ServiceApi.ServiceBrokerDelete(ctx, name)
	if err != nil {
		return diag.Errorf("Could not delete tsuru service broker: %s", err.Error())
	}

	return nil
}

func expandServiceBrokerConfig(config []interface{}) *tsuru.ServiceBrokerConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	indexedConfig := config[0].(map[string]interface{})
	brokerConfig := &tsuru.ServiceBrokerConfig{}

	if insecure, ok := indexedConfig["insecure"].(bool); ok {
		brokerConfig.Insecure = insecure
	}

	if contexts, ok := indexedConfig["context"].(map[string]interface{}); ok && len(contexts) > 0 {
		context := make(map[string]string)
		for key, val := range contexts {
			context[key] = fmt.Sprint(val)
		}
		brokerConfig.Context = context
	}

	if cacheInSeconds, ok := indexedConfig["cache_expiration_seconds"].(int); ok && cacheInSeconds > 0 {
		expiration := int32(cacheInSeconds)
		brokerConfig.CacheExpirationSeconds = expiration
	}

	if authConfigList, ok := indexedConfig["auth_config"].([]interface{}); ok && len(authConfigList) > 0 {
		authCfg := authConfigList[0].(map[string]interface{})
		brokerConfig.AuthConfig = *expandAuthConfig(authCfg)
	}

	return brokerConfig
}

func expandAuthConfig(inputAuthConfig map[string]interface{}) *tsuru.ServiceBrokerConfigAuthConfig {
	authConfig := &tsuru.ServiceBrokerConfigAuthConfig{}

	if basicAuthConfig, ok := inputAuthConfig["basic_auth_config"].([]interface{}); ok && len(basicAuthConfig) > 0 {
		indexedBasicAuth := basicAuthConfig[0].(map[string]interface{})
		authConfig.BasicAuthConfig = tsuru.ServiceBrokerConfigAuthConfigBasicAuthConfig{
			Username: indexedBasicAuth["username"].(string),
			Password: indexedBasicAuth["password"].(string),
		}
	}

	if bearerConfig, ok := inputAuthConfig["bearer_config"].([]interface{}); ok && len(bearerConfig) > 0 {
		indexedBearer := bearerConfig[0].(map[string]interface{})
		authConfig.BearerConfig = tsuru.ServiceBrokerConfigAuthConfigBearerConfig{
			Token: indexedBearer["token"].(string),
		}
	}

	return authConfig
}

func flattenServiceBrokerConfig(config *tsuru.ServiceBrokerConfig) []interface{} {
	if config == nil {
		return nil
	}

	outputConfig := make(map[string]interface{})

	outputConfig["insecure"] = config.Insecure

	if len(config.Context) > 0 {
		contextMap := make(map[string]interface{})
		for key, val := range config.Context {
			contextMap[key] = fmt.Sprintf("%v", val)
		}
		outputConfig["context"] = contextMap
	}

	if config.CacheExpirationSeconds > 0 {
		outputConfig["cache_expiration_seconds"] = int(config.CacheExpirationSeconds)
	}

	if config.AuthConfig != (tsuru.ServiceBrokerConfigAuthConfig{}) {
		outputConfig["auth_config"] = flattenAuthConfig(&config.AuthConfig)
	}

	return []interface{}{outputConfig}
}

func flattenAuthConfig(authConfig *tsuru.ServiceBrokerConfigAuthConfig) []interface{} {
	if authConfig == nil {
		return nil
	}

	auth := make(map[string]interface{})

	if authConfig.BasicAuthConfig != (tsuru.ServiceBrokerConfigAuthConfigBasicAuthConfig{}) {
		basicAuth := map[string]interface{}{
			"username": authConfig.BasicAuthConfig.Username,
			"password": authConfig.BasicAuthConfig.Password,
		}
		auth["basic_auth_config"] = []interface{}{basicAuth}
	}

	if authConfig.BearerConfig != (tsuru.ServiceBrokerConfigAuthConfigBearerConfig{}) {
		bearer := map[string]interface{}{
			"token": authConfig.BearerConfig.Token,
		}
		auth["bearer_config"] = []interface{}{bearer}
	}

	return []interface{}{auth}
}

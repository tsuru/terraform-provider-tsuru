package provider

import (
	"context"
	"reflect"

	yaml "github.com/ghodss/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruRouter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruRouterCreate,
		ReadContext:   resourceTsuruRouterRead,
		UpdateContext: resourceTsuruRouterUpdate,
		DeleteContext: resourceTsuruRouterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"readiness_gates": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of readiness gates associated with this router",
			},
			"config": {
				Type:        schema.TypeString,
				Description: "Configuration for router in YAML format",
				Optional:    true,
			},
		},
	}
}

func routerFromResourceData(d *schema.ResourceData) (tsuru.DynamicRouter, diag.Diagnostics) {
	config, err := parseRouterConfig(d.Get("config"))
	if err != nil {
		return tsuru.DynamicRouter{}, diag.Errorf("Could not decode config, err : %s", err.Error())
	}

	readinessGates := []string{}
	for _, item := range d.Get("readiness_gates").([]interface{}) {
		readinessGates = append(readinessGates, item.(string))
	}

	return tsuru.DynamicRouter{
		Name:           d.Get("name").(string),
		Type:           d.Get("type").(string),
		ReadinessGates: readinessGates,
		Config:         config,
	}, nil
}

func resourceTsuruRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	router, dg := routerFromResourceData(d)
	if dg != nil {
		return dg
	}

	_, err := provider.TsuruClient.RouterApi.RouterCreate(ctx, router)
	if err != nil {
		return diag.Errorf("Could not create tsuru router, err : %s", err.Error())
	}

	d.SetId(router.Name)

	return nil
}

func resourceTsuruRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("name").(string)
	typo := d.Get("type").(string)

	router, _, err := provider.TsuruClient.RouterApi.RouterList(ctx)

	if err != nil {
		return diag.Errorf("Could not read tsuru router, err : %s", err.Error())
	}

	for _, router := range router {
		if router.Name != name || router.Type != typo {
			continue
		}
		d.Set("name", router.Name)
		d.Set("type", router.Type)
		d.Set("readiness_gates", router.ReadinessGates)

		config, err := parseRouterConfig(d.Get("config"))
		if err != nil {
			return diag.Errorf("Could not decode config, err : %s", err.Error())
		}

		if !reflect.DeepEqual(config, router.Config) {
			b, err := yaml.Marshal(router.Config)
			if err != nil {
				return diag.Errorf("Could not encode config, err : %s", err.Error())
			}
			d.Set("config", string(b))
		}

		return nil
	}
	d.SetId("")
	return nil
}

func resourceTsuruRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	router, dg := routerFromResourceData(d)
	if dg != nil {
		return dg
	}

	_, err := provider.TsuruClient.RouterApi.RouterUpdate(ctx, d.Id(), router)
	if err != nil {
		return diag.Errorf("Could not update tsuru router: %q, err: %s", d.Id(), err.Error())
	}
	return nil
}

func resourceTsuruRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	_, err := provider.TsuruClient.RouterApi.RouterDelete(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Could not delete tsuru router, err: %s", err.Error())
	}

	return nil
}

func parseRouterConfig(data interface{}) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	rawConfig := data.(string)

	err := yaml.Unmarshal([]byte(rawConfig), &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

package tsuru

import (
	"context"

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
			"config": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		},
	}
}

func resourceTsuruRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)
	name := d.Get("name").(string)
	config := make(map[string]interface{})

	for key, value := range d.Get("config").(map[string]interface{}) {
		config[key] = value.(string)
	}
	_, err := provider.TsuruClient.RouterApi.RouterCreate(ctx, tsuru.DynamicRouter{
		Name:   name,
		Type:   d.Get("type").(string),
		Config: config,
	})

	if err != nil {
		return diag.Errorf("Could not create tsuru router, err : %s", err.Error())
	}

	d.SetId(name)

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
		d.Set("config", router.Config)
		return nil
	}
	return diag.Errorf("Could not find tsuru router, name: %s,type: %s", name, typo)

}

func resourceTsuruRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	_, err := provider.TsuruClient.RouterApi.RouterUpdate(ctx, d.Id(), tsuru.DynamicRouter{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Config: d.Get("config").(map[string]interface{}),
	})

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

package tsuru

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func dataSourceTsuruApp() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTsuruAppRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Unique name of pool",
				Required:    true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tsuru_provisioner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"team_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"internal_addresses": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"process": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"routers": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"addresses": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"opts": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTsuruAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Get("name").(string)

	app, _, err := provider.TsuruClient.AppApi.AppGet(ctx, name)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(name)

	d.Set("description", app.Description)
	d.Set("tags", app.Tags)
	d.Set("platform", app.Platform)
	d.Set("tags", app.Tags)
	d.Set("pool", app.Pool)
	d.Set("cluster", app.Cluster)
	d.Set("tsuru_provisioner", app.Provisioner)

	d.Set("internal_addresses", flattenInternalAddresses(app.InternalAddresses))
	d.Set("routers", flattenRouters(app.Routers))

	d.Set("team_owner", app.TeamOwner)
	d.Set("teams", app.Teams)

	return nil
}

func flattenInternalAddresses(addrs []tsuru.AppInternalAddresses) []interface{} {
	result := []interface{}{}

	for _, addr := range addrs {
		result = append(result, map[string]interface{}{
			"domain":   addr.Domain,
			"port":     addr.Port,
			"protocol": addr.Protocol,
			"version":  addr.Version,
			"process":  addr.Process,
		})
	}

	return result
}

func flattenRouters(routers []tsuru.AppRouters) []interface{} {
	result := []interface{}{}

	for _, router := range routers {
		result = append(result, map[string]interface{}{
			"addresses": router.Addresses,
			"name":      router.Name,
			"opts":      router.Opts,
		})
	}

	return result
}

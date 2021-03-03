package tsuru

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func resourceTsuruPlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruPlanCreate,
		ReadContext:   resourceTsuruPlanRead,
		//UpdateContext: resourceTsuruPlanUpdate,
		DeleteContext: resourceTsuruPlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: true,
			},
			"swap": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"cpumilli": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"default": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Default:  false,
			},
			"router": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "tsuru_router",
			},
		},
	}
}

func resourceTsuruPlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	plan := planResourceData(d)

	plan, _, err := provider.TsuruClient.PlanApi.PlanCreate(ctx, plan)

	if err != nil {
		return diag.Errorf("Could not create tsuru plan: %q, err: %s", plan.Name, err.Error())
	}
	d.SetId(plan.Name)
	return nil
}

func planResourceData(d *schema.ResourceData) tsuru.Plan {
	return tsuru.Plan{
		Name:     d.Get("name").(string),
		Memory:   int64(d.Get("memory").(int)),
		Swap:     int64(d.Get("swap").(int)),
		Cpumilli: int32(d.Get("cpumilli").(int)),
		Router:   d.Get("router").(string),
		Default:  d.Get("default").(bool),
	}
}

func resourceTsuruPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	//parts := strings.SplitN(d.Id(), "/", 2)
	name := d.Get("name").(string)

	plans, _, err := provider.TsuruClient.PlanApi.PlanList(ctx)

	for _, plan := range plans {
		if plan.Name != name {
			continue
		}

		d.Set("name", plan.Name)
		d.Set("memory", plan.Memory)
		d.Set("swap", plan.Swap)
		d.Set("cpumilli", plan.Cpumilli)
		d.Set("default", plan.Default)

		return nil
	}

	if err != nil {
		return diag.Errorf("Could not read tsuru plan: %q, err: %s", d.Id(), err.Error())
	}
	return nil

}

// func resourceTsuruPlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	provider := metal.(*tsuruProvider)

// }

func resourceTsuruPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	plan := planResourceData(d)
	fmt.Println(" test", plan.Name)
	_, err := provider.TsuruClient.PlanApi.DeletePlan(ctx, plan.Name)
	if err != nil {
		return diag.Errorf("Could not delete tsuru plan, err: %s", err.Error())
	}

	return nil
}

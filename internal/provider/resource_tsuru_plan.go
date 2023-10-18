package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"k8s.io/apimachinery/pkg/api/resource"
)

func resourceTsuruPlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruPlanCreate,
		ReadContext:   resourceTsuruPlanRead,
		DeleteContext: resourceTsuruPlanDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpu": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpu_burst": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default": {
							Type:        schema.TypeFloat,
							Optional:    true,
							ForceNew:    true,
							Description: "Factor of burst, ie: 1.1 means 10% of burst",
						},
						"max_allowed": {
							Type:        schema.TypeFloat,
							Optional:    true,
							ForceNew:    true,
							Description: "max allowed when user customizes the burst",
						},
					},
				},
			},
			"default": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
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

	return resourceTsuruPlanRead(ctx, d, meta)
}

func planResourceData(d *schema.ResourceData) tsuru.Plan {
	cpuString := d.Get("cpu").(string)
	cpuFormat := cpuFormat(cpuString)
	var cpuMilli int32
	if cpuFormat == "unit" {
		cpuMilli = cpuUnitToMilli(cpuString)
	} else if cpuFormat == "percent" {
		cpuMilli = cpuPercentToMilli(cpuString)
	} else if cpuFormat == "milli" {
		cpuMilli = cpuMilliInt32(cpuString)
	}

	memoryString := d.Get("memory").(string)
	memoryBytes, _ := parseMemoryQuantity(memoryString)

	cpuBurst := tsuru.PlanCpuBurst{}
	if m, ok := d.GetOk("cpu_burst"); ok {
		cpuBurst = cpuBurstFromResourceData(m)
	}

	return tsuru.Plan{
		Name:     d.Get("name").(string),
		Memory:   memoryBytes,
		Cpumilli: cpuMilli,
		Default:  d.Get("default").(bool),
		CpuBurst: cpuBurst,
	}
}

func resourceTsuruPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()
	cpu := d.Get("cpu").(string)
	cpuFormat := cpuFormat(cpu)

	plans, _, err := provider.TsuruClient.PlanApi.PlanList(ctx)
	if err != nil {
		return diag.Errorf("Could not read tsuru plans err: %s", err.Error())
	}

	for _, plan := range plans {
		if plan.Name != name {
			continue
		}

		d.Set("name", plan.Name)
		d.Set("memory", memoryBytesToString(plan.Memory))

		if cpuFormat == "unit" {
			d.Set("cpu", cpuMillisToUnitString(plan.Cpumilli))
		} else if cpuFormat == "percent" {
			d.Set("cpu", cpuMillisToPercentString(plan.Cpumilli))
		} else if cpuFormat == "milli" {
			d.Set("cpu", cpuMillisToString(plan.Cpumilli))
		}

		cpuBurst := map[string]interface{}{}
		if plan.CpuBurst.Default != 0 {
			cpuBurst["default"] = plan.CpuBurst.Default
		}
		if plan.CpuBurst.MaxAllowed != 0 {
			cpuBurst["max_allowed"] = plan.CpuBurst.MaxAllowed
		}
		if len(cpuBurst) > 0 {
			d.Set("cpu_burst", []interface{}{cpuBurst})
		} else {
			d.Set("cpu_burst", []interface{}{})
		}

		d.Set("default", plan.Default)

		return nil
	}

	d.SetId("")
	return nil
}

func resourceTsuruPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	plan := planResourceData(d)
	_, err := provider.TsuruClient.PlanApi.DeletePlan(ctx, plan.Name)
	if err != nil {
		return diag.Errorf("Could not delete tsuru plan, err: %s", err.Error())
	}

	return nil
}

func memoryBytesToString(m int64) string {
	return resource.NewQuantity(m, resource.BinarySI).String()
}

func cpuFormat(cpu string) string {
	if strings.HasSuffix(cpu, "%") {
		return "percent"
	}
	if strings.HasSuffix(cpu, "m") {
		return "milli"
	}

	return "unit"
}

func cpuMillisToPercentString(c int32) string {
	return fmt.Sprintf("%g%%", float32(c)/10.0)
}

func cpuMillisToUnitString(c int32) string {
	return fmt.Sprintf("%g", float32(c)/1000.0)
}

func cpuMillisToString(c int32) string {
	return fmt.Sprintf("%dm", c)
}

func cpuUnitToMilli(c string) int32 {
	v, _ := strconv.ParseFloat(c, 32)
	return int32(v * 1000)
}

func cpuPercentToMilli(c string) int32 {
	v, _ := strconv.ParseFloat(c[0:len(c)-1], 32)
	return int32(v * 10)
}

func cpuMilliInt32(c string) int32 {
	v, _ := strconv.ParseFloat(c[0:len(c)-1], 32)
	return int32(v)
}

func parseMemoryQuantity(m string) (numBytes int64, err error) {
	if v, parseErr := strconv.Atoi(m); parseErr == nil {
		return int64(v), nil
	}
	memoryQuantity, err := resource.ParseQuantity(m)
	if err != nil {
		return 0, err
	}

	numBytes, _ = memoryQuantity.AsInt64()
	return numBytes, nil
}

func cpuBurstFromResourceData(meta interface{}) tsuru.PlanCpuBurst {
	cpuBurst := tsuru.PlanCpuBurst{}

	m := meta.([]interface{})
	if len(m) == 0 || m[0] == nil {
		return cpuBurst
	}

	containerMap := m[0].(map[string]interface{})

	if v, ok := containerMap["default"]; ok {
		cpuBurst.Default = v.(float64)
	}

	if v, ok := containerMap["max_allowed"]; ok {
		cpuBurst.MaxAllowed = v.(float64)
	}

	return cpuBurst
}

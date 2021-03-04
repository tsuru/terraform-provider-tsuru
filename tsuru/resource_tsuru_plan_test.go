package tsuru

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccTsuruPlan_basic(t *testing.T) {
	fakeServer := echo.New()
	expectedPlans := []tsuru.Plan{
		{
			Name:     "plan1",
			Memory:   int64(64 * 1024 * 1024),
			Cpumilli: int32(100),
			Default:  true,
		},
		{
			Name:     "plan2",
			Memory:   int64(1024 * 1024 * 1024),
			Cpumilli: int32(1000),
		},
		{
			Name:     "plan3",
			Memory:   int64(1024 * 1024 * 1024),
			Cpumilli: int32(2000),
		},
	}

	fakeServer.POST("/1.0/plans", func(c echo.Context) error {
		p := tsuru.Plan{}
		err := c.Bind(&p)
		require.NoError(t, err)

		for _, expectedPlan := range expectedPlans {
			if expectedPlan.Name != p.Name {
				continue
			}

			assert.Equal(t, expectedPlan.Memory, p.Memory)
			assert.Equal(t, expectedPlan.Cpumilli, p.Cpumilli)
			assert.Equal(t, expectedPlan.Default, p.Default)
		}

		return c.JSON(200, p)
	})
	fakeServer.GET("/1.0/plans", func(c echo.Context) error {
		return c.JSON(http.StatusOK, expectedPlans)
	})
	fakeServer.DELETE("/1.0/plans/:name", func(c echo.Context) error {
		name := c.Param("name")
		for _, expectedPlan := range expectedPlans {
			if expectedPlan.Name == name {
				return c.NoContent(http.StatusNoContent)
			}
		}
		return c.NoContent(http.StatusNotFound)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruPlanConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tsuru_plan.plan1", "name", "plan1"),
					resource.TestCheckResourceAttr("tsuru_plan.plan2", "name", "plan2"),
					resource.TestCheckResourceAttr("tsuru_plan.plan3", "name", "plan3"),

					resource.TestCheckResourceAttr("tsuru_plan.plan1", "cpu", "100m"),
					resource.TestCheckResourceAttr("tsuru_plan.plan2", "cpu", "1"),
					resource.TestCheckResourceAttr("tsuru_plan.plan3", "cpu", "200%"),

					resource.TestCheckResourceAttr("tsuru_plan.plan1", "memory", "64Mi"),
					resource.TestCheckResourceAttr("tsuru_plan.plan2", "memory", "1Gi"),
					resource.TestCheckResourceAttr("tsuru_plan.plan3", "memory", "1Gi"),
				),
			},
		},
	})
}

func testAccTsuruPlanConfig_basic() string {
	return `
resource "tsuru_plan" "plan1" {
	name = "plan1"
	cpu = "100m"
	memory = "64Mi"
	default = true
}

resource "tsuru_plan" "plan2" {
	name = "plan2"
	cpu = "1"
	memory = "1Gi"
}

resource "tsuru_plan" "plan3" {
	name = "plan3"
	cpu = "200%"
	memory = "1Gi"
}
`
}

func TestMemoryToString(t *testing.T) {
	assert.Equal(t, "512Mi", memoryBytesToString(1024*1024*512))
	assert.Equal(t, "2Gi", memoryBytesToString(1024*1024*1024*2))
}

func TestCPUMillisToPercentString(t *testing.T) {
	assert.Equal(t, "100%", cpuMillisToPercentString(1000))
	assert.Equal(t, "20%", cpuMillisToPercentString(200))
}

func TestCPUMillisToUnitString(t *testing.T) {
	assert.Equal(t, "1", cpuMillisToUnitString(1000))
	assert.Equal(t, "0.2", cpuMillisToUnitString(200))
}

func TestCPUMillisToString(t *testing.T) {
	assert.Equal(t, "1000m", cpuMillisToString(1000))
	assert.Equal(t, "200m", cpuMillisToString(200))
}

func TestCpuUnitToMilli(t *testing.T) {
	assert.Equal(t, int32(4000), cpuUnitToMilli("4"))
	assert.Equal(t, int32(1000), cpuUnitToMilli("1"))
	assert.Equal(t, int32(200), cpuUnitToMilli("0.2"))
}

func TestCpuPercentToMilli(t *testing.T) {
	assert.Equal(t, int32(2000), cpuPercentToMilli("200%"))
	assert.Equal(t, int32(1000), cpuPercentToMilli("100%"))
	assert.Equal(t, int32(200), cpuPercentToMilli("20%"))
}

func TestCpuMilliInt32(t *testing.T) {
	assert.Equal(t, int32(2000), cpuMilliInt32("2000m"))
	assert.Equal(t, int32(1000), cpuMilliInt32("1000m"))
	assert.Equal(t, int32(200), cpuMilliInt32("200m"))
}

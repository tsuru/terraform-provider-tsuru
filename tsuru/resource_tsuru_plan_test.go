package tsuru

import (
	"fmt"
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
	fakeServer.POST("/1.0/plans", func(c echo.Context) error {
		p := tsuru.Plan{}
		err := c.Bind(&p)
		require.NoError(t, err)
		assert.Equal(t, p.Name, "test_plan")
		assert.Equal(t, p.Memory, int64(0))
		assert.Equal(t, p.Cpumilli, int32(0))
		assert.Equal(t, p.Swap, int64(0))
		assert.Equal(t, p.Router, "tsuru_router")

		return c.JSON(200, p)
	})
	fakeServer.GET("/1.0/plans", func(c echo.Context) error {
		//memory := c.Param("memory")

		return c.JSON(http.StatusOK, []tsuru.Plan{
			{

				Name:     "test_plan",
				Memory:   0,
				Cpumilli: 0,
				Swap:     0,
				Router:   "test_router",
			},
		})
	})
	fakeServer.DELETE("/1.0/plans/:name", func(c echo.Context) error {
		name := c.Param("name")
		require.Equal(t, name, "test_plan")
		return c.NoContent(http.StatusNoContent)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_plan.test_plan"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruPlanConfig_basic(server.URL, "test_plan"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test_plan"),
				),
			},
		},
	})
}

func testAccTsuruPlanConfig_basic(fakeServer, name string) string {
	return fmt.Sprintf(`
resource "tsuru_plan"  "test_plan"   {
	name = "%s"

	router = "tsuru_router"
}
`, name)
}

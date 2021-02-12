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

func TestAccTsuruRouter_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("/1.8/routers", func(c echo.Context) error {
		p := &tsuru.DynamicRouter{}
		err := c.Bind(&p)
		require.NoError(t, err)
		assert.Equal(t, p.Name, "test_router")
		assert.Equal(t, p.Type, "router")
		assert.Equal(t, p.Config, map[string]interface{}{"config": "test_config"})

		return nil
	})
	fakeServer.GET("/1.3/routers", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []*tsuru.DynamicRouter{
			{
				Name:   "test_router",
				Type:   "router",
				Config: map[string]interface{}{"config": "test_config"},
			},
		})
	})
	fakeServer.DELETE("/1.8/routers/:name", func(c echo.Context) error {
		name := "test_router"
		require.Equal(t, name, "test_router")
		return c.NoContent(http.StatusNoContent)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_router.test_router"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruRouterConfig_basic(server.URL, "test_router"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test_router"),
					resource.TestCheckResourceAttr(resourceName, "type", "router"),
					resource.TestCheckResourceAttr(resourceName, "config.config", "test_config"),
				),
			},
		},
	})
}

func testAccTsuruRouterConfig_basic(fakeServer, name string) string {
	return fmt.Sprintf(`
resource "tsuru_router"  "test_router"   {
	name = "%s" 
	type = "router"
	config = {
		"config" = "test_config"
	}
	
}
`, name)
}

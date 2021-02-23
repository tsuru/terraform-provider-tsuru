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
		assert.Equal(t, "test_router", p.Name)
		assert.Equal(t, "router", p.Type)
		assert.Equal(t, map[string]interface{}{
			"url": "testing",
			"headers": map[string]interface{}{
				"x-my-header": "test",
			},
		}, p.Config)

		return nil
	})
	fakeServer.GET("/1.3/routers", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []*tsuru.DynamicRouter{
			{
				Name: "test_router",
				Type: "router",
				Config: map[string]interface{}{
					"url": "testing",
					"headers": map[string]interface{}{
						"x-my-header": "test",
					},
				},
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

	config = <<-EOT
	url: "testing"
	headers:
	  "x-my-header": test
	EOT
}
`, name)
}

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

func TestAccTsuruCluster_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("1.3/provisioner/clusters", func(c echo.Context) error {
		p := &tsuru.Cluster{}
		err := c.Bind(&p)
		require.NoError(t, err)
		assert.Equal(t, p.Name, "name")
		assert.Equal(t, p.Addresses, "adresses")
		assert.Equal(t, p.Cacert, "ca_certificate")
		assert.Equal(t, p.Clientcert, "client_certificate")
		assert.Equal(t, p.Clientkey, "client_key")
		assert.Equal(t, p.CustomData, "custom_data")

		return nil
	})
	fakeServer.GET("/1.8/provisioner/clusters/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.JSON(http.StatusOK, &tsuru.Cluster{
			Name:        name,
			Provisioner: "provisioner",
		})
	})
	fakeServer.DELETE("/cluster/name", func(c echo.Context) error {
		name := c.Param("name")
		require.Equal(t, name, "name")
		return c.NoContent(http.StatusNoContent)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_cluster.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "tsuru_cluster.test",
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruClusterConfig_basic(server.URL, "test_carlos"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr("tsuru_cluster.test", "addresses", "name"),
					resource.TestCheckResourceAttr("tsuru_cluster.test", "custom_data", "ca_certificate"),
				),
			},
		},
	})
}

func testAccTsuruClusterConfig_basic(fakeServer, name string) string {
	return fmt.Sprintf(`
resource "tsuru_cluster"  "test_carlos"   {
	name = "%s"
}
`, name)
}

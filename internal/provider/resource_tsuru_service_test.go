package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAccTsuruService_basic(t *testing.T) {
	fakeServer := echo.New()

	fakeServer.POST("/1.0/services", func(c echo.Context) error {
		id := c.FormValue("id")
		endpoint := c.FormValue("endpoint")
		team := c.FormValue("team")
		encoding := c.FormValue("encoding")

		assert.Equal(t, "my-service", id)
		assert.Equal(t, "http://my-service.example.com", endpoint)
		assert.Equal(t, "my-team", team)
		assert.Equal(t, "form", encoding)

		return c.NoContent(http.StatusCreated)
	})

	fakeServer.GET("/1.0/services/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "my-service", name)
		return c.JSON(http.StatusOK, []map[string]interface{}{
			{
				"service":   "my-service",
				"instances": []string{},
			},
		})
	})

	fakeServer.PUT("/1.0/services/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "my-service", name)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.DELETE("/1.0/services/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "my-service", name)
		return c.NoContent(http.StatusNoContent)
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
				Config: testAccTsuruServiceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tsuru_service.my_service", "name", "my-service"),
					resource.TestCheckResourceAttr("tsuru_service.my_service", "endpoint", "http://my-service.example.com"),
					resource.TestCheckResourceAttr("tsuru_service.my_service", "team", "my-team"),
					resource.TestCheckResourceAttr("tsuru_service.my_service", "multi_cluster", "false"),
					resource.TestCheckResourceAttr("tsuru_service.my_service", "encoding", "form"),
				),
			},
		},
	})
}

func TestAccTsuruService_withEncoding(t *testing.T) {
	fakeServer := echo.New()

	fakeServer.POST("/1.0/services", func(c echo.Context) error {
		id := c.FormValue("id")
		encoding := c.FormValue("encoding")

		assert.Equal(t, "my-service", id)
		assert.Equal(t, "json", encoding)

		return c.NoContent(http.StatusCreated)
	})

	fakeServer.GET("/1.0/services/:name", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []map[string]interface{}{
			{
				"service":   "my-service",
				"instances": []string{},
			},
		})
	})

	fakeServer.PUT("/1.0/services/:name", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	fakeServer.DELETE("/1.0/services/:name", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
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
				Config: testAccTsuruServiceConfig_withEncoding(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tsuru_service.my_service", "encoding", "json"),
				),
			},
		},
	})
}

func testAccTsuruServiceConfig_basic() string {
	return `
resource "tsuru_service" "my_service" {
	name     = "my-service"
	endpoint = "http://my-service.example.com"
	team     = "my-team"
}
`
}

func testAccTsuruServiceConfig_withEncoding() string {
	return `
resource "tsuru_service" "my_service" {
	name     = "my-service"
	endpoint = "http://my-service.example.com"
	team     = "my-team"
	encoding = "json"
}
`
}

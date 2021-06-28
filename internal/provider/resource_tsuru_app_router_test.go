// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccResourceTsuruAppRouter(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.3/routers", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.PlanRouter{{Name: "some-router"}})
	})

	fakeServer.GET("/1.5/apps/:app/routers", func(c echo.Context) error {
		routers := []tsuru.AppRouter{}
		if iterationCount == 1 {
			routers = append(routers, tsuru.AppRouter{
				Name: "some-router",
				Opts: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			})
		}
		return c.JSON(http.StatusOK, routers)
	})

	fakeServer.POST("/1.5/apps/:app/routers", func(c echo.Context) error {
		app := c.Param("app")
		router := tsuru.AppRouter{}
		c.Bind(&router)
		assert.Equal(t, "app01", app)
		assert.Equal(t, "some-router", router.Name)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.5/apps/:app/routers/:router", func(c echo.Context) error {
		app := c.Param("app")
		router := c.Param("router")
		assert.Equal(t, "app01", app)
		assert.Equal(t, "some-router", router)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_router.router"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppRouter_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "name", "some-router"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppRouter_basic() string {
	return `
	resource "tsuru_app_router" "router" {
		app = "app01"
		name = "some-router"
		options = {
			"key1" = "value1"
			"key2" = "value2"
		}
	}
`
}

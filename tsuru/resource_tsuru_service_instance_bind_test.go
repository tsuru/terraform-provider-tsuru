// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

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

func TestAccResourceServiceInstanceBind(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/services/:service/instances/:instance", func(c echo.Context) error {
		service := c.Param("service")
		if service != "service01" {
			return nil
		}

		app := &tsuru.ServiceInstanceInfo{
			Apps:      []string{},
			Teamowner: "my-team",
			Teams:     []string{},
		}

		if iterationCount == 1 {
			app.Apps = []string{
				"app01",
			}
		}

		return c.JSON(http.StatusOK, app)
	})

	fakeServer.PUT("/1.0/services/:service/instances/:instance/:app", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		app := c.Param("app")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "app01", app)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.0/services/:service/instances/:instance/:app", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		app := c.Param("app")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "app01", app)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_service_instance_bind.instance_bind"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceInstanceBind_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_name", "service01"),
					resource.TestCheckResourceAttr(resourceName, "service_instance", "my-instance"),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "restart_on_update", "false"),
				),
			},
		},
	})
}

func testAccResourceServiceInstanceBind_basic() string {
	return `
	resource "tsuru_service_instance_bind" "instance_bind" {
		service_name = "service01"
		service_instance = "my-instance"
		app = "app01"
		restart_on_update = false
	}
`
}

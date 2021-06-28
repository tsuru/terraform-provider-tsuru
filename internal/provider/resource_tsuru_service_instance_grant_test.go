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

func TestAccResourceServiceInstanceGrant(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/services/:service/instances/:instance", func(c echo.Context) error {
		service := c.Param("service")
		if service != "service01" {
			return nil
		}

		app := &tsuru.ServiceInstanceInfo{
			Teamowner: "my-team",
			Teams:     []string{},
		}

		if iterationCount == 1 {
			app.Teams = []string{
				"mysupport-team",
			}
		}

		return c.JSON(http.StatusOK, app)
	})

	fakeServer.PUT("/1.0/services/:service/instances/permission/:instance/:team", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		team := c.Param("team")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "mysupport-team", team)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.0/services/:service/instances/permission/:instance/:team", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		team := c.Param("team")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "mysupport-team", team)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_service_instance_grant.instance_grant"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceInstanceGrant_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_name", "service01"),
					resource.TestCheckResourceAttr(resourceName, "service_instance", "my-instance"),
					resource.TestCheckResourceAttr(resourceName, "team", "mysupport-team"),
				),
			},
		},
	})
}

func testAccResourceServiceInstanceGrant_basic() string {
	return `
	resource "tsuru_service_instance_grant" "instance_grant" {
		service_name = "service01"
		service_instance = "my-instance"
		team = "mysupport-team"
	}
`
}

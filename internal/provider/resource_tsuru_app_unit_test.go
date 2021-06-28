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

func TestAccResourceTsuruAppUnit(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		app := &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
			Units: []tsuru.Unit{
				{Processname: "web"},
				{Processname: "web"},
				{Processname: "web"},
			},
		}

		if iterationCount == 1 {
			app.Units = []tsuru.Unit{
				{Processname: "web"},
				{Processname: "web"},
				{Processname: "web"},
				{Processname: "web"},
				{Processname: "web"},
			}
		}

		return c.JSON(http.StatusOK, app)
	})

	fakeServer.PUT("/1.0/apps/:app/units", func(c echo.Context) error {
		app := c.Param("app")
		delta := tsuru.UnitsDelta{}
		c.Bind(&delta)
		assert.Equal(t, "app01", app)
		assert.Equal(t, "web", delta.Process)
		assert.Equal(t, "2", delta.Units)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.0/apps/:app/units", func(c echo.Context) error {
		app := c.Param("app")
		delta := tsuru.UnitsDelta{}
		c.Bind(&delta)
		assert.Equal(t, "app01", app)
		assert.Equal(t, "web", delta.Process)
		assert.Equal(t, "5", delta.Units)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_unit.units"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppUnit_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "units_count", "5"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppUnit_basic() string {
	return `
	resource "tsuru_app_unit" "units" {
		app = "app01"
		process = "web"
		units_count = 5
	}
`
}

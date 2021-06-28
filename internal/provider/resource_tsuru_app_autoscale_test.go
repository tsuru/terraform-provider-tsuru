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

func TestAccResourceTsuruAppAutoscale(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
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
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "800m",
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscale_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "min_units", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "cpu_average", "800m"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscale_basic() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "800m"
	}
`
}

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

func TestConflictOnTsuruAppGrant(t *testing.T) {
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
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		}

		if iterationCount == 1 {
			app.Teams = []string{
				"mysupport-team",
			}
		}

		return c.JSON(http.StatusOK, app)
	})

	fakeServer.PUT("/1.0/apps/:app/teams/:team", func(c echo.Context) error {
		team := c.Param("team")
		assert.Equal(t, "mysupport-team", team)
		iterationCount++
		return c.JSON(http.StatusConflict, nil)
	})

	fakeServer.DELETE("/1.0/apps/:app/teams/:team", func(c echo.Context) error {
		team := c.Param("team")
		assert.Equal(t, "mysupport-team", team)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_grant.team"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppGrant_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "team", "mysupport-team"),
				),
			},
		},
	})
}

func TestAccResourceTsuruAppGrant(t *testing.T) {
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
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		}

		if iterationCount == 1 {
			app.Teams = []string{
				"mysupport-team",
			}
		}

		return c.JSON(http.StatusOK, app)
	})

	fakeServer.PUT("/1.0/apps/:app/teams/:team", func(c echo.Context) error {
		team := c.Param("team")
		assert.Equal(t, "mysupport-team", team)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.0/apps/:app/teams/:team", func(c echo.Context) error {
		team := c.Param("team")
		assert.Equal(t, "mysupport-team", team)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_grant.team"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppGrant_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "team", "mysupport-team"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppGrant_basic() string {
	return `
	resource "tsuru_app_grant" "team" {
		app = "app01"
		team = "mysupport-team"
	}
`
}

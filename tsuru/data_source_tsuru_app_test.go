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
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccDatasourceTsuruApp_basic(t *testing.T) {
	fakeServer := echo.New()

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")

		if name == "app01" {

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
				Routers: []tsuru.AppRouters{
					{
						Addresses: []string{
							"web.app01.tsuru.io",
							"v2.web.app01.tsuru.io",
						},
						Opts: map[string]interface{}{
							"enable-feature-x": "true",
						},
						Name: "external-router",
					},
					{
						Addresses: []string{
							"web.app01.local",
							"v2.web.app01.local",
						},
						Opts: map[string]interface{}{
							"enable-feature-y": "true",
						},
						Name: "vpn-router",
					},
				},
				InternalAddresses: []tsuru.AppInternalAddresses{
					{
						Domain:   "myapp-web.namespace.svc.cluster.local",
						Port:     8888,
						Process:  "web",
						Version:  "",
						Protocol: "TCP",
					},
					{
						Domain:   "myapp-subscriber.namespace.svc.cluster.local",
						Port:     8888,
						Process:  "subscriber",
						Version:  "",
						Protocol: "TCP",
					},
					{
						Domain:   "myapp-web-v2.namespace.svc.cluster.local",
						Port:     8888,
						Process:  "web",
						Version:  "2",
						Protocol: "TCP",
					},
				},
			})
		}
		return nil
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
				Config: testAccDatasourceTsuruAppConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "name", "app01"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "description", "my beautiful application"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "team_owner", "myteam"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "teams.0", "mysupport-team"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "teams.1", "mysponsors"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "cluster", "my-cluster-01"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "pool", "my-pool"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "tsuru_provisioner", "kubernetes"),

					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.0.addresses.0", "web.app01.tsuru.io"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.0.addresses.1", "v2.web.app01.tsuru.io"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.0.opts.enable-feature-x", "true"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.0.name", "external-router"),

					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.1.addresses.0", "web.app01.local"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.1.addresses.1", "v2.web.app01.local"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.1.opts.enable-feature-y", "true"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "routers.1.name", "vpn-router"),

					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.0.domain", "myapp-web.namespace.svc.cluster.local"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.0.port", "8888"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.0.process", "web"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.0.version", ""),

					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.1.domain", "myapp-subscriber.namespace.svc.cluster.local"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.1.port", "8888"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.1.process", "subscriber"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.1.version", ""),

					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.2.domain", "myapp-web-v2.namespace.svc.cluster.local"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.2.port", "8888"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.2.protocol", "TCP"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.2.process", "web"),
					resource.TestCheckResourceAttr("data.tsuru_app.app01", "internal_addresses.2.version", "2"),
				),
			},
		},
	})
}

func testAccDatasourceTsuruAppConfig_basic() string {
	return `
	data "tsuru_app" "app01" {
		name = "app01"
	}
	  
`
}

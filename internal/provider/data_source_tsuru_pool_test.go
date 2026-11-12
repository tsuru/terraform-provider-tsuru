// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccDatasourceTsuruPool_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.GET("/pools/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name == "broken-pool" {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "boom"})
		}
		if name != "my-pool" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "pool not found"})
		}
		return c.JSON(http.StatusOK, &tsuru.Pool{
			Name:        name,
			Provisioner: "kubernetes",
			Default:     true,
			Teams:       []string{"team-a", "team-b"},
			Labels: map[string]string{
				"my-label": "value",
			},
			Allowed: map[string][]string{
				"team":   {"team-a", "team-b"},
				"router": {"ingress-nginx"},
				"plan":   {"c1m1", "c2m2"},
			},
		})
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	defer server.Close()
	os.Setenv("TSURU_TARGET", server.URL)

	dataSourceName := "data.tsuru_pool.pool"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "tsuru_pool" "pool" {
	name = "my-pool"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "my-pool"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "my-pool"),
					resource.TestCheckResourceAttr(dataSourceName, "tsuru_provisioner", "kubernetes"),
					resource.TestCheckResourceAttr(dataSourceName, "default", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "public", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "teams.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "teams.0", "team-a"),
					resource.TestCheckResourceAttr(dataSourceName, "teams.1", "team-b"),
					resource.TestCheckResourceAttr(dataSourceName, "labels.my-label", "value"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.0.field", "plan"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.0.values.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.0.values.1", "c2m2"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.1.field", "router"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.1.values.0", "ingress-nginx"),
					resource.TestCheckResourceAttr(dataSourceName, "allowed.2.field", "team"),
				),
			},
			{
				Config: `
data "tsuru_pool" "pool" {
	name = "missing-pool"
}
`,
				ExpectError: regexp.MustCompile(`tsuru pool "missing-pool" not found`),
			},
			{
				Config: `
data "tsuru_pool" "pool" {
	name = "broken-pool"
}
`,
				ExpectError: regexp.MustCompile(`Could not read tsuru pool: "broken-pool"`),
			},
		},
	})
}

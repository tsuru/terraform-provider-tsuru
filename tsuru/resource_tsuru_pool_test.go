// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

func TestAccTsuruPool_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("/1.0/pools", func(c echo.Context) error {
		p := &tsuru.PoolCreateData{}
		err := c.Bind(&p)
		require.NoError(t, err)
		assert.Equal(t, p.Name, "my-pool")
		assert.Equal(t, p.Provisioner, "kubernetes")
		assert.Equal(t, p.Labels, map[string]string{
			"my-label": "value",
		})
		assert.False(t, p.Public)
		assert.False(t, p.Default)

		return nil
	})
	fakeServer.GET("/pools/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.JSON(http.StatusOK, &tsuru.Pool{
			Name:        name,
			Provisioner: "kubernetes",
			Labels: map[string]string{
				"my-label": "value",
			},
		})
	})
	fakeServer.DELETE("/pools/:name", func(c echo.Context) error {
		name := c.Param("name")
		require.Equal(t, name, "my-pool")
		return c.NoContent(http.StatusNoContent)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("method=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}

	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_pool.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "tsuru_pool.test",
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruPoolConfig_basic(server.URL, "my-pool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr("tsuru_pool.test", "name", "my-pool"),
					resource.TestCheckResourceAttr("tsuru_pool.test", "tsuru_provisioner", "kubernetes"),
				),
			},
		},
	})
}

func testAccTsuruPoolConfig_basic(fakeServer, name string) string {
	return fmt.Sprintf(`
resource "tsuru_pool" "test" {
	name = "%s"
	labels = {
		"my-label" = "value"
	}
}
`, name)
}

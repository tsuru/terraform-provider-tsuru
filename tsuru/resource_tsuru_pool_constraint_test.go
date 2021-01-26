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

func TestAccTsuruPoolConstraint_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.PUT("/1.3/constraints", func(c echo.Context) error {
		p := &tsuru.PoolConstraintSet{}
		err := c.Bind(p)
		require.NoError(t, err)
		assert.Equal(t, p.PoolExpr, "my-pool")
		assert.Equal(t, p.Field, "router")
		if len(p.Values) > 1 {
			assert.Equal(t, p.Values, []string{"ingress", "load-balancer"})
		}
		assert.False(t, p.Blacklist)
		return nil
	})
	fakeServer.GET("/1.3/constraints", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []*tsuru.PoolConstraint{
			{
				PoolExpr:  "my-pool",
				Field:     "router",
				Values:    []string{"ingress", "load-balancer"},
				Blacklist: false,
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

	resourceName := "tsuru_pool_constraint.test_routers"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "tsuru_pool_constraint.test",
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruPoolConstraintConfig_basic(server.URL, "my-pool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr("tsuru_pool_constraint.test_routers", "pool_expr", "my-pool"),
					resource.TestCheckResourceAttr("tsuru_pool_constraint.test_routers", "field", "router"),
				),
			},
		},
	})
}

func testAccTsuruPoolConstraintConfig_basic(fakeServer, poolExpr string) string {
	return fmt.Sprintf(`
resource "tsuru_pool_constraint" "test_routers" {
	pool_expr = "%s"
	field = "router"
	values = [
		"load-balancer",
		"ingress"
	]
}
`, poolExpr)
}

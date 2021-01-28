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

func TestAccTsuruClusterPool_basic(t *testing.T) {
	fakeServer := echo.New()
	iterationCount := 0
	fakeServer.POST("/1.4/provisioner/clusters/:cluster", func(c echo.Context) error {
		p := &tsuru.Cluster{}
		err := c.Bind(p)
		require.NoError(t, err)
		assert.Equal(t, p.Name, c.Param("cluster"))
		if iterationCount == 0 {
			assert.Equal(t, p.Pools, []string{"other-pool", "my-pool"})
		} else if iterationCount == 1 {
			assert.Equal(t, p.Pools, []string{"other-pool"})
		}

		iterationCount++
		return nil
	})
	fakeServer.GET("/1.8/provisioner/clusters/:cluster", func(c echo.Context) error {
		cluster := c.Param("cluster")
		pools := []string{"other-pool"}
		if iterationCount == 1 {
			pools = append(pools, "my-pool")
		}
		return c.JSON(http.StatusOK, tsuru.Cluster{
			Name:  cluster,
			Pools: pools,
		})
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("method=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}

	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_cluster_pool.cluster-pool"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "tsuru_pool_constraint.test",
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruClusterPoolConfig_basic(server.URL, "my-cluster", "my-pool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool", "my-pool"),
					resource.TestCheckResourceAttr(resourceName, "cluster", "my-cluster"),
				),
			},
		},
	})
}

func testAccTsuruClusterPoolConfig_basic(fakeServer, cluster, pool string) string {
	return fmt.Sprintf(`
resource "tsuru_cluster_pool" "cluster-pool" {
	cluster = "%s"
	pool = "%s"
}
`, cluster, pool)
}

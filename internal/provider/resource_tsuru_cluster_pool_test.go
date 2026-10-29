// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestAccTsuruClusterPool_retryOnEventLockKeepsConcurrentPools(t *testing.T) {
	var mu sync.Mutex
	pools := []string{"other-pool"}
	armed := true
	locks := 0

	fakeServer := echo.New()
	fakeServer.POST("/1.4/provisioner/clusters/:cluster", func(c echo.Context) error {
		mu.Lock()
		defer mu.Unlock()
		if armed {
			// a concurrent operation holds the lock and changes the pools before releasing it
			armed = false
			locks++
			pools = append(pools, fmt.Sprintf("concurrent-%d", locks))
			return c.String(http.StatusInternalServerError, "event locked: cluster(my-cluster) running \"cluster.update\"")
		}
		p := &tsuru.Cluster{}
		require.NoError(t, c.Bind(p))
		pools = p.Pools
		return nil
	})
	fakeServer.GET("/1.8/provisioner/clusters/:cluster", func(c echo.Context) error {
		mu.Lock()
		defer mu.Unlock()
		return c.JSON(http.StatusOK, tsuru.Cluster{
			Name:  c.Param("cluster"),
			Pools: append([]string{}, pools...),
		})
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("method=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}

	server := httptest.NewServer(fakeServer)
	defer server.Close()
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_cluster_pool.cluster-pool"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruClusterPoolConfig_basic(server.URL, "my-cluster", "my-pool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					func(*terraform.State) error {
						mu.Lock()
						defer mu.Unlock()
						assert.Equal(t, []string{"other-pool", "concurrent-1", "my-pool"}, pools)
						armed = true // lock the destroy as well
						return nil
					},
				),
			},
		},
	})

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, locks)
	assert.Equal(t, []string{"other-pool", "concurrent-1", "concurrent-2"}, pools)
}

func testAccTsuruClusterPoolConfig_basic(fakeServer, cluster, pool string) string {
	return fmt.Sprintf(`
resource "tsuru_cluster_pool" "cluster-pool" {
	cluster = "%s"
	pool = "%s"
}
`, cluster, pool)
}

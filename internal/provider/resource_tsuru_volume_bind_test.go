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

func TestAccResourceVolumeBind(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.4/volumes/:volume", func(c echo.Context) error {
		v := tsuru.Volume{
			Name:  "volume01",
			Binds: []tsuru.VolumeBind{},
		}

		if iterationCount == 1 {
			v.Binds = append(v.Binds, tsuru.VolumeBind{
				Id: tsuru.VolumeBindId{
					App:        "app01",
					Mountpoint: "/mnt/my-volume",
					Volume:     "volume01",
				},
				Readonly: false,
			})
		}
		return c.JSON(http.StatusOK, v)
	})

	fakeServer.POST("/1.4/volumes/:volume/bind", func(c echo.Context) error {
		volume := c.Param("volume")
		v := tsuru.VolumeBindData{}
		c.Bind(&v)
		assert.Equal(t, "volume01", volume)
		assert.Equal(t, "app01", v.App)
		assert.Equal(t, "/mnt/my-volume", v.Mountpoint)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.4/volumes/:volume/bind", func(c echo.Context) error {
		volume := c.Param("volume")
		v := tsuru.VolumeBindData{}
		c.Bind(&v)
		assert.Equal(t, "volume01", volume)
		assert.Equal(t, "app01", v.App)
		assert.Equal(t, "/mnt/my-volume", v.Mountpoint)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_volume_bind.volume-bind"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVolumeBind_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "volume", "volume01"),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "mount_point", "/mnt/my-volume"),
					resource.TestCheckResourceAttr(resourceName, "restart_on_update", "false"),
				),
			},
		},
	})
}

func testAccResourceVolumeBind_basic() string {
	return `
	resource "tsuru_volume_bind" "volume-bind" {
		volume = "volume01"
		app = "app01"
		mount_point = "/mnt/my-volume"
		restart_on_update = false
	}
`
}

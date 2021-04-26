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
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccResourceVolume(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.POST("/1.4/volumes", func(c echo.Context) error {
		v := tsuru.Volume{}
		c.Bind(&v)
		assert.Equal(t, "volume01", v.Name)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.GET("/1.4/volumes/:volume", func(c echo.Context) error {
		if iterationCount == 1 {
			v := tsuru.Volume{
				Name:      "volume01",
				TeamOwner: "my-team",
				Plan:      tsuru.VolumePlan{Name: "plan01"},
				Pool:      "pool01",
				Opts: map[string]string{
					"key": "value",
				},
			}
			return c.JSON(http.StatusOK, v)
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.DELETE("/1.4/volumes/:volume", func(c echo.Context) error {
		volume := c.Param("volume")
		assert.Equal(t, "volume01", volume)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_volume.volume"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVolume_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "volume01"),
					resource.TestCheckResourceAttr(resourceName, "owner", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "plan", "plan01"),
					resource.TestCheckResourceAttr(resourceName, "pool", "pool01"),
				),
			},
		},
	})
}

func testAccResourceVolume_basic() string {
	return `
	resource "tsuru_volume" "volume" {
		name = "volume01"
		owner = "my-team"
		plan = "plan01"
		pool = "pool01"
		options = {
			"key" = "value"
		}
	}
`
}

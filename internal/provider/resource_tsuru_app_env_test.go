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

func TestAccResourceTsuruAppEnv(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:app/env", func(c echo.Context) error {
		envs := []tsuru.EnvVar{}
		if iterationCount == 1 {
			envs = append(envs, tsuru.EnvVar{
				Name:      "env1",
				Value:     "10",
				Public:    true,
				ManagedBy: "terraform",
			})
			envs = append(envs, tsuru.EnvVar{
				Name:      "env2",
				Value:     "12",
				Public:    false,
				ManagedBy: "terraform",
			})
		}
		return c.JSON(http.StatusOK, envs)
	})

	fakeServer.POST("/1.0/apps/:app/env", func(c echo.Context) error {
		envs := tsuru.EnvSetData{}
		c.Bind(&envs)
		iterationCount++

		if iterationCount == 1 {
			assert.Equal(t, 2, len(envs.Envs))
		} else if iterationCount == 2 {
			assert.Equal(t, 0, len(envs.Envs))
		}
		assert.Equal(t, "terraform", envs.ManagedBy)
		assert.Equal(t, true, envs.Norestart)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_env.env"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppEnv_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.env1", "10"),
					resource.TestCheckResourceAttr(resourceName, "private_environment_variables.env2", "12"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppEnv_basic() string {
	return `
	resource "tsuru_app_env" "env" {
		app = "app01"
		restart_on_update = false
		environment_variables = {
			env1 = "10"
		}

		private_environment_variables = {
			env2 = "12"
		}
	}
`
}

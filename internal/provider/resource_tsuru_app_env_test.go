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
		assert.Equal(t, 2, len(envs.Envs))
		assert.Equal(t, "terraform", envs.ManagedBy)
		assert.Equal(t, true, envs.Norestart)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.0/apps/:app/env", func(c echo.Context) error {
		envs := c.QueryParam("env")
		noRestart := c.QueryParam("norestart")
		assert.Equal(t, "env1", envs)
		assert.Equal(t, "true", noRestart)
		return c.NoContent(http.StatusOK)
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
					resource.TestCheckResourceAttr(resourceName, "environment_variable.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "environment_variable.*", map[string]string{
						"name":  "env1",
						"value": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "environment_variable.*", map[string]string{
						"name":            "env2",
						"sensitive_value": "12",
					}),
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
		environment_variable {
			name = "env1"
			value = "10"
		}

		environment_variable {
			name = "env2"
			sensitive_value = "12"
		}
	}
`
}

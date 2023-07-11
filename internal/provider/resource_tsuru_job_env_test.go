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

func TestAccResourceTsuruJobEnv(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.16/jobs/:job/env", func(c echo.Context) error {
		envs := []tsuru.EnvVar{}
		switch iterationCount {
		case 1:
			envs = []tsuru.EnvVar{
				{
					Name:      "env1",
					Value:     "10",
					Public:    true,
					ManagedBy: "terraform",
				}, {
					Name:      "env3",
					Value:     "private_value",
					Public:    false,
					ManagedBy: "terraform",
				},
			}
		case 2:
			envs = []tsuru.EnvVar{
				{
					Name:      "env1",
					Value:     "15",
					Public:    true,
					ManagedBy: "terraform",
				}, {
					Name:      "env2",
					Value:     "20",
					Public:    true,
					ManagedBy: "terraform",
				}, {
					Name:      "env3",
					Value:     "private_value",
					Public:    false,
					ManagedBy: "terraform",
				},
			}
		case 3:
			envs = []tsuru.EnvVar{
				{
					Name:      "env1",
					Value:     "15",
					Public:    true,
					ManagedBy: "terraform",
				}, {
					Name:      "env3",
					Value:     "new_private_value",
					Public:    false,
					ManagedBy: "terraform",
				},
			}
		}

		return c.JSON(http.StatusOK, envs)
	})

	fakeServer.POST("/1.13/jobs/:job/env", func(c echo.Context) error {
		envs := tsuru.EnvSetData{}
		c.Bind(&envs)
		iterationCount++

		switch iterationCount {
		case 1:
			assert.Equal(t, 2, len(envs.Envs))
		case 2:
			assert.Equal(t, 3, len(envs.Envs))
		case 3:
			assert.Equal(t, 2, len(envs.Envs))
		default:
			assert.Equal(t, 0, len(envs.Envs))
		}
		assert.Equal(t, "terraform", envs.ManagedBy)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_job_env.env"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "tsuru_job_env" "env" {
						job = "job01"
						environment_variables = {
							env1 = "10"
						}
						private_environment_variables = {
							env3 = "private_value"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "job", "job01"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.env1", "10"),
					resource.TestCheckResourceAttr(resourceName, "private_environment_variables.env3", "private_value"),
				),
			},
			{
				Config: `
					resource "tsuru_job_env" "env" {
						job = "job01"
						environment_variables = {
							env1 = "15",
							env2 = "20"
						}
						private_environment_variables = {
							env3 = "private_value"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "job", "job01"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.env1", "15"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.env2", "20"),
					resource.TestCheckResourceAttr(resourceName, "private_environment_variables.env3", "private_value"),
				),
			},
			{
				Config: `
					resource "tsuru_job_env" "env" {
						job = "job01"
						environment_variables = {
							env1 = "15"
						}
						private_environment_variables = {
							env3 = "new_private_value"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "job", "job01"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.env1", "15"),
					resource.TestCheckResourceAttr(resourceName, "private_environment_variables.env3", "new_private_value"),
				),
			},
		},
	})
}

func testAccResourceTsuruJobEnv_basic() string {
	return `
	resource "tsuru_job_env" "env" {
		job = "job01"
		environment_variables = {
			env1 = "10"
		}
		private_environment_variables = {
			env2 = "20"
		}
	}
`
}

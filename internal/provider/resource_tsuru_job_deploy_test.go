// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceTsuruJobDeploy(t *testing.T) {
	fakeServer := echo.New()

	fakeServer.POST("/1.23/jobs/:job/deploy", func(c echo.Context) error {
		c.Response().Header().Set("X-Tsuru-Eventid", "abc-123")

		formParams, err := c.FormParams()
		if err != nil {
			return err
		}
		assert.Equal(t, url.Values{
			"image":   {"fake-repo/job:0.1.0"},
			"message": {"deploy via terraform"},
			"origin":  {"image"}},
			formParams)

		return c.String(http.StatusOK, "OK")
	})

	iterationCount := 0
	fakeServer.GET("/1.1/events/:eventID", func(c echo.Context) error {
		iterationCount++

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Running": iterationCount < 2,
			"EndTime": "2023-01-04T19:26:20.946Z",
			"EndCustomData": map[string]interface{}{
				"Kind": 3,
				"Data": "GwAAAAJpbWFnZQALAAAAdGVzdDoxLjIuMwAA",
			},
		})
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_job_deploy.deploy"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruJobDeploy_basic(server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "job", "my-job"),
					resource.TestCheckResourceAttr(resourceName, "image", "fake-repo/job:0.1.0"),
					resource.TestCheckResourceAttr(resourceName, "status", "finished"),
					resource.TestCheckResourceAttr(resourceName, "output_image", "test:1.2.3"),
				),
			},
		},
	})
}

func TestAccResourceTsuruJobDeployFailed(t *testing.T) {
	fakeServer := echo.New()

	fakeServer.POST("/1.23/jobs/:job/deploy", func(c echo.Context) error {
		c.Response().Header().Set("X-Tsuru-Eventid", "abc-123")

		formParams, err := c.FormParams()
		if err != nil {
			return err
		}
		assert.Equal(t, url.Values{
			"image":   {"fake-repo/job:0.1.0"},
			"message": {"deploy via terraform"},
			"origin":  {"image"}},
			formParams)

		return c.String(http.StatusOK, "OK")
	})

	fakeServer.GET("/1.1/events/:eventID", func(c echo.Context) error {

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Running": false,
			"Error":   "deploy failed",
			"EndTime": "2023-01-04T19:26:20.946Z",
		})
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceTsuruJobDeploy_basic(server.URL),
				ExpectError: regexp.MustCompile("deploy failed, see details of event ID: abc-123"),
			},
		},
	})
}

func testAccResourceTsuruJobDeploy_basic(serverURL string) string {
	return fmt.Sprintf(`

	provider "tsuru" {
		host = "%s"
	}

	resource "tsuru_job_deploy" "deploy" {
		job = "my-job"
		image = "fake-repo/job:0.1.0"
	}
`, serverURL)
}

// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func writePlatformStreamOK(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/x-json-stream")
	c.Response().WriteHeader(http.StatusOK)
	_, err := c.Response().Write([]byte(`{"Message":"ok\n"}` + "\n"))
	return err
}

func TestAccTsuruPlatform_basic(t *testing.T) {
	disabled := false
	rebuildCount := 0

	fakeServer := echo.New()
	fakeServer.POST("/1.0/platforms", func(c echo.Context) error {
		err := c.Request().ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		assert.Equal(t, "my-platform", c.Request().FormValue("name"))

		file, _, err := c.Request().FormFile("dockerfile_content")
		require.NoError(t, err)
		defer file.Close()
		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "FROM tsuru/base-platform\n", string(content))

		return writePlatformStreamOK(c)
	})
	fakeServer.PUT("/1.0/platforms/:platform", func(c echo.Context) error {
		err := c.Request().ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		assert.Equal(t, "my-platform", c.Param("platform"))

		if d := c.Request().FormValue("disabled"); d != "" {
			disabled = d == "true"
		}

		if file, _, err := c.Request().FormFile("dockerfile_content"); err == nil {
			defer file.Close()
			rebuildCount++
		}

		return writePlatformStreamOK(c)
	})
	fakeServer.GET("/1.6/platforms/:platform", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &tsuru.PlatformInfo{
			Platform: tsuru.Platform{
				Name:     c.Param("platform"),
				Disabled: disabled,
			},
			Images: []string{"my-platform:v1"},
		})
	})
	fakeServer.DELETE("/1.0/platforms/:platform", func(c echo.Context) error {
		assert.Equal(t, "my-platform", c.Param("platform"))
		return c.NoContent(http.StatusOK)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_platform.my_platform"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruPlatformConfig_basic("my-platform", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "my-platform"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "images.0", "my-platform:v1"),
				),
			},
			{
				Config: testAccTsuruPlatformConfig_basic("my-platform", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
					func(s *terraform.State) error {
						assert.Equal(t, 0, rebuildCount, "no rebuild should happen when only disabled changes")
						return nil
					},
				),
			},
			{
				Config: testAccTsuruPlatformConfig_withTrigger("my-platform", true, "trigger-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					func(s *terraform.State) error {
						assert.Equal(t, 1, rebuildCount, "changing update_trigger should force a rebuild")
						return nil
					},
				),
			},
		},
	})
}

func TestAccTsuruPlatform_disabledOnCreate(t *testing.T) {
	var createDisabled string
	putCalls := 0

	fakeServer := echo.New()
	fakeServer.POST("/1.0/platforms", func(c echo.Context) error {
		err := c.Request().ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		createDisabled = c.Request().FormValue("disabled")

		return writePlatformStreamOK(c)
	})
	fakeServer.PUT("/1.0/platforms/:platform", func(c echo.Context) error {
		putCalls++
		return writePlatformStreamOK(c)
	})
	fakeServer.GET("/1.6/platforms/:platform", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &tsuru.PlatformInfo{
			Platform: tsuru.Platform{
				Name:     c.Param("platform"),
				Disabled: true,
			},
		})
	})
	fakeServer.DELETE("/1.0/platforms/:platform", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_platform.my_platform"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruPlatformConfig_basic("my-platform", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
					func(s *terraform.State) error {
						assert.Equal(t, "true", createDisabled, "create request should send disabled=true")
						assert.Equal(t, 0, putCalls, "no follow-up PUT should be issued when disabled is set at creation")
						return nil
					},
				),
			},
		},
	})
}

func testAccTsuruPlatformConfig_basic(name string, disabled bool) string {
	return fmt.Sprintf(`
resource "tsuru_platform" "my_platform" {
	name     = %q
	disabled = %t
	dockerfile = <<-EOT
	FROM tsuru/base-platform
	EOT
}
`, name, disabled)
}

func testAccTsuruPlatformConfig_withTrigger(name string, disabled bool, trigger string) string {
	return fmt.Sprintf(`
resource "tsuru_platform" "my_platform" {
	name           = %q
	disabled       = %t
	update_trigger = %q
	dockerfile = <<-EOT
	FROM tsuru/base-platform
	EOT
}
`, name, disabled, trigger)
}

func TestAccTsuruPlatform_image(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("/1.0/platforms", func(c echo.Context) error {
		err := c.Request().ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		assert.Equal(t, "my-platform", c.Request().FormValue("name"))

		file, _, err := c.Request().FormFile("dockerfile_content")
		require.NoError(t, err)
		defer file.Close()
		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "FROM tsuru/base-platform:v2\n", string(content))

		return writePlatformStreamOK(c)
	})
	fakeServer.GET("/1.6/platforms/:platform", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &tsuru.PlatformInfo{
			Platform: tsuru.Platform{Name: c.Param("platform")},
			Images:   []string{"my-platform:v1"},
		})
	})
	fakeServer.DELETE("/1.0/platforms/:platform", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_platform.my_platform"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: `
resource "tsuru_platform" "my_platform" {
	name  = "my-platform"
	image = "tsuru/base-platform:v2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "my-platform"),
					resource.TestCheckResourceAttr(resourceName, "image", "tsuru/base-platform:v2"),
				),
			},
		},
	})
}

func TestAccTsuruPlatform_imageConflictsWithDockerfile(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "tsuru_platform" "my_platform" {
	name       = "my-platform"
	image      = "tsuru/base-platform:v2"
	dockerfile = "FROM tsuru/base-platform:v2"
}
`,
				ExpectError: regexp.MustCompile("conflicts with"),
			},
		},
	})
}

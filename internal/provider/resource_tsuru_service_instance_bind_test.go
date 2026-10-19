// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccResourceServiceInstanceBind(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/services/:service/instances/:instance", func(c echo.Context) error {
		service := c.Param("service")
		if service != "service01" {
			return nil
		}

		app := &tsuru.ServiceInstanceInfo{
			Apps:      []string{},
			Teamowner: "my-team",
			Teams:     []string{},
		}

		if iterationCount == 1 {
			app.Apps = []string{
				"app01",
			}
		}

		return c.JSON(http.StatusOK, app)
	})

	fakeServer.PUT("/1.13/services/:service/instances/:instance/apps/:app", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		app := c.Param("app")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "app01", app)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.13/services/:service/instances/:instance/apps/:app", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		app := c.Param("app")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "app01", app)
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_service_instance_bind.instance_bind"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceInstanceBind_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_name", "service01"),
					resource.TestCheckResourceAttr(resourceName, "service_instance", "my-instance"),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "restart_on_update", "false"),
				),
			},
		},
	})
}

func TestAccResourceServiceInstanceJobBind(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/services/:service/instances/:instance", func(c echo.Context) error {
		service := c.Param("service")
		if service != "service01" {
			return nil
		}

		instance := &tsuru.ServiceInstanceInfo{
			Apps:      []string{},
			Jobs:      []string{},
			Teamowner: "my-team",
			Teams:     []string{},
		}

		if iterationCount == 1 {
			instance.Jobs = []string{
				"job01",
			}
		}

		return c.JSON(http.StatusOK, instance)
	})

	fakeServer.PUT("/1.13/services/:service/instances/:instance/jobs/:job", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		job := c.Param("job")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "job01", job)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.13/services/:service/instances/:instance/jobs/:job", func(c echo.Context) error {
		service := c.Param("service")
		instance := c.Param("instance")
		job := c.Param("job")
		assert.Equal(t, "service01", service)
		assert.Equal(t, "my-instance", instance)
		assert.Equal(t, "job01", job)
		iterationCount--
		return c.NoContent(http.StatusOK)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_service_instance_bind.instance_bind"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceInstanceJobBind_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_name", "service01"),
					resource.TestCheckResourceAttr(resourceName, "service_instance", "my-instance"),
					resource.TestCheckResourceAttr(resourceName, "job", "job01"),
				),
			},
		},
	})
}

func serviceInstanceBindReadTestProvider(t *testing.T, instance tsuru.ServiceInstanceInfo) *tsuruProvider {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/1.0/services/service/instances/db", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(instance)
	}))
	t.Cleanup(server.Close)
	return &tsuruProvider{TsuruClient: tsuru.NewAPIClient(&tsuru.Configuration{
		BasePath:      server.URL,
		HTTPClient:    server.Client(),
		DefaultHeader: map[string]string{},
	})}
}

func TestResourceServiceInstanceBindRead(t *testing.T) {
	t.Run("JobWithSameNameAsApp", func(t *testing.T) {
		resource := resourceTsuruServiceInstanceBind()
		data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"service_name":     "service",
			"service_instance": "db",
			"job":              "shared",
		})
		data.SetId("service::db::tsuru-job::shared")
		provider := serviceInstanceBindReadTestProvider(t, tsuru.ServiceInstanceInfo{
			Apps: []string{"shared"},
			Jobs: []string{"shared"},
		})

		diagnostics := resource.ReadContext(context.Background(), data, provider)

		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		require.Equal(t, "service::db::tsuru-job::shared", data.Id())
		require.Equal(t, "shared", data.Get("job"))
		require.Empty(t, data.Get("app"))
	})

	t.Run("ClearsAppFromJobState", func(t *testing.T) {
		resource := resourceTsuruServiceInstanceBind()
		// Older provider versions could populate both fields when names matched.
		data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"app": "shared",
			"job": "shared",
		})
		data.SetId("service::db::tsuru-job::shared")
		provider := serviceInstanceBindReadTestProvider(t, tsuru.ServiceInstanceInfo{
			Apps: []string{"shared"},
			Jobs: []string{"shared"},
		})

		diagnostics := resource.ReadContext(context.Background(), data, provider)

		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		require.Equal(t, "service::db::tsuru-job::shared", data.Id())
		require.Equal(t, "shared", data.Get("job"))
		require.Empty(t, data.Get("app"))
	})

	t.Run("MissingJobWithSameNameAsApp", func(t *testing.T) {
		resource := resourceTsuruServiceInstanceBind()
		data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"job": "shared",
		})
		data.SetId("service::db::tsuru-job::shared")
		provider := serviceInstanceBindReadTestProvider(t, tsuru.ServiceInstanceInfo{
			Apps: []string{"shared"},
		})

		diagnostics := resource.ReadContext(context.Background(), data, provider)

		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		require.Empty(t, data.Id(), "the app binding must not keep a missing job binding in state")
	})

	t.Run("MissingAppWithSameNameAsJob", func(t *testing.T) {
		resource := resourceTsuruServiceInstanceBind()
		data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"app": "shared",
		})
		data.SetId("service::db::shared")
		provider := serviceInstanceBindReadTestProvider(t, tsuru.ServiceInstanceInfo{
			Jobs: []string{"shared"},
		})

		diagnostics := resource.ReadContext(context.Background(), data, provider)

		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		require.Empty(t, data.Id(), "the job binding must not keep a missing app binding in state")
	})

	t.Run("AppWithSameNameAsJob", func(t *testing.T) {
		resource := resourceTsuruServiceInstanceBind()
		data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"app": "shared",
		})
		data.SetId("service::db::shared")
		provider := serviceInstanceBindReadTestProvider(t, tsuru.ServiceInstanceInfo{
			Apps: []string{"shared"},
			Jobs: []string{"shared"},
		})

		diagnostics := resource.ReadContext(context.Background(), data, provider)

		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		require.Equal(t, "service::db::shared", data.Id())
		require.Equal(t, "shared", data.Get("app"))
		require.Empty(t, data.Get("job"))
	})

	t.Run("ImportedJobWithSameNameAsApp", func(t *testing.T) {
		resource := resourceTsuruServiceInstanceBind()
		// Import starts with only the ID; read must populate the target attributes.
		data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{})
		data.SetId("service::db::tsuru-job::shared")
		provider := serviceInstanceBindReadTestProvider(t, tsuru.ServiceInstanceInfo{
			Apps: []string{"shared"},
			Jobs: []string{"shared"},
		})

		diagnostics := resource.ReadContext(context.Background(), data, provider)

		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		require.Equal(t, "service::db::tsuru-job::shared", data.Id())
		require.Equal(t, "service", data.Get("service_name"))
		require.Equal(t, "db", data.Get("service_instance"))
		require.Equal(t, "shared", data.Get("job"))
		require.Empty(t, data.Get("app"))
	})
}

func TestAccServiceInstanceBindError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "tsuru_service_instance_bind" "instance_bind" {
					service_name = "service01"
					service_instance = "my-instance"
					app = "app01"
					job = "job01"
				}`,
				ExpectError: regexp.MustCompile("only one of `app,job` can be specified, but `app,job` were specified"),
			},
		},
	})
}

func testAccResourceServiceInstanceBind_basic() string {
	return `
	resource "tsuru_service_instance_bind" "instance_bind" {
		service_name = "service01"
		service_instance = "my-instance"
		app = "app01"
		restart_on_update = false
	}
`
}

func testAccResourceServiceInstanceJobBind_basic() string {
	return `
	resource "tsuru_service_instance_bind" "instance_bind" {
		service_name = "service01"
		service_instance = "my-instance"
		job = "job01"
	}
`
}

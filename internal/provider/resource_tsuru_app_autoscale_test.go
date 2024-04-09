// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccResourceTsuruAppAutoscalePercentage(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "800m",
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscale_percentage(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "min_units", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "cpu_average", "80%"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscale_percentage() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "80%"
	}
`
}

func TestAccResourceTsuruAppAutoscaleNumber(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "750m",
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscale_number(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "min_units", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "cpu_average", "75"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscale_number() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "75"
	}
`
}

func TestAccResourceTsuruAppAutoscaleMilli(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "800m",
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscale_milli(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "min_units", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "cpu_average", "800m"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscale_milli() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "800m"
	}
`
}

func TestAccResourceTsuruAppAutoscaleWithSchedules(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "800m",
				Schedules: []tsuru.AutoScaleSchedule{
					{
						MinReplicas: 2,
						Start:       "15 7 * * *",
						End:         "0 17 * * *",
						Timezone:    "America/Sao_Paulo",
					},
					{
						MinReplicas: 3,
						Start:       "0 5 * * *",
						End:         "30 5 * * *",
						Timezone:    "UTC",
					},
				},
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscale_schedules(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "min_units", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_units", "10"),
					resource.TestCheckResourceAttr(resourceName, "cpu_average", "80%"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscale_schedules() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "80%"

		schedule {
			min_replicas = 2
			start = "15 7 * * *"
			end = "0 17 * * *"
			timezone = "America/Sao_Paulo"
		}

		schedule {
			min_replicas = 3
			start = "0 5 * * *"
			end = "30 5 * * *"
			timezone = "UTC"
		}
	}
`
}

func TestAccResourceTsuruAppAutoscaleWithPrometheus(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "800m",
				Prometheus: []tsuru.AutoScalePrometheus{
					{
						Name:              "prom_metric",
						Threshold:         2.5,
						Query:             "sum(rate(my_query{app='my-app'}[5m]))",
						PrometheusAddress: "http://ronaldo",
					},
				},
			}})
		} else if iterationCount == 2 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:    "web",
				MinUnits:   3,
				MaxUnits:   10,
				AverageCPU: "800m",
				Prometheus: []tsuru.AutoScalePrometheus{
					{
						Name:              "prom_metric",
						Threshold:         2.5,
						Query:             "sum(rate(my_query{app='my-app'}[5m]))",
						PrometheusAddress: "http://my-prometheus.namespace.svc.cluster.local:9090",
					},
				},
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscalePrometheus(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.name", "prom_metric"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.threshold", "2.5"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.query", "sum(rate(my_query{app='my-app'}[5m]))"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.custom_address", ""),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.prometheus_address", "http://ronaldo"),
				),
			},
			{
				Config: testAccResourceTsuruAppAutoscalePrometheusWithCustomAddress(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.name", "prom_metric"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.threshold", "2.5"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.query", "sum(rate(my_query{app='my-app'}[5m]))"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.custom_address", "http://my-prometheus.namespace.svc.cluster.local:9090"),
					resource.TestCheckResourceAttr(resourceName, "prometheus.0.prometheus_address", "http://my-prometheus.namespace.svc.cluster.local:9090"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscalePrometheus() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "80%"

		prometheus {
			name           = "prom_metric"
			threshold      = 2.5
			query          = "sum(rate(my_query{app='my-app'}[5m]))"
		}
	}
`
}

func testAccResourceTsuruAppAutoscalePrometheusWithCustomAddress() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10
		cpu_average = "80%"

		prometheus {
			name           = "prom_metric"
			threshold      = 2.5
			query          = "sum(rate(my_query{app='my-app'}[5m]))"
			custom_address = "http://my-prometheus.namespace.svc.cluster.local:9090"
		}
	}
`
}

func TestAccResourceTsuruAppAutoscaleWithoutCPU(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		return c.JSON(http.StatusOK, &tsuru.App{
			Name:        name,
			Description: "my beautiful application",
			TeamOwner:   "myteam",
			Teams: []string{
				"mysupport-team",
				"mysponsors",
			},
			Cluster:     "my-cluster-01",
			Pool:        "my-pool",
			Provisioner: "kubernetes",
			Deploys:     2,
		})

	})

	fakeServer.GET("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		if iterationCount == 1 {
			return c.JSON(http.StatusOK, []tsuru.AutoScaleSpec{{
				Process:  "web",
				MinUnits: 3,
				MaxUnits: 10,
				Schedules: []tsuru.AutoScaleSchedule{
					{
						MinReplicas: 2,
						Start:       "15 7 * * *",
						End:         "0 17 * * *",
						Timezone:    "America/Sao_Paulo",
					},
					{
						MinReplicas: 3,
						Start:       "0 5 * * *",
						End:         "30 5 * * *",
						Timezone:    "UTC",
					},
				},
			}})
		}
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.POST("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		autoscale := tsuru.AutoScaleSpec{}
		c.Bind(&autoscale)
		assert.Equal(t, "web", autoscale.Process)
		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": "true"})
	})

	fakeServer.DELETE("/1.9/apps/:app/units/autoscale", func(c echo.Context) error {
		p := c.QueryParam("process")
		assert.Equal(t, "web", p)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app_autoscale.autoscale"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruAppAutoscale_withoutCpu(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "app01"),
					resource.TestCheckResourceAttr(resourceName, "process", "web"),
					resource.TestCheckResourceAttr(resourceName, "min_units", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_units", "10"),
				),
			},
		},
	})
}

func testAccResourceTsuruAppAutoscale_withoutCpu() string {
	return `
	resource "tsuru_app_autoscale" "autoscale" {
		app = "app01"
		process = "web"
		min_units = 3
		max_units = 10

		schedule {
			min_replicas = 2
			start = "15 7 * * *"
			end = "0 17 * * *"
			timezone = "America/Sao_Paulo"
		}

		schedule {
			min_replicas = 3
			start = "0 5 * * *"
			end = "30 5 * * *"
			timezone = "UTC"
		}
	}
`
}

func TestAccTsuruAutoscaleSetShouldErrorWithoutCPUOrScheduleOrPrometheus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "tsuru_app_autoscale" "autoscale" {
					app = "app01"
					process = "web"
					min_units = 3
					max_units = 10
				}`,
				ExpectError: regexp.MustCompile("one of `cpu_average,prometheus,schedule` must be specified"),
			},
		},
	})
}

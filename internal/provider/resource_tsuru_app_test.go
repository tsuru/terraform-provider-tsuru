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

func TestAccResourceTsuruApp(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/platforms", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Platform{{Name: "python"}})
	})

	fakeServer.GET("/1.0/pools", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Pool{{Name: "prod"}})
	})

	fakeServer.GET("/1.0/plans", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Plan{{Name: "c2m4"}})
	})

	fakeServer.POST("/1.0/apps", func(c echo.Context) error {
		app := tsuru.InputApp{}
		c.Bind(&app)
		assert.Equal(t, "app01", app.Name)
		assert.Equal(t, "my app description", app.Description)
		assert.Equal(t, "python", app.Platform)
		assert.Equal(t, "c2m4", app.Plan)
		assert.Equal(t, "my-team", app.TeamOwner)
		assert.Equal(t, "prod", app.Pool)
		assert.Equal(t, []string{"tagA", "tagB"}, app.Tags)
		iterationCount++
		return c.JSON(http.StatusOK, tsuru.AppCreateResponse{Status: "created"})
	})

	fakeServer.GET("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "app01" {
			return nil
		}

		if iterationCount == 1 {
			app := &tsuru.App{
				Name:        name,
				Description: "my app description",
				TeamOwner:   "my-team",
				Platform:    "python",
				Plan:        tsuru.Plan{Name: "c2m4"},
				Cluster:     "my-cluster-01",
				Pool:        "prod",
				Provisioner: "kubernetes",
				Tags:        []string{"tagA", "tagB"},
				Metadata: tsuru.Metadata{
					Annotations: []tsuru.MetadataItem{{Name: "annotation1", Value: "some really long value"}},
					Labels:      []tsuru.MetadataItem{{Name: "label1", Value: "value1"}},
				},
				Deploys: 2,
				Units: []tsuru.Unit{
					{Processname: "web"},
				},
				InternalAddresses: []tsuru.AppInternalAddresses{
					{
						Version:  "10",
						Port:     8888,
						Process:  "web",
						Domain:   "app01.namespace.svc.cluster.local",
						Protocol: "TCP",
					},
				},
				Routers: []tsuru.AppRouters{
					{
						Name: "default-router",
						Addresses: []string{
							"my-app.router.io",
						},
					},
				},
				Processes: []tsuru.AppProcess{
					{
						Name: "web",
						Plan: "c2m2",
						Metadata: tsuru.Metadata{
							Labels: []tsuru.MetadataItem{
								{
									Name:  "weblabel",
									Value: "value",
								},
							},
							Annotations: []tsuru.MetadataItem{
								{
									Name:  "webannotation",
									Value: "nice",
								},
							},
						},
					},
					{
						Name: "worker",
						Metadata: tsuru.Metadata{
							Labels: []tsuru.MetadataItem{
								{
									Name:  "workerlabel",
									Value: "value",
								},
							},
							Annotations: []tsuru.MetadataItem{
								{
									Name:  "workerannotation",
									Value: "nice",
								},
							},
						},
					},
				},
			}
			return c.JSON(http.StatusOK, app)
		}

		if iterationCount == 2 {
			app := &tsuru.App{
				Name:        name,
				Description: "my app description",
				TeamOwner:   "my-team",
				Platform:    "python",
				Plan:        tsuru.Plan{Name: "c2m4"},
				Cluster:     "my-cluster-01",
				Pool:        "prod",
				Provisioner: "kubernetes",
				Tags:        []string{"tagA", "tagB"},
				Metadata: tsuru.Metadata{
					Labels: []tsuru.MetadataItem{{Name: "label1", Value: "value1"}, {Name: "label3", Value: "value3"}},
				},
				Deploys: 2,
				Units: []tsuru.Unit{
					{Processname: "web"},
				},
				InternalAddresses: []tsuru.AppInternalAddresses{
					{
						Version:  "10",
						Port:     8888,
						Process:  "web",
						Domain:   "app01.namespace.svc.cluster.local",
						Protocol: "TCP",
					},
				},
				Routers: []tsuru.AppRouters{
					{
						Name: "default-router",
						Addresses: []string{
							"my-app.router.io",
						},
					},
				},
				Processes: []tsuru.AppProcess{
					{
						Name: "web",
						Plan: "c2m2",
						Metadata: tsuru.Metadata{
							Labels: []tsuru.MetadataItem{
								{
									Name:  "weblabel",
									Value: "value",
								},
							},
							Annotations: []tsuru.MetadataItem{
								{
									Name:  "webannotation",
									Value: "nice",
								},
							},
						},
					},
					{
						Name: "worker",
						Metadata: tsuru.Metadata{
							Labels: []tsuru.MetadataItem{
								{
									Name:  "workerlabel",
									Value: "value",
								},
							},
							Annotations: []tsuru.MetadataItem{
								{
									Name:  "workerannotation",
									Value: "nice",
								},
							},
						},
					},
				},
			}
			return c.JSON(http.StatusOK, app)
		}

		return c.JSON(http.StatusNotFound, nil)
	})

	fakeServer.PUT("/1.0/apps/:name", func(c echo.Context) error {
		iterationCount++
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.DELETE("/1.0/apps/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "app01", name)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_app.app"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruApp_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "app01"),
					resource.TestCheckResourceAttr(resourceName, "description", "my app description"),
					resource.TestCheckResourceAttr(resourceName, "platform", "python"),
					resource.TestCheckResourceAttr(resourceName, "plan", "c2m4"),
					resource.TestCheckResourceAttr(resourceName, "team_owner", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "pool", "prod"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tagA"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tagB"),
					resource.TestCheckResourceAttr(resourceName, "router.0.name", "default-router"),
					resource.TestCheckResourceAttr(resourceName, "internal_address.0.domain", "app01.namespace.svc.cluster.local"),
				),
			},
			{
				Config: testAccResourceTsuruApp_metadataAfterUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "app01"),
					resource.TestCheckResourceAttr(resourceName, "description", "my app description"),
					resource.TestCheckResourceAttr(resourceName, "platform", "python"),
					resource.TestCheckResourceAttr(resourceName, "plan", "c2m4"),
					resource.TestCheckResourceAttr(resourceName, "team_owner", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "pool", "prod"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tagA"),
				),
			},
			{
				Config:      testAccResourceTsuruApp_invalidProcessOrder(),
				ExpectError: regexp.MustCompile("please, sort app processes alphabetically"),
			},
		},
	})
}

func testAccResourceTsuruApp_basic() string {
	return `
	resource "tsuru_app" "app" {
		name = "app01"
		description = "my app description"
		platform = "python"
		plan = "c2m4"
		team_owner = "my-team"
		pool = "prod"
		tags = ["tagA", "tagB"]
		metadata {
			labels = {
				"label1" = "value1"
			}
			annotations = {
				"annotation1": "some really long value"
			}
		}

		process {
			name = "web"
			plan = "c2m2"
			metadata {
				labels = {
					"weblabel" = "value"
				}
				annotations = {
					"webannotation": "nice"
				}
			}
		}

		process {
			name = "worker"

			metadata {
				labels = {
					"workerlabel" = "value"
				}
				annotations = {
					"workerannotation": "nice"
				}
			}
		}
	}
`
}

func testAccResourceTsuruApp_metadataAfterUpdate() string {
	return `
	resource "tsuru_app" "app" {
		name = "app01"
		description = "my app description"
		platform = "python"
		plan = "c2m4"
		team_owner = "my-team"
		pool = "prod"
		tags = ["tagA", "tagB"]
		metadata {
			labels = {
				"label1" = "value1"
				"label3" = "value3"
			}
		}

		process {
			name = "web"
			plan = "c2m2"
			metadata {
				labels = {
					"weblabel" = "value"
				}
				annotations = {
					"webannotation": "nice"
				}
			}
		}

		process {
			name = "worker"

			metadata {
				labels = {
					"workerlabel" = "value"
				}
				annotations = {
					"workerannotation": "nice"
				}
			}
		}
	}
`
}

func testAccResourceTsuruApp_invalidProcessOrder() string {
	return `
	resource "tsuru_app" "app" {
		name = "app01"
		description = "my app description"
		platform = "python"
		plan = "c2m4"
		team_owner = "my-team"
		pool = "prod"
		tags = ["tagA", "tagB"]

		process {
			name = "worker"
			
			metadata {
				labels = {
					"workerlabel" = "value"
				}
				annotations = {
					"workerannotation" = "nice"
				}
			}
		}

		process {
			name = "web"
			plan = "c2m2"
			metadata {
				labels = {
					"weblabel" = "value"
				}
				annotations = {
					"webannotation" = "nice"
				}
			}
		}
	}
    `
}

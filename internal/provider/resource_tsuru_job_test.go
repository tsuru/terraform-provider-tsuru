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
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccResourceTsuruJob(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/pools", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Pool{{Name: "prod"}})
	})

	fakeServer.GET("/1.0/plans", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Plan{{Name: "c1m1"}})
	})

	fakeServer.POST("/1.13/jobs", func(c echo.Context) error {
		job := tsuru.InputJob{}
		c.Bind(&job)
		assert.Equal(t, "job01", job.Name)
		assert.Equal(t, "my job description", job.Description)
		assert.Equal(t, "c1m1", job.Plan)
		assert.Equal(t, "my-team", job.TeamOwner)
		assert.Equal(t, "prod", job.Pool)
		assert.Equal(t, "* * * * *", job.Schedule)
		assert.Equal(t, []string{"sleep", "600"}, job.Container.Command)
		assert.Equal(t, "tsuru/scratch:latest", job.Container.Image)

		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"jobName": job.Name,
		})
	})

	fakeServer.GET("/1.13/jobs/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "job01" {
			return nil
		}

		if iterationCount == 1 {
			job := &tsuru.Job{
				Name:        name,
				Description: "my job description",
				TeamOwner:   "my-team",
				Plan:        tsuru.Plan{Name: "c1m1"},
				Pool:        "prod",
				Spec: tsuru.JobSpec{
					Schedule: "* * * * *",
					Container: tsuru.InputJobContainer{
						Image: "tsuru/scratch:latest",
						Command: []string{
							"sleep",
							"600",
						},
					},
				},
			}
			return c.JSON(http.StatusOK, tsuru.JobInfo{Job: *job})
		}

		return c.JSON(http.StatusNotFound, nil)
	})

	fakeServer.PUT("/1.13/jobs/:name", func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.DELETE("/1.13/jobs/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "job01", name)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_job.job"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruJob_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "job01"),
					resource.TestCheckResourceAttr(resourceName, "description", "my job description"),
					resource.TestCheckResourceAttr(resourceName, "plan", "c1m1"),
					resource.TestCheckResourceAttr(resourceName, "team_owner", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "pool", "prod"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "* * * * *"),
				),
			},
		},
	})
}

func TestAccResourceTsuruJobComplete(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.GET("/1.0/pools", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Pool{{Name: "prod"}})
	})

	fakeServer.GET("/1.0/plans", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []tsuru.Plan{{Name: "c1m1"}})
	})

	fakeServer.POST("/1.13/jobs", func(c echo.Context) error {
		job := tsuru.InputJob{}
		c.Bind(&job)
		assert.Equal(t, "job01", job.Name)
		assert.Equal(t, "my job description", job.Description)
		assert.Equal(t, "c1m1", job.Plan)
		assert.Equal(t, "my-team", job.TeamOwner)
		assert.Equal(t, "prod", job.Pool)
		assert.Equal(t, "* * * * *", job.Schedule)
		assert.Equal(t, []string{"sleep", "600"}, job.Container.Command)
		assert.Equal(t, "tsuru/scratch:latest", job.Container.Image)
		require.NotNil(t, job.ConcurrencyPolicy)
		assert.Equal(t, "Forbid", *job.ConcurrencyPolicy)

		require.NotNil(t, job.ActiveDeadlineSeconds)
		assert.Equal(t, int64(300), *job.ActiveDeadlineSeconds)

		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"jobName": job.Name,
		})
	})

	fakeServer.GET("/1.13/jobs/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name != "job01" {
			return nil
		}

		concurrencyPolicy := "Forbid"
		activeDeadLineSeconds := int64(300)

		if iterationCount == 1 {
			job := &tsuru.Job{
				Name:        name,
				Description: "my job description",
				TeamOwner:   "my-team",
				Plan:        tsuru.Plan{Name: "c1m1"},
				Pool:        "prod",
				Spec: tsuru.JobSpec{
					Schedule: "* * * * *",
					Container: tsuru.InputJobContainer{
						Image: "tsuru/scratch:latest",
						Command: []string{
							"sleep",
							"600",
						},
					},
					ConcurrencyPolicy:     &concurrencyPolicy,
					ActiveDeadlineSeconds: &activeDeadLineSeconds,
				},
			}
			return c.JSON(http.StatusOK, tsuru.JobInfo{Job: *job})
		}

		return c.JSON(http.StatusNotFound, nil)
	})

	fakeServer.PUT("/1.13/jobs/:name", func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.DELETE("/1.13/jobs/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "job01", name)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_job.job"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruJob_complete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "job01"),
					resource.TestCheckResourceAttr(resourceName, "description", "my job description"),
					resource.TestCheckResourceAttr(resourceName, "plan", "c1m1"),
					resource.TestCheckResourceAttr(resourceName, "team_owner", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "pool", "prod"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "* * * * *"),
				),
			},
		},
	})
}

func testAccResourceTsuruJob_basic() string {
	return `
	resource "tsuru_job" "job" {
		name = "job01"
		description = "my job description"
		plan = "c1m1"
		team_owner = "my-team"
		pool = "prod"
		schedule = "* * * * *"
		container {
			image   = "tsuru/scratch:latest"
			command = ["sleep", 600]
		}
	}
`
}

func testAccResourceTsuruJob_complete() string {
	return `
	resource "tsuru_job" "job" {
		name = "job01"
		description = "my job description"
		plan = "c1m1"
		team_owner = "my-team"
		pool = "prod"
		schedule = "* * * * *"
		container {
			image   = "tsuru/scratch:latest"
			command = ["sleep", 600]
		}


		active_deadline_seconds = 300
		concurrency_policy      = "Forbid"
	}
`
}

// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

type fakeTeamStore struct {
	mu    sync.Mutex
	teams map[string][]string
}

func newFakeTeamServer(t *testing.T, store *fakeTeamStore) *echo.Echo {
	fakeServer := echo.New()
	fakeServer.POST("/1.0/teams", func(c echo.Context) error {
		args := tsuru.TeamCreateArgs{}
		require.NoError(t, c.Bind(&args))
		store.mu.Lock()
		defer store.mu.Unlock()
		if _, ok := store.teams[args.Name]; ok {
			return c.JSON(http.StatusConflict, map[string]string{"error": "team already exists"})
		}
		store.teams[args.Name] = args.Tags
		return c.NoContent(http.StatusCreated)
	})
	fakeServer.GET("/1.0/teams", func(c echo.Context) error {
		store.mu.Lock()
		defer store.mu.Unlock()
		if len(store.teams) == 0 {
			return c.NoContent(http.StatusNoContent)
		}
		teams := []tsuru.Team{}
		for name, tags := range store.teams {
			teams = append(teams, tsuru.Team{Name: name, Tags: tags, Permissions: []string{"team"}})
		}
		return c.JSON(http.StatusOK, teams)
	})
	fakeServer.PUT("/1.6/teams/:name", func(c echo.Context) error {
		name := c.Param("name")
		args := tsuru.TeamUpdateArgs{}
		require.NoError(t, c.Bind(&args))
		store.mu.Lock()
		defer store.mu.Unlock()
		if _, ok := store.teams[name]; !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "team not found"})
		}
		if args.Newname != "" {
			delete(store.teams, name)
			name = args.Newname
		}
		store.teams[name] = args.Tags
		return c.NoContent(http.StatusOK)
	})
	fakeServer.DELETE("/1.0/teams/:name", func(c echo.Context) error {
		name := c.Param("name")
		store.mu.Lock()
		defer store.mu.Unlock()
		if _, ok := store.teams[name]; !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "team not found"})
		}
		delete(store.teams, name)
		return c.NoContent(http.StatusOK)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	return fakeServer
}

func TestAccTsuruTeam_basic(t *testing.T) {
	store := &fakeTeamStore{teams: map[string][]string{
		"other-team": {"unmanaged"},
	}}
	server := httptest.NewServer(newFakeTeamServer(t, store))
	defer server.Close()
	os.Setenv("TSURU_TARGET", server.URL)

	checkStoredTeam := func(name string, tags ...string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			store.mu.Lock()
			defer store.mu.Unlock()
			stored, ok := store.teams[name]
			if !ok {
				return fmt.Errorf("team %q not found in fake server", name)
			}
			assert.Equal(t, tags, stored)
			return nil
		}
	}

	resourceName := "tsuru_team.team"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			store.mu.Lock()
			defer store.mu.Unlock()
			for name := range store.teams {
				if name != "other-team" {
					return fmt.Errorf("team %q still exists", name)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruTeamConfig("my-team", "backend", "payments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "name", "my-team"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "backend"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "payments"),
					checkStoredTeam("my-team", "backend", "payments"),
				),
			},
			{
				Config: testAccTsuruTeamConfig("my-team", "backend"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "backend"),
					checkStoredTeam("my-team", "backend"),
				),
			},
			{
				Config: testAccTsuruTeamConfig("renamed-team", "backend"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "renamed-team"),
					resource.TestCheckResourceAttr(resourceName, "name", "renamed-team"),
					checkStoredTeam("renamed-team", "backend"),
					func(s *terraform.State) error {
						store.mu.Lock()
						defer store.mu.Unlock()
						assert.NotContains(t, store.teams, "my-team")
						return nil
					},
				),
			},
			{
				Config: testAccTsuruTeamConfig("renamed-team"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					checkStoredTeam("renamed-team"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "renamed-team",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTsuruTeam_removedOutsideTerraform(t *testing.T) {
	store := &fakeTeamStore{teams: map[string][]string{}}
	server := httptest.NewServer(newFakeTeamServer(t, store))
	defer server.Close()
	os.Setenv("TSURU_TARGET", server.URL)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruTeamConfig("my-team"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists("tsuru_team.team"),
					func(s *terraform.State) error {
						store.mu.Lock()
						defer store.mu.Unlock()
						delete(store.teams, "my-team")
						return nil
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTsuruTeamConfig(name string, tags ...string) string {
	quoted := make([]string, len(tags))
	for i, tag := range tags {
		quoted[i] = fmt.Sprintf("%q", tag)
	}
	return fmt.Sprintf(`
resource "tsuru_team" "team" {
	name = %q
	tags = [%s]
}
`, name, strings.Join(quoted, ", "))
}

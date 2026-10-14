// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roleDefaultRequest struct {
	Method  string
	Payload roleDefaultValues
}

func TestAccTsuruRoleDefault_lifecycle(t *testing.T) {
	var mutex sync.Mutex
	roleEvents := map[string]map[string]bool{}
	var roleDefaultRequests []roleDefaultRequest

	recordRoleDefaultRequest := func(method string, payload roleDefaultValues) {
		mutex.Lock()
		defer mutex.Unlock()
		cloned := roleDefaultValues{
			UserCreate: append([]string(nil), payload.UserCreate...),
			TeamCreate: append([]string(nil), payload.TeamCreate...),
		}
		sort.Strings(cloned.UserCreate)
		sort.Strings(cloned.TeamCreate)
		roleDefaultRequests = append(roleDefaultRequests, roleDefaultRequest{Method: method, Payload: cloned})
	}

	applyRoleDefaultPayload := func(payload roleDefaultValues, add bool) {
		mutex.Lock()
		defer mutex.Unlock()
		assign := func(evtName string, roles []string) {
			for _, roleName := range roles {
				if add {
					if roleEvents[roleName] == nil {
						roleEvents[roleName] = map[string]bool{}
					}
					roleEvents[roleName][evtName] = true
				} else if roleEvents[roleName] != nil {
					delete(roleEvents[roleName], evtName)
				}
			}
		}
		assign(roleEventUserCreate, payload.UserCreate)
		assign(roleEventTeamCreate, payload.TeamCreate)
	}

	fakeServer := echo.New()
	fakeServer.POST("/1.0/role/default", func(c echo.Context) error {
		var payload roleDefaultValues
		require.NoError(t, c.Bind(&payload))
		recordRoleDefaultRequest(http.MethodPost, payload)
		applyRoleDefaultPayload(payload, true)
		return c.NoContent(http.StatusOK)
	})
	fakeServer.DELETE("/1.0/role/default", func(c echo.Context) error {
		var payload roleDefaultValues
		require.NoError(t, c.Bind(&payload))
		recordRoleDefaultRequest(http.MethodDelete, payload)
		applyRoleDefaultPayload(payload, false)
		return c.NoContent(http.StatusOK)
	})
	fakeServer.GET("/1.0/role/default", func(c echo.Context) error {
		mutex.Lock()
		defer mutex.Unlock()
		roles := []tsuruRoleWithEvents{}
		for roleName, events := range roleEvents {
			var evts []string
			for evt, enabled := range events {
				if enabled {
					evts = append(evts, evt)
				}
			}
			if len(evts) == 0 {
				continue
			}
			sort.Strings(evts)
			roles = append(roles, tsuruRoleWithEvents{Name: roleName, Events: evts})
		}
		sort.Slice(roles, func(i, j int) bool { return roles[i].Name < roles[j].Name })
		return c.JSON(http.StatusOK, roles)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("method=%s, path=%s, err=%s", c.Request().Method, c.Path(), err)
	}

	roleEvents["legacy-admin"] = map[string]bool{roleEventUserCreate: true}

	server := httptest.NewServer(fakeServer)
	defer server.Close()
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_role_default.this"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			mutex.Lock()
			defer mutex.Unlock()
			last := roleDefaultRequests[len(roleDefaultRequests)-1]
			require.Equal(t, http.MethodDelete, last.Method)
			require.Equal(t, roleDefaultValues{
				UserCreate: []string{"billing-viewer", "team-member"},
			}, last.Payload)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "tsuru_role_default" "this" {
	user_create = ["team-member"]
	team_create = ["team-owner"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_create.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_create.*", "team-member"),
					resource.TestCheckResourceAttr(resourceName, "team_create.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "team_create.*", "team-owner"),
					func(_ *terraform.State) error {
						mutex.Lock()
						defer mutex.Unlock()
						require.Contains(t, roleDefaultRequests, roleDefaultRequest{
							Method:  http.MethodDelete,
							Payload: roleDefaultValues{UserCreate: []string{"legacy-admin"}},
						})
						require.NotContains(t, roleEvents["legacy-admin"], roleEventUserCreate)
						roleDefaultRequests = nil
						return nil
					},
				),
			},
			{
				Config: `
resource "tsuru_role_default" "this" {
	user_create = ["team-member", "billing-viewer"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_create.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_create.*", "team-member"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_create.*", "billing-viewer"),
					resource.TestCheckResourceAttr(resourceName, "team_create.#", "0"),
					func(_ *terraform.State) error {
						mutex.Lock()
						defer mutex.Unlock()
						require.Equal(t, []roleDefaultRequest{
							{
								Method:  http.MethodDelete,
								Payload: roleDefaultValues{TeamCreate: []string{"team-owner"}},
							},
							{
								Method:  http.MethodPost,
								Payload: roleDefaultValues{UserCreate: []string{"billing-viewer"}},
							},
						}, roleDefaultRequests)
						return nil
					},
				),
			},
			{
				Config: `
resource "tsuru_role_default" "this" {
	user_create = ["team-member", "billing-viewer"]
}
`,
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: "something-else",
				ExpectError:   regexp.MustCompile(`singleton resource`),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "default",
				ImportStateVerify: true,
			},
		},
	})

	mutex.Lock()
	defer mutex.Unlock()
	assert.NotContains(t, roleEvents["team-owner"], roleEventTeamCreate)
}

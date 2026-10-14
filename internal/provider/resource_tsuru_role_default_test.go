// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
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
	Payload map[string][]string
}

func TestAccTsuruRoleDefault_lifecycle(t *testing.T) {
	var mutex sync.Mutex
	roleEvents := map[string]map[string]bool{}
	var roleDefaultRequests []roleDefaultRequest

	recordRoleDefaultRequest := func(method string, payload map[string][]string) {
		mutex.Lock()
		defer mutex.Unlock()
		cloned := make(map[string][]string, len(payload))
		for event, roles := range payload {
			copied := append([]string(nil), roles...)
			sort.Strings(copied)
			cloned[event] = copied
		}
		roleDefaultRequests = append(roleDefaultRequests, roleDefaultRequest{Method: method, Payload: cloned})
	}

	applyRoleDefaultPayload := func(payload map[string][]string, add bool) {
		mutex.Lock()
		defer mutex.Unlock()
		for evtName, roles := range payload {
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
	}

	fakeServer := echo.New()
	fakeServer.POST("/1.0/role/default", func(c echo.Context) error {
		var payload map[string][]string
		require.NoError(t, c.Bind(&payload))
		recordRoleDefaultRequest(http.MethodPost, payload)
		applyRoleDefaultPayload(payload, true)
		return c.NoContent(http.StatusOK)
	})
	fakeServer.DELETE("/1.0/role/default", func(c echo.Context) error {
		var payload map[string][]string
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
							Method: http.MethodDelete,
							Payload: map[string][]string{
								roleEventUserCreate: {"legacy-admin"},
							},
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
								Method: http.MethodDelete,
								Payload: map[string][]string{
									roleEventTeamCreate: {"team-owner"},
								},
							},
							{
								Method: http.MethodPost,
								Payload: map[string][]string{
									roleEventUserCreate: {"billing-viewer"},
								},
							},
						}, roleDefaultRequests)
						return nil
					},
				),
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

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"fmt"
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

func TestTsuruServiceInstance_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("/1.0/services/rpaasv2/instances", func(c echo.Context) error {
		si := &tsuru.ServiceInstance{}
		err := c.Bind(&si)
		require.NoError(t, err)
		assert.Equal(t, "my-reverse-proxy", si.Name)
		assert.Equal(t, "rpaasv2", si.ServiceName)
		assert.Equal(t, "my-team", si.TeamOwner)
		assert.Equal(t, "My Reverse Proxy", si.Description)
		assert.Equal(t, "c2m2", si.PlanName)
		assert.Equal(t, []string{
			"tag_a",
			"tag_b",
		}, si.Tags)
		assert.Equal(t, map[string]string{
			"value":      "10",
			"otherValue": "false",
		}, si.Parameters)

		return nil
	})
	fakeServer.GET("/1.0/services/rpaasv2/instances/my-reverse-proxy", func(c echo.Context) error {
		return c.JSON(http.StatusOK, tsuru.ServiceInstanceInfo{
			Teamowner: "my-team",
			Planname:  "c2m2",
			Pool:      "some-pool",
		})
	})
	fakeServer.DELETE("/1.0/services/rpaasv2/instances/my-reverse-proxy", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_service_instance.my_reverse_proxy"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruServiceInstanceConfig_basic(server.URL, "my-reverse-proxy"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "my-reverse-proxy"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "rpaasv2"),
				),
			},
		},
	})
}

func testAccTsuruServiceInstanceConfig_basic(fakeServer, name string) string {
	return fmt.Sprintf(`
resource "tsuru_service_instance"  "my_reverse_proxy"   {
	service_name = "rpaasv2"

	name = "%s"
	owner = "my-team"
	description = "My Reverse Proxy"
	pool = "some-pool"

	plan = "c2m2"
	tags = ["tag_a", "tag_b"]
	parameters = {
		"value" = "10"
		"otherValue" = "false"
	}
}
`, name)
}

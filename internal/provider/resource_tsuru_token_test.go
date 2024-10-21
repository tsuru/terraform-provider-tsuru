package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccResourceTsuruToken(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.POST("/1.6/tokens", func(c echo.Context) error {
		teamToken := tsuru.TeamTokenCreateArgs{}
		err := c.Bind(&teamToken)
		require.NoError(t, err)

		if iterationCount == 0 {
			assert.Equal(t, "my-simple-token", teamToken.TokenId)
			assert.Equal(t, "My description", teamToken.Description)
			assert.Equal(t, "team-dev", teamToken.Team)
			assert.Equal(t, int64(86400), teamToken.ExpiresIn)
		}

		if iterationCount == 1 {
			assert.Equal(t, "my-simple-token", teamToken.TokenId)
			assert.Equal(t, "My new description", teamToken.Description)
			assert.Equal(t, "team-dev", teamToken.Team)
			assert.Equal(t, int64(43200), teamToken.ExpiresIn)
		}

		iterationCount++
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":   "success",
			"token_id": teamToken.TokenId,
		})
	})

	fakeServer.GET("/1.7/tokens/:token", func(c echo.Context) error {
		token := c.Param("token")
		if iterationCount == 1 {
			if token != "my-simple-token" {
				return nil
			}

			teamToken := tsuru.TeamToken{
				Token:        "string-token",
				TokenId:      token,
				Description:  "My description",
				CreatedAt:    time.Now(),
				ExpiresAt:    time.Now().Add(time.Hour * 24),
				LastAccess:   time.Now().Add(-time.Hour * 1),
				CreatorEmail: "creator@example.com",
				Team:         "team-dev",
				Roles: []tsuru.RoleInstance{
					{
						Name:         "role-name",
						Contextvalue: "role-context",
					},
				},
			}
			return c.JSON(http.StatusOK, teamToken)
		}

		if iterationCount == 2 {
			if token != "my-simple-token" {
				return nil
			}

			teamToken := tsuru.TeamToken{
				Token:        "string-token",
				TokenId:      token,
				Description:  "My new description",
				CreatedAt:    time.Now(),
				ExpiresAt:    time.Now().Add(time.Hour * 24),
				LastAccess:   time.Now().Add(-time.Hour * 1),
				CreatorEmail: "creator@example.com",
				Team:         "team-dev",
				Roles: []tsuru.RoleInstance{
					{
						Name:         "role-name",
						Contextvalue: "role-context",
					},
				},
			}
			return c.JSON(http.StatusOK, teamToken)
		}

		return c.JSON(http.StatusNotFound, nil)
	})

	fakeServer.PUT("/1.6/tokens/:token", func(c echo.Context) error {
		token := c.Param("token")
		teamTokem := tsuru.TeamToken{}
		c.Bind(&teamTokem)
		assert.Equal(t, "my-simple-token", token)
		assert.Equal(t, "My new description", teamTokem.Description)
		iterationCount++
		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.DELETE("/1.6/tokens/:token", func(c echo.Context) error {
		token := c.Param("token")
		assert.Equal(t, "my-simple-token", token)
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}

	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_token.team_token"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruToken_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "token_id", "my-simple-token"),
					resource.TestCheckResourceAttr(resourceName, "description", "My description"),
					resource.TestCheckResourceAttr(resourceName, "team", "team-dev"),
					resource.TestCheckResourceAttr(resourceName, "expires", "24h"),
				),
			},
			{
				Config: testAccResourceTsuruToken_complete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "token_id", "my-simple-token"),
					resource.TestCheckResourceAttr(resourceName, "description", "My new description"),
					resource.TestCheckResourceAttr(resourceName, "team", "team-dev"),
					resource.TestCheckResourceAttr(resourceName, "expires", "12h"),
					resource.TestCheckResourceAttr(resourceName, "regenerate_on_update", "true"),
				),
			},
		},
	})

}

func testAccResourceTsuruToken_basic() string {
	return `
	resource "tsuru_token" "team_token" {
		token_id = "my-simple-token"
		description = "My description"
		team = "team-dev"
		expires = "24h"
	}
`
}

func testAccResourceTsuruToken_complete() string {
	return `
	resource "tsuru_token" "team_token" {
		token_id = "my-simple-token"
		description = "My new description"
		team = "team-dev"
		expires = "12h"
		regenerate_on_update = "true"
	}
`
}

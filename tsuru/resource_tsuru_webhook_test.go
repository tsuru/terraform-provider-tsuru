package tsuru

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestAccTsuruWebhook_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("/1.6/events/webhooks", func(c echo.Context) error {
		w := tsuru.Webhook{}
		c.Bind(&w)

		if w.Name == "webhook1" {
			assert.Equal(t, "my event", w.Description)
			assert.Equal(t, "myteam", w.TeamOwner)

			assert.Equal(t, []string{
				"target01",
				"target02",
			}, w.EventFilter.TargetTypes)
			assert.Equal(t, []string{
				"targetvalue01",
				"targetvalue02",
			}, w.EventFilter.TargetValues)
			assert.Equal(t, []string{
				"kind_name",
			}, w.EventFilter.KindNames)
			assert.Equal(t, []string{
				"kind_type",
			}, w.EventFilter.KindTypes)
			assert.False(t, w.EventFilter.ErrorOnly)
			assert.True(t, w.EventFilter.SuccessOnly)

			assert.Equal(t, "http://blah.io/webhook", w.Url)
			assert.Equal(t, "http://myproxy.com", w.ProxyUrl)
			assert.Equal(t, map[string][]string{
				"X-Token": {"my-token"},
			}, w.Headers)
			assert.Equal(t, "POST", w.Method)
			assert.Equal(t, "body-test", w.Body)
			assert.True(t, w.Insecure)
		}
		return nil
	})
	fakeServer.GET("/1.6/events/webhooks/:name", func(c echo.Context) error {
		name := c.Param("name")

		if name == "webhook1" {

			return c.JSON(http.StatusOK, &tsuru.Webhook{
				Name:        name,
				Description: "my event",
				TeamOwner:   "myteam",

				EventFilter: tsuru.WebhookEventFilter{
					TargetTypes: []string{"target01", "target02"},
					TargetValues: []string{
						"targetvalue01",
						"targetvalue02",
					},
					KindNames: []string{
						"kind_name",
					},
					KindTypes: []string{
						"kind_type",
					},
					ErrorOnly:   false,
					SuccessOnly: true,
				},

				Url:      "http://blah.io/webhook",
				ProxyUrl: "http://myproxy.com",
				Headers: map[string][]string{
					"X-Token": {"my-token"},
				},
				Method:   "POST",
				Body:     "body-test",
				Insecure: true,
			})
		}
		return nil
	})

	fakeServer.DELETE("/1.6/events/webhooks/:name", func(c echo.Context) error {
		assert.Equal(t, "webhook1", c.Param("name"))
		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruWebhookConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tsuru_webhook.webhook1", "name", "webhook1"),
					resource.TestCheckResourceAttr("tsuru_webhook.webhook1", "event_filter.0.target_types.0", "target01"),
				),
			},
		},
	})
}

func testAccTsuruWebhookConfig_basic() string {
	return `
	resource "tsuru_webhook" "webhook1" {
		name        = "webhook1"
		description = "my event"
		team_owner  = "myteam"
	  
		event_filter {
		  target_types = [
			"target01",
			"target02",
		  ]
		  target_values = [
			"targetvalue01",
			"targetvalue02",
		  ]
	  
		  kind_types = [
			"kind_type"
		  ]
	  
		  kind_names = [
			"kind_name"
		  ]
	  
		  error_only   = false
		  success_only = true
		}
	  
		url       = "http://blah.io/webhook"
		proxy_url = "http://myproxy.com"
		headers = {
		  "X-Token" = "my-token"
		}
	  
		method   = "POST"
		body     = "body-test"
		insecure = true
	  }
	  
`
}

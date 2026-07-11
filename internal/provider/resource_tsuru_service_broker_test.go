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

func newBasicAuthConfigedBroker() *tsuru.ServiceBrokerConfig {
	return &tsuru.ServiceBrokerConfig{
		Insecure:               false,
		CacheExpirationSeconds: 3600,
		Context: map[string]string{
			"test_context_1": "TEST_CONTEXT_1",
			"test_context_2": "TEST_CONTEXT_2",
		},
		AuthConfig: tsuru.ServiceBrokerConfigAuthConfig{
			BasicAuthConfig: tsuru.ServiceBrokerConfigAuthConfigBasicAuthConfig{
				Username: "test_username",
				Password: "test_password",
			},
		},
	}
}

func newBearerAuthConfigedBroker() *tsuru.ServiceBrokerConfig {
	return &tsuru.ServiceBrokerConfig{
		Insecure:               true,
		CacheExpirationSeconds: 3600,
		Context: map[string]string{
			"test_context_1": "TEST_CONTEXT_1",
			"test_context_2": "TEST_CONTEXT_2",
		},
		AuthConfig: tsuru.ServiceBrokerConfigAuthConfig{
			BearerConfig: tsuru.ServiceBrokerConfigAuthConfigBearerConfig{
				Token: "SENSITIVE_TOKEN",
			},
		},
	}
}

func TestAccResourceTsuruServiceBroker(t *testing.T) {
	fakeServer := echo.New()

	iterationCount := 0

	fakeServer.POST("/1.7/brokers", func(c echo.Context) error {
		sb := &tsuru.ServiceBroker{}
		err := c.Bind(&sb)
		require.NoError(t, err)
		if iterationCount == 0 {
			assert.Equal(t, "test_name", sb.Name)
			assert.Equal(t, "https://broker.example.com", sb.URL)
			assert.Equal(t, *newBasicAuthConfigedBroker(), sb.Config)
		}
		iterationCount++

		return nil
	})

	fakeServer.GET("/1.7/brokers", func(c echo.Context) error {
		brokers := []tsuru.ServiceBroker{}
		broker := tsuru.ServiceBroker{}

		if iterationCount == 1 {
			broker = tsuru.ServiceBroker{
				Name:   "test_name",
				URL:    "https://broker.example.com",
				Config: *newBasicAuthConfigedBroker(),
			}
		}

		if iterationCount == 2 {
			broker.Name = "test_name"
			broker.URL = "https://broker.example.com"
			broker.Config = *newBearerAuthConfigedBroker()
		}

		brokers = append(brokers, broker)

		return c.JSON(http.StatusOK, &tsuru.ServiceBrokerList{
			Brokers: brokers,
		})
	})

	fakeServer.PUT("/1.7/brokers/:name", func(c echo.Context) error {
		name := c.Param("name")
		broker := tsuru.ServiceBroker{}
		c.Bind(&broker)
		assert.Equal(t, "test_name", name)
		assert.Equal(t, *newBearerAuthConfigedBroker(), broker.Config)
		iterationCount++

		return c.JSON(http.StatusOK, nil)
	})

	fakeServer.DELETE("/1.7/brokers/:name", func(c echo.Context) error {
		name := c.Param("name")
		assert.Equal(t, "test_name", name)

		return c.NoContent(http.StatusNoContent)
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}

	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_service_broker.test_broker"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTsuruServiceBroker_basicAuth(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test_name"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://broker.example.com"),
					resource.TestCheckResourceAttr(resourceName, "config.0.insecure", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.cache_expiration_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "config.0.context.test_context_1", "TEST_CONTEXT_1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.context.test_context_2", "TEST_CONTEXT_2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.auth_config.0.basic_auth_config.0.username", "test_username"),
					resource.TestCheckResourceAttr(resourceName, "config.0.auth_config.0.basic_auth_config.0.password", "test_password"),
				),
			},
			{
				Config: testAccResourceTsuruServiceBroker_bearerAuth(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test_name"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://broker.example.com"),
					resource.TestCheckResourceAttr(resourceName, "config.0.insecure", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.cache_expiration_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "config.0.context.test_context_1", "TEST_CONTEXT_1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.context.test_context_2", "TEST_CONTEXT_2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.auth_config.0.bearer_config.0.token", "SENSITIVE_TOKEN"),
				),
			},
		},
	})
}

func testAccResourceTsuruServiceBroker_basicAuth() string {
	return `
resource "tsuru_service_broker" "test_broker" {
    name = "test_name"
    url  = "https://broker.example.com"
    config {
        insecure = false
        cache_expiration_seconds = 3600
        context = {
            test_context_1 = "TEST_CONTEXT_1"
            test_context_2 = "TEST_CONTEXT_2"
        }
        auth_config {
            basic_auth_config {
                username = "test_username"
                password = "test_password"
            }
        }
    }
}
`
}

func testAccResourceTsuruServiceBroker_bearerAuth() string {
	return `
resource "tsuru_service_broker" "test_broker" {
    name = "test_name"
    url  = "https://broker.example.com"
    config {
        cache_expiration_seconds = 3600
        insecure = true
        context = {
            test_context_1 = "TEST_CONTEXT_1"
            test_context_2 = "TEST_CONTEXT_2"
        }
        auth_config {
            bearer_config {
                token = "SENSITIVE_TOKEN"
            }
        }
    }
}
`
}

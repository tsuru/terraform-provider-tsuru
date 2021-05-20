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

func TestAccTsuruCluster_basic(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.POST("/1.3/provisioner/clusters", func(c echo.Context) error {
		p := &tsuru.Cluster{}
		err := c.Bind(&p)
		require.NoError(t, err)
		assert.Equal(t, p.Name, "test_cluster")
		assert.Equal(t, p.Addresses, []string{"https://mycluster.local"})
		assert.Equal(t, p.Pools, []string{"my-pool"})
		assert.Equal(t, p.CustomData, map[string]string{"token": "test_token"})

		return nil
	})
	fakeServer.GET("/1.8/provisioner/clusters/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.JSON(http.StatusOK, &tsuru.Cluster{
			Name:        name,
			Provisioner: "kubernetes",
			Pools:       []string{"my-pool"},
			Addresses:   []string{"https://mycluster.local"},
			CustomData: map[string]string{
				"token": "test_token",
			},
		})
	})
	fakeServer.DELETE("/1.3/provisioner/clusters/:name", func(c echo.Context) error {
		name := c.Param("name")
		require.Equal(t, name, "test_cluster")
		return c.NoContent(http.StatusNoContent)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_cluster.test_cluster"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruClusterConfig_basic(server.URL, "test_cluster"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "custom_data.token", "test_token"),
					//resource.TestCheckResourceAttr("tsuru_cluster.test_cluster", "tsuru_provisioner", "kubernetes"),
				),
			},
		},
	})
}

func testAccTsuruClusterConfig_basic(fakeServer, name string) string {
	return fmt.Sprintf(`
resource "tsuru_cluster"  "test_cluster"   {
	name = "%s"
	tsuru_provisioner = "kubernetes"
	addresses = [
		"https://mycluster.local"
	]
	initial_pools = [
		"my-pool"
	]
	custom_data = {
		"token" = "test_token"
	}
}
`, name)
}

func TestAccTsuruCluster_kubeConfig(t *testing.T) {
	testKubeConfig := tsuru.ClusterKubeConfig{
		Cluster: tsuru.ClusterKubeConfigCluster{
			Server:                   "https://mycluster.local",
			CertificateAuthorityData: "server-cert",
			TlsServerName:            "mycluster.local",
			InsecureSkipTlsVerify:    true,
		},
		User: tsuru.ClusterKubeConfigUser{
			AuthProvider: tsuru.ClusterKubeConfigUserAuthprovider{
				Name: "tsuru",
				Config: map[string]string{
					"tsuru-flag-01": "result",
				},
			},
			Token:                 "token",
			Username:              "username",
			Password:              "password",
			ClientCertificateData: "client-cert",
			ClientKeyData:         "client-key",
			Exec: &tsuru.ClusterKubeConfigUserExec{
				ApiVersion: "api-version",
				Command:    "tsuru",
				Args:       []string{"cluster", "login"},
				Env: []tsuru.ClusterKubeConfigUserExecEnv{
					{
						Name:  "TSURU_TOKEN",
						Value: "token",
					},
				},
			},
		},
	}

	simpleKubeConfig := tsuru.ClusterKubeConfig{
		Cluster: tsuru.ClusterKubeConfigCluster{
			Server:                   "https://simplecluster.local",
			CertificateAuthorityData: "server-cert",
		},
		User: tsuru.ClusterKubeConfigUser{
			AuthProvider: tsuru.ClusterKubeConfigUserAuthprovider{
				Name: "mycloud",
			},
		},
	}

	fakeServer := echo.New()
	fakeServer.POST("/1.3/provisioner/clusters", func(c echo.Context) error {
		p := &tsuru.Cluster{}
		err := c.Bind(&p)
		require.NoError(t, err)

		if p.Name == "test_cluster" {
			assert.Nil(t, p.CustomData)
			assert.True(t, p.Default)
			assert.Equal(t, testKubeConfig, p.KubeConfig)
			assert.Equal(t, "http://myproxy.io:3128", p.HttpProxy)
		} else if p.Name == "simple_cluster" {
			assert.Nil(t, p.CustomData)
			assert.False(t, p.Default)
			assert.Equal(t, simpleKubeConfig, p.KubeConfig)
		}

		return nil
	})
	fakeServer.GET("/1.8/provisioner/clusters/:name", func(c echo.Context) error {
		name := c.Param("name")

		if name == "test_cluster" {
			return c.JSON(http.StatusOK, &tsuru.Cluster{
				Name:        name,
				Provisioner: "kubernetes",
				CustomData: map[string]string{
					"token": "test_token",
				},
				Default:    true,
				KubeConfig: testKubeConfig,
				HttpProxy:  "http://myproxy.io:3128",
			})
		}

		if name == "simple_cluster" {
			return c.JSON(http.StatusOK, &tsuru.Cluster{
				Name:        name,
				Provisioner: "kubernetes",
				KubeConfig:  simpleKubeConfig,
			})
		}

		return c.NoContent(http.StatusNotFound)
	})
	fakeServer.DELETE("/1.3/provisioner/clusters/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name == "test_cluster" || name == "simple_cluster" {
			return c.NoContent(http.StatusNoContent)
		}
		return c.NoContent(http.StatusNotFound)
	})
	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("methods=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}
	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_cluster.test_cluster"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     resourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruClusterConfig_kubeConfig(),

				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kube_config.0.cluster.0.server", "https://mycluster.local"),
				),
			},
			{
				ImportState:       true,
				ResourceName:      resourceName,
				ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kube_config.0.cluster.0.server", "https://mycluster.local"),
				),
			},
		},
	})

	simpleResourceName := "tsuru_cluster.simple_cluster"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     simpleResourceName,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTsuruClusterConfig_kubeConfigSimple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(simpleResourceName),
					resource.TestCheckResourceAttr(simpleResourceName, "kube_config.0.cluster.0.server", "https://simplecluster.local"),
				),
			},
			{
				ImportState:       true,
				ResourceName:      simpleResourceName,
				ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(simpleResourceName),
					resource.TestCheckResourceAttr(simpleResourceName, "kube_config.0.cluster.0.server", "https://simplecluster.local"),
				),
			},
		},
	})
}

func testAccTsuruClusterConfig_kubeConfig() string {
	return `
resource "tsuru_cluster" "test_cluster"   {
	name = "test_cluster"
	tsuru_provisioner = "kubernetes"
	default = true

	http_proxy = "http://myproxy.io:3128"

	kube_config {
		cluster {
			server                     = "https://mycluster.local"
			tls_server_name            = "mycluster.local"
			insecure_skip_tls_verify   = true
			certificate_authority_data = "server-cert"
		}
		user {
			auth_provider {
				name = "tsuru"
				config = {
					"tsuru-flag-01" = "result"
				}
			}

			client_certificate_data = "client-cert"
			client_key_data = "client-key"

			token = "token"
			username = "username"
			password = "password"

			exec {
				api_version = "api-version"
				command = "tsuru"
				args = ["cluster", "login"]

				env {
					name = "TSURU_TOKEN"
					value = "token"
				}
			}
		}
	}
}`
}

func testAccTsuruClusterConfig_kubeConfigSimple() string {
	return `
resource "tsuru_cluster" "simple_cluster"   {
	name = "simple_cluster"
	tsuru_provisioner = "kubernetes"
	kube_config {
		cluster {
			server                     = "https://simplecluster.local"
			certificate_authority_data = "server-cert"
		}

		user {
			auth_provider {
				name = "mycloud"
			}
		}
	}
}
`
}

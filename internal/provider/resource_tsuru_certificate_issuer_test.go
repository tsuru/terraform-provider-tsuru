// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

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

func TestAccTsuruCertificateIssuer(t *testing.T) {
	fakeServer := echo.New()
	fakeServer.PUT("/1.24/apps/:app/certissuer", func(c echo.Context) error {
		p := &tsuru.CertIssuerSetData{}
		err := c.Bind(p)
		require.NoError(t, err)
		assert.Equal(t, "my-app", c.Param("app"))
		assert.Equal(t, "my-cname.org", p.Cname)
		assert.Equal(t, "lets-encrypt", p.Issuer)

		return nil
	})

	fakeServer.DELETE("/1.24/apps/:app/certissuer", func(c echo.Context) error {
		assert.Equal(t, "my-app", c.Param("app"))
		assert.Equal(t, "my-cname.org", c.QueryParam("cname"))

		return nil
	})
	fakeServer.GET("/1.24/apps/:app/certificate", func(c echo.Context) error {
		assert.Equal(t, "my-app", c.Param("app"))

		return c.JSON(http.StatusOK, tsuru.AppCertificates{
			Routers: map[string]tsuru.AppCertificatesRouters{
				"https-router": {
					Cnames: map[string]tsuru.AppCertificatesCnames{
						"my-cname.org": {
							Issuer:      "lets-encrypt",
							Certificate: "123",
						},

						"other-cname.org": {
							Issuer:      "self-signed",
							Certificate: "321",
						},
					},
				},
				"other-https-router": {
					Cnames: map[string]tsuru.AppCertificatesCnames{
						"my-cname.org": {
							Issuer:      "rapid-ssl",
							Certificate: "123",
						},

						"other-cname.org": {
							Issuer:      "self-signed",
							Certificate: "4321",
						},
					},
				},
			},
		})
	})

	fakeServer.HTTPErrorHandler = func(err error, c echo.Context) {
		t.Errorf("method=%s, path=%s, err=%s", c.Request().Method, c.Path(), err.Error())
	}

	server := httptest.NewServer(fakeServer)
	os.Setenv("TSURU_TARGET", server.URL)

	resourceName := "tsuru_certificate_issuer.cert"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "tsuru_certificate_issuer.cert",
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:             testAccTsuruCertificateIssuer("my-app", "my-cname.org", "lets-encrypt"),
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "my-app"),
					resource.TestCheckResourceAttr(resourceName, "cname", "my-cname.org"),
					resource.TestCheckResourceAttr(resourceName, "issuer", "lets-encrypt"),
				),
			},
		},
	})
}

func testAccTsuruCertificateIssuer(app, cname, issuer string) string {
	return fmt.Sprintf(`
resource "tsuru_certificate_issuer" "cert" {
	app    = %q
	cname  = %q
	issuer = %q
}
`, app, cname, issuer)
}

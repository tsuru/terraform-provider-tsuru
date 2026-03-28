// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/client"
	"github.com/tsuru/go-tsuruclient/pkg/config"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Description: "Target to tsuru API",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TSURU_HOST", nil),
			},
			"token": {
				Type:        schema.TypeString,
				Description: "Token to authenticate on tsuru API (optional)",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TSURU_TOKEN", nil),
			},
			"skip_cert_verification": {
				Type:        schema.TypeBool,
				Description: "Disable certificate verification",
				Default:     false,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TSURU_SKIP_CERT_VERIFICATION", nil),
			},
			"full_management_of_user_environment_variables": {
				Type:        schema.TypeBool,
				Description: "Use `true` to manage all user environment variables. (Default: false)",
				Default:     false,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TSURU_FULL_MANAGEMENT_OF_USER_ENVIRONMENT_VARIABLES", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tsuru_service_instance_bind":  resourceTsuruServiceInstanceBind(),
			"tsuru_service_instance_grant": resourceTsuruServiceInstanceGrant(),
			"tsuru_service_instance":       resourceTsuruServiceInstance(),

			"tsuru_volume_bind": resourceTsuruVolumeBind(),
			"tsuru_volume":      resourceTsuruVolume(),

			"tsuru_app_autoscale": resourceTsuruApplicationAutoscale(),
			"tsuru_app_env":       resourceTsuruApplicationEnvironment(),
			"tsuru_app_unit":      resourceTsuruApplicationUnits(),
			"tsuru_app_cname":     resourceTsuruApplicationCName(),
			"tsuru_app_router":    resourceTsuruApplicationRouter(),
			"tsuru_app_grant":     resourceTsuruApplicationGrant(),
			"tsuru_app_deploy":    resourceTsuruApplicationDeploy(),
			"tsuru_app":           resourceTsuruApplication(),

			"tsuru_certificate_issuer": resourceTsuruCertificateIssuer(),

			"tsuru_job":        resourceTsuruJob(),
			"tsuru_job_env":    resourceTsuruJobEnvironment(),
			"tsuru_job_deploy": resourceTsuruJobDeploy(),

			"tsuru_router":          resourceTsuruRouter(),
			"tsuru_plan":            resourceTsuruPlan(),
			"tsuru_webhook":         resourceTsuruWebhook(),
			"tsuru_pool_constraint": resourceTsuruPoolConstraint(),
			"tsuru_pool":            resourceTsuruPool(),
			"tsuru_cluster_pool":    resourceTsuruClusterPool(),
			"tsuru_cluster":         resourceTsuruCluster(),
			"tsuru_token":           resourceTsuruToken(),
			"tsuru_platform":        resourceTsuruPlatform(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"tsuru_app": dataSourceTsuruApp(),
		},
	}
	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(ctx, d, p.TerraformVersion)
	}

	return p
}

type tsuruProvider struct {
	Host               string
	Token              string
	TsuruClient        *tsuru.APIClient
	FullManagementEnvs bool
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	userAgent := fmt.Sprintf("HashiCorp/1.0 Terraform/%s", terraformVersion)

	cfg := &tsuru.Configuration{
		DefaultHeader: map[string]string{},
		UserAgent:     userAgent,
	}

	if d.Get("skip_cert_verification").(bool) {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		cfg.HTTPClient = &http.Client{
			Transport: transport,
		}
	}

	var err error
	host := d.Get("host").(string)
	if host == "" {
		host = os.Getenv("TSURU_TARGET")
	}
	if host == "" {
		host, err = config.GetTarget()
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	cfg.BasePath = host
	os.Setenv("TSURU_TARGET", host)

	token := d.Get("token").(string)
	if token != "" {
		cfg.DefaultHeader["Authorization"] = token
	}

	client, err := client.ClientFromEnvironment(cfg)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	fullManagementEnvs := d.Get("full_management_of_user_environment_variables").(bool)

	return &tsuruProvider{
		Host:               host,
		Token:              token,
		TsuruClient:        client,
		FullManagementEnvs: fullManagementEnvs,
	}, nil
}

func logTsuruStream(in io.Reader) {
	reader := bufio.NewScanner(in)
	for reader.Scan() {
		log.Println("[INFO] ", reader.Text())
	}
}

// Copyright 2023 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	tsuruClientConfig "github.com/tsuru/tsuru-client/tsuru/config"
)

func resourceTsuruApplicationDeploy() *schema.Resource {
	return &schema.Resource{
		Description:   "Perform an application deploy. Currently, only supporting deploys via prebuilt container images; in order to deploy via tsuru platforms please use tsuru-client",
		CreateContext: resourceTsuruApplicationDeployDo,
		UpdateContext: resourceTsuruApplicationDeployDo,
		ReadContext:   resourceTsuruApplicationDeployRead,
		DeleteContext: resourceTsuruApplicationDeployDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Description: "Application name",
				Required:    true,
				ForceNew:    true,
			},
			"image": {
				Type:        schema.TypeString,
				Description: "Docker Image",
				Required:    true,
			},

			"new_version": {
				Type:        schema.TypeBool,
				Description: "Creates a new version for the current deployment while preserving existing versions",
				Optional:    true,
				Default:     false,
			},

			"override_old_versions": {
				Type:        schema.TypeBool,
				Description: "Force replace all deployed versions by this new deploy",
				Optional:    true,
				Default:     false,
			},

			"wait": {
				Type:        schema.TypeBool,
				Description: "Wait for the rollout of deploy",
				Optional:    true,
				Default:     true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: "after apply may be three kinds of statuses: running or failed or finished",
				Computed:    true,
			},

			"output_image": {
				Type:        schema.TypeString,
				Description: "Image generated after success of deploy",
				Computed:    true,
			},
		},
	}
}

func resourceTsuruApplicationDeployDo(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	if !d.HasChange("image") {
		return nil
	}

	app := d.Get("app").(string)

	values := url.Values{}
	values.Set("origin", "image")
	values.Set("image", d.Get("image").(string))
	values.Set("message", "deploy via terraform")
	values.Set("new-version", strconv.FormatBool(d.Get("new_version").(bool)))
	values.Set("override-versions", strconv.FormatBool(d.Get("override_old_versions").(bool)))

	var buf bytes.Buffer
	buf.WriteString(values.Encode())

	url := fmt.Sprintf("%s/1.0/apps/%s/deploy", provider.Host, app)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	token := provider.Token
	if token == "" {
		token = deployToken()
	}
	req.Header.Set("Authorization", token)

	wait := d.Get("wait").(bool)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println("[DEBUG] failed to request deploy", err)
		return diag.FromErr(err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("Could not deploy, status code: %d, message: %s", resp.StatusCode, string(body))
	}

	eventID := resp.Header.Get("X-Tsuru-Eventid")
	d.SetId(eventID)

	if wait {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			log.Println("[DEBUG]", scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal("[ERROR]", err)
		}

		err = waitForEventComplete(ctx, provider, eventID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceTsuruApplicationDeployRead(ctx, d, meta)
}

func waitForEventComplete(ctx context.Context, provider *tsuruProvider, eventID string) error {
	deadline := time.Now().UTC().Add(time.Minute * 2)

	for {
		e, _, err := provider.TsuruClient.EventApi.EventInfo(ctx, eventID)

		if err != nil {
			return err
		}

		if e.Running {
			log.Println("[DEBUG] event is still running, pooling in the next 20 seconds")
			time.Sleep(time.Second * 20)
			continue
		}

		if e.Error != "" {
			return errors.New(e.Error + ", see details of event ID: " + eventID)
		}

		if time.Now().UTC().After(deadline) {
			return errors.New("event is still running after deploy")
		}

		return nil
	}

}

func resourceTsuruApplicationDeployRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	id := d.Id()

	e, _, err := provider.TsuruClient.EventApi.EventInfo(ctx, id)

	if err != nil {
		return diag.FromErr(err)
	}

	status := ""
	if e.Running {
		status = "running"
	} else if e.Error != "" {
		status = "error"
	} else if !e.EndTime.IsZero() {
		status = "finished"
	}

	d.Set("status", status)

	data, err := decodeRawBSONMap(e.EndCustomData)
	if err == nil {
		image, found := data["image"]
		if found {
			d.Set("output_image", image.(string))
		} else {
			d.Set("output_image", "")
		}
	} else {
		log.Println("[ERROR] found error decoding endCustomData", err)
	}

	return nil
}

func resourceTsuruApplicationDeployDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("[DEBUG] delete a deploy is a no-op by terraform")
	return nil
}

func decodeRawBSONMap(input tsuru.EventStartCustomData) (map[string]interface{}, error) {
	b, err := base64.StdEncoding.DecodeString(input.Data)
	if err != nil {
		return nil, err
	}
	r := bson.Raw{
		Kind: byte(input.Kind),
		Data: b,
	}
	data := map[string]interface{}{}
	err = r.Unmarshal(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func deployToken() string {
	if token, tokenErr := tsuruClientConfig.DefaultTokenProvider.Token(); tokenErr == nil && token != "" {
		return "bearer " + token
	}

	return ""
}

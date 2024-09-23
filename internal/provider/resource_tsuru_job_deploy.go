// Copyright 2023 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruJobDeploy() *schema.Resource {
	return &schema.Resource{
		Description:   "Perform an job deploy. Currently, only supporting deploys via prebuilt container images; in order to deploy via tsuru platforms please use tsuru-client",
		CreateContext: resourceTsuruJobDeployDo,
		UpdateContext: resourceTsuruJobDeployDo,
		ReadContext:   resourceTsuruJobDeployRead,
		DeleteContext: resourceTsuruJobDeployDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"job": {
				Type:        schema.TypeString,
				Description: "Job name",
				Required:    true,
				ForceNew:    true,
			},
			"image": {
				Type:        schema.TypeString,
				Description: "Docker Image",
				Required:    true,
			},
			"wait": {
				Type:        schema.TypeBool,
				Description: "Wait for the rollout of deploy",
				Optional:    true,
				Default:     true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "After apply may be three kinds of statuses: running or failed or finished",
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

func resourceTsuruJobDeployDo(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	if !d.HasChange("image") {
		return nil
	}

	job := d.Get("job").(string)

	values := url.Values{}
	values.Set("origin", "image")
	values.Set("image", d.Get("image").(string))
	values.Set("message", "deploy via terraform")

	var buf bytes.Buffer
	buf.WriteString(values.Encode())

	url := fmt.Sprintf("%s/1.23/jobs/%s/deploy", provider.Host, job)
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

	return resourceTsuruJobDeployRead(ctx, d, meta)
}

func resourceTsuruJobDeployRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceTsuruJobDeployDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Println("[DEBUG] delete a deploy is a no-op by terraform")
	return nil
}

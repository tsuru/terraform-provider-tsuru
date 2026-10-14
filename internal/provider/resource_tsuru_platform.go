// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTsuruPlatform() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a tsuru platform. Note: the tsuru API does not return the Dockerfile content back, so it cannot be refreshed from the API; terraform will keep whatever value is set in the configuration.",
		CreateContext: resourceTsuruPlatformCreate,
		ReadContext:   resourceTsuruPlatformRead,
		UpdateContext: resourceTsuruPlatformUpdate,
		DeleteContext: resourceTsuruPlatformDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Platform name",
			},
			"dockerfile": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Content of the Dockerfile used to build the platform image. Required to create the platform and whenever a new image build is needed.",
				ExactlyOneOf:  []string{"dockerfile", "image"},
				ConflictsWith: []string{"image"},
			},
			"image": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Base image to build the platform from. This is a convenience alternative to `dockerfile`: it generates a Dockerfile with a single `FROM ${image}` instruction.",
				ConflictsWith: []string{"dockerfile"},
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the platform is disabled, preventing new apps from being created with it",
			},
			"update_trigger": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Arbitrary value that, when changed, forces the platform image to be rebuilt from the current dockerfile content without requiring any other field to change",
			},
			"images": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of images built for this platform",
			},
		},
	}
}

func resourceTsuruPlatformCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := d.Get("name").(string)
	dockerfile, err := platformDockerfileContent(d)
	if err != nil {
		return diag.FromErr(err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", name); err != nil {
		return diag.FromErr(err)
	}
	if err := writePlatformDockerfile(writer, dockerfile); err != nil {
		return diag.FromErr(err)
	}
	if err := writer.Close(); err != nil {
		return diag.FromErr(err)
	}

	requestURL := strings.TrimRight(provider.Host, "/") + "/1.0/platforms"
	if dg := executePlatformStreamRequest(ctx, provider, http.MethodPost, requestURL, &body, writer.FormDataContentType()); dg != nil {
		return dg
	}

	d.SetId(name)

	// tsuru's platform Create ignores the disabled form field (it always stores the
	// platform as enabled), so disabling it requires a separate request after creation.
	if d.Get("disabled").(bool) {
		if dg := setPlatformDisabled(ctx, provider, name, true); dg != nil {
			return dg
		}
	}

	return resourceTsuruPlatformRead(ctx, d, meta)
}

func resourceTsuruPlatformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	name := d.Id()

	info, _, err := provider.TsuruClient.PlatformApi.PlatformInfo(ctx, name)
	if err != nil {
		if apiErr, ok := err.(interface{ StatusCode() int }); ok && apiErr.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Could not read tsuru platform %q, err: %s", name, err.Error())
	}

	d.Set("name", info.Platform.Name)
	d.Set("disabled", info.Platform.Disabled)
	d.Set("images", info.Images)

	return nil
}

func resourceTsuruPlatformUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	name := d.Id()

	if d.HasChange("dockerfile") || d.HasChange("image") || d.HasChange("update_trigger") {
		dockerfile, err := platformDockerfileContent(d)
		if err != nil {
			return diag.FromErr(err)
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if err := writer.WriteField("disabled", strconv.FormatBool(d.Get("disabled").(bool))); err != nil {
			return diag.FromErr(err)
		}
		if err := writePlatformDockerfile(writer, dockerfile); err != nil {
			return diag.FromErr(err)
		}
		if err := writer.Close(); err != nil {
			return diag.FromErr(err)
		}

		requestURL := strings.TrimRight(provider.Host, "/") + "/1.0/platforms/" + url.PathEscape(name)
		if dg := executePlatformStreamRequest(ctx, provider, http.MethodPut, requestURL, &body, writer.FormDataContentType()); dg != nil {
			return dg
		}
	} else if d.HasChange("disabled") {
		if dg := setPlatformDisabled(ctx, provider, name, d.Get("disabled").(bool)); dg != nil {
			return dg
		}
	}

	return resourceTsuruPlatformRead(ctx, d, meta)
}

func resourceTsuruPlatformDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	name := d.Id()

	_, err := provider.TsuruClient.PlatformApi.PlatformDelete(ctx, name)
	if err != nil && !isNotFoundError(err) {
		return diag.Errorf("Could not delete tsuru platform %q, err: %s", name, err.Error())
	}

	return nil
}

// platformDockerfileContent returns the Dockerfile content to send to the tsuru API, either the
// raw content of "dockerfile" or a generated "FROM ${image}" when "image" is used instead.
func platformDockerfileContent(d *schema.ResourceData) (string, error) {
	if image := d.Get("image").(string); image != "" {
		return fmt.Sprintf("FROM %s\n", image), nil
	}

	dockerfile := d.Get("dockerfile").(string)
	if strings.TrimSpace(dockerfile) == "" {
		return "", errors.New("either dockerfile or image must be provided")
	}

	return dockerfile, nil
}

func writePlatformDockerfile(writer *multipart.Writer, content string) error {
	part, err := writer.CreateFormFile("dockerfile_content", "Dockerfile")
	if err != nil {
		return err
	}
	_, err = part.Write([]byte(content))
	return err
}

func setPlatformDisabled(ctx context.Context, provider *tsuruProvider, name string, disabled bool) diag.Diagnostics {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("disabled", strconv.FormatBool(disabled)); err != nil {
		return diag.FromErr(err)
	}
	if err := writer.Close(); err != nil {
		return diag.FromErr(err)
	}

	requestURL := strings.TrimRight(provider.Host, "/") + "/1.0/platforms/" + url.PathEscape(name)
	return executePlatformStreamRequest(ctx, provider, http.MethodPut, requestURL, &body, writer.FormDataContentType())
}

// executePlatformStreamRequest performs a request against platform endpoints that respond with an
// application/x-json-stream body, where each line is a JSON object that may carry an Error field.
func executePlatformStreamRequest(ctx context.Context, provider *tsuruProvider, method, requestURL string, body io.Reader, contentType string) diag.Diagnostics {
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return diag.FromErr(err)
	}
	for key, values := range provider.HTTPHeaders {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Accept", "application/x-json-stream")

	response, err := provider.HTTPClient.Do(request)
	if err != nil {
		return diag.FromErr(err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(response.Body)
		return diag.Errorf("tsuru API returned status code %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	decoder := json.NewDecoder(response.Body)
	for {
		var msg struct {
			Message string
			Error   string
		}
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return diag.Errorf("failed to decode tsuru API response stream, err: %s", err.Error())
		}
		if msg.Message != "" {
			log.Print("[DEBUG] ", strings.TrimRight(msg.Message, "\n"))
		}
		if msg.Error != "" {
			return diag.Errorf("tsuru API returned an error: %s", msg.Error)
		}
	}

	return nil
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

const ID_SEPARATOR = "::"

type MaxRetriesError struct {
	Message string
	Meta    interface{}
}

func (e *MaxRetriesError) Error() string {
	return e.Message
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	openAPIError, ok := err.(tsuru_client.GenericOpenAPIError)
	return ok && openAPIError.StatusCode() == http.StatusNotFound
}

func isRetryableError(err []byte) bool {
	e := string(err)
	return strings.Contains(e, "event locked")
}

func tsuruRetry(ctx context.Context, d *schema.ResourceData, f func() error) error {
	return resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := f()
		if err != nil {
			var apiError tsuru_client.GenericOpenAPIError
			if errors.As(err, &apiError) {
				if isRetryableError(apiError.Body()) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
}

func createID(input []string) string {
	return strings.TrimSpace(strings.Join(input, ID_SEPARATOR))
}

func IDtoParts(input string, expectedLength int) ([]string, error) {
	output := strings.Split(input, ID_SEPARATOR)
	if len(output) < expectedLength {
		return nil, errors.Errorf("Mismatched length %d on input ID %s expected %d", len(output), input, expectedLength)
	}
	return output, nil
}

// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import (
	"strings"

	"github.com/pkg/errors"
)

const ID_SEPARATOR = "::"

func isRetryableError(err []byte) bool {
	e := string(err)
	return strings.Contains(e, "event locked")
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

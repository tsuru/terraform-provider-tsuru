// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import "strings"

const ID_SEPARATOR = "::"

func isRetryableError(err []byte) bool {
	e := string(err)
	if strings.Contains(e, "event locked") {
		return true
	}
	return false
}

func createID(input []string) string {
	return strings.TrimSpace(strings.Join(input, ID_SEPARATOR))
}

func IDtoParts(input string) []string {
	return strings.Split(input, ID_SEPARATOR)
}

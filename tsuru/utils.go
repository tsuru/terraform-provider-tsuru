// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsuru

import "strings"

func isLocked(err string) bool {
	if strings.Contains(err, "event locked") {
		return true
	}
	return false
}

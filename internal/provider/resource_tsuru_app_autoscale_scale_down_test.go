// Copyright 2024 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestFluentDown(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		scaleDownRead  tsuru_client.AutoScaleSpecBehaviorScaleDown
		scaleDownInput interface{}
		expected       interface{}
	}{
		{
			scaleDownRead: tsuru_client.AutoScaleSpecBehaviorScaleDown{
				UnitsPolicyValue:      3,
				PercentagePolicyValue: 10,
				StabilizationWindow:   300,
			},
			scaleDownInput: []interface{}{},
			expected:       nil,
		},

		{
			scaleDownRead: tsuru_client.AutoScaleSpecBehaviorScaleDown{
				UnitsPolicyValue:      3,
				PercentagePolicyValue: 10,
				StabilizationWindow:   300,
			},
			scaleDownInput: []interface{}{
				map[string]interface{}{"units": 3},
			},
			expected: []map[string]interface{}{{
				"units": int32(3),
			}},
		},
		{
			scaleDownRead: tsuru_client.AutoScaleSpecBehaviorScaleDown{
				UnitsPolicyValue:      3,
				PercentagePolicyValue: 10,
				StabilizationWindow:   300,
			},
			scaleDownInput: []interface{}{
				map[string]interface{}{"units": 3},
				map[string]interface{}{"stabilization_window": 300},
				map[string]interface{}{"percentage": 10},
			},
			expected: []map[string]interface{}{{
				"units":                int32(3),
				"stabilization_window": int32(300),
				"percentage":           int32(10),
			}},
		},
		{
			scaleDownRead: tsuru_client.AutoScaleSpecBehaviorScaleDown{
				UnitsPolicyValue:      21,
				PercentagePolicyValue: 21,
				StabilizationWindow:   21,
			},
			scaleDownInput: []interface{}{
				map[string]interface{}{"units": 3},
				map[string]interface{}{"stabilization_window": 300},
				map[string]interface{}{"percentage": 10},
			},
			expected: []map[string]interface{}{{
				"units":                int32(21),
				"stabilization_window": int32(21),
				"percentage":           int32(21),
			}},
		},
		{
			scaleDownRead: tsuru_client.AutoScaleSpecBehaviorScaleDown{
				UnitsPolicyValue:      21,
				PercentagePolicyValue: 21,
				StabilizationWindow:   21,
			},
			scaleDownInput: []interface{}{},
			expected: []map[string]interface{}{{
				"units":                int32(21),
				"stabilization_window": int32(21),
				"percentage":           int32(21),
			}},
		},
	}
	for _, test := range tests {
		readToDiff := newFlattenScaleDown(test.scaleDownRead, test.scaleDownInput).execute()
		assert.Equal(test.expected, readToDiff)
	}
}

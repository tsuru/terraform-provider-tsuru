// Copyright 2024 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"fmt"
	"reflect"

	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

type flattenScaleDown struct {
	PERCENTAGE_VALUE           int32
	PERCENTAGE_LABEL           string
	STABILIZATION_WINDOW_VALUE int32
	STABILIZATION_WINDOW_LABEL string
	UNITS_VALUE                int32
	UNITS_LABEL                string
	ScaleDownRead              tsuru_client.AutoScaleSpecBehaviorScaleDown
	Proposed                   interface{}
}

func newFlattenScaleDown(scaleDownRead tsuru_client.AutoScaleSpecBehaviorScaleDown, proposed interface{}) *flattenScaleDown {
	return &flattenScaleDown{
		PERCENTAGE_VALUE:           10,
		PERCENTAGE_LABEL:           "percentage",
		STABILIZATION_WINDOW_VALUE: 300,
		STABILIZATION_WINDOW_LABEL: "stabilization_window",
		UNITS_VALUE:                3,
		UNITS_LABEL:                "units",
		ScaleDownRead:              scaleDownRead,
		Proposed:                   proposed,
	}
}

func (fsd *flattenScaleDown) execute() interface{} {
	if fsd.ScaleDownRead == (tsuru_client.AutoScaleSpecBehaviorScaleDown{}) {
		return nil
	}
	proposedList, err := fsd.convertToMapSlice(fsd.Proposed)
	if err != nil {
		return []map[string]interface{}{{
			"percentage":           fsd.ScaleDownRead.PercentagePolicyValue,
			"stabilization_window": fsd.ScaleDownRead.StabilizationWindow,
			"units":                fsd.ScaleDownRead.UnitsPolicyValue,
		}}
	}
	if value, ok := fsd.noInputParameters(proposedList); ok {
		return value
	}
	return fsd.withInputParameters(proposedList)
}

func (fsd *flattenScaleDown) withInputParameters(proposedList []map[string]interface{}) (value []map[string]interface{}) {
	scaleDownCurrent := []map[string]interface{}{{}}
	percentage, ok := fsd.findScaleDownInProposedList(proposedList, fsd.PERCENTAGE_LABEL)
	if ok && percentage != 0 || fsd.ScaleDownRead.PercentagePolicyValue != int32(fsd.PERCENTAGE_VALUE) {
		scaleDownCurrent[0][fsd.PERCENTAGE_LABEL] = fsd.ScaleDownRead.PercentagePolicyValue
	}
	stabilizationWindow, ok := fsd.findScaleDownInProposedList(proposedList, fsd.STABILIZATION_WINDOW_LABEL)
	if ok && stabilizationWindow != 0 || fsd.ScaleDownRead.StabilizationWindow != int32(fsd.STABILIZATION_WINDOW_VALUE) {
		scaleDownCurrent[0][fsd.STABILIZATION_WINDOW_LABEL] = fsd.ScaleDownRead.StabilizationWindow
	}
	units, ok := fsd.findScaleDownInProposedList(proposedList, fsd.UNITS_LABEL)
	if ok && units != 0 || fsd.ScaleDownRead.UnitsPolicyValue != int32(fsd.UNITS_VALUE) {
		scaleDownCurrent[0][fsd.UNITS_LABEL] = fsd.ScaleDownRead.UnitsPolicyValue
	}
	return scaleDownCurrent
}

func (fsd *flattenScaleDown) noInputParameters(proposedList []map[string]interface{}) (value interface{}, ok bool) {
	if len(proposedList) != 0 {
		return nil, false
	}
	scaleDownCurrent := []map[string]interface{}{{}}
	if fsd.ScaleDownRead.PercentagePolicyValue != fsd.PERCENTAGE_VALUE {
		scaleDownCurrent[0][fsd.PERCENTAGE_LABEL] = fsd.ScaleDownRead.PercentagePolicyValue
	}
	if fsd.ScaleDownRead.StabilizationWindow != fsd.STABILIZATION_WINDOW_VALUE {
		scaleDownCurrent[0][fsd.STABILIZATION_WINDOW_LABEL] = fsd.ScaleDownRead.StabilizationWindow
	}
	if fsd.ScaleDownRead.UnitsPolicyValue != fsd.UNITS_VALUE {
		scaleDownCurrent[0][fsd.UNITS_LABEL] = fsd.ScaleDownRead.UnitsPolicyValue
	}
	if fsd.isScaleDownEmpty(scaleDownCurrent) {
		return nil, true
	}
	return scaleDownCurrent, true
}

func (fsd *flattenScaleDown) findScaleDownInProposedList(proposedList []map[string]interface{}, key string) (value int, ok bool) {
	for _, item := range proposedList {
		if v, ok := item[key]; ok {
			return v.(int), true
		}
	}
	return 0, false
}

func (fsd *flattenScaleDown) convertToMapSlice(input interface{}) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	if reflect.TypeOf(input).Kind() != reflect.Slice {
		return nil, fmt.Errorf("scale down: invalid input type, slice expected")
	}
	for _, item := range input.([]interface{}) {
		if mapItem, ok := item.(map[string]interface{}); ok {
			result = append(result, mapItem)
		} else {
			return []map[string]interface{}{}, nil
		}
	}
	return result, nil
}

func (fsd *flattenScaleDown) isScaleDownEmpty(param []map[string]interface{}) bool {
	if len(param) != 1 {
		return false
	}
	if _, ok := param[0][fsd.PERCENTAGE_LABEL]; ok {
		return false
	}
	if _, ok := param[0][fsd.STABILIZATION_WINDOW_LABEL]; ok {
		return false
	}
	if _, ok := param[0][fsd.UNITS_LABEL]; ok {
		return false
	}
	return true
}

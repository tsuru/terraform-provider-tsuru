package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tsuru_client "github.com/tsuru/go-tsuruclient/pkg/tsuru"
)

func TestMarkRemovedMetadataItemWhenOldIsEmpty(t *testing.T) {
	oldMetadataItemsList := []tsuru_client.MetadataItem{}
	newMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}}

	resultMetadataItemList := markRemovedMetadataItemAsDeleted(oldMetadataItemsList, newMetadataItemsList)

	expectedList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedMetadataItemWhenNewIsEmpty(t *testing.T) {
	oldMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}}
	newMetadataItemsList := []tsuru_client.MetadataItem{}

	resultMetadataItemList := markRemovedMetadataItemAsDeleted(oldMetadataItemsList, newMetadataItemsList)

	expectedList := []tsuru_client.MetadataItem{{
		Name:   "label_name",
		Value:  "label_value",
		Delete: true,
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedMetadataItemWhenOldIsReplaced(t *testing.T) {
	oldMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}}
	newMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "new_label_value",
	}}

	resultMetadataItemList := markRemovedMetadataItemAsDeleted(oldMetadataItemsList, newMetadataItemsList)

	expectedList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "new_label_value",
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedMetadataItemWhenOldIsRemvedNewIsAdded(t *testing.T) {
	oldMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}}
	newMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "new_label_name",
		Value: "new_label_value",
	}}

	resultMetadataItemList := markRemovedMetadataItemAsDeleted(oldMetadataItemsList, newMetadataItemsList)

	expectedList := []tsuru_client.MetadataItem{{
		Name:  "new_label_name",
		Value: "new_label_value",
	}, {
		Name:   "label_name",
		Value:  "label_value",
		Delete: true,
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedMetadataItemWhenOldShouldNotChange(t *testing.T) {
	oldMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}}
	newMetadataItemsList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}, {
		Name:  "new_label_name",
		Value: "new_label_value",
	}}

	resultMetadataItemList := markRemovedMetadataItemAsDeleted(oldMetadataItemsList, newMetadataItemsList)

	expectedList := []tsuru_client.MetadataItem{{
		Name:  "label_name",
		Value: "label_value",
	}, {
		Name:  "new_label_name",
		Value: "new_label_value",
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

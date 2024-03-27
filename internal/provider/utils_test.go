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

func TestMarkRemovedMetadataItemWhenOldIsRemovedNewIsAdded(t *testing.T) {
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

func TestMarkRemovedProcessAsDefaultPlanWhenCreatingNewProcess(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{}
	newProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}

	resultMetadataItemList := markRemovedProcessAsDefaultPlan(oldProcessList, newProcessList)

	expectedList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedProcessAsDefaultPlanWhenUpdatingProcessPlan(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}
	newProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c2m2",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}

	resultMetadataItemList := markRemovedProcessAsDefaultPlan(oldProcessList, newProcessList)

	expectedList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c2m2",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedProcessAsDefaultPlanWhenDeletingProcess(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}
	newProcessList := []tsuru_client.AppProcess{}

	resultMetadataItemList := markRemovedProcessAsDefaultPlan(oldProcessList, newProcessList)

	expectedList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "$default",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:   "label_name",
				Value:  "label_value",
				Delete: true,
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:   "annotation_name",
				Value:  "annotation_value",
				Delete: true,
			}},
		},
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedProcessAsDefaultPlanWhenUpdatingProcessMetadata(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name",
				Value: "label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "annotation_value",
			}},
		},
	}}
	newProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "new_label_name",
				Value: "new_label_value",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "new_annotation_value",
			}},
		},
	}}

	resultMetadataItemList := markRemovedProcessAsDefaultPlan(oldProcessList, newProcessList)

	expectedList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "new_label_name",
				Value: "new_label_value",
			}, {
				Name:   "label_name",
				Value:  "label_value",
				Delete: true,
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name",
				Value: "new_annotation_value",
			}},
		},
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestMarkRemovedProcessAsDefaultPlanWhenDeletingProcessMetadata(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_1",
				Value: "label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "annotation_value_1",
			}},
		},
	}, {
		Name: "process2",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_2",
				Value: "label_value_2",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_2",
				Value: "annotation_value_2",
			}},
		},
	}, {
		Name: "process3",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_3",
				Value: "label_value_3",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_3",
				Value: "annotation_value_3",
			}},
		},
	}}
	newProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "annotation_value_1",
			}},
		},
	}, {
		Name: "process2",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_2",
				Value: "label_value_2",
			}},
		},
	}, {
		Name: "process3",
		Plan: "c1m1",
	}}

	resultMetadataItemList := markRemovedProcessAsDefaultPlan(oldProcessList, newProcessList)

	expectedList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:   "label_name_1",
				Value:  "label_value_1",
				Delete: true,
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "annotation_value_1",
			}},
		},
	}, {
		Name: "process2",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_2",
				Value: "label_value_2",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:   "annotation_name_2",
				Value:  "annotation_value_2",
				Delete: true,
			}},
		},
	}, {
		Name: "process3",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:   "label_name_3",
				Value:  "label_value_3",
				Delete: true,
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:   "annotation_name_3",
				Value:  "annotation_value_3",
				Delete: true,
			}},
		},
	}}
	assert.Equal(t, expectedList, resultMetadataItemList)
}

func TestCheckProcessesListsChanges(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{{
		Name: "process0",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_0",
				Value: "label_value_0",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_0",
				Value: "annotation_value_0",
			}},
		},
	}, {
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_1",
				Value: "label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "annotation_value_1",
			}},
		},
	}}
	newProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "new_label_name_1",
				Value: "new_label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "new_annotation_value_1",
			}},
		},
	}, {
		Name: "process2",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_2",
				Value: "label_value_2",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_2",
				Value: "annotation_value_2",
			}},
		},
	}}

	old, new, both := checkProcessesListsChanges(oldProcessList, newProcessList)

	expectedOld := []tsuru_client.AppProcess{{
		Name: "process0",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_0",
				Value: "label_value_0",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_0",
				Value: "annotation_value_0",
			}},
		},
	}}
	expectedNew := []tsuru_client.AppProcess{{
		Name: "process2",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_2",
				Value: "label_value_2",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_2",
				Value: "annotation_value_2",
			}},
		},
	}}
	expectedBoth := []ChangedProcess{{
		Old: tsuru_client.AppProcess{
			Name: "process1",
			Plan: "c1m1",
			Metadata: tsuru_client.Metadata{
				Labels: []tsuru_client.MetadataItem{{
					Name:  "label_name_1",
					Value: "label_value_1",
				}},
				Annotations: []tsuru_client.MetadataItem{{
					Name:  "annotation_name_1",
					Value: "annotation_value_1",
				}},
			},
		},
		New: tsuru_client.AppProcess{
			Name: "process1",
			Plan: "c1m1",
			Metadata: tsuru_client.Metadata{
				Labels: []tsuru_client.MetadataItem{{
					Name:  "new_label_name_1",
					Value: "new_label_value_1",
				}},
				Annotations: []tsuru_client.MetadataItem{{
					Name:  "annotation_name_1",
					Value: "new_annotation_value_1",
				}},
			},
		},
	}}
	assert.Equal(t, expectedOld, old)
	assert.Equal(t, expectedNew, new)
	assert.Equal(t, expectedBoth, both)
}

func TestCheckProcessesListsChangesWhenOldEmpty(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{}
	newProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "new_label_name_1",
				Value: "new_label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "new_annotation_value_1",
			}},
		},
	}}

	old, new, both := checkProcessesListsChanges(oldProcessList, newProcessList)

	expectedOld := []tsuru_client.AppProcess{}
	expectedNew := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "new_label_name_1",
				Value: "new_label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "new_annotation_value_1",
			}},
		},
	}}
	expectedBoth := []ChangedProcess{}
	assert.Equal(t, expectedOld, old)
	assert.Equal(t, expectedNew, new)
	assert.Equal(t, expectedBoth, both)
}

func TestCheckProcessesListsChangesWhenNewEmpty(t *testing.T) {
	oldProcessList := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_1",
				Value: "label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "annotation_value_1",
			}},
		},
	}}
	newProcessList := []tsuru_client.AppProcess{}

	old, new, both := checkProcessesListsChanges(oldProcessList, newProcessList)

	expectedOld := []tsuru_client.AppProcess{{
		Name: "process1",
		Plan: "c1m1",
		Metadata: tsuru_client.Metadata{
			Labels: []tsuru_client.MetadataItem{{
				Name:  "label_name_1",
				Value: "label_value_1",
			}},
			Annotations: []tsuru_client.MetadataItem{{
				Name:  "annotation_name_1",
				Value: "annotation_value_1",
			}},
		},
	}}
	expectedNew := []tsuru_client.AppProcess{}
	expectedBoth := []ChangedProcess{}

	assert.Equal(t, expectedOld, old)
	assert.Equal(t, expectedNew, new)
	assert.Equal(t, expectedBoth, both)
}

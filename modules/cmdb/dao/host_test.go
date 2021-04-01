package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeHostLabel(t *testing.T) {
	var result string
	var hostLabels string

	hostLabels = ""
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "", result)

	hostLabels = "MESOS_ATTRIBUTES="
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any;"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any,test"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any,test", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any,test;"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any,test", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any,test;test_key:"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any,test", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any,test;test_key:test_value"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any,test", result)

	hostLabels = "MESOS_ATTRIBUTES=dice_tags:any,test;test_key:test_value;"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any,test", result)

	hostLabels = "K8S_ATTRIBUTES=dice_tags:any,test,service-stateless,platform,location-es"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "any,test,stateless-service,platform,location-es", result)

	hostLabels = "K8S_ATTRIBUTES=dice_tags:org-xxx,service-stateful,pack"
	result = MakeHostLabel(hostLabels)
	assert.Equal(t, "org-xxx,stateful-service,pack-job", result)

}

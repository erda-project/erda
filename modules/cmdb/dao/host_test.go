// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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

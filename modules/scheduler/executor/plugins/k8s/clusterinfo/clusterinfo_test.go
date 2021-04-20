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

package clusterinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNetportalURL(t *testing.T) {
	var (
		testURL1 = "inet://127.0.0.1/master.mesos"
		testURL2 = "inet://127.0.0.2?ssl=on&direct=on/master.mesos/service/marathon?test=on"
	)

	netportal, err := parseNetportalURL(testURL1)
	assert.Nil(t, err)
	assert.Equal(t, "inet://127.0.0.1", netportal)

	netportal, err = parseNetportalURL(testURL2)
	assert.Nil(t, err)
	assert.Equal(t, "inet://127.0.0.2?ssl=on&direct=on", netportal)
}

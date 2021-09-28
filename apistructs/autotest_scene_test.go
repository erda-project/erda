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

package apistructs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/envconf"
)

func TestAutoTestRunWait(t *testing.T) {
	jsonStr := `{"waitTimeSec": 2}`
	envMap := map[string]string{
		"ACTION_WAIT_TIME_SEC": "2",
	}

	var (
		jsonWait AutoTestRunWait
		envWait  AutoTestRunWait
	)
	err := json.Unmarshal([]byte(jsonStr), &jsonWait)
	assert.NoError(t, err)
	err = envconf.Load(&envWait, envMap)
	assert.NoError(t, err)
	assert.Equal(t, 2, jsonWait.WaitTimeSec)
	assert.Equal(t, 2, envWait.WaitTimeSec)
}

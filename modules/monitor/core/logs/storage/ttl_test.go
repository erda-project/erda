// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"testing"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/stretchr/testify/assert"
)

func mockMysqlStore() *mysqlStore {
	ms := &mysqlStore{
		defaultTTLSec: 60,
		ttlValue:      map[string]int{},
		Log:           logrusx.New(),
	}
	return ms
}

func TestPopulateTTLValue(t *testing.T) {
	ms := mockMysqlStore()
	list := []*MonitorConfig{
		{OrgName: "erda", Names: "*", Filters: "", Config: []byte(`{"ttl":"1h0m0s"}`)},
	}

	ttlmap := ms.populateTTLValue(list)
	val, ok := ttlmap["erda"]

	assert.True(t, ok)
	assert.Equal(t, 3600, val)
}

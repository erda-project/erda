// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
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
	ass := assert.New(t)
	// normal
	ttlmap := ms.populateTTLValue([]*MonitorConfig{
		{OrgName: "erda", Names: "*", Filters: "", Config: []byte(`{"ttl":"1h0m0s"}`)},
	})
	val, ok := ttlmap["erda"]

	ass.True(ok)
	ass.Equal(3600, val)

	// zero ttl
	ttlmap = ms.populateTTLValue([]*MonitorConfig{
		{OrgName: "erda", Names: "*", Filters: "", Config: []byte(`{"ttl":""}`)},
	})
	_, ok = ttlmap["erda"]
	ass.False(ok)

	// invalid ttl
	ttlmap = ms.populateTTLValue([]*MonitorConfig{
		{OrgName: "erda", Names: "*", Filters: "", Config: []byte(`{"ttl":"abc"}`)},
	})
	_, ok = ttlmap["erda"]
	ass.False(ok)
}

func Test_mysqlStore_GetSecondByKey(t *testing.T) {
	ms := mockMysqlStore()
	list := []*MonitorConfig{
		{OrgName: "erda", Names: "*", Filters: "", Config: []byte(`{"ttl":"1h0m0s"}`)},
	}
	ttlmap := ms.populateTTLValue(list)
	ms.ttlValue = ttlmap

	assert.Equal(t, 3600, ms.GetSecondByKey("erda"))
}

func Test_mysqlStore_loadLogsTTL(t *testing.T) {
	ms := mockMysqlStore()
	m := newMockMysql()
	ms.mysql = m.DB()

	err := ms.loadLogsTTL()
	assert.Nil(t, err)
}

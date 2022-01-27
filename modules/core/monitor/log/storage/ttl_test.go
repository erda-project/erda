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
		Log:           logrusx.New(),
	}
	return ms
}

func Test_mysqlStore_loadLogsTTL(t *testing.T) {
	ms := mockMysqlStore()
	m := newMockMysql()
	ms.mysql = m.DB()

	err := ms.loadLogsTTL()
	assert.Nil(t, err)

	second := ms.GetSecond("container", map[string]string{"dice_org_name": "erda", "dice_workspace": "prod"})
	assert.Equal(t, 360*3600, second)

	assert.Equal(t, 60, ms.GetSecond("container", map[string]string{"dice_org_name": "erda"}))
}

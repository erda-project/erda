// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"os"
	"testing"

	"github.com/erda-project/erda/bundle"
)

const (
	testOrgId = "1"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	{
		os.Setenv("OPS_ADDR", "ops.default.svc.cluster.local:9027")
		os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
	}
	bdl = bundle.New(bundle.WithCMDB())

}

// func TestGetResourcesInDescribeMetricMetaList(t *testing.T) {
// 	resp, err := GetDescribeMetricLast(testOrgId, "acs_rocketmq", "ConsumerLag")
// 	assert.Nil(t, err)
// 	t.Log(len(resp))
// 	for _, item := range resp {
// 		t.Logf("resp: %+v\n", item)
// 	}
// }
//
// func TestGetDescribeMetricLast(t *testing.T) {
// 	resp, err := ListMetricMeta(testOrgId, "acs_rocketmq")
// 	assert.Nil(t, err)
// 	data, _ := json.Marshal(resp)
// 	t.Logf("resp: %+v\n", string(data))
// }
//
// func TestGetDescribeProjectMeta(t *testing.T) {
// 	resp, err := ListProjectMeta(testOrgId, nil)
// 	assert.Nil(t, err)
// 	t.Logf("resp: %+v\n", resp)
// }
//
// func TestGetDescribeSomeProjectMeta(t *testing.T) {
// 	resp, err := ListProjectMeta(testOrgId, []string{"RDS", "ECS"})
// 	assert.Nil(t, err)
// 	assert.Equal(t, 2, len(resp))
// 	t.Logf("resp: %+v\n", resp)
// }
//
// func TestListOrgClusterPair(t *testing.T) {
// 	resp, err := ListOrgInfos()
// 	assert.Nil(t, err)
// 	t.Logf("resp: %+v", resp)
// }

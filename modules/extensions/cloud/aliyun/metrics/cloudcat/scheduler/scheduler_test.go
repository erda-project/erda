// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package scheduler

import (
	"fmt"
	"os"
	"testing"

	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/extensions/cloud/aliyun/metrics/cloudcat/api"
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
}

var testInfo = api.OrgInfo{
	OrgId:   "1",
	OrgName: "orgname",
}

// func TestNew(t *testing.T) {
// 	fw := &fakeWriter{}
// 	sc, err := New(testInfo, &globals.Config{}, fw)
// 	assert.Nil(t, err)
// 	assert.NotZero(t, sc)
// 	t.Logf("schedulers: %+v", sc)
// }
//
// func TestStart(t *testing.T) {
// 	fw := &fakeWriter{}
// 	cfg := &globals.Config{
// 		GatherWindow: time.Second * 10,
// 		Products:     []string{"ECS", "RDS"},
// 		Output: kafka.ProducerConfig{
// 			Parallelism: 3,
// 		},
// 	}
// 	sc, err := New(testInfo, cfg, fw)
// 	assert.Nil(t, err)
// 	assert.NotZero(t, sc)
// 	if err != nil {
// 		return
// 	}
// 	sc.Start()
// }
//
// func TestMetaChange(t *testing.T) {
// 	fw := &fakeWriter{}
// 	cfg := &globals.Config{
// 		GatherWindow: time.Second * 30,
// 		Products:     []string{"ECS"},
// 		Output: kafka.ProducerConfig{
// 			Parallelism: 3,
// 		},
// 	}
// 	sc, err := New(testInfo, cfg, fw)
// 	assert.Nil(t, err)
// 	assert.NotZero(t, sc)
// 	if err != nil {
// 		return
// 	}
// 	go func() { sc.Start() }()
// 	time.Sleep(time.Second)
// 	sc.grabberChangedSig <- 0
// 	select {}
// }

type fakeWriter struct {
}

func (fw *fakeWriter) Write(data interface{}) error {
	item := data.(*kafka.Message)
	fmt.Printf("topic: %s data: %s\n", *item.Topic, string(item.Data))
	return nil
}

func (fw *fakeWriter) WriteN(data ...interface{}) (int, error) {
	return 0, nil
}

func (fw *fakeWriter) Close() error {
	return nil
}

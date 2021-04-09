// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cloudcat

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/extensions/cloud/aliyun/metrics/cloudcat/globals"
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

func createManager() *manager {
	m := &manager{}
	m.writer = &fakeWriter{}
	m.Cfg = &globals.Config{GatherWindow: time.Minute * 1, Products: []string{"VPC", "WAF", "CMS"}}
	_ = m.initScheduler()
	return m
}

// func TestMetaChange(t *testing.T) {
// 	m := createManager()
// 	m.start()
// 	time.Sleep(time.Second * 2)
// 	m.schedulerChangedSig <- 0
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

func TestXXX(t *testing.T) {
	fmt.Println(strings.Split("", ","))
}

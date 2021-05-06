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

package actionagent

import (
	"fmt"
	"testing"
)

//func TestWatchFiles(t *testing.T) {
//	logrus.SetLevel(logrus.DebugLevel)
//	logrus.Debug("begin watch files")
//
//	fw, err := filewatch.New()
//	assert.NoError(t, err)
//	ctx, cancel := context.WithCancel(context.Background())
//	agent := Agent{
//		EasyUse: EasyUse{
//			EnablePushLog2Collector: true,
//			CollectorAddr:           "collector.default.svc.cluster.local:7076",
//			TaskLogID:               "agent-push-1",
//
//			RunMultiStdoutFilePath: "/tmp/stdout",
//			RunMultiStderrFilePath: "/tmp/stderr",
//		},
//		Ctx:         ctx,
//		Cancel:      cancel,
//		FileWatcher: fw,
//	}
//
//	go agent.watchFiles()
//
//	// mock logic done
//	time.Sleep(time.Millisecond * 10) // very fast action done
//
//	// teardown
//	agent.PreStop()
//
//	time.Sleep(time.Second * 1)
//}

func TestSlice(t *testing.T) {
	f := func(flogs *[]string, line string) {
		*flogs = append(*flogs, line)
	}
	var logs = &[]string{}
	f(logs, "l1")
	fmt.Println(logs)
	f(logs, "l2")
	fmt.Println(logs)
}

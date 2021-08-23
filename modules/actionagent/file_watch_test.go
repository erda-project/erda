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

package actionagent

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/alecthomas/assert"
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

func TestMatchStdErrLine(t *testing.T) {
	regexpList := []*regexp.Regexp{}
	regexpStrList := []string{"^[a-z]*can't*"}
	for i := range regexpStrList {
		reg, err := regexp.Compile(regexpStrList[i])
		assert.NoError(t, err)
		regexpList = append(regexpList, reg)
	}
	assert.False(t, matchStdErrLine(regexpList, "failed to open file test.text"))
	assert.True(t, matchStdErrLine(regexpList, "can't open '/jfoejfoijlkjlj/jflejwf.txt': No such file or directory"))
}

func TestDesensitizeLine(t *testing.T) {
	var txtBlackList = []string{"abcdefg", "kskdudh&"}
	tt := []struct {
		line, want string
	}{
		{"admin", "admin"},
		{"abcdefg", "******"},
		{"admin123", "admin123"},
		{"docker login -u abcdefg -p kskdudh&", "docker login -u ****** -p ******"},
	}
	for _, v := range tt {
		desensitizeLine(txtBlackList, &v.line)
		assert.Equal(t, v.want, v.line)
	}

}

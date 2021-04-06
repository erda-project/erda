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
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/actionagent/filewatch"
	"github.com/erda-project/erda/pkg/filehelper"
)

func (agent *Agent) watchFiles() {

	watcher, err := filewatch.New()
	if err != nil {
		agent.AppendError(err)
		return
	}
	agent.FileWatcher = watcher

	// ${METAFILE}
	watcher.RegisterFullHandler(agent.EasyUse.ContainerMetaFile, metaFileFullHandler(agent))
	// stdout && stderr
	watcher.RegisterTailHandler(agent.EasyUse.RunMultiStdoutFilePath, stdoutTailHandler(agent))
	watcher.RegisterTailHandler(agent.EasyUse.RunMultiStderrFilePath, stderrTailHandler(agent))

	if agent.EasyUse.EnablePushLog2Collector {
		go agent.asyncPushCollectorLog()
	}
}

func (agent *Agent) stopWatchFiles() {
	if agent.FileWatcher != nil {
		agent.FileWatcher.Close()
	}
}

// metaFileFullHandler 全量处理 metafile
func metaFileFullHandler(agent *Agent) filewatch.FullHandler {
	return func(r io.ReadCloser) error {
		agent.LockPushedMetaFileMap.Lock()
		defer agent.LockPushedMetaFileMap.Unlock()
		// 一次性读取
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		logrus.Debugf("监听到了 METAFILE 改动，时间: %v，全量内容: %s", time.Now().Format(time.RFC3339), string(b))
		cb := &Callback{}
		err = cb.HandleMetaFile(b)
		if err != nil {
			return err
		}
		err = agent.callbackToPipelinePlatform(cb)
		if err != nil {
			return err
		}
		updatePushedMetadata(cb, agent)
		return nil
	}
}

// stdoutTailHandler 以 tail 方式处理 stdout
func stdoutTailHandler(agent *Agent) filewatch.TailHandler {
	return func(line string, allLines []string) error {
		// add your logic here
		// logrus.Printf("stdout tailed a line: %s\n", line)

		// meta
		tailHandlerForMeta(line, agent.EasyUse.ContainerMetaFile)

		// push to collector
		if agent.EasyUse.EnablePushLog2Collector {
			tailHandlerForPushCollectorLog(line, apistructs.CollectorLogPushStreamStdout, stdoutLogs, agent.EasyUse.TaskLogID, &stdoutLock)
		}

		return nil
	}
}

// stderrTailHandler 以 tail 方式处理 stderr
func stderrTailHandler(agent *Agent) filewatch.TailHandler {
	return func(line string, allLines []string) error {
		// add your logic here
		// logrus.Printf("stderr tailed a line: %s\n", line)

		// meta
		tailHandlerForMeta(line, agent.EasyUse.ContainerMetaFile)

		// push to collector
		if agent.EasyUse.EnablePushLog2Collector {
			tailHandlerForPushCollectorLog(line, apistructs.CollectorLogPushStreamStderr, stderrLogs, agent.EasyUse.TaskLogID, &stderrLock)
		}

		return nil
	}
}

const metaPrefix = "action meta:"

func tailHandlerForMeta(line, metafile string) {
	isMetaLog, k, v := getMetaFromLogLine(line)
	if isMetaLog {
		// write to metaFile, use universal callback logic
		if err := filehelper.Append(metafile, fmt.Sprintf("\n%s=%s\n", k, v)); err != nil {
			logrus.Warnf("failed to append meta to metafile, k: %s, v: %s, err: %v", k, v, err)
		}
	}
}

// getMetaFromLogLine 从日志行中分析 meta，格式如下：
// action meta: key=value
//考虑到 value 可能较复杂，每行只支持声明一个 meta
func getMetaFromLogLine(line string) (bool, string, string) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, metaPrefix) {
		return false, "", ""
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, metaPrefix))
	kv := strings.SplitN(line, "=", 2)
	var k string
	var v string
	if len(kv) > 0 {
		k = strings.TrimSpace(kv[0])
	}
	if len(kv) > 1 {
		v = strings.TrimSpace(kv[1])
	}
	return true, k, v
}

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

package logic

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/httpclient"
)

func init() {
	go asyncPushCollectorLog()
}

const (
	CtxKeyCollectorLogID = "logID"
	CtxKeyLogger         = "logger"
)

// clog means collector log.
func clog(ctx context.Context) *logrus.Entry {
	return ctx.Value(CtxKeyLogger).(*logrus.Entry)
}

type actionLogFormatter struct {
	logrus.TextFormatter
}

func (f *actionLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	_bytes, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	return append([]byte("[Action Log] "), _bytes...), nil
}

// newLogger return logger.
func newLogger() *logrus.Logger {
	// set logrus
	l := logrus.New()
	l.SetFormatter(&actionLogFormatter{
		logrus.TextFormatter{
			ForceColors:            true,
			DisableTimestamp:       true,
			DisableLevelTruncation: true,
		},
	})
	l.SetOutput(ioutil.Discard)
	l.AddHook(&CollectorHook{})
	return l
}

type CollectorHook struct{}

func (c *CollectorHook) Levels() []logrus.Level { return logrus.AllLevels }
func (c *CollectorHook) Fire(entry *logrus.Entry) error {
	logLock.Lock()
	defer logLock.Unlock()
	*logLines = append(*logLines, apistructs.LogPushLine{
		ID:        entry.Context.Value(CtxKeyCollectorLogID).(string),
		Source:    string(apistructs.DashboardSpotLogSourceJob),
		Stream:    &apistructs.CollectorLogPushStreamStdout,
		Timestamp: time.Now().UnixNano(),
		Content:   entry.Message,
	})
	return nil
}

const (
	asyncPushInterval = time.Second * 3
)

var logLines = &[]apistructs.LogPushLine{}
var logLock = &sync.Mutex{}

// asyncPushCollectorLog async push log to collector
func asyncPushCollectorLog() {
	handleFunc := func() {
		logLock.Lock()
		defer logLock.Unlock()

		if logLines == nil || len(*logLines) == 0 {
			return
		}
		err := pushCollectorLog(logLines)
		if err != nil {
			logrus.Error(err)
			// not refresh logs, try together at next time
			return
		}
		// refresh logs if success
		*logLines = nil
	}

	ticker := time.NewTicker(asyncPushInterval)
	for range ticker.C {
		handleFunc()
	}
}

func pushCollectorLog(logLines *[]apistructs.LogPushLine) error {
	var respBody bytes.Buffer
	resp, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Post(conf.CollectorAddr()).
		Path("/collect/logs/job").
		JSONBody(logLines).
		Header("Content-Type", "application/json").
		Do().
		Body(&respBody)
	if err != nil {
		return fmt.Errorf("failed to push log to collector, err: %v", err)
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to push log to collector, resp body: %s", respBody.String())
	}
	return nil
}

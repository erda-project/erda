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

package spotlog

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common/unmarshalwork"
)

var (
	idKeys = []string{"TERMINUS_DEFINE_TAG", "terminus_define_tag", "MESOS_TASK_ID", "mesos_task_id"}
)

var (
	logLevelInfoValue = "INFO"
	logStreamStdout   = "stdout"
)

var (
	logLevelTag       = "level"
	logReqIDTag       = "request_id"
	logReqIDTagV1     = "request-id"
	logTraceTag       = "trace_id"
	logDiceOrgTag     = "dice_org_name"
	logOrgTag         = "org_name"
	logMonitorKeyTag  = "monitor_log_key"
	logTerminusKeyTag = "terminus_log_key"
)

func ParseSpotLog(buf []byte, callback func(m *log.Log) error) error {
	uw := newUnmarshalWork(buf, callback)
	uw.wg.Add(1)
	unmarshalwork.Schedule(uw)
	uw.wg.Wait()
	if uw.err != nil {
		return fmt.Errorf("parse spotMetric err: %w", uw.err)
	}
	return nil
}

type unmarshalWork struct {
	buf      []byte
	err      error
	wg       sync.WaitGroup
	callback func(m *log.Log) error
}

func newUnmarshalWork(buf []byte, callback func(m *log.Log) error) *unmarshalWork {
	return &unmarshalWork{buf: buf, callback: callback}
}

func (uw *unmarshalWork) Unmarshal() {
	defer uw.wg.Done()
	data := &log.LabeledLog{}
	if err := common.JsonDecoder.Unmarshal(uw.buf, data); err != nil {
		uw.err = err
		return
	}
	normalize(&data.Log)
	if err := Validate(data); err != nil {
		uw.err = err
		return
	}

	if err := uw.callback(&data.Log); err != nil {
		uw.err = err
		return
	}
}

var (
	// ErrIDEmpty .
	ErrIDEmpty = errors.New("id empty")
)

func Validate(l *log.LabeledLog) error {
	if len(l.ID) <= 0 {
		return ErrIDEmpty
	}
	return nil
}

func normalize(data *log.Log) {
	if data.Tags == nil {
		data.Tags = make(map[string]string)
	}

	if data.Time != nil {
		data.Timestamp = data.Time.UnixNano()
		data.Time = nil
	}

	// setup level
	if level, ok := data.Tags[logLevelTag]; ok {
		data.Tags[logLevelTag] = strings.ToUpper(level)
	} else {
		data.Tags[logLevelTag] = logLevelInfoValue
	}

	// setup request id
	if reqID, ok := data.Tags[logReqIDTag]; ok {
		data.Tags[logTraceTag] = reqID
	} else if reqID, ok = data.Tags[logTraceTag]; ok {
		data.Tags[logReqIDTag] = reqID
	} else if reqID, ok = data.Tags[logReqIDTagV1]; ok {
		data.Tags[logReqIDTag] = reqID
		data.Tags[logTraceTag] = reqID
		delete(data.Tags, logReqIDTagV1)
	}

	// setup org name
	if _, ok := data.Tags[logDiceOrgTag]; !ok {
		if org, ok := data.Tags[logOrgTag]; ok {
			data.Tags[logDiceOrgTag] = org
		}
	}

	// setup log key for compatibility
	key, ok := data.Tags[logMonitorKeyTag]
	if !ok {
		key, ok = data.Tags[logTerminusKeyTag]
		if ok {
			data.Tags[logMonitorKeyTag] = key
		}
	}

	// setup log id
	for _, key := range idKeys {
		if val, ok := data.Tags[key]; ok {
			data.ID = val
			break
		}
	}

	// setup default stream
	if data.Stream == "" {
		data.Stream = logStreamStdout
	}
}

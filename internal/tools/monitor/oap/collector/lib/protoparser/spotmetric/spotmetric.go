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

package spotmetric

import (
	"fmt"
	"strings"
	"sync"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common/unmarshalwork"
)

func ParseSpotMetric(buf []byte, callback func(m *metric.Metric) error) error {
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
	callback func(m *metric.Metric) error
}

func newUnmarshalWork(buf []byte, callback func(m *metric.Metric) error) *unmarshalWork {
	return &unmarshalWork{buf: buf, callback: callback}
}

func (uw *unmarshalWork) Unmarshal() {
	defer uw.wg.Done()
	data := &metric.Metric{}
	if err := common.JsonDecoder.Unmarshal(uw.buf, data); err != nil {
		uw.err = err
		return
	}
	if v, ok := data.Tags[lib.OrgNameKey]; ok {
		data.OrgName = v
	}

	if strings.IndexAny(data.Name, "-") != -1 {
		data.Name = strings.ReplaceAll(data.Name, "-", "_")
	}

	if err := uw.callback(data); err != nil {
		uw.err = err
		return
	}
}

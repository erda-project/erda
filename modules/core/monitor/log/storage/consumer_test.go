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

package storage

import (
	"testing"
	"time"

	logmodule "github.com/erda-project/erda/modules/core/monitor/log"
	"github.com/stretchr/testify/assert"
)

func Test_provider_processLog(t *testing.T) {
	mp := mockProvider()
	ass := assert.New(t)

	// default
	log := &logmodule.Log{
		ID:        "aaa",
		Source:    "container",
		Stream:    "",
		Offset:    0,
		Timestamp: 0,
		Content:   "",
		Tags:      nil,
	}
	mp.processLog(log)
	ass.Equal("INFO", log.Tags["level"])
	ass.Equal("stdout", log.Stream)

	// upper level
	log = &logmodule.Log{
		ID:        "aaa",
		Source:    "container",
		Stream:    "",
		Offset:    0,
		Timestamp: 0,
		Content:   "",
		Tags:      map[string]string{"level": "error"},
	}
	mp.processLog(log)
	ass.Equal("ERROR", log.Tags["level"])
	ass.Equal("stdout", log.Stream)

	// id keys
	mp.Cfg.Output.IDKeys = []string{"id_key1"}
	log = &logmodule.Log{
		ID:        "aaa",
		Source:    "container",
		Stream:    "",
		Offset:    0,
		Timestamp: 0,
		Content:   "",
		Tags:      map[string]string{"id_key1": "bbb"},
	}
	mp.processLog(log)
	ass.Equal("bbb", log.ID)
}

func Test_provider_invokeV2(t *testing.T) {
	mp := mockProvider()
	mw := &mockWriter{}
	mp.output = mw
	ass := assert.New(t)

	log := &logmodule.Log{
		ID:        "aaa",
		Source:    "container",
		Stream:    "stdout",
		Offset:    1024,
		Timestamp: time.Now().UnixNano(),
		Content:   "hello world",
		Tags:      map[string]string{"level": "INFO"},
	}

	value, err := json.Marshal(log)
	ass.Nil(err)
	err = mp.invoke(nil, value, nil, time.Now())
	ass.Nil(err)
	ass.Equal(&logmodule.LogMeta{
		ID:     log.ID,
		Source: log.Source,
		Tags:   log.Tags,
	}, mw.datas[0])
	ass.Equal(log, mw.datas[1].(*logmodule.Log))

	// bad value
	err = mp.invoke(nil, []byte(`bad value`), nil, time.Now())
	ass.Error(err)
}

type mockWriter struct {
	datas []interface{}
}

func (m *mockWriter) Write(data interface{}) error {
	m.datas = append(m.datas, data)
	return nil
}

func (m *mockWriter) WriteN(data ...interface{}) (int, error) {
	m.datas = append(m.datas, data...)
	return len(data), nil
}

func (m *mockWriter) Close() error {
	return nil
}

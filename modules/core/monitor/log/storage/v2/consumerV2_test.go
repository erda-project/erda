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

package storagev2

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-proto-go/core/monitor/log/storage/pb"
	logmodule "github.com/erda-project/erda/modules/core/monitor/log"
	"github.com/golang/protobuf/proto"
)

func Test_provider_processLogV2(t *testing.T) {
	mp := mockProvider()
	ass := assert.New(t)

	// default
	log := &pb.Log{
		Id:        "aaa",
		Source:    "container",
		Stream:    "",
		Offset:    0,
		Timestamp: 0,
		Content:   "",
		Tags:      nil,
		Labels:    nil,
	}
	mp.processLogV2(log)
	ass.Equal("INFO", log.Tags["level"])
	ass.Equal("stdout", log.Stream)

	// upper level
	log = &pb.Log{
		Id:        "aaa",
		Source:    "container",
		Stream:    "",
		Offset:    0,
		Timestamp: 0,
		Content:   "",
		Tags:      map[string]string{"level": "error"},
		Labels:    nil,
	}
	mp.processLogV2(log)
	ass.Equal("ERROR", log.Tags["level"])
	ass.Equal("stdout", log.Stream)

	// id keys
	mp.Cfg.Output.IDKeys = []string{"id_key1"}
	log = &pb.Log{
		Id:        "aaa",
		Source:    "container",
		Stream:    "",
		Offset:    0,
		Timestamp: 0,
		Content:   "",
		Tags:      map[string]string{"id_key1": "bbb"},
		Labels:    nil,
	}
	mp.processLogV2(log)
	ass.Equal("bbb", log.Id)
}

func Test_provider_invokeV2(t *testing.T) {
	mp := mockProvider()
	mw := &mockWriter{}
	mp.output = mw
	ass := assert.New(t)

	log := &pb.Log{
		Id:        "aaa",
		Source:    "container",
		Stream:    "stdout",
		Offset:    1024,
		Timestamp: time.Now().UnixNano(),
		Content:   "hello world",
		Tags:      map[string]string{"level": "INFO"},
		Labels:    nil,
	}

	lb := &pb.LogBatch{
		Logs: []*pb.Log{log},
	}
	value, err := proto.Marshal(lb)
	ass.Nil(err)
	err = mp.invokeV2(nil, value, nil, time.Now())
	ass.Nil(err)
	ass.Equal(&logmodule.LogMeta{
		ID:     log.Id,
		Source: log.Source,
		Tags:   log.Tags,
	}, mw.datas[0])
	proto.Equal(log, mw.datas[1].(*pb.Log))

	// bad value
	err = mp.invokeV2(nil, []byte(`bad value`), nil, time.Now())
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

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

package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/modules/monitor/core/collector/v2/outputs/console"
	"github.com/erda-project/erda/modules/monitor/core/logs/pb"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestCollectLogsV2(t *testing.T) {
	coll := &provider{
		output: &console.Output{
			Writer:      io.Discard,
			DecoderFunc: console.DefaultDecoderFunc,
		},
		Logger: logrusx.New(),
		Cfg: &config{
			Limiter: limitConfig{
				RequestBodySize: "1K",
			},
		},
	}

	e := echo.New()
	h := coll.collectLogsV2

	// protobuf body
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/collect/:source", bytes.NewReader(mockProtobufBody(1)))
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("source")
	ctx.SetParamValues("log")
	req.Header.Add("Content-Type", "application/x-protobuf")
	req.Header.Add("Content-Encoding", "gzip")
	assert.Nil(t, h(ctx))

	// invalid content-type
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v2/collect/:source", bytes.NewReader(mockProtobufBody(1)))
	ctx = e.NewContext(req, rec)
	ctx.SetParamNames("source")
	ctx.SetParamValues("log")
	req.Header.Add("Content-Type", "application/x-protobu")
	req.Header.Add("Content-Encoding", "gzip")
	assert.Error(t, h(ctx))
}

func BenchmarkCollectLogsV2Protobuf(b *testing.B) {
	coll := &provider{
		output: &console.Output{
			Writer: io.Discard,
		},
		Logger: logrusx.New(),
	}

	e := echo.New()
	h := coll.collectLogsV2
	buf := mockProtobufBody(200)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/collect/:source", nil)
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("source")
	ctx.SetParamValues("log")
	req.Header.Add("Content-Type", "application/x-protobuf")
	req.Header.Add("Content-Encoding", "gzip")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewReader(buf))
		_ = h(ctx)
	}
}

func mockLogBatch(count int) *pb.LogBatch {
	logs := make([]*pb.Log, count)
	for i := range logs {
		logs[i] = &pb.Log{
			Id:      "77e90e85233cb3ec1bfd7655633248056a3fc03e092596c405a5b158cc8885b2",
			Source:  "container",
			Stream:  "stdout",
			Offset:  17420730,
			Content: "\u001B[37mDEBU\u001B[0m[2021-04-22 14:18:52.265950181] finished handle request GET /health (took 107.411µs) ",
			Tags: map[string]string{
				"dice_cluster_name": "terminus-dev",
				"pod_name":          "dice-qa-7cb5b7fd4-494zb",
				"pod_namespace":     "default",
				"container_name":    "qa",
				"dice_component":    "qa",
			},
			Timestamp: 1415792726371000000,
			Labels:    map[string]string{},
		}
	}
	lb := &pb.LogBatch{
		Logs: logs,
	}
	return lb
}

func mockProtobufBody(count int) []byte {
	lb := mockLogBatch(count)
	buf, err := lb.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	buf, err = compress(buf)
	if err != nil {
		log.Fatal(err)
	}
	return buf
}

func b64Encode(buf []byte) []byte {
	res := make([]byte, 0, len(buf))
	base64.StdEncoding.Encode(res, buf)
	return res
}

func compress(buf []byte) ([]byte, error) {
	gz, _ := NewGziper(9)
	return gz.Compress(buf)
}

type gziper struct {
	buf    *bytes.Buffer
	writer *gzip.Writer
}

func NewGziper(level int) (*gziper, error) {
	buf := bytes.NewBuffer(nil)
	writer, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return nil, err
	}

	return &gziper{
		buf:    buf,
		writer: writer,
	}, nil
}

func (g *gziper) Compress(data []byte) ([]byte, error) {
	defer func() {
		g.buf.Reset()
		g.writer.Reset(g.buf)
	}()

	if _, err := g.writer.Write(data); err != nil {
		return nil, err
	}
	if err := g.writer.Flush(); err != nil {
		return nil, err
	}
	if err := g.writer.Close(); err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(nil)
	if _, err := io.Copy(b, g.buf); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

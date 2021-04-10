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
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
	"unsafe"

	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda/modules/monitor/core/logs"
	"github.com/erda-project/erda/modules/monitor/core/logs/schema"
)

func (p *provider) createLogStatementBuilder() cassandra.StatementBuilder {
	var buf bytes.Buffer
	return &LogStatement{
		p:          p,
		gzipWriter: gzip.NewWriter(&buf),
	}
}

type LogStatement struct {
	gzipWriter *gzip.Writer
	p          *provider
}

func (ls *LogStatement) GetStatement(data interface{}) (string, []interface{}, error) {
	switch data.(type) {
	case *logs.Log:
		return ls.p.getLogStatement(data.(*logs.Log), ls.gzipWriter)
	case *logs.LogMeta:
		return ls.p.getMetaStatement(data.(*logs.LogMeta))
	default:
		return "", nil, fmt.Errorf("value %#v must be Log or LogMeta", data)
	}
}

func (p *provider) getLogStatement(log *logs.Log, reusedWriter *gzip.Writer) (string, []interface{}, error) {
	ttl := p.ttl.GetSecondByKey(log.Tags[diceOrgNameKey])

	var requestID *string // request_id 字段不存在时为null，所以使用指针
	if rid, ok := log.Tags["request-id"]; ok {
		requestID = &rid
	}

	table := schema.DefaultBaseLogTable
	if org, ok := log.Tags[diceOrgNameKey]; ok {
		table = schema.BaseLogWithOrgName(org)
	}

	content, err := gzipContentV2(log.Content, reusedWriter)
	if err != nil {
		return "", nil, err
	}
	// nolint
	cql := fmt.Sprintf(`INSERT INTO %s (source, id, stream, time_bucket, timestamp, offset, content, level, request_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) USING TTL ?;`, table)
	return cql, []interface{}{
		log.Source,
		log.ID,
		log.Stream,
		trncateDate(log.Timestamp),
		log.Timestamp,
		log.Offset,
		content,
		log.Tags["level"],
		requestID,
		ttl,
	}, nil
}

func (p *provider) getMetaStatement(meta *logs.LogMeta) (string, []interface{}, error) {
	ttl := p.ttl.GetSecondByKey(meta.Tags[diceOrgNameKey])
	cql := `INSERT INTO spot_prod.base_log_meta (source, id, tags) VALUES (?, ?, ?) USING TTL ?;`
	return cql, []interface{}{
		meta.Source,
		meta.ID,
		meta.Tags,
		ttl,
	}, nil
}

func trncateDate(unixNano int64) int64 {
	const day = time.Hour * 24
	return unixNano - unixNano%int64(day)
}

func gzipContent(content string) ([]byte, error) {
	reader, err := compressWithPipe(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func compressWithPipe(reader io.Reader) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)

	var err error
	go func() {
		_, err = io.Copy(gzipWriter, reader)
		gzipWriter.Close()
		// subsequent reads from the read half of the pipe will
		// return no bytes and the error err, or EOF if err is nil.
		pipeWriter.CloseWithError(err)
	}()

	return pipeReader, err
}

func gzipContentV2(content string, reusedWriter *gzip.Writer) ([]byte, error) {
	reader, err := compressWithPipeV2(bytes.NewReader(*(*[]byte)(unsafe.Pointer(&content))), reusedWriter)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func compressWithPipeV2(reader io.Reader, reusedWriter *gzip.Writer) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	reusedWriter.Reset(pipeWriter)

	go func() {
		_, err := io.Copy(reusedWriter, reader)
		if err != nil {
			fmt.Printf("gzip copy failed: %s\n", err)
		}
		reusedWriter.Close()
		pipeWriter.CloseWithError(err)
	}()

	return pipeReader, nil
}

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

package query

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"

	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log/schema"
)

var ErrEmptyLogMeta = errors.New("empty log meta record")

type CQLQueryInf interface {
	Query(builder *qb.SelectBuilder, binding qb.M, dest interface{}) error
	// Update(builder *qb.UpdateBuilder)
}

type cassandraQuery struct {
	session *gocql.Session
}

func (c *cassandraQuery) Query(builder *qb.SelectBuilder, binding qb.M, dest interface{}) error {
	stmt, names := builder.ToCql()
	cql := gocqlx.Query(c.session.Query(stmt), names).BindMap(binding)
	return cql.SelectRelease(dest)
}

func (p *provider) getLogItems(r *RequestCtx) ([]*pb.LogItem, error) {
	// 请求ID关联的日志
	if len(r.RequestID) > 0 {
		logs, err := p.queryRequestLog(
			p.getTableNameWithFilters(map[string]interface{}{
				"tags['dice_application_id']": r.ApplicationID,
			}),
			r.RequestID,
		)
		if err != nil {
			return nil, err
		}
		return logs, nil
	}

	// 基础日志
	if r.Count == 0 {
		return nil, nil
	}
	return p.queryBaseLog(
		p.getTableNameWithFilters(map[string]interface{}{
			"source": r.Source,
			"id":     r.ID,
		}),
		r.Source,
		r.ID,
		r.Stream,
		r.Start,
		r.End,
		r.Count,
	)
}

func (p *provider) queryBaseLogMetaWithFilters(filters map[string]interface{}) (*LogMeta, error) {
	var res []*LogMeta
	cqlBuilder := qb.Select(LogMetaTableName).Limit(1)
	for key := range filters {
		cqlBuilder = cqlBuilder.Where(qb.Eq(key))
	}
	if err := p.cqlQuery.Query(cqlBuilder, filters, &res); err != nil {
		return nil, fmt.Errorf("retrive %s fialed: %w", LogMetaTableName, err)
	}
	if len(res) == 0 {
		return nil, ErrEmptyLogMeta
	}
	return res[0], nil
}

func (p *provider) queryRequestLog(table, requestID string) ([]*pb.LogItem, error) {
	var list []*SavedLog
	if err := p.cqlQuery.Query(
		qb.Select(table).Where(qb.Eq("request_id")),
		qb.M{"request_id": requestID},
		&list,
	); err != nil {
		return nil, fmt.Errorf("retrive %s failed: %w", table, err)
	}

	logs, err := convertToLogList(list)
	if err != nil {
		return nil, err
	}

	// todo. for back forward compatibility, prepare remove in version 3.21
	if table == schema.DefaultBaseLogTable {
		return logs, nil
	}
	oldLogs, err := p.queryRequestLog(schema.DefaultBaseLogTable, requestID)
	if err != nil {
		return nil, err
	}

	logs = append(logs, oldLogs...)
	sort.Sort(Logs(logs))
	return logs, nil
}

func (p *provider) queryBaseLog(table, source, id, stream string, start, end, count int64) ([]*pb.LogItem, error) {
	orderBy, limit := qb.ASC, uint(count)
	if count < 0 {
		limit = uint(-1 * count)
		orderBy = qb.DESC
	}

	var logs []*pb.LogItem
	for {
		if start >= end {
			break
		}
		var bucket int64
		if orderBy == qb.ASC {
			bucket = trncateDate(start)
		} else {
			bucket = trncateDate(end)
		}
		slogs, err := p.queryBaseLogInBucket(table, source, id, stream, bucket, start+1, end, orderBy, limit)
		if err != nil {
			return nil, err
		}
		list, err := convertToLogList(slogs)
		if err != nil {
			return nil, err
		}
		logs = append(logs, list...)
		if len(logs) >= int(limit) {
			logs = logs[0:limit]
			break
		}
		if orderBy == qb.ASC {
			start = bucket + int64(time.Hour)*24
		} else {
			end = bucket - 1
		}
	}
	sort.Sort(Logs(logs))
	return logs, nil
}

func (p *provider) walkSavedLogs(table, source, id, stream string, start, end int64, fn func([]*SavedLog) error) error {
	timespan, tail := int64(p.Cfg.Download.TimeSpan), end
	for {
		if start >= tail {
			break
		}
		if start+timespan > tail {
			end = tail
		} else {
			end = start + timespan
		}
		bucket := trncateDate(start)
		list, err := p.queryBaseLogInBucket(table, source, id, stream, bucket, start, end, qb.ASC, 0)
		if err != nil {
			return err
		}
		if len(list) > 0 {
			err = fn(list)
			if err != nil {
				return err
			}
		}
		start = end
	}
	return nil
}

func (p *provider) queryBaseLogInBucket(
	table, source, id, stream string,
	bucket, start, end int64,
	order qb.Order, limit uint,
) ([]*SavedLog, error) {
	var logs []*SavedLog
	if err := p.cqlQuery.Query(
		qb.Select(table).
			Where(
				qb.Eq("source"),
				qb.Eq("id"),
				qb.Eq("stream"),
				qb.Eq("time_bucket"),
				qb.GtOrEqNamed("timestamp", "start"),
				qb.LtNamed("timestamp", "end")).
			OrderBy("timestamp", order).OrderBy("offset", order).
			Limit(limit),
		qb.M{
			"source":      source,
			"id":          id,
			"stream":      stream,
			"time_bucket": bucket,
			"start":       start,
			"end":         end,
		},
		&logs,
	); err != nil {
		return nil, fmt.Errorf("retrive %s failed: %w", table, err)
	}

	// todo. for back forward compatibility, prepare remove in version 3.21
	if table == schema.DefaultBaseLogTable {
		return logs, nil
	}
	oldLogs, err := p.queryBaseLogInBucket(schema.DefaultBaseLogTable, source, id, stream, bucket, start, end, order, limit)
	if err != nil {
		return nil, err
	}
	logs = append(logs, oldLogs...)
	return logs, nil
}

func convertToLogList(list []*SavedLog) ([]*pb.LogItem, error) {
	var logs []*pb.LogItem
	for _, log := range list {
		data, err := wrapLogData(log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, data)
	}
	return logs, nil
}

func trncateDate(unixNano int64) int64 {
	const day = time.Hour * 24
	return unixNano - unixNano%int64(day)
}

func wrapLogData(sl *SavedLog) (*pb.LogItem, error) {
	content, err := gunzipContent(sl.Content)
	if err != nil {
		return nil, err
	}
	return &pb.LogItem{
		Source:     sl.Source,
		Id:         sl.ID,
		Stream:     sl.Stream,
		TimeBucket: strconv.FormatInt(sl.TimeBucket, 10),
		Timestamp:  strconv.FormatInt(sl.Timestamp, 10),
		Offset:     strconv.FormatInt(sl.Offset, 10),
		Content:    string(content),
		Level:      sl.Level,
		RequestId:  sl.RequestID,
	}, nil
}

func gunzipContent(content []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	res, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type qresult struct {
	stmt   string
	values []interface{}
}

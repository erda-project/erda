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

package query

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	"github.com/erda-project/erda/modules/monitor/core/logs/schema"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

const (
	maxDay       = 7
	maxIdsLength = 20
)

func (p *provider) queryBaseLogMetaWithFilters(filters map[string]interface{}) (res []*LogMeta, err error) {
	cqlBuilder := qb.Select(LogMetaTableName).Limit(10)
	for key := range filters {
		cqlBuilder = cqlBuilder.Where(qb.Eq(key))
	}
	stmt, names := cqlBuilder.ToCql()
	cql := gocqlx.Query(p.session.Query(stmt), names).BindMap(filters)
	p.Logger.Debugf("cql=%+v", cql)
	if err := cql.SelectRelease(&res); err != nil {
		return nil, fmt.Errorf("query cassandra failed. err=%s", err)
	}
	return
}

func (p *provider) queryRequestLog(table, requestID string) ([]*Log, error) {
	stmt, names := qb.Select(table).
		Where(qb.Eq("request_id")).ToCql()
	var list []*SavedLog
	if err := gocqlx.Query(p.session.Query(stmt), names).
		BindMap(qb.M{"request_id": requestID}).
		SelectRelease(&list); err != nil {
		return nil, err
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

func (p *provider) queryBaseLog(table, source, id, stream string, start, end, count int64) ([]*Log, error) {
	if count == 0 {
		return nil, nil
	}
	orderBy, limit := qb.ASC, uint(count)
	if count < 0 {
		limit = uint(-1 * count)
		orderBy = qb.DESC
	}

	var logs []*Log
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

func (p *provider) queryBaseLogInBucket(table, source, id, stream string, bucket, start, end int64, order qb.Order, limit uint) ([]*SavedLog, error) {
	stmt, names := qb.Select(table).
		Where(
			qb.Eq("source"),
			qb.Eq("id"),
			qb.Eq("stream"),
			qb.Eq("time_bucket"),
			qb.GtOrEqNamed("timestamp", "start"),
			qb.LtNamed("timestamp", "end")).
		OrderBy("timestamp", order).OrderBy("offset", order).
		Limit(limit).ToCql()
	var logs []*SavedLog
	cql := gocqlx.Query(p.session.Query(stmt), names).BindMap(qb.M{
		"source":      source,
		"id":          id,
		"stream":      stream,
		"time_bucket": bucket,
		"start":       start,
		"end":         end,
	})
	p.Logger.Debugf("log query. cql=%+v", cql.String())
	if err := cql.SelectRelease(&logs); err != nil {
		return nil, err
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

func convertToLogList(list []*SavedLog) ([]*Log, error) {
	var logs []*Log
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

func wrapLogData(sl *SavedLog) (*Log, error) {
	content, err := gunzipContent(sl.Content)
	if err != nil {
		return nil, err
	}
	return &Log{
		Source:     sl.Source,
		ID:         sl.ID,
		Stream:     sl.Stream,
		TimeBucket: strconv.FormatInt(sl.TimeBucket, 10),
		Timestamp:  strconv.FormatInt(sl.Timestamp, 10),
		Offset:     strconv.FormatInt(sl.Offset, 10),
		Content:    string(content),
		Level:      sl.Level,
		RequestID:  sl.RequestID,
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

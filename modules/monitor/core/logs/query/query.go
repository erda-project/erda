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
	"math"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda/modules/monitor/core/logs/schema"
	"github.com/olivere/elastic"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

const (
	maxDay       = 7
	maxIdsLength = 20
)

func (p *provider) queryBaseLogMetaWithFilters(filters map[string]interface{}) (res []*LogMeta, err error) {
	cqlBuilder := qb.Select(LogMetaTableName).Limit(10)
	for key, _ := range filters {
		cqlBuilder = cqlBuilder.Where(qb.Eq(key))
	}
	stmt, names := cqlBuilder.ToCql()
	cql := gocqlx.Query(p.session.Query(stmt), names).BindMap(filters)
	p.L.Debugf("cql=%+v", cql)
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
	timespan, tail := int64(p.C.Download.TimeSpan), end
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
	p.L.Debugf("log query. cql=%+v", cql.String())
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

func wrapLogMeta(meta *SaveLogMeta) *LogMeta {
	return &LogMeta{
		Source: meta.Source,
		ID:     meta.ID,
		Tags:   meta.Tags,
	}
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

func (p *provider) removeBaseLog(ids []string, timeBuckets []int64, orgName string) error {
	var values []interface{}
	stmt := "DELETE FROM " + schema.BaseLogWithOrgName(orgName) + " WHERE source IN ('container', 'job') AND stream IN ('stdout', 'stderr') "
	stmt += "AND time_bucket IN ("
	for idx, item := range timeBuckets {
		if idx > 0 {
			stmt += ", "
		}
		stmt += "?"
		values = append(values, item)
	}
	stmt += ") "

	qress := p.doLoopQuery(stmt, ids)
	for _, item := range qress {
		tmp := make([]interface{}, len(values))
		copy(tmp, values)
		tmp = append(tmp, item.values...)
		if err := p.session.Query(item.stmt, tmp...).Exec(); err != nil {
			return fmt.Errorf("stmt=%s, values=%+v. err=%s", item.stmt, tmp, err)
		}
	}
	return nil
}

type qresult struct {
	stmt   string
	values []interface{}
}

func (p *provider) doLoopQuery(baseQry string, ids []string) []*qresult {
	res := []*qresult{}
	start, end := 0, maxIdsLength
loop:
	values := []interface{}{}
	qry := baseQry + "AND id IN ("
	if end >= len(ids) {
		end = len(ids) - 1
	}
	for idx, item := range ids[start:end] {
		if idx > 0 {
			qry += ", "
		}
		qry += "?"
		values = append(values, item)
	}
	qry += ") "
	res = append(res, &qresult{qry, values})

	start, end = start+maxIdsLength, end+maxIdsLength
	if start >= len(ids) {
		return res
	} else {
		goto loop
	}
}

func (p *provider) getContainerIds(filters url.Values, end int64, orgName string) (cids []string, err error) {
	largeSize, cnt := 5000, 1
	boolq := elastic.NewBoolQuery()
	if len(orgName) > 0 && orgName != "prod" {
		boolq = boolq.Must(elastic.NewTermQuery("tags.org_name", orgName))
	}
	for _, k := range []string{"container_id", "application_id", "runtime_id", "workspace", "instance"} {
		if filters.Get(k) != "" {
			boolq = boolq.Must(elastic.NewTermQuery("tags."+k, filters.Get(k)))
		}
	}
	boolq = boolq.Must(elastic.NewRangeQuery("timestamp").Lte(time.Duration(end) * time.Millisecond))

doquery:
	res, err := p.q.SearchRaw([]string{"spot-docker_container_status*"}, elastic.NewSearchSource().Size(0).Query(boolq).
		Aggregation("unique_cids", elastic.NewTermsAggregation().Size(largeSize).Field("tags.container_id")))
	data, ok := res.Aggregations.Range("unique_cids")
	if !ok {
		return nil, fmt.Errorf("no result")
	}
	if data.SumOfOtherDocCount > 0 {
		largeSize = largeSize * int(math.Pow(2, float64(cnt)))
		cnt++
		goto doquery
	}
	for _, item := range data.Buckets {
		cids = append(cids, item.Key)
	}
	return
}

func (p *provider) getTimeBuckets(offset int64) []int64 {
	timeBuckets := []int64{}
	t := time.Unix((offset-offset%1000)/1000, offset%1000*100000)
	for i := 1; i <= maxDay; i++ {
		t := trncateDate(t.AddDate(0, 0, -1*i).UnixNano())
		timeBuckets = append(timeBuckets, t)
	}
	return timeBuckets
}

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

package dbclient

import (
	"fmt"

	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

type recordsReader struct {
	db         *dbengine.DBEngine
	conditions []string
	limit      int
	offset     int
}

type recordsWriter struct {
	db *dbengine.DBEngine
}

func (c *DBClient) RecordsReader() *recordsReader {
	return &recordsReader{db: c.DBEngine, conditions: []string{}, limit: 0, offset: -1}
}

func (r *recordsReader) PageNum(n int) *recordsReader {
	r.offset = n
	return r
}

func (r *recordsReader) PageSize(n int) *recordsReader {
	r.limit = n
	return r
}

func (r *recordsReader) ByIDs(ids ...string) *recordsReader {
	r.conditions = append(r.conditions, fmt.Sprintf("id in (%s)", strutil.Join(ids, ",")))
	return r
}

func (r *recordsReader) ByPipelineIDs(ids ...string) *recordsReader {
	r.conditions = append(r.conditions, fmt.Sprintf("pipeline_id in (%s)", strutil.Join(ids, ",")))
	return r
}

func (r *recordsReader) ByRecordTypes(tps ...string) *recordsReader {
	tpsStr := strutil.Map(tps, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("record_type in (%s)", strutil.Join(tpsStr, ",")))
	return r
}

func (r *recordsReader) ByStatuses(statuses ...string) *recordsReader {
	statusesStr := strutil.Map(statuses, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("status in (%s)", strutil.Join(statusesStr, ",")))
	return r
}

func (r *recordsReader) ByOrgID(orgid string) *recordsReader {
	r.conditions = append(r.conditions, fmt.Sprintf("org_id = \"%s\"", orgid))
	return r
}

func (r *recordsReader) ByClusterNames(clusternames ...string) *recordsReader {
	clusternamesStr := strutil.Map(clusternames, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("cluster_name in (%s) ", strutil.Join(clusternamesStr, ",")))
	return r
}

func (r *recordsReader) ByUserIDs(userids ...string) *recordsReader {
	useridsStr := strutil.Map(userids, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("user_id in (%s)", strutil.Join(useridsStr, ",")))
	return r
}

func (r *recordsReader) ByCreateTime(beforeNSecs int) *recordsReader {
	r.conditions = append(r.conditions, fmt.Sprintf("created_at < now() - interval %d second", beforeNSecs))
	return r
}

func (r *recordsReader) ByUpdateTime(beforeNSecs int) *recordsReader {
	r.conditions = append(r.conditions, fmt.Sprintf("updated_at < now() - interval %d second", beforeNSecs))
	return r
}

func (r *recordsReader) Limit(n int) *recordsReader {
	r.limit = n
	return r
}
func (r *recordsReader) Count() (int64, error) {
	var count int64
	err := r.db.Model(&Record{}).Where(strutil.Join(r.conditions, " AND ", true)).Count(&count).Error
	return count, err
}

func (r *recordsReader) Do() ([]Record, error) {
	records := []Record{}
	expr := r.db.Where(strutil.Join(r.conditions, " AND ", true)).Order("created_at desc")
	if r.limit != 0 {
		expr = expr.Limit(r.limit)
	}
	if r.offset != -1 {
		expr = expr.Offset(r.offset)
	}
	if err := expr.Find(&records).Error; err != nil {
		r.conditions = []string{}
		return nil, err
	}
	r.conditions = []string{}
	return records, nil
}

func (c *DBClient) RecordsWriter() *recordsWriter {
	return &recordsWriter{db: c.DBEngine}
}

func (w *recordsWriter) Create(s *Record) (uint64, error) {
	db := w.db.Save(s)
	return s.ID, db.Error
}
func (w *recordsWriter) Update(s Record) error {
	return w.db.Model(&s).Updates(s).Error
}
func (w *recordsWriter) Delete(ids ...uint64) error {
	return w.db.Delete(Record{}, "id in (?)", ids).Error
}

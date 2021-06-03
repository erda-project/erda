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

	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

type addonManageReader struct {
	db         *dbengine.DBEngine
	conditions []string
	limit      int
	offset     int
}

type addonManageWriter struct {
	db *dbengine.DBEngine
}

// read condition
func (c *DBClient) AddonManageReader() *addonManageReader {
	return &addonManageReader{db: c.DBEngine, conditions: []string{}, limit: 0, offset: -1}
}

func (r *addonManageReader) ByAddonIDs(rids ...string) *addonManageReader {
	render := strutil.Map(rids, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("addon_id in (%s)", strutil.Join(render, ",")))
	return r
}

func (r *addonManageReader) ByProjectIDs(pids ...string) *addonManageReader {
	render := strutil.Map(pids, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("project_id in (%s)", strutil.Join(render, ",")))
	return r
}

func (r *addonManageReader) ByOrgIDs(oids ...string) *addonManageReader {
	render := strutil.Map(oids, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("org_id in (%s)", strutil.Join(render, ",")))
	return r
}

// read
func (r *addonManageReader) Do() ([]AddonManagement, error) {
	ams := []AddonManagement{}
	expr := r.db.Where(strutil.Join(r.conditions, " AND ", true)).Order("create_time desc")
	if r.limit != 0 {
		expr = expr.Limit(r.limit)
	}
	if r.offset != -1 {
		expr = expr.Offset(r.offset)
	}
	if err := expr.Find(&ams).Error; err != nil {
		r.conditions = []string{}
		return nil, err
	}
	r.conditions = []string{}
	return ams, nil
}

// write
func (c *DBClient) AddonManageWriter() *addonManageWriter {
	return &addonManageWriter{db: c.DBEngine}
}

func (w *addonManageWriter) Create(a *AddonManagement) (uint64, error) {
	db := w.db.Save(a)
	return a.ID, db.Error
}

func (w *addonManageWriter) Update(a AddonManagement) error {
	return w.db.Model(&a).Updates(a).Error
}

func (w *addonManageWriter) Delete(ids ...string) error {
	render := strutil.Map(ids, func(s string) string { return "\"" + s + "\"" })
	return w.db.Delete(AddonManagement{}, fmt.Sprintf("id in (%s)", strutil.Join(render, ","))).Error
}

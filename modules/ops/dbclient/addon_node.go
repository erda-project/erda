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
	"strings"
	"time"

	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// AddonNode addon node信息
type AddonNode struct {
	ID         string `gorm:"type:varchar(64)"`
	InstanceID string `gorm:"type:varchar(64)"` // AddonInstance 主键
	Namespace  string `gorm:"type:text"`
	NodeName   string
	CPU        float64
	Mem        uint64
	Deleted    string    `gorm:"column:is_deleted"` // Y: 已删除 N: 未删除
	CreatedAt  time.Time `gorm:"column:create_time"`
	UpdatedAt  time.Time `gorm:"column:update_time"`
}

// TableName 数据库表名
func (AddonNode) TableName() string {
	return "tb_middle_node"
}

type addonNodeReader struct {
	db         *dbengine.DBEngine
	conditions []string
	limit      int
	offset     int
}

type addonNodeWriter struct {
	db *dbengine.DBEngine
}

// read condition
func (c *DBClient) AddonNodeReader() *addonNodeReader {
	return &addonNodeReader{db: c.DBEngine, conditions: []string{}, limit: 0, offset: -1}
}

func (r *addonNodeReader) ByAddonIDs(aids ...string) *addonNodeReader {
	render := strutil.Map(aids, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("instance_id in (%s)", strutil.Join(render, ",")))
	return r
}

func (r *addonNodeReader) ByDeleteStatus(status ...string) *addonNodeReader {
	render := strutil.Map(status, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("is_deleted in (%s)", strutil.Join(render, ",")))
	return r
}

// read
func (r *addonNodeReader) Do() ([]AddonNode, error) {
	ams := []AddonNode{}
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
func (c *DBClient) AddonNodeWriter() *addonNodeWriter {
	return &addonNodeWriter{db: c.DBEngine}
}

func (w *addonNodeWriter) Create(a *AddonNode) (string, error) {
	db := w.db.Save(a)
	return a.ID, db.Error
}

func (w *addonNodeWriter) Update(a AddonNode) error {
	return w.db.Model(&a).Updates(a).Error
}

func (w *addonNodeWriter) Updates(cpu float64, mem uint64, addonIDs ...string) error {
	render := strutil.Map(addonIDs, func(s string) string { return "\"" + s + "\"" })
	return w.db.Model(AddonNode{}).Where(fmt.Sprintf("instance_id in (%s)", strutil.Join(render, ","))).
		Updates(map[string]interface{}{"cpu": cpu, "mem": mem}).Error
}

func (w *addonNodeWriter) Delete(ids ...string) error {
	render := strutil.Map(ids, func(s string) string { return "\"" + s + "\"" })
	return w.db.Delete(AddonNode{}, fmt.Sprintf("id in (%s)", strutil.Join(render, ","))).Error
}

type AddonNodeList []AddonNode

func (r AddonNodeList) Len() int {
	return len(r)
}

func (r AddonNodeList) Less(i, j int) bool {
	return strings.Compare(r[i].NodeName, r[j].NodeName) < 0
}

func (r AddonNodeList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

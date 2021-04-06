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

package instanceinfo

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

type podReader struct {
	db         *dbengine.DBEngine
	conditions []string
	limit      int
}

type podWriter struct {
	db *dbengine.DBEngine
}

func (c *Client) PodReader() *podReader {
	return &podReader{db: c.db, conditions: []string{}, limit: 0}
}
func (r *podReader) ByCluster(clustername string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("cluster = \"%s\"", clustername))
	return r
}
func (r *podReader) ByNamespace(ns string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("namespace = \"%s\"", ns))
	return r
}
func (r *podReader) ByName(name string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("name = \"%s\"", name))
	return r
}
func (r *podReader) ByOrgName(name string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("org_name = \"%s\"", name))
	return r
}
func (r *podReader) ByOrgID(id string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("org_id = \"%s\"", id))
	return r
}
func (r *podReader) ByProjectName(name string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("project_name = \"%s\"", name))
	return r
}
func (r *podReader) ByProjectID(id string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("project_id = \"%s\"", id))
	return r
}
func (r *podReader) ByApplicationName(name string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("application_name = \"%s\"", name))
	return r
}
func (r *podReader) ByApplicationID(id string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("application_id = \"%s\"", id))
	return r
}
func (r *podReader) ByRuntimeName(name string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("runtime_name = \"%s\"", name))
	return r
}
func (r *podReader) ByRuntimeID(id string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("runtime_id = \"%s\"", id))
	return r
}
func (r *podReader) ByService(name string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("service_name = \"%s\"", name))
	return r
}
func (r *podReader) ByServiceType(tp string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("service_type = \"%s\"", tp))
	return r
}
func (r *podReader) ByAddonID(id string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("addon_id = \"%s\"", id))
	return r
}
func (r *podReader) ByWorkspace(ws string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("workspace = \"%s\"", ws))
	return r
}
func (r *podReader) ByPhase(phase string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("phase = \"%s\"", phase))
	return r
}
func (r *podReader) ByPhases(phases ...string) *podReader {
	phasesStr := strutil.Map(phases, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("phase in (%s)", strutil.Join(phasesStr, ",")))
	return r
}
func (r *podReader) ByK8SNamespace(namespace string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("k8s_namespace = \"%s\"", namespace))
	return r
}
func (r *podReader) ByPodName(podname string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("pod_name = \"%s\"", podname))
	return r
}
func (r *podReader) ByUid(uid string) *podReader {
	r.conditions = append(r.conditions, fmt.Sprintf("uid = \"%s\"", uid))
	return r
}
func (r *podReader) ByUpdatedTime(beforeNSecs int) *podReader {
	//使用scheduler时间查询，避免sceduler跟数据库时间不一致导致实例误GC
	now := time.Now().Format("2006-01-02 15:04:05")
	r.conditions = append(r.conditions, fmt.Sprintf("updated_at < '%s' - interval %d second", now, beforeNSecs))
	return r
}

func (r *podReader) Limit(n int) *podReader {
	r.limit = n
	return r
}
func (r *podReader) Do() ([]PodInfo, error) {
	podinfo := []PodInfo{}
	expr := r.db.Where(strutil.Join(r.conditions, " AND ", true)).Order("started_at desc")
	if r.limit != 0 {
		expr = expr.Limit(r.limit)
	}
	if err := expr.Find(&podinfo).Error; err != nil {
		r.conditions = []string{}
		return nil, err
	}
	r.conditions = []string{}
	return podinfo, nil
}

func (c *Client) PodWriter() *podWriter {
	return &podWriter{db: c.db}
}
func (w *podWriter) Create(s *PodInfo) error {
	return w.db.Save(s).Error
}
func (w *podWriter) Update(s PodInfo) error {
	return w.db.Model(&s).Updates(s).Update("updated_at", time.Now()).Error
}
func (w *podWriter) Delete(ids ...uint64) error {
	return w.db.Delete(PodInfo{}, "id in (?)", ids).Error
}

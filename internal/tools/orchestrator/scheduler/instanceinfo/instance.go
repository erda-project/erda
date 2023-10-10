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

package instanceinfo

import (
	"time"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

type InstanceReader struct {
	db         *dbengine.DBEngine
	conditions []string
	values     []interface{}
	limit      int
}

type instanceWriter struct {
	db *dbengine.DBEngine
}

func (c *Client) InstanceReader() *InstanceReader {
	return &InstanceReader{db: c.db, conditions: []string{}, limit: 0}
}

func (r *InstanceReader) ByCluster(clustername string) *InstanceReader {
	r.conditions = append(r.conditions, "cluster = ?")
	r.values = append(r.values, clustername)
	return r
}
func (r *InstanceReader) ByNamespace(ns string) *InstanceReader {
	r.conditions = append(r.conditions, "namespace = ?")
	r.values = append(r.values, ns)
	return r
}
func (r *InstanceReader) ByName(name string) *InstanceReader {
	r.conditions = append(r.conditions, "name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByOrgName(name string) *InstanceReader {
	r.conditions = append(r.conditions, "org_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByOrgID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "org_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *InstanceReader) ByProjectName(name string) *InstanceReader {
	r.conditions = append(r.conditions, "project_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByProjectID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "project_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *InstanceReader) ByApplicationName(name string) *InstanceReader {
	r.conditions = append(r.conditions, "application_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByEdgeApplicationName(name string) *InstanceReader {
	r.conditions = append(r.conditions, "edge_application_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByEdgeSite(name string) *InstanceReader {
	r.conditions = append(r.conditions, "edge_site = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByApplicationID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "application_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *InstanceReader) ByRuntimeName(name string) *InstanceReader {
	r.conditions = append(r.conditions, "runtime_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByRuntimeID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "runtime_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *InstanceReader) ByService(name string) *InstanceReader {
	r.conditions = append(r.conditions, "service_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *InstanceReader) ByWorkspace(ws string) *InstanceReader {
	r.conditions = append(r.conditions, "workspace = ?")
	r.values = append(r.values, ws)
	return r
}
func (r *InstanceReader) ByContainerID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "container_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *InstanceReader) ByServiceType(tp string) *InstanceReader {
	r.conditions = append(r.conditions, "service_type = ?")
	r.values = append(r.values, tp)
	return r
}
func (r *InstanceReader) ByPhase(phase string) *InstanceReader {
	r.conditions = append(r.conditions, "phase = ?")
	r.values = append(r.values, phase)
	return r
}
func (r *InstanceReader) ByPhases(phases ...string) *InstanceReader {
	r.conditions = append(r.conditions, "phase IN (?)")
	r.values = append(r.values, phases)
	return r
}
func (r *InstanceReader) ByFinishedTime(beforeNday int) *InstanceReader {
	r.conditions = append(r.conditions, "finished_at < now() - interval ？ day")
	r.values = append(r.values, beforeNday)
	return r
}
func (r *InstanceReader) ByUpdatedTime(beforeNSecs int) *InstanceReader {
	// Use scheduler time query to avoid the inconsistency between sceduler and database time and cause the instance to GC by mistake
	now := time.Now().Format("2006-01-02 15:04:05")
	r.conditions = append(r.conditions, "updated_at < "+now+" - interval ? second")
	r.values = append(r.values, beforeNSecs)
	return r
}
func (r *InstanceReader) ByTaskID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "task_id = ?")
	r.values = append(r.values, id)
	return r
}

func (r *InstanceReader) ByNotTaskID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "task_id <> ?")
	r.values = append(r.values, id)
	return r
}

func (r *InstanceReader) ByAddonID(id string) *InstanceReader {
	r.conditions = append(r.conditions, "addon_id = ?")
	r.values = append(r.values, id)
	return r
}

func (r *InstanceReader) ByInstanceIP(ips ...string) *InstanceReader {
	r.conditions = append(r.conditions, "container_ip IN (?)")
	r.values = append(r.values, ips)
	return r
}

func (r *InstanceReader) ByHostIP(ips ...string) *InstanceReader {
	r.conditions = append(r.conditions, "host_ip IN (?)")
	r.values = append(r.values, ips)
	return r
}

func (r *InstanceReader) ByMetaLike(s string) *InstanceReader {
	r.conditions = append(r.conditions, "meta LIKE ?")
	r.values = append(r.values, "%"+s+"%")
	return r
}

func (r *InstanceReader) Limit(n int) *InstanceReader {
	r.limit = n
	return r
}
func (r *InstanceReader) Do() ([]InstanceInfo, error) {
	instanceinfo := []InstanceInfo{}
	expr := r.db.Order("started_at desc")
	for k := range r.conditions {
		expr = expr.Where(r.conditions[k], r.values[k])
	}
	if r.limit != 0 {
		expr = expr.Limit(r.limit)
	}
	if err := expr.Find(&instanceinfo).Error; err != nil {
		r.conditions = []string{}
		return nil, err
	}
	r.conditions = []string{}
	return instanceinfo, nil
}

func (c *Client) InstanceWriter() *instanceWriter {
	return &instanceWriter{db: c.db}
}
func (w *instanceWriter) Create(s *InstanceInfo) error {
	return w.db.Save(s).Error
}
func (w *instanceWriter) Update(s InstanceInfo) error {
	return w.db.Model(&s).Updates(s).Update("updated_at", time.Now()).Error
}
func (w *instanceWriter) Delete(ids ...uint64) error {
	return w.db.Delete(InstanceInfo{}, "id in (?)", ids).Error
}

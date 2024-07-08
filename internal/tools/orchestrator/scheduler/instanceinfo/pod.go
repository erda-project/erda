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

type PodReader struct {
	db         *dbengine.DBEngine
	conditions []string
	values     []interface{}
	limit      int
}

type podWriter struct {
	db *dbengine.DBEngine
}

func (c *Client) PodReader() *PodReader {
	return &PodReader{db: c.db, conditions: []string{}, limit: 0}
}
func (r *PodReader) ByCluster(clustername string) *PodReader {
	r.conditions = append(r.conditions, "cluster = ?")
	r.values = append(r.values, clustername)
	return r
}
func (r *PodReader) ByNamespace(ns string) *PodReader {
	r.conditions = append(r.conditions, "namespace = ?")
	r.values = append(r.values, ns)
	return r
}
func (r *PodReader) ByName(name string) *PodReader {
	r.conditions = append(r.conditions, "name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *PodReader) ByOrgName(name string) *PodReader {
	r.conditions = append(r.conditions, "org_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *PodReader) ByOrgID(id string) *PodReader {
	r.conditions = append(r.conditions, "org_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *PodReader) ByProjectName(name string) *PodReader {
	r.conditions = append(r.conditions, "project_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *PodReader) ByProjectID(id string) *PodReader {
	r.conditions = append(r.conditions, "project_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *PodReader) ByApplicationName(name string) *PodReader {
	r.conditions = append(r.conditions, "application_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *PodReader) ByApplicationID(id string) *PodReader {
	r.conditions = append(r.conditions, "application_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *PodReader) ByRuntimeName(name string) *PodReader {
	r.conditions = append(r.conditions, "runtime_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *PodReader) ByRuntimeID(id string) *PodReader {
	r.conditions = append(r.conditions, "runtime_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *PodReader) ByService(name string) *PodReader {
	r.conditions = append(r.conditions, "service_name = ?")
	r.values = append(r.values, name)
	return r
}
func (r *PodReader) ByServiceType(tp string) *PodReader {
	r.conditions = append(r.conditions, "service_type = ?")
	r.values = append(r.values, tp)
	return r
}
func (r *PodReader) ByAddonID(id string) *PodReader {
	r.conditions = append(r.conditions, "addon_id = ?")
	r.values = append(r.values, id)
	return r
}
func (r *PodReader) ByWorkspace(ws string) *PodReader {
	r.conditions = append(r.conditions, "workspace = ?")
	r.values = append(r.values, ws)
	return r
}
func (r *PodReader) ByPhase(phase string) *PodReader {
	r.conditions = append(r.conditions, "phase = ?")
	r.values = append(r.values, phase)
	return r
}
func (r *PodReader) ByPhases(phases ...string) *PodReader {
	r.conditions = append(r.conditions, "phase in (?)")
	r.values = append(r.values, phases)
	return r
}
func (r *PodReader) ByK8SNamespace(namespace string) *PodReader {
	r.conditions = append(r.conditions, "k8s_namespace = ?")
	r.values = append(r.values, namespace)
	return r
}
func (r *PodReader) ByPodName(podname string) *PodReader {
	r.conditions = append(r.conditions, "pod_name = ?")
	r.values = append(r.values, podname)
	return r
}
func (r *PodReader) ByUid(uid string) *PodReader {
	r.conditions = append(r.conditions, "uid = ?")
	r.values = append(r.values, uid)
	return r
}
func (r *PodReader) ByUpdatedTime(beforeNSecs int) *PodReader {
	// Use scheduler time query to avoid the inconsistency between sceduler and database time and cause the instance to GC by mistake
	now := time.Now().Format("2006-01-02 15:04:05")
	r.conditions = append(r.conditions, "updated_at < "+now+" - interval ? second")
	r.values = append(r.values, beforeNSecs)
	return r
}

func (r *PodReader) Limit(n int) *PodReader {
	r.limit = n
	return r
}
func (r *PodReader) Do() ([]PodInfo, error) {
	podinfo := []PodInfo{}
	expr := r.db.Order("started_at desc")
	for k := range r.conditions {
		expr = expr.Where(r.conditions[k], r.values[k])
	}
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

package instanceinfo

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

type instanceReader struct {
	db         *dbengine.DBEngine
	conditions []string
	limit      int
}

type instanceWriter struct {
	db *dbengine.DBEngine
}

func (c *Client) InstanceReader() *instanceReader {
	return &instanceReader{db: c.db, conditions: []string{}, limit: 0}
}

func (r *instanceReader) ByCluster(clustername string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("cluster = \"%s\"", clustername))
	return r
}
func (r *instanceReader) ByNamespace(ns string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("namespace = \"%s\"", ns))
	return r
}
func (r *instanceReader) ByName(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByOrgName(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("org_name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByOrgID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("org_id = \"%s\"", id))
	return r
}
func (r *instanceReader) ByProjectName(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("project_name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByProjectID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("project_id = \"%s\"", id))
	return r
}
func (r *instanceReader) ByApplicationName(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("application_name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByEdgeApplicationName(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("edge_application_name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByEdgeSite(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("edge_site = \"%s\"", name))
	return r
}
func (r *instanceReader) ByApplicationID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("application_id = \"%s\"", id))
	return r
}
func (r *instanceReader) ByRuntimeName(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("runtime_name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByRuntimeID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("runtime_id = \"%s\"", id))
	return r
}
func (r *instanceReader) ByService(name string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("service_name = \"%s\"", name))
	return r
}
func (r *instanceReader) ByWorkspace(ws string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("workspace = \"%s\"", ws))
	return r
}
func (r *instanceReader) ByContainerID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("container_id = \"%s\"", id))
	return r
}
func (r *instanceReader) ByServiceType(tp string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("service_type = \"%s\"", tp))
	return r
}
func (r *instanceReader) ByPhase(phase string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("phase = \"%s\"", phase))
	return r
}
func (r *instanceReader) ByPhases(phases ...string) *instanceReader {
	phasesStr := strutil.Map(phases, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("phase in (%s)", strutil.Join(phasesStr, ",")))
	return r
}
func (r *instanceReader) ByFinishedTime(beforeNday int) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("finished_at < now() - interval %d day", beforeNday))
	return r
}
func (r *instanceReader) ByUpdatedTime(beforeNSecs int) *instanceReader {
	//使用scheduler时间查询，避免sceduler跟数据库时间不一致导致实例误GC
	now := time.Now().Format("2006-01-02 15:04:05")
	r.conditions = append(r.conditions, fmt.Sprintf("updated_at < '%s' - interval %d second", now, beforeNSecs))
	return r
}
func (r *instanceReader) ByTaskID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("task_id = \"%s\"", id))
	return r
}

func (r *instanceReader) ByNotTaskID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("task_id <> \"%s\"", id))
	return r
}

func (r *instanceReader) ByAddonID(id string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("addon_id = \"%s\"", id))
	return r
}

func (r *instanceReader) ByInstanceIP(ips ...string) *instanceReader {
	ipsStr := strutil.Map(ips, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("container_ip in (%s)", strutil.Join(ipsStr, ",")))
	return r
}

func (r *instanceReader) ByHostIP(ips ...string) *instanceReader {
	ipsStr := strutil.Map(ips, func(s string) string { return "\"" + s + "\"" })
	r.conditions = append(r.conditions, fmt.Sprintf("host_ip in (%s)", strutil.Join(ipsStr, ",")))
	return r
}

func (r *instanceReader) ByMetaLike(s string) *instanceReader {
	r.conditions = append(r.conditions, fmt.Sprintf("meta LIKE '%%%s%%'", s))
	return r
}

func (r *instanceReader) Limit(n int) *instanceReader {
	r.limit = n
	return r
}
func (r *instanceReader) Do() ([]InstanceInfo, error) {
	instanceinfo := []InstanceInfo{}
	expr := r.db.Where(strutil.Join(r.conditions, " AND ", true)).Order("started_at desc")
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

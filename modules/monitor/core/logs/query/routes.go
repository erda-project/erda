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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/core/logs/schema"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.GET("/api/logs", p.queryLog)
	routes.GET("/api/logs/actions/download", p.downloadLog)
	routes.POST("/api/logs/actions/delete", p.deleteLog)

	// runtime
	p.getApplicationID = permission.QueryValue("applicationId")
	routes.GET("/api/runtime/logs", p.queryRuntimeLog, permission.Intercepter(
		permission.ScopeApp, p.getApplicationID,
		pkg.ResourceRuntime, permission.ActionGet,
	))
	routes.GET("/api/runtime/logs/actions/download", p.downloadRuntimeLog, permission.Intercepter(
		permission.ScopeApp, p.getApplicationID,
		pkg.ResourceRuntime, permission.ActionGet,
	))
	// org
	p.checkOrgCluster = permission.OrgIDByCluster("clusterName")
	routes.GET("/api/orgCenter/logs", p.queryOrgLog, permission.Intercepter(
		permission.ScopeOrg, p.checkContainerLog,
		pkg.ResourceOrgCenter, permission.ActionGet,
	))
	routes.GET("/api/orgCenter/logs/actions/download", p.downloadOrgLog, permission.Intercepter(
		permission.ScopeOrg, p.checkContainerLog,
		pkg.ResourceOrgCenter, permission.ActionGet,
	))
	return nil
}

// Request .
type RequestCtx struct {
	RequestID     string `form:"requestId"`
	LogID         string `form:"requestId"`
	Source        string `form:"source"`
	ID            string `form:"id"`
	Stream        string `form:"stream" default:"stdout"`
	Start         int64  `form:"start"`
	End           int64  `form:"end"`
	Count         int64  `form:"count"`
	ApplicationID string `from:"applicationId"`
	ClusterName   string `from:"clusterName"`
}

// Response .
type Response struct {
	Lines []*Log `json:"lines"`
}

const (
	defaultStream      = "stdout"
	defaultCount       = 50
	maxCount, minCount = 200, -200
	maxTimeRange       = 7 * 24 * int64(time.Hour)
	// day                = 24 * int64(time.Hour)
)

func normalizeRequest(r *RequestCtx) error {
	if len(r.RequestID) <= 0 {
		r.RequestID = r.LogID
	}
	if len(r.RequestID) > 0 {
		return nil // 直接查询所有 trace 日志
	}

	if len(r.Source) <= 0 || len(r.ID) <= 0 {
		return fmt.Errorf("missing parameter source or id")
	}
	if len(r.Stream) <= 0 {
		r.Stream = defaultStream
	}
	if r.End <= 0 {
		r.End = time.Now().UnixNano()
	}
	if r.Start <= 0 {
		r.Start = r.End - maxTimeRange
		if r.Start < 0 {
			r.Start = 0
		}
	}
	if r.End < r.Start {
		return fmt.Errorf("start must be less than end")
	} else if r.End-r.Start > maxTimeRange {
		return fmt.Errorf("time range is too large")
	}
	if r.Count < minCount {
		r.Count = minCount
	} else if r.Count > maxCount {
		r.Count = maxCount
	} else if r.Count == 0 {
		r.Count = defaultCount
	}
	return nil
}

func (p *provider) queryRuntimeLog(r *RequestCtx) interface{} {
	result, err := p.checkLogMeta(r.Source, r.ID, "dice_application_id", r.ApplicationID)
	if err != nil {
		return api.Errors.Internal(err)
	} else if !result {
		return api.Success(&Response{[]*Log{}})
	}
	return p.queryLog(r)
}

func (p *provider) queryOrgLog(r *RequestCtx) interface{} {
	result, err := p.checkLogMeta(r.Source, r.ID, "dice_cluster_name", r.ClusterName)
	if err != nil {
		return api.Errors.Internal(err)
	} else if !result {
		return api.Success(&Response{[]*Log{}})
	}
	return p.queryLog(r)
}

func (p *provider) downloadRuntimeLog(w http.ResponseWriter, r *RequestCtx) interface{} {
	result, err := p.checkLogMeta(r.Source, r.ID, "dice_application_id", r.ApplicationID)
	if err != nil {
		return api.Errors.Internal(err)
	} else if !result {
		return nil
	}
	return p.downloadLog(w, r)
}

func (p *provider) downloadOrgLog(w http.ResponseWriter, r *RequestCtx) interface{} {
	result, err := p.checkLogMeta(r.Source, r.ID, "dice_cluster_name", r.ClusterName)
	if err != nil {
		return api.Errors.Internal(err)
	} else if !result {
		return nil
	}
	return p.downloadLog(w, r)
}

func (p *provider) getTableNameWithFilters(filters map[string]interface{}) string {
	table := schema.DefaultBaseLogTable
	metas, err := p.queryBaseLogMetaWithFilters(filters)
	if err != nil || len(metas) == 0 {
		return table
	}
	if v, ok := metas[0].Tags["dice_org_name"]; ok {
		table = schema.BaseLogWithOrgName(v)
	}
	return table
}

func (p *provider) checkLogMeta(source, id, key, value string) (bool, error) {
	if source != "container" { // permission check only for container
		return true, nil
	}
	metas, err := p.queryBaseLogMetaWithFilters(map[string]interface{}{
		"source": source,
		"id":     id,
	})
	if err != nil || len(metas) == 0 {
		return false, err
	}
	return metas[0].Tags[key] == value, nil
}

func (p *provider) queryLog(r *RequestCtx) interface{} {
	if err := normalizeRequest(r); err != nil {
		return api.Errors.InvalidParameter(err)
	}

	// 请求ID关联的日志
	if len(r.RequestID) > 0 {
		logs, err := p.queryRequestLog(
			p.getTableNameWithFilters(map[string]interface{}{
				"tags['dice_application_id']": r.ApplicationID,
			}),
			r.RequestID,
		)
		if err != nil {
			return api.Errors.Internal(err)
		}
		return api.Success(&Response{logs})
	}

	// 基础日志
	if r.Count == 0 {
		return api.Success(&Response{})
	}
	logs, err := p.queryBaseLog(
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
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(&Response{logs})
}

func (p *provider) downloadLog(w http.ResponseWriter, r *RequestCtx) interface{} {
	if err := normalizeRequest(r); err != nil {
		return api.Errors.InvalidParameter(err)
	}

	filename := strings.Replace(r.ID, ".", "_", -1) + "_" + r.Stream
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("charset", "utf-8")
	w.Header().Set("Content-Disposition", "attachment;filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")

	flusher := w.(http.Flusher)
	err := p.walkSavedLogs(
		p.getTableNameWithFilters(map[string]interface{}{
			"source": r.Source,
			"id":     r.ID,
		}),
		r.Source,
		r.ID,
		r.Stream,
		r.Start,
		r.End,
		func(logs []*SavedLog) error {
			for _, log := range logs {
				content, err := gunzipContent(log.Content)
				if err != nil {
					return err
				}
				w.Write(content)
				w.Write([]byte("\n"))
			}
			flusher.Flush()
			return nil
		},
	)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return nil
}

func (p *provider) deleteLog(r *http.Request, params struct {
	Offset  string `query:"offset" validate:"required"`
	OrgName string `query:"org_name" default:"prod"` // default spot_prod.base_log
}) interface{} {
	offset, err := ConvertStringToMS(params.Offset, 0)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}

	ids, err := p.getContainerIds(r.URL.Query(), offset, params.OrgName)
	if err != nil {
		return api.Errors.Internal(fmt.Errorf("get containers failed. err=%s", err))
	}
	buckets := p.getTimeBuckets(offset)
	if len(ids) == 0 || len(buckets) == 0 {
		return api.Errors.Internal(fmt.Errorf("invalid data. ids=%+v, buckets=%+v", ids, buckets))
	}

	if err := p.removeBaseLog(ids, buckets, params.OrgName); err != nil {
		return api.Errors.Internal(fmt.Errorf("delete failed. err=%s", err))
	}
	return api.Success("delete successfully!")
}

func ConvertStringToMS(value string, now int64) (int64, error) {
	if value == "" {
		return 0, nil
	}

	if now == 0 {
		now = time.Now().UnixNano() / int64(time.Millisecond)
	}
	if len(value) > 0 && unicode.IsDigit([]rune(value)[0]) {
		ts, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timestamp %s", value)
		}
		return ts, nil
	}
	if strings.HasPrefix(value, "before_") {
		d, err := getMillisecond(value[len("before_"):])
		if err != nil {
			return 0, nil
		}
		return now - d, nil
	} else if strings.HasPrefix(value, "after_") {
		d, err := getMillisecond(value[len("after_"):])
		if err != nil {
			return 0, nil
		}
		return now + d, nil
	}
	return now, nil
}

func getMillisecond(value string) (int64, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %s", value)
	}
	return int64(d) / int64(time.Millisecond), nil
}

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

package query

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log/schema"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
	api "github.com/erda-project/erda/pkg/common/httpapi"
)

func (p *provider) intRoutes(routes httpserver.Router) {
	routes.GET("/api/logs/actions/download", p.downloadLog)

	// runtime
	routes.GET("/api/runtime/logs/actions/download", p.downloadRuntimeLog, permission.Intercepter(
		permission.ScopeApp, permission.QueryValue("applicationId"),
		common.ResourceRuntime, permission.ActionGet,
	))

	// org
	routes.GET("/api/orgCenter/logs/actions/download", p.downloadOrgLog, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDByCluster("clusterName"),
		common.ResourceOrgCenter, permission.ActionGet,
	))
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

const (
	defaultStream      = "stdout"
	defaultCount       = 50
	maxCount, minCount = 200, -200
	maxTimeRange       = 7 * 24 * int64(time.Hour)
)

func convertToRequestCtx(req interface{}) (*RequestCtx, error) {
	switch req.(type) {
	case *pb.GetLogRequest:
		v := req.(*pb.GetLogRequest)
		return &RequestCtx{
			RequestID:     v.RequestId,
			LogID:         v.RequestId,
			Source:        v.Source,
			ID:            v.Id,
			Stream:        v.Stream,
			Start:         v.Start,
			End:           v.End,
			Count:         v.Count,
			ApplicationID: "",
			ClusterName:   "",
		}, nil
	case *pb.GetLogByRuntimeRequest:
		v := req.(*pb.GetLogByRuntimeRequest)
		return &RequestCtx{
			RequestID:     v.RequestId,
			LogID:         v.RequestId,
			Source:        v.Source,
			ID:            v.Id,
			Stream:        v.Stream,
			Start:         v.Start,
			End:           v.End,
			Count:         v.Count,
			ApplicationID: v.ApplicationId,
			ClusterName:   "",
		}, nil
	case *pb.GetLogByOrganizationRequest:
		v := req.(*pb.GetLogByOrganizationRequest)
		return &RequestCtx{
			RequestID:     v.RequestId,
			LogID:         v.RequestId,
			Source:        v.Source,
			ID:            v.Id,
			Stream:        v.Stream,
			Start:         v.Start,
			End:           v.End,
			Count:         v.Count,
			ApplicationID: "",
			ClusterName:   v.ClusterName,
		}, nil
	}
	return &RequestCtx{}, errors.New("invalid request type")
}

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
	meta, err := p.queryBaseLogMetaWithFilters(filters)
	if err != nil {
		return table
	}
	if v, ok := meta.Tags["dice_org_name"]; ok {
		table = schema.BaseLogWithOrgName(v)
	}
	return table
}

func (p *provider) checkLogMeta(source, id, key, value string) (bool, error) {
	if source != "container" { // permission check only for container
		return true, nil
	}
	meta, err := p.queryBaseLogMetaWithFilters(map[string]interface{}{
		"source": source,
		"id":     id,
	})
	if errors.Is(err, ErrEmptyLogMeta) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return meta.Tags[key] == value, nil
}

func (p *provider) downloadLog(w http.ResponseWriter, r *RequestCtx) interface{} {
	if err := normalizeRequest(r); err != nil {
		return api.Errors.InvalidParameter(err)
	}

	meta, _ := p.queryBaseLogMetaWithFilters(map[string]interface{}{
		"source": r.Source,
		"id":     r.ID,
	})
	filename := getFilename(r, meta)
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

func getFilename(r *RequestCtx, meta *LogMeta) string {
	sep, filenamePrefix := "_", ""
	if meta == nil {
		filenamePrefix = strings.Replace(r.ID, ".", sep, -1)
	} else {
		filenamePrefix = strings.Replace(meta.ID, ".", sep, -1)

		if val, ok := meta.Tags["pod_name"]; ok {
			filenamePrefix = val
		}
		if val, ok := meta.Tags["dice_application_name"]; ok {
			filenamePrefix = val
		}
		if val, ok := meta.Tags["dice_service_name"]; ok {
			filenamePrefix = val
		}
	}
	return strings.Join([]string{filenamePrefix, r.Stream, strconv.Itoa(int(time.Now().Unix()))}, sep) + ".log"
}

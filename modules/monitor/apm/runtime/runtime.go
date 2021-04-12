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

package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/erda-project/erda-infra/providers/httpserver"
	apm "github.com/erda-project/erda/modules/monitor/apm/common"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/erda-project/erda/modules/monitor/common/permission"
	"github.com/olivere/elastic"
)

type Vo struct {
	RuntimeId     string `query:"runtimeId"`
	RuntimeName   string `query:"runtimeName"`
	TerminusKey   string `query:"terminusKey"`
	ApplicationId string `query:"applicationId"`
}

const (
	ApplicationServiceNode = "application_service_node"
)

func getRuntimePermission(db *db.DB) httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.TkFromParams(db),
		apm.Monitor, permission.ActionGet,
	)
}

func getProjectPermission() httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.ProjectIdFromParams(),
		apm.Monitor, permission.ActionGet,
	)
}

func getInstancePermission(db *db.DB) httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.TkFromParams(db),
		apm.Monitor, permission.ActionGet,
	)
}

func (runtime *provider) getInstanceByTk(key string) (db.Monitor, error) {
	return runtime.db.Monitor.GetInstanceByTk(key)
}

func (runtime *provider) getTkByProjectIdAndWorkspace(projectId, workspace string) (string, error) {
	return runtime.db.Monitor.GetTkByProjectIdAndWorkspace(projectId, workspace)
}

func (runtime *provider) getRuntime(params Vo) (*Info, error) {
	ctx := context.Background()
	boolQuery := elastic.NewBoolQuery()
	if params.RuntimeId == "" {
		boolQuery.Filter(elastic.NewTermQuery(apm.TagsTerminusKey, params.TerminusKey)).
			Filter(elastic.NewTermQuery(apm.TagsApplicationId, params.ApplicationId)).
			Filter(elastic.NewTermQuery(apm.TagsRuntimeName, params.RuntimeName))
	} else {
		boolQuery.Filter(elastic.NewTermQuery(apm.TagsRuntimeId, params.RuntimeId))
	}
	nowMs := time.Now().UnixNano() / 1e6
	indices := apm.CreateEsIndices(ApplicationServiceNode, nowMs-1, nowMs)

	searchSource := elastic.NewSearchSource().Query(boolQuery).Size(1)

	do, err := runtime.es.Search(indices...).SearchSource(searchSource).Do(ctx)
	if err != nil {
		return nil, err
	}
	source := do.Hits.Hits[0].Source
	runtimeInfo, err := parseToRuntime(source)
	return runtimeInfo, err
}

type Info struct {
	TerminusKey     string `json:"terminus_key"`
	Workspace       string `json:"Workspace"`
	ProjectId       string `json:"project_id"`
	ProjectName     string `json:"project_name"`
	ApplicationId   string `json:"application_id"`
	ApplicationName string `json:"application_name"`
	RuntimeId       string `json:"runtime_id"`
	RuntimeName     string `json:"runtime_name"`
}

func parseToRuntime(hits *json.RawMessage) (*Info, error) {
	var runtimeInfo Info
	m := make(map[string]interface{})
	err := json.Unmarshal(*hits, &m)
	if err != nil {
		return nil, err
	}
	tags := m[apm.Tags]
	tagsJson, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tagsJson, &runtimeInfo)
	if err != nil {
		return nil, err
	}
	return &runtimeInfo, err
}

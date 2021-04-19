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

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// 项目 + env 日志查询
	routes.GET("/api/micro_service/:addon/logs/statistic/histogram", p.logStatistic)
	routes.GET("/api/micro_service/:addon/logs/search", p.logSearch)
	routes.GET("/api/micro_service/logs/tags/tree", p.logMSTagsTree)

	// 企业日志查询
	routes.GET("/api/org/logs/statistic/histogram", p.logStatistic)
	routes.GET("/api/org/logs/search", p.logSearch)
	routes.GET("/api/org/logs/tags/tree", p.orgLogTagsTree)
	return nil
}

func (p *provider) buildLogFilters(r *http.Request) []*Tag {
	if r.URL == nil {
		return nil
	}
	var filters []*Tag
	for k, vs := range r.URL.Query() {
		if strings.HasPrefix(k, "tags.") {
			k = k[len("tags."):]
			if len(k) <= 0 {
				continue
			}
			for _, v := range vs {
				filters = append(filters, &Tag{
					Key:   k,
					Value: v,
				})
			}
		}
	}
	return filters
}

func (p *provider) checkTime(start, end int64) error {
	if end <= start {
		return fmt.Errorf("end must after start")
	}
	if end-start > 7*24*60*60*1000 {
		return fmt.Errorf("too large time span")
	}
	return nil
}

func (p *provider) logStatistic(r *http.Request, params struct {
	Start       int64  `query:"start" validate:"gte=1"`
	End         int64  `query:"end" validate:"gte=1"`
	Query       string `query:"query"`
	Points      int64  `query:"points"`
	Interval    int64  `query:"interval"`
	Debug       bool   `query:"debug"`
	Addon       string `param:"addon"`
	ClusterName string `query:"clusterName"`
}) interface{} {
	orgID := api.OrgID(r)
	orgid, err := strconv.ParseInt(orgID, 10, 64)
	if err != nil {
		return api.Errors.InvalidParameter("invalid Org-ID")
	}
	if params.Points <= 0 {
		params.Points = 60
	}
	filters := p.buildLogFilters(r)
	data, err := p.StatisticLogs(&LogStatisticRequest{
		LogRequest: LogRequest{
			OrgID:       orgid,
			ClusterName: params.ClusterName,
			Addon:       params.Addon,
			Start:       params.Start,
			End:         params.End,
			Filters:     filters,
			Query:       params.Query,
			Debug:       params.Debug,
			Lang:        api.Language(r),
		},
		Points:   params.Points,
		Interval: params.Interval,
	})
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) logSearch(r *http.Request, params struct {
	Start       int64  `query:"start" validate:"gte=1"`
	End         int64  `query:"end" validate:"gte=1"`
	Size        int64  `query:"size"`
	Query       string `query:"query"`
	Sort        string `query:"sort"`
	Debug       bool   `query:"debug"`
	Addon       string `param:"addon"`
	ClusterName string `query:"clusterName"`
}) interface{} {
	orgID := api.OrgID(r)
	orgid, err := strconv.ParseInt(orgID, 10, 64)
	if err != nil {
		return api.Errors.InvalidParameter("invalid Org-ID")
	}
	if params.Size <= 0 {
		params.Size = 50
	}
	err = p.checkTime(params.Start, params.End)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	filters := p.buildLogFilters(r)
	logs, err := p.SearchLogs(&LogSearchRequest{
		LogRequest: LogRequest{
			OrgID:       orgid,
			ClusterName: params.ClusterName,
			Addon:       params.Addon,
			Start:       params.Start,
			End:         params.End,
			Filters:     filters,
			Query:       params.Query,
			Debug:       params.Debug,
			Lang:        api.Language(r),
		},
		Size: params.Size,
		Sort: params.Sort,
	})
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(logs)
}

func (p *provider) logMSTagsTree(r *http.Request) interface{} {
	return api.Success(p.GetTagsTree("micro_service", api.Language(r)))
}
func (p *provider) orgLogTagsTree(r *http.Request) interface{} {
	return api.Success(p.GetTagsTree("org", api.Language(r)))
}

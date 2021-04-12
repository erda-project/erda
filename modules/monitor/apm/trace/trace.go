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

package trace

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
)

func (p *provider) traceDebugRecords(r *http.Request, params struct {
	ScopeId string `param:"scopeId" validate:"required"`
	Offset  int64  `query:"offset"`
	Limit   int64  `query:"limit" default:"20"`
}) interface{} {

	if params.Limit <= 0 {
		params.Limit = 20
	} else if params.Limit > 200 {
		params.Limit = 200
	}

	if params.Offset == 0 {
		params.Offset = time.Now().Unix()
	}

	return nil
}

func (p *provider) traceOne(r *http.Request, params struct {
	TraceId string `param:"traceId" validate:"required"`
	ScopeId string `param:"scopeId" validate:"required"`
	Limit   int64  `query:"limit" default:"1000"`
}) interface{} {
	spans := p.spanq.SelectSpans(params.TraceId, params.Limit)
	return api.Success(spans)
}

func (p *provider) traces(r *http.Request, params struct {
	ScopeId       string `param:"scopeId" validate:"required"`
	ApplicationId int64  `query:"applicationId" default:"-1"`
	Status        int    `query:"status" default:"0"` // -1 error, 0 both, 1 success
	Start         int64  `query:"start"`
	End           int64  `query:"end"`
	Limit         int64  `query:"limit" default:"20"`
}) interface{} {

	if params.End == 0 {
		params.End = time.Now().UnixNano() / 1e6
		h, _ := time.ParseDuration("-1h")
		params.Start = time.Now().Add(h).UnixNano() / 1e6
	}
	options := url.Values{}
	options.Set("start", strconv.FormatInt(params.Start, 10))
	options.Set("end", strconv.FormatInt(params.End, 10))

	metricParams := make(map[string]interface{})
	metricParams["terminus_key"] = params.ScopeId
	metricParams["limit"] = strconv.FormatInt(params.Limit, 10)
	var where bytes.Buffer
	if params.ApplicationId != -1 {
		metricParams["applications_ids"] = strconv.FormatInt(params.ApplicationId, 10)
		where.WriteString("applications_ids::field=$applications_ids AND ")
	}
	if params.Status != 0 {
		if params.Status == -1 {
			metricParams["errors_sum"] = strconv.Itoa(params.Status)
			where.WriteString(" errors_sum::field=$errors_sum AND ")
		} else if params.Status == 1 {
			metricParams["errors_sum"] = strconv.Itoa(params.Status)
			where.WriteString(" errors_sum::field>$errors_sum AND ")
		}
	}
	str := fmt.Sprintf("SELECT sum(errors_sum),min(start_time_min),max(end_time_max),last(labels_distinct),"+
		"trace_id::tag FROM trace WHERE terminus_key::tag=$terminus_key GROUP BY trace_id::tag ORDER BY "+
		"max(start_time_min::field) LIMIT %s", strconv.FormatInt(params.Limit, 10))
	response, err := p.metricq.Query(
		metricq.InfluxQL,
		str,
		metricParams,
		options)
	if err != nil {
		return api.Errors.Internal(err)
	}

	result := make([]map[string]interface{}, 0, params.Limit)
	for _, r := range response.ResultSet.Rows {
		itemResult := make(map[string]interface{})
		itemResult["startTime"] = r[1].(float64) / 10000
		itemResult["elapsed"] = math.Abs(r[1].(float64) - r[2].(float64))
		itemResult["tracesId"] = r[4]
		itemResult["labels"] = r[3]
		result = append(result, itemResult)
	}

	return api.Success(result)
}

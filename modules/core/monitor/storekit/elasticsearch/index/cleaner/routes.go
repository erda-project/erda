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

package cleaner

import (
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

func (p *provider) intRoutes(routes httpserver.Router, prefix string) error {
	routes.GET(prefix+"/cleaner/rollover-body", p.getRolloverBody)
	routes.GET(prefix+"/cleaner/clean-expired", p.cleanExpiredIndices)
	routes.GET(prefix+"/cleaner/clean-by-disk", p.cleanByDiskUsage)
	routes.GET(prefix+"/cleaner/is-leader", p.checkIsLeader)
	return nil
}

func (p *provider) getRolloverBody(r *http.Request) interface{} {
	return p.rolloverBodyForDiskClean
}

func (p *provider) checkIsLeader(r *http.Request) interface{} {
	return p.election.IsLeader()
}

func (p *provider) cleanExpiredIndices(r *http.Request, params struct {
	TimeOffset string `query:"timeOffset"`
}) interface{} {
	var duration time.Duration
	if len(params.TimeOffset) > 0 {
		d, err := time.ParseDuration(params.TimeOffset)
		if err != nil {
			return err
		}
		duration = d
	}
	err := p.checkAndDeleteIndices(r.Context(), time.Now().Add(duration), func(*loader.IndexEntry) bool { return true })
	if err != nil {
		return err
	}
	return true
}

func (p *provider) cleanByDiskUsage(r *http.Request, params struct {
	TargetUsagePercent     float64 `query:"targetPercent"`
	ThresholdPercent       float64 `query:"thresholdPercent"`
	MinIndicesStorePercent float64 `query:"minIndicesStorePercent"`
}) interface{} {
	config := p.Cfg.DiskClean //deep copy
	if params.TargetUsagePercent > 0 {
		config.LowDiskUsagePercent = params.TargetUsagePercent
	}
	if params.ThresholdPercent > 0 {
		config.HighDiskUsagePercent = params.ThresholdPercent
	}
	if params.MinIndicesStorePercent > 0 {
		config.MinIndicesStorePercent = params.MinIndicesStorePercent
	}
	err := p.checkDiskUsage(r.Context(), config)
	if err != nil {
		return err
	}
	return true
}

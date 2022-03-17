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

package resourcegc

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

const bufferTime = 3600

// sometimes the pipeline is in downtime or restart time
// then the etcd lease of gc may expire at this time
// and then there is no instance get lease, which results in some namespaces pod not being gc
func (p *provider) compensateGCNamespaces(ctx context.Context) {
	p.doWaitGCCompensate(false)
	p.doWaitGCCompensate(true)

	ticker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ticker.C:
			p.doWaitGCCompensate(false)
			p.doWaitGCCompensate(true)
		case <-ctx.Done():
			return
		}
	}
}

func (p *provider) doWaitGCCompensate(isSnippetPipeline bool) {
	var pageNum = 1
	for {
		needGCPipelines, total, err := p.getNeedGCPipelines(pageNum, isSnippetPipeline)
		pageNum += 1
		if err != nil {
			p.errorf("getNeedGcPipelines error: %v", err)
			continue
		} else {
			if total <= 0 {
				break
			}

			for _, pt := range needGCPipelines {
				// random pageSize to disperse query pressure
				p.WaitGC(pt.Extra.Namespace, pt.ID, uint64(rand.Intn(1000)))
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func (p *provider) getNeedGCPipelines(pageNum int, isSnippet bool) ([]spec.Pipeline, int, error) {
	var pipelineResults []spec.Pipeline

	var req apistructs.PipelinePageListRequest
	req.PageNum = pageNum
	req.PageSize = 100
	req.LargePageSize = true
	req.AscCols = []string{"id"}
	for _, end := range apistructs.PipelineEndStatuses {
		req.Statuses = append(req.Statuses, end.String())
	}
	req.AllSources = true
	req.IncludeSnippet = isSnippet
	pipelines, _, _, _, err := p.dbClient.PageListPipelines(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to compensate pipeline req %v err: %v", req, err)
	} else {
		for _, pt := range pipelines {
			if !pt.Status.IsEndStatus() {
				continue
			}

			if pt.Extra.CompleteReconcilerGC || pt.Extra.CompleteReconcilerTeardown {
				continue
			}

			ttl := pt.GetResourceGCTTL()
			if ttl <= 0 {
				ttl = defaultGCTime
			}

			var endTime = pt.TimeEnd
			if endTime == nil {
				endTime = pt.TimeUpdated
			}

			if endTime == nil || endTime.IsZero() {
				continue
			}

			if uint64(time.Now().Unix()-endTime.Unix()) < (ttl + bufferTime) {
				continue
			}

			pipelineResults = append(pipelineResults, pt)
		}
	}
	return pipelineResults, len(pipelines), nil
}

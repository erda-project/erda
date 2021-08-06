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

package reconciler

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

const bufferTime = 3600

// sometimes the pipeline is in downtime or restart time
// then the etcd lease of gc may expire at this time
// and then there is no instance get lease, which results in some namespaces pod not being gc
func (r *Reconciler) CompensateGCNamespaces() {
	r.doWaitGCCompensate(false)
	r.doWaitGCCompensate(true)

	time.AfterFunc(24*time.Hour, func() {
		r.CompensateGCNamespaces()
	})
}

func (r *Reconciler) doWaitGCCompensate(isSnippetPipeline bool) {
	var pageNum = 1
	for {
		needGCPipelines, total, err := r.getNeedGCPipelines(pageNum, isSnippetPipeline)
		pageNum += 1
		if err != nil {
			logrus.Errorf("getNeedGcPipelines error: %v", err)
			continue
		} else {
			if total <= 0 {
				break
			}

			for _, p := range needGCPipelines {
				// random pageSize to disperse query pressure
				r.waitGC(p.Extra.Namespace, p.ID, uint64(rand.Intn(1000)))
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func (r *Reconciler) getNeedGCPipelines(pageNum int, isSnippet bool) ([]spec.Pipeline, int, error) {
	var pipelineResults []spec.Pipeline

	var req apistructs.PipelinePageListRequest
	req.PageNum = pageNum
	req.PageSize = 1000
	req.LargePageSize = true
	req.AscCols = []string{"id"}
	for _, end := range apistructs.PipelineEndStatuses {
		req.Statuses = append(req.Statuses, end.String())
	}
	req.AllSources = true
	req.IncludeSnippet = isSnippet
	pipelines, _, _, _, err := r.dbClient.PageListPipelines(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to compensate pipeline req %v err: %v", req, err)
	} else {
		for _, p := range pipelines {
			if !p.Status.IsEndStatus() {
				continue
			}

			if p.Extra.CompleteReconcilerGC || p.Extra.CompleteReconcilerTeardown {
				continue
			}

			ttl := p.GetResourceGCTTL()
			if ttl <= 0 {
				ttl = defaultGCTime
			}

			var endTime = p.TimeEnd
			if endTime == nil {
				endTime = p.TimeUpdated
			}

			if endTime == nil || endTime.IsZero() {
				continue
			}

			if uint64(time.Now().Unix()-endTime.Unix()) < (ttl + bufferTime) {
				continue
			}

			pipelineResults = append(pipelineResults, p)
		}
	}
	return pipelineResults, len(pipelines), nil
}

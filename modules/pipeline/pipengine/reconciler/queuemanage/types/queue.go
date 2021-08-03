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

package types

import (
	"github.com/erda-project/erda-proto-go/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type Queue interface {
	QueueValidator
	Start(stopCh chan struct{})
	ID() string
	IsStrictMode() bool
	Usage() pb.QueueUsage
	Update(pq *apistructs.PipelineQueue)
	RangePendingQueue()
	AddPipelineIntoQueue(p *spec.Pipeline, doneCh chan struct{})
	PopOutPipeline(p *spec.Pipeline)
	BatchUpdatePipelinePriorityInQueue(pipelines []*spec.Pipeline) error
}

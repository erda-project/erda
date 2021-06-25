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

package flink

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var Kind = types.Kind("flink")

type Flink struct {
	name      types.Name
	addr      string
	options   map[string]string
	client    *httpclient.HTTPClient
	enableTag bool
	cluster   apistructs.ClusterInfo
}

func (f *Flink) Kind() types.Kind {
	return Kind
}

func (f *Flink) Name() types.Name {
	return f.name
}

type FlinkCreateRequest struct {
	EntryClass  string `json:"entryClass"`
	ProgramArgs string `json:"programArgs"`
}

type FlinkCreateResponse struct {
	JobId string `json:"jobid"`
}

type FlinkGetResponse struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	StartTime   int64  `json:"start-time"`
	CurrentTime int64  `json:"now"`
	EndTime     int64  `json:"end-time"`
}

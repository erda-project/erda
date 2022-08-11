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

package details_apis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda-proto-go/core/monitor/alertdetail/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type alertDetailService struct {
	p                *provider
	ClickhouseSource ClickhouseSource
}

func (a *alertDetailService) QuerySystemPodMetrics(ctx context.Context, request *pb.QuerySystemPodMetricsRequest) (*pb.QuerySystemPodMetricsResponse, error) {
	start := time.Now().Add(-5 * time.Minute).UnixMilli()
	end := time.Now().UnixMilli()
	if request.Timestamp != 0 {
		start = request.Timestamp - 30*60*1000
		end = request.Timestamp + 30*60*1000
	}
	orgIDStr := apis.GetOrgID(ctx)
	if orgIDStr == "" {
		return nil, fmt.Errorf("orgID can not be empty")
	}
	var (
		pod *PodInfo
		err error
	)
	if !a.p.C.QueryMetricFromCk {
		pod, err = a.p.getPodInfo(request.ClusterName, request.Name, start, end)
	} else {
		pod, err = a.ClickhouseSource.GetPodInfo(ctx, request.ClusterName, request.Name, start, end)
	}
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err := json.Marshal(pod)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QuerySystemPodMetricsResponse{
		Data: &pb.PodInfo{
			Summary:   nil,
			Instances: make([]*pb.PodInfoInstanse, 0),
		},
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

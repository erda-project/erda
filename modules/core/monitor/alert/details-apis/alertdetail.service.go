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

package details_apis

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-proto-go/core/monitor/alertdetail/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type alertDetailService struct {
	p *provider
}

func (a *alertDetailService) QuerySystemPodMetrics(ctx context.Context, request *pb.QuerySystemPodMetricsRequest) (*pb.QuerySystemPodMetricsResponse, error) {
	start := request.Timestamp - 30*60*1000
	end := request.Timestamp + 30*60*1000
	pod, err := a.p.getPodInfo(request.ClusterName, request.Name, start, end)
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

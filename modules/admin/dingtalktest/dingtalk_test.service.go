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

package dingtalktest

import (
	"context"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)
import pb "github.com/erda-project/erda-proto-go/admin/pb"

type dingTalkTestService struct {
	Log logs.Logger
	bdl *bundle.Bundle
}

func (s *dingTalkTestService) SendTestMessage(ctx context.Context, request *pb.DingTalkTestRequest) (*pb.DingTalkTestResponse, error) {
	eventRequest := &apistructs.EventBoxRequest{
		Sender:  "admin",
		Content: "Hello, this is a test message.",
		Labels: map[string]interface{}{
			"DINGDING": []apistructs.Target{
				{
					Receiver: request.GetWebhook(),
					Secret:   request.GetSecret(),
				},
			},
		},
	}
	err := s.bdl.CreateEventNotify(eventRequest)
	response := &pb.DingTalkTestResponse{
		Success: true,
	}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	}
	return response, nil
}

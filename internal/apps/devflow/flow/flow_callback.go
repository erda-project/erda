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

package flow

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/apps/devflow/flow/pb"
	rulepb "github.com/erda-project/erda-proto-go/dop/rule/pb"
)

func (s *Service) FlowCallBack(ctx context.Context, req *pb.FlowCallbackRequest) (*pb.FlowCallbackResponse, error) {
	params := req.Content.Params
	s.p.Log.Debugf("devflow callback content params %v, req %v", params, req)
	marshal, err := json.Marshal(req.Content)
	if err != nil {
		return nil, err
	}
	flowContent := make(map[string]interface{})
	if err := json.Unmarshal(marshal, &flowContent); err != nil {
		return nil, err
	}
	env, err := structpb.NewValue(flowContent)
	if err != nil {
		return nil, err
	}
	eventType := "dev_flow"
	res, err := s.p.RuleExecutor.Fire(ctx, &rulepb.FireRequest{
		Scope:     "project",
		ScopeID:   req.ProjectID,
		EventType: eventType,
		Env: map[string]*structpb.Value{
			eventType: env,
		},
	})
	if err != nil {
		s.p.Log.Errorf("devflow fire rule err %v", err)
		return nil, err
	}
	s.p.Log.Infof("%v rule executor result %v", eventType, res)
	return &pb.FlowCallbackResponse{}, nil
}

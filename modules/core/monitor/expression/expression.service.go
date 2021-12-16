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

package expression

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda-proto-go/core/monitor/expression/pb"
)

type expressionService struct {
	p *provider
}

func (e *expressionService) GetAllAlertExpression(ctx context.Context, request *pb.GetAllAlertExpressionRequest) (*pb.GetAllAlertExpressionResponse, error) {
	expressions, err := e.p.alertDB.GetAllAlertExpression()
	if err != nil {
		return nil, err
	}
	result := &pb.GetAllAlertExpressionResponse{
		Data: make([]*pb.AlertExpression, len(expressions)),
	}
	data, err := json.Marshal(expressions)
	if err != nil {
		return result, err
	}
	fmt.Println(string(data))
	err = json.Unmarshal(data, &result.Data)
	if err != nil {
		return result, err
	}
	return result, nil
}

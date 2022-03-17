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

package eventbox

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda-proto-go/core/messenger/eventbox/pb"
	inputhttp "github.com/erda-project/erda/modules/messenger/eventbox/input/http"
	"github.com/erda-project/erda/pkg/common/errors"
)

type eventBoxService struct {
	HttpI *inputhttp.HttpInput
}

func (e *eventBoxService) CreateMessage(ctx context.Context, request *pb.CreateMessageRequest) (*pb.CreateMessageResponse, error) {
	resp, err := e.HttpI.CreateMessage(ctx, request, nil)
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	var httpResp pb.HTTPResponse
	err = json.Unmarshal(data, &httpResp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.CreateMessageResponse{
		Data: &httpResp,
	}, nil
}

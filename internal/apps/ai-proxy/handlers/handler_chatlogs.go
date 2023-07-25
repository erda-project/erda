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

package handlers

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ChatLogsHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (c *ChatLogsHandler) GetChatLogs(ctx context.Context, req *pb.GetChatLogsReq) (*pb.GetChatLogsRespData, error) {
	_, ok := getUserId(ctx, req)
	if !ok {
		return nil, UserPermissionDenied
	}
	// todo: validate userId

	sessionId, ok := getSessionId(ctx, req)
	if !ok {
		return nil, InvalidSessionId
	}
	if req.GetPageSize() == 0 {
		req.PageSize = 10
	}
	if req.GetPageNum() == 0 {
		req.PageNum = 1
	}
	total, chatLogs, err := c.Dao.PagingChatLogs(sessionId, int(req.GetPageNum()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	return &pb.GetChatLogsRespData{
		Total: uint64(total),
		List:  chatLogs,
	}, nil
}

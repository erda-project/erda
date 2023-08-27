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

package handler_session

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type SessionHandler struct {
	DAO dao.DAO
}

func (h *SessionHandler) Create(ctx context.Context, req *pb.SessionCreateRequest) (*pb.Session, error) {
	return h.DAO.SessionClient().Create(ctx, req)
}

func (h *SessionHandler) Get(ctx context.Context, req *pb.SessionGetRequest) (*pb.Session, error) {
	return h.DAO.SessionClient().Get(ctx, req)
}

func (h *SessionHandler) Delete(ctx context.Context, req *pb.SessionDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.SessionClient().Delete(ctx, req)
}

func (h *SessionHandler) Update(ctx context.Context, req *pb.SessionUpdateRequest) (*pb.Session, error) {
	return h.DAO.SessionClient().Update(ctx, req)
}

func (h *SessionHandler) Archive(ctx context.Context, req *pb.SessionArchiveRequest) (*pb.Session, error) {
	return h.DAO.SessionClient().Archive(ctx, req)
}

func (h *SessionHandler) UnArchive(ctx context.Context, req *pb.SessionUnArchiveRequest) (*pb.Session, error) {
	return h.DAO.SessionClient().UnArchive(ctx, req)
}

func (h *SessionHandler) Reset(ctx context.Context, req *pb.SessionResetRequest) (*pb.Session, error) {
	return h.DAO.SessionClient().Reset(ctx, req)
}

func (h *SessionHandler) GetChatLogs(ctx context.Context, req *pb.SessionChatLogGetRequest) (*pb.SessionChatLogGetResponse, error) {
	return h.DAO.SessionClient().GetChatLogs(ctx, req)
}

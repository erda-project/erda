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
	"database/sql"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/common/apis"
)

type SessionsHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (s *SessionsHandler) CreateSession(ctx context.Context, req *pb.Session) (*pb.CreateSessionRespData, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user

	if req.GetName() == "" {
		return nil, InvalidSessionName
	}
	if req.GetTopic() == "" {
		return nil, InvalidSessionTopic
	}
	if req.GetContextLength() > 20 {
		req.ContextLength = 20
	}
	if req.GetSource() == "" {
		return nil, InvalidSessionSource
	}
	if req.GetModel() == "" {
		return nil, InvalidSessionModel
	}
	if req.GetTemperature() < 0 {
		req.Temperature = 0
	}
	if req.GetTemperature() > 2 {
		req.Temperature = 2
	}

	id, err := s.Dao.CreateSession(userId, req.GetName(), req.GetTopic(), req.GetContextLength(), req.GetSource(), req.GetModel(), req.GetTemperature())
	if err != nil {
		return nil, err
	}
	return &pb.CreateSessionRespData{
		Id: id,
	}, nil
}

func (s *SessionsHandler) UpdateSession(ctx context.Context, req *pb.Session) (*common.VoidResponse, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user

	if req.GetName() == "" {
		return nil, InvalidSessionName
	}
	if req.GetTopic() == "" {
		return nil, InvalidSessionTopic
	}
	if req.GetContextLength() > 20 {
		req.ContextLength = 20
	}
	if req.GetModel() == "" {
		return nil, InvalidSessionModel
	}
	if req.GetTemperature() < 0 {
		req.Temperature = 0
	}
	if req.GetTemperature() > 2 {
		req.Temperature = 2
	}

	var updates = map[string]interface{}{
		"name":           req.GetName(),
		"topic":          req.GetTopic(),
		"context_length": req.GetContextLength(),
		"is_archived":    req.GetIsArchived(),
		"model":          req.GetModel(),
		"temperature":    req.GetTemperature(),
	}
	if req.GetResetAt() != nil {
		updates["reset_at"] = sql.NullTime{Time: req.GetResetAt().AsTime(), Valid: true}
	}
	if err := s.Dao.UpdateSession(req.GetId(), updates); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) ResetSession(ctx context.Context, req *pb.ResetSessionReq) (*common.VoidResponse, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	if req.GetId() == "" {
		return nil, InvalidSessionId
	}
	if req.GetResetAt() == nil {
		return nil, InvalidSessionResetAt
	}
	var updates = map[string]interface{}{
		"reset_at": req.GetResetAt(),
	}
	if err := s.Dao.UpdateSession(req.GetId(), updates); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) ArchiveSession(ctx context.Context, req *pb.ArchiveSessionReq) (*common.VoidResponse, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	if req.GetId() == "" {
		return nil, InvalidSessionId
	}
	var updates = map[string]interface{}{
		"is_archived": req.GetIsArchived(),
	}
	if err := s.Dao.UpdateSession(req.GetId(), updates); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) DeleteSession(ctx context.Context, req *pb.LocateSessionCondition) (*common.VoidResponse, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	if req.GetId() == "" {
		return nil, InvalidSessionId
	}
	if err := s.Dao.DeleteSession(req.GetId()); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) ListSessions(ctx context.Context, req *pb.ListSessionsReq) (*pb.ListSessionsRespData, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	var where = map[string]any{"user_id": userId}
	var source = req.GetSource()
	if source == "" {
		source = apis.GetHeader(ctx, vars.XErdaAIProxySource)
	}
	if source != "" {
		where["source"] = source
	}
	total, sessions, err := s.Dao.ListSessions(where)
	if err != nil {
		return nil, err
	}
	return &pb.ListSessionsRespData{
		Total: uint64(total),
		List:  sessions,
	}, nil
}

func (s *SessionsHandler) GetSession(ctx context.Context, req *pb.LocateSessionCondition) (*pb.Session, error) {
	var userId = req.GetUserId()
	if userId == "" {
		userId = apis.GetUserID(ctx)
	}
	if userId == "" {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	if req.GetId() == "" {
		return nil, InvalidSessionId
	}
	return s.Dao.GetSession(req.GetId())
}

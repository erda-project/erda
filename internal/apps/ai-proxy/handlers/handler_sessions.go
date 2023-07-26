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
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

type SessionsHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (s *SessionsHandler) CreateSession(ctx context.Context, req *pb.Session) (*pb.CreateSessionRespData, error) {
	userId, ok := getUserId(ctx, req)
	if !ok {
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
	// todo: hard code yet
	if req.GetSource() != "erda.cloud" {
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
	return &pb.CreateSessionRespData{Id: id}, nil
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

	var session models.AIProxySessions
	var setters = []models.Setter{
		session.FieldName().Set(req.GetName()),
		session.FieldTopic().Set(req.GetTopic()),
		session.FieldContextLength().Set(req.GetContextLength()),
		session.FieldIsArchived().Set(req.GetIsArchived()),
		session.FieldModel().Set(req.GetModel()),
		session.FieldTemperature().Set(req.GetTemperature()),
	}
	if req.GetResetAt() != nil {
		setters = append(setters, session.FieldResetAt().Set(req.GetResetAt().AsTime()))
	}
	if err := s.Dao.UpdateSession(req.GetId(), setters...); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) ResetSession(ctx context.Context, req *pb.LocateSessionCondition) (*common.VoidResponse, error) {
	_, ok := getUserId(ctx, req)
	if !ok {
		return nil, UserPermissionDenied
	}
	// todo: validate user

	if req.GetId() == "" {
		return nil, InvalidSessionId
	}

	if err := s.Dao.UpdateSession(req.GetId(), new(models.AIProxySessions).FieldResetAt().Set(time.Now())); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) ArchiveSession(ctx context.Context, req *pb.LocateSessionCondition) (*common.VoidResponse, error) {
	_, ok := getUserId(ctx, req)
	if !ok {
		return nil, UserPermissionDenied
	}
	// todo: validate user

	if req.GetId() == "" {
		return nil, InvalidSessionId
	}

	if err := s.Dao.UpdateSession(req.GetId(), new(models.AIProxySessions).FieldIsArchived().Set(true)); err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) DeleteSession(ctx context.Context, req *pb.LocateSessionCondition) (*common.VoidResponse, error) {
	_, ok := getUserId(ctx, req)
	if !ok {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	if req.GetId() == "" {
		return nil, InvalidSessionId
	}
	_, err := new(models.AIProxySessions).Deleter(s.Dao.Q()).Delete()
	if err != nil {
		return nil, err
	}
	return &common.VoidResponse{}, nil
}

func (s *SessionsHandler) ListSessions(ctx context.Context, req *pb.ListSessionsReq) (*pb.ListSessionsRespData, error) {
	// try to get userId
	userId, ok := getUserId(ctx, req)
	if !ok {
		return nil, UserPermissionDenied
	}
	// todo: validate user

	// try to get request source
	source, ok := getSource(ctx, req)
	if !ok {
		return nil, InvalidSessionSource
	}

	var sessions models.AIProxySessionsList
	total, err := (&sessions).Pager(s.Dao.Q()).
		Where(
			sessions.FieldUserID().Equal(userId),
			sessions.FieldSource().Equal(source),
		).
		Paging(20, 1, sessions.FieldUpdatedAt().DESC())
	if err != nil {
		return nil, err
	}
	return &pb.ListSessionsRespData{
		Total: uint64(total),
		List:  sessions.ToProtobuf(),
	}, nil
}

func (s *SessionsHandler) GetSession(ctx context.Context, req *pb.LocateSessionCondition) (*pb.Session, error) {
	_, ok := getUserId(ctx, req)
	if !ok {
		return nil, UserPermissionDenied
	}
	// todo: validate user
	if req.GetId() == "" {
		return nil, InvalidSessionId
	}
	session, ok, err := s.Dao.GetSession(req.GetId())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, new(errorresp.APIError).NotFound()
	}
	return session, nil
}

func getUserId(ctx context.Context, req interface{ GetUserId() string }) (string, bool) {
	if userId := req.GetUserId(); userId != "" {
		return userId, true
	}
	if userId := apis.GetUserID(ctx); userId != "" {
		return userId, true
	}
	if userId := apis.GetHeader(ctx, vars.XAIProxyUserId); userId != "" {
		return userId, true
	}
	return "", false
}

func getSource(ctx context.Context, req interface{ GetSource() string }) (string, bool) {
	if source := req.GetSource(); source != "" {
		return source, true
	}
	if source := apis.GetHeader(ctx, vars.XAIProxySource); source != "" {
		return source, true
	}
	return "", false
}

func getSessionId(ctx context.Context, req interface{ GetSessionId() string }) (string, bool) {
	if sessionId := req.GetSessionId(); sessionId != "" {
		return sessionId, true
	}
	sessionId := apis.GetHeader(ctx, vars.XAIProxySessionId)
	return sessionId, sessionId != ""
}

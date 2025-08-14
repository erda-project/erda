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

package handler_mcp_server

import (
	"context"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/cache"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type MCPHandler struct {
	DAO      dao.DAO
	listener *cache.EventListener
}

func NewMCPHandler(d dao.DAO) *MCPHandler {
	handler := MCPHandler{
		DAO:      d,
		listener: cache.NewEventListener(),
	}
	taskFunc := func() {
		logrus.Infof("will sync mcp server from database to cache")

		mcpServers, err := handler.DAO.MCPServerClient().ListAll(context.Background(), false)
		if err != nil {
			logrus.Errorf("mcp server list error: %v", err)
			return
		}

		// clean cache
		handler.listener.Send(context.Background(), &cache.Event{
			Method: cache.EventMethodClear,
		})
		time.Sleep(1 * time.Second)

		for _, server := range mcpServers {
			logrus.Infof("load mcp server %+v, endpoint: %+v", server.Name, server.Endpoint)
			handler.listener.Send(context.Background(), &cache.Event{
				Method: cache.EventMethodSet,
				Name:   server.Name,
				Tag:    server.Version,
				Data: &cache.McpServerInfo{
					Endpoint:      server.Endpoint,
					Tag:           server.Version,
					Name:          server.Name,
					Version:       server.Version,
					TransportType: server.TransportType,
				},
			})
		}
		logrus.Infof("sync mcp server done")
	}
	// first time run task
	taskFunc()

	go func() {
		// run task every 1:00 am
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, now.Location())

		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		duration := next.Sub(now)
		logrus.Infof("sync task will run after: %+v", duration)

		time.AfterFunc(duration, func() {
			ticker := time.NewTicker(24 * time.Hour)
			for range ticker.C {
				taskFunc()
			}
		})
	}()

	handler.listener.OnEvent(context.Background())
	return &handler
}

func (m *MCPHandler) Update(ctx context.Context, req *pb.MCPServerUpdateRequest) (*pb.MCPServerUpdateResponse, error) {
	return m.DAO.MCPServerClient().Update(ctx, req)
}

func (m *MCPHandler) Delete(ctx context.Context, req *pb.MCPServerDeleteRequest) (*pb.MCPServerDeleteResponse, error) {
	return m.DAO.MCPServerClient().Delete(ctx, req)
}

func (m *MCPHandler) Version(ctx context.Context, req *pb.MCPServerVersionRequest) (*pb.MCPServerVersionResponse, error) {
	total, servers, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
		PageNum:            int(req.PageNum),
		PageSize:           int(req.PageSize),
		Name:               req.Name,
		IncludeUnpublished: req.IncludeUnpublished,
	})
	if err != nil {
		return nil, err
	}

	return &pb.MCPServerVersionResponse{
		Total: total,
		List:  servers,
	}, nil
}

func (m *MCPHandler) Register(ctx context.Context, req *pb.MCPServerRegisterRequest) (*pb.MCPServerRegisterResponse, error) {
	return m.DAO.MCPServerClient().CreateOrUpdate(ctx, req)
}

func (m *MCPHandler) Publish(ctx context.Context, req *pb.MCPServerActionPublishRequest) (*pb.MCPServerActionPublishResponse, error) {
	return m.DAO.MCPServerClient().Publish(ctx, req)
}

func (m *MCPHandler) Get(ctx context.Context, req *pb.MCPServerGetRequest) (*pb.MCPServerGetResponse, error) {
	return m.DAO.MCPServerClient().Get(ctx, req)
}

func (m *MCPHandler) List(ctx context.Context, req *pb.MCPServerListRequest) (*pb.MCPServerListResponse, error) {
	total, servers, err := m.DAO.MCPServerClient().List(ctx, &mcp_server.ListOptions{
		IncludeUnpublished: req.IncludeUnpublished,
		PageNum:            int(req.PageNum),
		PageSize:           int(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	return &pb.MCPServerListResponse{
		Total: total,
		List:  servers,
	}, nil
}

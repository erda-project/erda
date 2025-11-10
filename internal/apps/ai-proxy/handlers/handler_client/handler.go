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

package handler_client

import (
	"context"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type ClientHandler struct {
	DAO dao.DAO
}

func (h *ClientHandler) Create(ctx context.Context, req *pb.ClientCreateRequest) (*pb.Client, error) {
	return h.DAO.ClientClient().Create(ctx, req)
}

func (h *ClientHandler) Get(ctx context.Context, req *pb.ClientGetRequest) (*pb.Client, error) {
	client, err := h.DAO.ClientClient().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	desensitizeClient(ctx, client)
	return client, nil
}

func (h *ClientHandler) Delete(ctx context.Context, req *pb.ClientDeleteRequest) (*commonpb.VoidResponse, error) {
	return h.DAO.ClientClient().Delete(ctx, req)
}

func (h *ClientHandler) Update(ctx context.Context, req *pb.ClientUpdateRequest) (*pb.Client, error) {
	return h.DAO.ClientClient().Update(ctx, req)
}

func (h *ClientHandler) Paging(ctx context.Context, req *pb.ClientPagingRequest) (*pb.ClientPagingResponse, error) {
	resp, err := h.DAO.ClientClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	// data sensitive
	for _, item := range resp.List {
		desensitizeClient(ctx, item)
	}
	return resp, nil
}

func desensitizeClient(ctx context.Context, item *pb.Client) {
	// pass for: admin
	if auth.IsAdmin(ctx) {
		return
	}
	// pass for: the client itself
	if auth.IsClient(ctx) {
		authed := auth.GetClient(ctx)
		if authed != nil && authed.Id == item.Id {
			return
		}
	}
	// hide sensitive data
	item.AccessKeyId = ""
	item.SecretKeyId = ""
	if item.Metadata != nil {
		item.Metadata.Secret = nil
	}
}

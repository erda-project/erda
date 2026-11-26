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

package token

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	tokenstore "github.com/erda-project/erda/pkg/oauth2/tokenstore/mysqltokenstore"
	"github.com/erda-project/erda/pkg/secret"
)

type TokenService struct {
	logger logs.Logger

	dao             Dao
	sysAdminChecker sysAdminChecker
}

type sysAdminChecker interface {
	IsSysAdmin(userID string) (bool, error)
}

func (s *TokenService) QueryTokens(ctx context.Context, req *pb.QueryTokensRequest) (*pb.QueryTokensResponse, error) {
	reveal, err := s.prepareQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	objs, total, err := s.dao.QueryToken(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.Token, len(objs))
	for i, obj := range objs {
		res[i] = obj.ToPbToken()
		if !reveal {
			maskToken(res[i])
		}
	}
	return &pb.QueryTokensResponse{Data: res, Total: total}, nil
}

func (s *TokenService) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	obj, err := s.dao.GetToken(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeTokenAccess(ctx, obj.CreatorID); err != nil {
		return nil, err
	}

	res := obj.ToPbToken()
	if !s.canReveal(ctx) {
		maskToken(res)
	}
	return &pb.GetTokenResponse{Data: res}, nil
}

func (s *TokenService) prepareQuery(ctx context.Context, req *pb.QueryTokensRequest) (bool, error) {
	// Exact access-key lookups from authenticated internal callers must see the real value.
	if req.Access != "" && (apis.IsInternalClient(ctx) || apis.GetUserID(ctx) == "") {
		return true, nil
	}

	hasFilter := req.Access != "" || req.CreatorId != "" || (req.Scope != "" && req.ScopeId != "")

	if apis.IsInternalClient(ctx) {
		if !hasFilter {
			return false, fmt.Errorf("token query requires at least one explicit filter")
		}
		return true, nil
	}

	userID := apis.GetUserID(ctx)
	if userID != "" {
		admin, err := s.isSysAdmin(userID)
		if err != nil {
			return false, err
		}
		if admin {
			return false, nil
		}
		if req.CreatorId != "" && req.CreatorId != userID {
			return false, fmt.Errorf("cannot query tokens created by another user")
		}
		req.CreatorId = userID
		return false, nil
	}

	// Calls without an identity are in-process service calls. They still need an explicit filter.
	if !hasFilter {
		return false, fmt.Errorf("token query requires at least one explicit filter")
	}
	return true, nil
}

func (s *TokenService) canReveal(ctx context.Context) bool {
	// External callers always receive masked tokens. Internal and in-process callers may need
	// the real value for authentication or for writing it into a one-time credential response.
	return apis.IsInternalClient(ctx) || apis.GetUserID(ctx) == ""
}

func (s *TokenService) isSysAdmin(userID string) (bool, error) {
	if s.sysAdminChecker == nil {
		return false, nil
	}
	return s.sysAdminChecker.IsSysAdmin(userID)
}

func (s *TokenService) authorizeTokenID(ctx context.Context, id string) error {
	if apis.IsInternalClient(ctx) || apis.GetUserID(ctx) == "" {
		return nil
	}
	obj, err := s.dao.GetToken(ctx, &pb.GetTokenRequest{Id: id})
	if err != nil {
		return err
	}
	return s.authorizeTokenAccess(ctx, obj.CreatorID)
}

func (s *TokenService) authorizeTokenAccess(ctx context.Context, creatorID string) error {
	if apis.IsInternalClient(ctx) {
		return nil
	}
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil
	}
	admin, err := s.isSysAdmin(userID)
	if err != nil {
		return err
	}
	if admin || creatorID == userID {
		return nil
	}
	return fmt.Errorf("token not found")
}

func maskToken(token *pb.Token) {
	if token == nil {
		return
	}
	token.AccessKey = maskTokenValue(token.AccessKey)
	token.SecretKey = ""
}

func maskTokenValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "********"
	}
	return value[:4] + "********" + value[len(value)-4:]
}

func (s *TokenService) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	userID := apis.GetUserID(ctx)
	token, err := ToModelToken(userID, req)
	if err != nil {
		return nil, err
	}
	obj, err := s.dao.CreateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &pb.CreateTokenResponse{Data: obj.ToPbToken()}, nil
}

func (s *TokenService) UpdateToken(ctx context.Context, req *pb.UpdateTokenRequest) (*pb.UpdateTokenResponse, error) {
	if err := s.authorizeTokenID(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.UpdateTokenResponse{}, s.dao.UpdateToken(ctx, req)
}

func (s *TokenService) DeleteToken(ctx context.Context, req *pb.DeleteTokenRequest) (*pb.DeleteTokenResponse, error) {
	if err := s.authorizeTokenID(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteTokenResponse{}, s.dao.DeleteToken(ctx, req)
}

func ToModelToken(userID string, req *pb.CreateTokenRequest) (tokenstore.TokenStoreItem, error) {
	tokenType := req.Type
	switch tokenType {
	case tokenstore.AccessKey.String():
		pair := secret.CreateAkSkPair()
		return tokenstore.TokenStoreItem{
			AccessKey:   pair.AccessKeyID,
			SecretKey:   pair.SecretKey,
			Description: req.Description,
			Scope:       req.Scope,
			ScopeId:     req.ScopeId,
			CreatorID:   req.CreatorId,
			Type:        req.Type,
		}, nil
	case tokenstore.PAT.String():
		if userID == "" {
			if req.Scope == string(apistructs.OrgScope) {
				userID = req.CreatorId
			} else {
				return tokenstore.TokenStoreItem{}, fmt.Errorf("user id is empty")
			}
		}
		return tokenstore.TokenStoreItem{
			AccessKey:   uuid.UUID(),
			Description: req.Description,
			Scope:       req.Scope,
			ScopeId:     req.ScopeId,
			CreatorID:   userID,
			Type:        req.Type,
		}, nil
	default:
		return tokenstore.TokenStoreItem{}, fmt.Errorf("token type: %s is not valid", tokenType)
	}
}

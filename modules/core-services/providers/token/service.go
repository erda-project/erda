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
	context "context"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/token/pb"
)

type TokenService struct {
	logger logs.Logger

	dao Dao
}

func (s *TokenService) QueryTokens(ctx context.Context, req *pb.QueryTokensRequest) (*pb.QueryTokensResponse, error) {
	objs, total, err := s.dao.QueryToken(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.Token, len(objs))
	for i, obj := range objs {
		res[i] = obj.ToPbToken()
	}
	return &pb.QueryTokensResponse{Data: res, Total: total}, nil
}

func (s *TokenService) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	obj, err := s.dao.GetToken(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.GetTokenResponse{Data: obj.ToPbToken()}, nil
}

func (s *TokenService) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	obj, err := s.dao.CreateToken(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.CreateTokenResponse{Data: obj.ToPbToken()}, nil
}

func (s *TokenService) UpdateToken(ctx context.Context, req *pb.UpdateTokenRequest) (*pb.UpdateTokenResponse, error) {
	return &pb.UpdateTokenResponse{}, s.dao.UpdateToken(ctx, req)
}

func (s *TokenService) DeleteToken(ctx context.Context, req *pb.DeleteTokenRequest) (*pb.DeleteTokenResponse, error) {
	return &pb.DeleteTokenResponse{}, s.dao.DeleteToken(ctx, req)
}

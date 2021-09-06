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

package accesskey

import (
	context "context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
)

type accessKeyService struct {
	p *provider
}

func (s *accessKeyService) QueryAccessKeys(ctx context.Context, req *pb.QueryAccessKeysRequest) (*pb.QueryAccessKeysResponse, error) {
	objs, total, err := s.p.dao.QueryAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.AccessKeysItem, len(objs))
	for i, obj := range objs {
		res[i] = &pb.AccessKeysItem{
			Id:          obj.ID,
			AccessKey:   obj.AccessKey,
			SecretKey:   obj.SecretKey,
			Status:      obj.Status,
			SubjectType: obj.SubjectType,
			Subject:     obj.Subject,
			Description: obj.Description,
			CreatedAt:   timestamppb.New(obj.CreatedAt),
		}
	}
	return &pb.QueryAccessKeysResponse{Data: res, Total: total}, nil
}

func (s *accessKeyService) GetAccessKey(ctx context.Context, req *pb.GetAccessKeyRequest) (*pb.GetAccessKeyResponse, error) {
	obj, err := s.p.dao.GetAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccessKeyResponse{Data: &pb.AccessKeysItem{
		Id:          obj.ID,
		AccessKey:   obj.AccessKey,
		SecretKey:   obj.SecretKey,
		Status:      obj.Status,
		SubjectType: obj.SubjectType,
		Subject:     obj.Subject,
		Description: obj.Description,
		CreatedAt:   timestamppb.New(obj.CreatedAt),
	}}, nil
}

func (s *accessKeyService) CreateAccessKey(ctx context.Context, req *pb.CreateAccessKeyRequest) (*pb.CreateAccessKeyResponse, error) {
	obj, err := s.p.dao.CreateAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAccessKeyResponse{Data: &pb.AccessKeysItem{
		Id:          obj.ID,
		AccessKey:   obj.AccessKey,
		SecretKey:   obj.SecretKey,
		Status:      obj.Status,
		SubjectType: obj.SubjectType,
		Subject:     obj.Subject,
		Description: obj.Description,
		CreatedAt:   timestamppb.New(obj.CreatedAt),
	}}, nil
}

func (s *accessKeyService) UpdateAccessKey(ctx context.Context, req *pb.UpdateAccessKeyRequest) (*pb.UpdateAccessKeyResponse, error) {
	return &pb.UpdateAccessKeyResponse{}, s.p.dao.UpdateAccessKey(ctx, req)
}

func (s *accessKeyService) DeleteAccessKey(ctx context.Context, req *pb.DeleteAccessKeyRequest) (*pb.DeleteAccessKeyResponse, error) {
	return &pb.DeleteAccessKeyResponse{}, s.p.dao.DeleteAccessKey(ctx, req)
}

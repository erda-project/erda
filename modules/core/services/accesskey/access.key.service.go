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

	pb "github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
)

type accessKeyService struct {
	p *provider
}

func (s *accessKeyService) QueryAccessKeys(ctx context.Context, req *pb.QueryAccessKeysRequest) (*pb.QueryAccessKeysResponse, error) {
	objs, err := s.p.dao.QueryAccessKey(ctx, req)
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
	return &pb.QueryAccessKeysResponse{Data: res}, nil
}

func (s *accessKeyService) GetAccessKey(ctx context.Context, req *pb.GetAccessKeysRequest) (*pb.GetAccessKeysResponse, error) {
	obj, err := s.p.dao.GetAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccessKeysResponse{Data: &pb.AccessKeysItem{
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

func (s *accessKeyService) CreateAccessKeys(ctx context.Context, req *pb.CreateAccessKeysRequest) (*pb.CreateAccessKeysResponse, error) {
	obj, err := s.p.dao.CreateAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAccessKeysResponse{Data: &pb.AccessKeysItem{
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

func (s *accessKeyService) UpdateAccessKeys(ctx context.Context, req *pb.UpdateAccessKeysRequest) (*pb.UpdateAccessKeysResponse, error) {
	return nil, s.p.dao.UpdateAccessKey(ctx, req)
}

func (s *accessKeyService) DeleteAccessKeys(ctx context.Context, req *pb.DeleteAccessKeysRequest) (*pb.DeleteAccessKeysResponse, error) {
	return nil, s.p.dao.DeleteAccessKey(ctx, req)
}

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
package credential

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda-proto-go/msp/credential/pb"
)

////go:generate mockgen -destination=./credential_register_test.go -package exporter github.com/erda-project/erda-infra/pkg/transport Register
////go:generate mockgen -destination=./credential_ak_test.go -package exporter github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb AccessKeyServiceServer
func Test_accessKeyService_QueryAccessKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)

	akService.EXPECT().CreateAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(&akpb.CreateAccessKeyResponse{
		Data: &akpb.AccessKeysItem{},
	}, nil)

	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		AccessKeyService:     akService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.CreateAccessKey(context.Background(), &pb.CreateAccessKeyRequest{
		SubjectType: akpb.SubjectTypeEnum_SYSTEM,
		Subject:     "xddd",
		Description: "cdddd",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_DeleteAccessKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)
	akService.EXPECT().DeleteAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		AccessKeyService:     akService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.DeleteAccessKey(context.Background(), &pb.DeleteAccessKeyRequest{
		Id: "13eef468-1d0b-42ce-aa7b-b499545bf92d",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_QueryAccessKeys1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)
	akService.EXPECT().QueryAccessKeys(gomock.Any(), gomock.Any()).AnyTimes().Return(&akpb.QueryAccessKeysResponse{
		Data: make([]*akpb.AccessKeysItem, 0),
	}, nil)
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		AccessKeyService:     akService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.QueryAccessKeys(context.Background(), &pb.QueryAccessKeysRequest{
		SubjectType: 3,
		Subject:     "22",
		PageSize:    1,
		PageNo:      3,
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_GetAccessKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)
	akService.EXPECT().GetAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(&akpb.GetAccessKeyResponse{
		Data: &akpb.AccessKeysItem{},
	}, nil)
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		AccessKeyService:     akService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.GetAccessKey(context.Background(), &pb.GetAccessKeyRequest{
		Id: "13eef468-1d0b-42ce-aa7b-b499545bf92d",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

func Test_accessKeyService_DownloadAccessKeyFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	akService := NewMockAccessKeyServiceServer(ctrl)
	akService.EXPECT().GetAccessKey(gomock.Any(), gomock.Any()).AnyTimes().Return(&akpb.GetAccessKeyResponse{
		Data: &akpb.AccessKeysItem{
			Id:          "ssss",
			AccessKey:   "dddd",
			SecretKey:   "dddd",
			Status:      0,
			SubjectType: 0,
			Subject:     "ccc",
			Description: "aaa",
			CreatedAt:   &timestamppb.Timestamp{},
		},
	}, nil)
	pro := &provider{
		Cfg:                  &config{},
		Register:             NewMockRegister(ctrl),
		credentialKeyService: &accessKeyService{},
		AccessKeyService:     akService,
	}
	pro.credentialKeyService.p = pro
	_, err := pro.credentialKeyService.DownloadAccessKeyFile(context.Background(), &pb.DownloadAccessKeyFileRequest{
		Id: "ssss",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}

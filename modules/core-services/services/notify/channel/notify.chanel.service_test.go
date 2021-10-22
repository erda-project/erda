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

package channel

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"bou.ke/monkey"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/services/notify/channel/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/notify/channel/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/ucauth"
)

func Test_notifyChannelService_CreateNotifyChannel(t *testing.T) {

	type args struct {
		ctx context.Context
		req *pb.CreateNotifyChannelRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.CreateNotifyChannelRequest{Name: "", Type: "ali_short_message", Config: nil}}, true},
		{"case2", args{req: &pb.CreateNotifyChannelRequest{Name: "test", Type: "", Config: map[string]*structpb.Value{}}}, true},
		{"case3", args{req: &pb.CreateNotifyChannelRequest{Name: "test", Type: "error", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case4", args{req: &pb.CreateNotifyChannelRequest{Name: "create_error", Type: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case5", args{req: &pb.CreateNotifyChannelRequest{Name: "test", Type: "ali_short_message", Config: nil}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ncs *notifyChannelService
			monkey.PatchInstanceMethod(reflect.TypeOf(ncs), "ConfigValidate", func(ncs *notifyChannelService, channelType string, c map[string]*structpb.Value) error {
				if channelType == "error" {
					return errors.New("not support")
				}
				return nil
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(ncs), "CovertToPbNotifyChannel", func(ncs *notifyChannelService, lang i18n.LanguageCodes, channel *model.NotifyChannel) *pb.NotifyChannel {
				return nil
			})

			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetByName", func(ncdb *db.NotifyChannelDB, name string) (*model.NotifyChannel, error) {
				if name == "" {
					return nil, errors.New("err")
				}
				if name == "test" {
					return &model.NotifyChannel{Id: "xx", Name: name}, nil
				}
				return nil, nil
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "Create", func(ncdb *db.NotifyChannelDB, notifyChannel *model.NotifyChannel) (*model.NotifyChannel, error) {
				if notifyChannel.Name == "create_error" {
					return nil, errors.New("create_error")
				}
				return notifyChannel, nil
			})

			monkey.Patch(apis.GetUserID, func(ctx context.Context) string {
				return "1"
			})
			monkey.Patch(apis.GetOrgID, func(ctx context.Context) string {
				return "1"
			})
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return i18n.LanguageCodes{{Code: "zh"}}
			})
			var b *ucauth.UCClient
			monkey.PatchInstanceMethod(reflect.TypeOf(b), "GetCurrentUser", func(b *bundle.Bundle, userID string) (*apistructs.UserInfo, error) {
				return &apistructs.UserInfo{ID: "1", Name: "Test"}, nil
			})

			s := &notifyChannelService{p: &provider{uc: b}}
			_, err := s.CreateNotifyChannel(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNotifyChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func Test_notifyChannelService_GetNotifyChannels(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetNotifyChannelsRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.GetNotifyChannelsRequest{Page: 1, PageSize: 10}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(apis.GetOrgID, func(ctx context.Context) string {
				return "1"
			})
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return i18n.LanguageCodes{{Code: "zh"}}
			})
			var ncs *notifyChannelService
			monkey.PatchInstanceMethod(reflect.TypeOf(ncs), "CovertToPbNotifyChannel", func(ncs *notifyChannelService, lang i18n.LanguageCodes, channel *model.NotifyChannel) *pb.NotifyChannel {
				nc := pb.NotifyChannel{}
				err := copier.CopyWithOption(&nc, channel, copier.Option{IgnoreEmpty: true, DeepCopy: true})
				if err != nil {
					return nil
				}
				return &nc
			})

			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "ListByPage", func(ncdb *db.NotifyChannelDB, offset, pageSize int64, orgId string) (int64, []model.NotifyChannel, error) {
				var channels []model.NotifyChannel
				for i := 0; i < 10; i++ {
					channels = append(channels, model.NotifyChannel{Id: strconv.Itoa(i), Name: strconv.Itoa(i), Type: strconv.Itoa(i)})
				}
				return 10, channels, nil
			})

			s := &notifyChannelService{}
			_, err := s.GetNotifyChannels(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNotifyChannels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_notifyChannelService_UpdateNotifyChannel(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.UpdateNotifyChannelRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.UpdateNotifyChannelRequest{Name: "create_error", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case2", args{req: &pb.UpdateNotifyChannelRequest{Id: "error", Name: "create_error", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case3", args{req: &pb.UpdateNotifyChannelRequest{Id: "nil", Name: "create_error", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case4", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, false},
		{"case5", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "error", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case6", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "exist", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case7", args{req: &pb.UpdateNotifyChannelRequest{Id: "not_same", Name: "current", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
		{"case8", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "current", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, false},
		{"case9", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "test", Type: "", ChannelProviderType: "ali_short_message", Config: nil}}, false},
		{"case10", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "test", Type: "short_message", ChannelProviderType: "ali_short_message", Config: nil}}, false},
		{"case11", args{req: &pb.UpdateNotifyChannelRequest{Id: "test", Name: "test", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, false},
		{"case12", args{req: &pb.UpdateNotifyChannelRequest{Id: "error", Name: "test", Type: "short_message", ChannelProviderType: "ali_short_message", Config: map[string]*structpb.Value{"AccessKeyId": structpb.NewStringValue("xx"), "AccessKeySecret": structpb.NewStringValue("xx"), "SignName": structpb.NewStringValue("xx"), "TemplateCode": structpb.NewStringValue("xx")}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return i18n.LanguageCodes{{Code: "zh"}}
			})
			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetById", func(ncdb *db.NotifyChannelDB, id string) (*model.NotifyChannel, error) {
				if id == "error" {
					return nil, errors.New("error")
				}
				if id == "nil" {
					return nil, nil
				}
				return &model.NotifyChannel{Id: id}, nil
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "UpdateById", func(ncdb *db.NotifyChannelDB, notifyChannel *model.NotifyChannel) (*model.NotifyChannel, error) {
				if notifyChannel.Id == "error" {
					return nil, errors.New("error")
				}
				return notifyChannel, nil
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetCountByName", func(ncdb *db.NotifyChannelDB, name string) (int64, error) {
				if name == "error" {
					return 0, errors.New("error")
				}
				if name == "exist" {
					return 2, nil
				}
				if name == "current" {
					return 1, nil
				}
				return 0, nil
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetByName", func(ncdb *db.NotifyChannelDB, name string) (*model.NotifyChannel, error) {
				return &model.NotifyChannel{Id: "test", Name: name}, nil
			})
			var ncs *notifyChannelService
			monkey.PatchInstanceMethod(reflect.TypeOf(ncs), "CovertToPbNotifyChannel", func(ncs *notifyChannelService, lang i18n.LanguageCodes, channel *model.NotifyChannel) *pb.NotifyChannel {
				nc := pb.NotifyChannel{}
				err := copier.CopyWithOption(&nc, channel, copier.Option{IgnoreEmpty: true, DeepCopy: true})
				if err != nil {
					return nil
				}
				return &nc
			})

			s := &notifyChannelService{}
			_, err := s.UpdateNotifyChannel(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateNotifyChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_notifyChannelService_GetNotifyChannel(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetNotifyChannelRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.GetNotifyChannelRequest{Id: ""}}, true},
		{"case2", args{req: &pb.GetNotifyChannelRequest{Id: "error"}}, true},
		{"case3", args{req: &pb.GetNotifyChannelRequest{Id: "test"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return i18n.LanguageCodes{{Code: "zh"}}
			})
			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetById", func(ncdb *db.NotifyChannelDB, id string) (*model.NotifyChannel, error) {
				if id == "error" {
					return nil, errors.New("error")
				}
				if id == "nil" {
					return nil, nil
				}
				return &model.NotifyChannel{Id: id}, nil
			})
			var ncs *notifyChannelService
			monkey.PatchInstanceMethod(reflect.TypeOf(ncs), "CovertToPbNotifyChannel", func(ncs *notifyChannelService, lang i18n.LanguageCodes, channel *model.NotifyChannel) *pb.NotifyChannel {
				nc := pb.NotifyChannel{}
				err := copier.CopyWithOption(&nc, channel, copier.Option{IgnoreEmpty: true, DeepCopy: true})
				if err != nil {
					return nil
				}
				return &nc
			})

			s := &notifyChannelService{}
			_, err := s.GetNotifyChannel(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNotifyChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_notifyChannelService_DeleteNotifyChannel(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.DeleteNotifyChannelRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.DeleteNotifyChannelRequest{Id: ""}}, true},
		{"case1", args{req: &pb.DeleteNotifyChannelRequest{Id: "error"}}, true},
		{"case1", args{req: &pb.DeleteNotifyChannelRequest{Id: "test"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "DeleteById", func(ncdb *db.NotifyChannelDB, id string) (*model.NotifyChannel, error) {
				if id == "error" {
					return nil, errors.New("error")
				}
				if id == "nil" {
					return nil, nil
				}
				return &model.NotifyChannel{Id: id}, nil
			})

			s := &notifyChannelService{}
			_, err := s.DeleteNotifyChannel(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteNotifyChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_notifyChannelService_ConfigValidate(t *testing.T) {
	type args struct {
		channelType string
		c           map[string]*structpb.Value
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{channelType: "ali_short_message", c: map[string]*structpb.Value{"accessKeyId": structpb.NewStringValue("xx"), "accessKeySecret": structpb.NewStringValue("xx"), "signName": structpb.NewStringValue("xx"), "templateCode": structpb.NewStringValue("xx")}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &notifyChannelService{}
			if err := s.ConfigValidate(tt.args.channelType, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("ConfigValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_notifyChannelService_GetNotifyChannelEnabled(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetNotifyChannelEnabledRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.GetNotifyChannelEnabledRequest{ScopeId: "", ScopeType: "", Type: "xx"}}, true},
		{"case2", args{req: &pb.GetNotifyChannelEnabledRequest{ScopeId: "xx", ScopeType: "", Type: ""}}, true},
		{"case3", args{req: &pb.GetNotifyChannelEnabledRequest{ScopeId: "error", ScopeType: "xx", Type: "xx"}}, true},
		{"case4", args{req: &pb.GetNotifyChannelEnabledRequest{ScopeId: "xx", ScopeType: "xx", Type: "xx"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return i18n.LanguageCodes{{Code: "zh"}}
			})
			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetByScopeAndType", func(ncdb *db.NotifyChannelDB, scopeId, scopeType, channelType string) (*model.NotifyChannel, error) {
				if scopeId == "error" {
					return nil, errors.New("error")
				}
				return &model.NotifyChannel{Id: "test"}, nil
			})
			var ncs *notifyChannelService
			monkey.PatchInstanceMethod(reflect.TypeOf(ncs), "CovertToPbNotifyChannel", func(ncs *notifyChannelService, lang i18n.LanguageCodes, channel *model.NotifyChannel) *pb.NotifyChannel {
				nc := pb.NotifyChannel{}
				err := copier.CopyWithOption(&nc, channel, copier.Option{IgnoreEmpty: true, DeepCopy: true})
				if err != nil {
					return nil
				}
				return &nc
			})
			s := &notifyChannelService{}
			_, err := s.GetNotifyChannelEnabled(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNotifyChannelEnabled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_notifyChannelService_UpdateNotifyChannelEnabled(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.UpdateNotifyChannelEnabledRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: ""}}, true},
		{"case2", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: "error", Enable: true}}, true},
		{"case3", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: "nil", Enable: false}}, true},
		{"case4", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: "type_error", Enable: true}}, true},
		{"case5", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: "type_exist", Enable: true}}, true},
		{"case6", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: "update_error", Enable: true}}, true},
		{"case7", args{req: &pb.UpdateNotifyChannelEnabledRequest{Id: "test", Enable: false}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return i18n.LanguageCodes{{Code: "zh"}}
			})
			var ncdb *db.NotifyChannelDB
			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetById", func(ncdb *db.NotifyChannelDB, id string) (*model.NotifyChannel, error) {
				if id == "error" {
					return nil, errors.New("error")
				}
				if id == "nil" {
					return nil, nil
				}
				if id == "type_error" {
					return &model.NotifyChannel{ScopeId: "test", ScopeType: "test", Type: "type_error", IsEnabled: false}, nil
				}
				if id == "type_exist" {
					return &model.NotifyChannel{ScopeId: "test", ScopeType: "test", Type: "type_exist", IsEnabled: false}, nil
				}

				if id == "update_error" {
					return &model.NotifyChannel{ScopeId: "test", ScopeType: "test", Type: "update_error", IsEnabled: false}, nil
				}
				return &model.NotifyChannel{ScopeId: "test", ScopeType: "test", Type: "test", IsEnabled: false}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "GetCountByScopeAndType", func(ncdb *db.NotifyChannelDB, scopeId, scopeType, channelType string) (int64, error) {
				if channelType == "type_error" {
					return 0, errors.New("error")
				}
				if channelType == "type_exist" {
					return 1, nil
				}
				return 0, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(ncdb), "UpdateById", func(ncdb *db.NotifyChannelDB, notifyChannel *model.NotifyChannel) (*model.NotifyChannel, error) {
				if notifyChannel.Type == "update_error" {
					return nil, errors.New("error")
				}
				return notifyChannel, nil
			})

			s := &notifyChannelService{}
			_, err := s.UpdateNotifyChannelEnabled(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateNotifyChannelEnabled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

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
	context "context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/erda-project/erda-proto-go/core/services/notify/channel/pb"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/notify/channel/chtype"
	"github.com/erda-project/erda/modules/core-services/services/notify/channel/db"
	"github.com/erda-project/erda/pkg/common/apis"
	pkgerrors "github.com/erda-project/erda/pkg/common/errors"
)

type notifyChannelService struct {
	p               *provider
	NotifyChannelDB *db.NotifyChannelDB
}

func (s *notifyChannelService) CreateNotifyChannel(ctx context.Context, req *pb.CreateNotifyChannelRequest) (*pb.CreateNotifyChannelResponse, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewMissingParameterError("name")
	}
	if req.Type == "" {
		return nil, pkgerrors.NewMissingParameterError("type")
	}
	err := s.ConfigValidate(req.Type, req.Config)
	if err != nil {
		return nil, err
	}

	ch, err := s.NotifyChannelDB.GetByName(req.Name)
	if err != nil {
		return nil, err
	}
	if ch != nil {
		return nil, pkgerrors.NewAlreadyExistsError("Channel name")
	}

	creatorId := apis.GetUserID(ctx)
	user, err := s.p.bdl.GetCurrentUser(creatorId)
	if user == nil || user.ID == "" {
		return nil, pkgerrors.NewNotFoundError("User")
	}
	orgId := apis.GetOrgID(ctx)
	if orgId == "" {
		return nil, pkgerrors.NewNotFoundError("Org")
	}
	config, err := json.Marshal(req.Config)
	if err != nil {
		return nil, err
	}
	channel, err := s.NotifyChannelDB.Create(&model.NotifyChannel{
		Id:          uuid.NewV4().String(),
		Name:        req.Name,
		Type:        req.Type,
		Config:      string(config),
		ScopeId:     orgId,
		ScopeType:   "org",
		CreatorId:   creatorId,
		CreatorName: user.Name,
		CreateAt:    time.Now(),
		UpdatedAt:   time.Now(),
		IsDeleted:   false,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateNotifyChannelResponse{Data: s.CovertToPbNotifyChannel(channel)}, nil
}

func (s *notifyChannelService) GetNotifyChannels(ctx context.Context, req *pb.GetNotifyChannelsRequest) (*pb.GetNotifyChannelsResponse, error) {
	orgId := apis.GetOrgID(ctx)
	if orgId == "" {
		return nil, pkgerrors.NewNotFoundError("Org")
	}
	total, channels, err := s.NotifyChannelDB.ListByPage(req.Page, req.PageSize, orgId)
	var pbChannels []*pb.NotifyChannel
	for _, channel := range channels {
		pbChannels = append(pbChannels, s.CovertToPbNotifyChannel(&channel))
	}

	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	return &pb.GetNotifyChannelsResponse{Page: req.Page, PageSize: req.PageSize, Total: total, Data: pbChannels}, nil
}

func (s *notifyChannelService) UpdateNotifyChannel(ctx context.Context, req *pb.UpdateNotifyChannelRequest) (*pb.UpdateNotifyChannelResponse, error) {
	if req.Id == "" {
		return nil, pkgerrors.NewMissingParameterError("id")
	}
	channel, err := s.NotifyChannelDB.GetById(req.Id)
	if err != nil {
		return nil, pkgerrors.NewDatabaseError(err)
	}
	if channel == nil {
		return nil, pkgerrors.NewDatabaseError(pkgerrors.NewNotFoundError(req.Id))
	}
	if req.Name != "" {
		countByName, err := s.NotifyChannelDB.GetCountByName(req.Name)
		if err != nil {
			return nil, pkgerrors.NewDatabaseError(err)
		}
		if countByName > 1 {
			return nil, pkgerrors.NewAlreadyExistsError("name")
		}
		if countByName == 1 {
			byName, err := s.NotifyChannelDB.GetByName(req.Name)
			if err != nil {
				return nil, pkgerrors.NewDatabaseError(err)
			}
			if byName != nil && byName.Id != req.Id {
				return nil, pkgerrors.NewAlreadyExistsError("name")
			}
		}
		channel.Name = req.Name
	}
	if req.Type != "" {
		channel.Type = req.Type
	}
	if req.Config != nil {
		err = s.ConfigValidate(req.Type, req.Config)
		if err != nil {
			return nil, err
		}
		config, err := json.Marshal(req.Config)
		if err != nil {
			return nil, err
		}
		channel.Config = string(config)
	}
	channel.UpdatedAt = time.Now()

	update, err := s.NotifyChannelDB.UpdateById(channel)
	if err != nil {
		return nil, pkgerrors.NewDatabaseError(err)
	}
	return &pb.UpdateNotifyChannelResponse{Data: s.CovertToPbNotifyChannel(update)}, nil
}

func (s *notifyChannelService) GetNotifyChannel(ctx context.Context, req *pb.GetNotifyChannelRequest) (*pb.GetNotifyChannelResponse, error) {
	if req.GetId() == "" {
		return nil, pkgerrors.NewMissingParameterError("id")
	}
	channel, err := s.NotifyChannelDB.GetById(req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetNotifyChannelResponse{Data: s.CovertToPbNotifyChannel(channel)}, nil
}

func (s *notifyChannelService) DeleteNotifyChannel(ctx context.Context, req *pb.DeleteNotifyChannelRequest) (*pb.DeleteNotifyChannelResponse, error) {
	if req.GetId() == "" {
		return nil, pkgerrors.NewMissingParameterError("id")
	}
	channel, err := s.NotifyChannelDB.DeleteById(req.Id)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	return &pb.DeleteNotifyChannelResponse{Id: channel.Id}, nil
}

func (s *notifyChannelService) ConfigValidate(channelType string, c map[string]*structpb.Value) error {
	switch channelType {
	case strings.ToLower(pb.Type_ALI_SHORT_MESSAGE.String()):
		bytes, err := json.Marshal(c)
		if err != nil {
			return errors.New("Json parser failed.")
		}
		var asm chtype.AliShortMessage
		err = json.Unmarshal(bytes, &asm)
		if err != nil {
			return errors.New("Json parser failed.")
		}
		return asm.Validate()
	default:
		return pkgerrors.NewInternalServerError(errors.New("Not support notify channel type"))
	}
}

func (s *notifyChannelService) ParamsValidate(req *pb.CreateNotifyChannelRequest) error {
	err := s.ConfigValidate(req.Type, req.Config)
	if err != nil {
		return err
	}
	return nil
}

func (s *notifyChannelService) CovertToPbNotifyChannel(channel *model.NotifyChannel) *pb.NotifyChannel {
	ncpb := pb.NotifyChannel{}
	err := copier.CopyWithOption(&ncpb, &channel, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return nil
	}
	var config map[string]*structpb.Value
	err = json.Unmarshal([]byte(channel.Config), &config)
	if err != nil {
		return nil
	}
	ncpb.Config = config
	layout := "2006-01-02 15:04:05"
	ncpb.CreateAt = channel.CreateAt.Format(layout)
	ncpb.UpdateAt = channel.UpdatedAt.Format(layout)
	return &ncpb
}

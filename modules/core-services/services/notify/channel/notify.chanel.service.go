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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/i18n"
	pb "github.com/erda-project/erda-proto-go/core/services/notify/channel/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/notify/channel/db"
	"github.com/erda-project/erda/modules/core-services/services/notify/channel/kind"
	"github.com/erda-project/erda/pkg/common/apis"
	pkgerrors "github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
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
	if req.ChannelProviderType == "" {
		return nil, pkgerrors.NewMissingParameterError("channelProviderType")
	}
	c, err := s.ConfigValidate(req.ChannelProviderType, req.Config)
	if err != nil {
		return nil, err
	}
	req.Config = c
	kmsKey, err := s.p.bdl.KMSCreateKey(apistructs.KMSCreateKeyRequest{
		CreateKeyRequest: kmstypes.CreateKeyRequest{
			PluginKind: kmstypes.PluginKind_ERDA_KMS,
		},
	})
	if err != nil {
		return nil, err
	}
	encryptSecret, err := s.p.bdl.KMSEncrypt(apistructs.KMSEncryptRequest{
		EncryptRequest: kmstypes.EncryptRequest{
			KeyID:           kmsKey.KeyMetadata.KeyID,
			PlaintextBase64: base64.StdEncoding.EncodeToString([]byte(req.Config["need_kms_data"].GetStringValue())),
		},
	})
	if err != nil {
		return nil, err
	}
	req.Config[c["need_kms_key"].GetStringValue()] = structpb.NewStringValue(encryptSecret.CiphertextBase64)
	delete(req.Config, "need_kms_key")
	delete(req.Config, "need_kms_data")

	ch, err := s.NotifyChannelDB.GetByName(req.Name)
	if err != nil {
		return nil, err
	}
	if ch != nil {
		return nil, pkgerrors.NewAlreadyExistsError("Channel name")
	}

	creatorId := apis.GetUserID(ctx)
	user, err := s.p.uc.GetUser(creatorId)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
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
		Id:              uuid.NewV4().String(),
		Name:            req.Name,
		Type:            req.Type,
		ChannelProvider: req.ChannelProviderType,
		Config:          string(config),
		ScopeId:         orgId,
		ScopeType:       "org",
		CreatorId:       creatorId,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		KmsKey:          kmsKey.KeyMetadata.KeyID,
		IsEnabled:       false, // default no enable
		IsDeleted:       false,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateNotifyChannelResponse{Data: s.CovertToPbNotifyChannel(apis.Language(ctx), channel)}, nil
}

func (s *notifyChannelService) GetNotifyChannels(ctx context.Context, req *pb.GetNotifyChannelsRequest) (*pb.GetNotifyChannelsResponse, error) {
	if req.PageNo < 1 {
		req.PageNo = 1
	}
	if req.PageSize < 15 {
		req.PageSize = 15
	}
	if req.PageSize > 60 {
		req.PageSize = 60
	}
	orgId := apis.GetOrgID(ctx)
	if orgId == "" {
		return nil, pkgerrors.NewNotFoundError("Org")
	}
	scopeType := "org"
	total, channels, err := s.NotifyChannelDB.ListByPage((req.PageNo-1)*req.PageSize, req.PageSize, orgId, scopeType)
	var pbChannels []*pb.NotifyChannel
	for _, channel := range channels {
		notifyChannel := s.CovertToPbNotifyChannel(apis.Language(ctx), &channel)
		c, err := s.ConfigValidate(notifyChannel.ChannelProviderType.Name, notifyChannel.Config)
		if err != nil {
			return nil, err
		}
		decrypt, err := s.p.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
			DecryptRequest: kmstypes.DecryptRequest{
				KeyID:            channel.KmsKey,
				CiphertextBase64: c["need_kms_data"].GetStringValue(),
			}})
		decodeString, err := base64.StdEncoding.DecodeString(decrypt.PlaintextBase64)
		if err != nil {
			return nil, err
		}
		notifyChannel.Config[c["need_kms_key"].GetStringValue()] = structpb.NewStringValue(string(decodeString))
		delete(notifyChannel.Config, "need_kms_key")
		delete(notifyChannel.Config, "need_kms_data")
		if err != nil {
			return nil, err
		}
		pbChannels = append(pbChannels, notifyChannel)
	}

	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	return &pb.GetNotifyChannelsResponse{PageNo: req.PageNo, PageSize: req.PageSize, Total: total, Data: pbChannels}, nil
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
	if req.ChannelProviderType != "" {
		channel.ChannelProvider = req.ChannelProviderType
	}

	needKmsKey := ""
	needKmsData := ""
	if req.Config != nil {
		c, err := s.ConfigValidate(req.ChannelProviderType, req.Config)
		if err != nil {
			return nil, err
		}
		encryptSecret, err := s.p.bdl.KMSEncrypt(apistructs.KMSEncryptRequest{
			EncryptRequest: kmstypes.EncryptRequest{
				KeyID:           channel.KmsKey,
				PlaintextBase64: base64.StdEncoding.EncodeToString([]byte(c["need_kms_data"].GetStringValue())),
			},
		})
		if err != nil {
			return nil, err
		}
		needKmsKey = c["need_kms_key"].GetStringValue()
		needKmsData = c["need_kms_data"].GetStringValue()
		req.Config[needKmsKey] = structpb.NewStringValue(encryptSecret.CiphertextBase64)
		delete(req.Config, "need_kms_key")
		delete(req.Config, "need_kms_data")

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
	notifyChannel := s.CovertToPbNotifyChannel(apis.Language(ctx), update)
	if needKmsData != "" && needKmsKey != "" {
		notifyChannel.Config[needKmsKey] = structpb.NewStringValue(needKmsData)
	}
	return &pb.UpdateNotifyChannelResponse{Data: notifyChannel}, nil
}

func (s *notifyChannelService) GetNotifyChannel(ctx context.Context, req *pb.GetNotifyChannelRequest) (*pb.GetNotifyChannelResponse, error) {
	if req.Id == "" {
		return nil, pkgerrors.NewMissingParameterError("id")
	}
	channel, err := s.NotifyChannelDB.GetById(req.Id)
	if err != nil {
		return nil, err
	}

	c := map[string]*structpb.Value{}
	err = json.Unmarshal([]byte(channel.Config), &c)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}

	cTemp, err := s.ConfigValidate(channel.ChannelProvider, c)
	decrypt, err := s.p.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
		DecryptRequest: kmstypes.DecryptRequest{
			KeyID:            channel.KmsKey,
			CiphertextBase64: cTemp["need_kms_data"].GetStringValue(),
		}})
	decodeString, err := base64.StdEncoding.DecodeString(decrypt.PlaintextBase64)
	if err != nil {
		return nil, err
	}
	cTemp[cTemp["need_kms_key"].GetStringValue()] = structpb.NewStringValue(string(decodeString))
	delete(cTemp, "need_kms_key")
	delete(cTemp, "need_kms_data")
	bytes, err := json.Marshal(cTemp)
	if err != nil {
		return nil, err
	}
	channel.Config = string(bytes)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}

	return &pb.GetNotifyChannelResponse{Data: s.CovertToPbNotifyChannel(apis.Language(ctx), channel)}, nil
}

func (s *notifyChannelService) DeleteNotifyChannel(ctx context.Context, req *pb.DeleteNotifyChannelRequest) (*pb.DeleteNotifyChannelResponse, error) {
	if req.Id == "" {
		return nil, pkgerrors.NewMissingParameterError("id")
	}
	channel, err := s.NotifyChannelDB.DeleteById(req.Id)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	return &pb.DeleteNotifyChannelResponse{Id: channel.Id}, nil
}

func (s *notifyChannelService) GetNotifyChannelTypes(ctx context.Context, req *pb.GetNotifyChannelTypesRequest) (*pb.GetNotifyChannelTypesResponse, error) {

	language := apis.Language(ctx)

	var shortMessageProviderTypes []*pb.NotifyChannelProviderType
	shortMessageProviderTypes = append(shortMessageProviderTypes, &pb.NotifyChannelProviderType{
		Name:        strings.ToLower(pb.ProviderType_ALIYUN_SMS.String()),
		DisplayName: s.p.I18n.Text(language, strings.ToLower(pb.ProviderType_ALIYUN_SMS.String())),
	})

	var types []*pb.NotifyChannelTypeResponse
	types = append(types, &pb.NotifyChannelTypeResponse{
		Name:        strings.ToLower(pb.Type_SHORT_MESSAGE.String()),
		DisplayName: s.p.I18n.Text(language, strings.ToLower(pb.Type_SHORT_MESSAGE.String())),
		Providers:   shortMessageProviderTypes,
	})

	return &pb.GetNotifyChannelTypesResponse{Data: types}, nil
}

func (s *notifyChannelService) GetNotifyChannelEnabled(ctx context.Context, req *pb.GetNotifyChannelEnabledRequest) (*pb.GetNotifyChannelEnabledResponse, error) {
	if req.ScopeId == "" {
		return nil, pkgerrors.NewMissingParameterError("scopeId")
	}
	if req.ScopeType == "" {
		req.ScopeType = "org"
	}
	if req.Type == "" {
		return nil, pkgerrors.NewMissingParameterError("type")
	}

	data, err := s.NotifyChannelDB.GetByScopeAndType(req.ScopeId, req.ScopeType, req.Type)

	var c map[string]*structpb.Value
	err = json.Unmarshal([]byte(data.Config), &c)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}

	cTemp, err := s.ConfigValidate(data.ChannelProvider, c)
	decrypt, err := s.p.bdl.KMSDecrypt(apistructs.KMSDecryptRequest{
		DecryptRequest: kmstypes.DecryptRequest{
			KeyID:            data.KmsKey,
			CiphertextBase64: cTemp["need_kms_data"].GetStringValue(),
		}})

	decodeString, err := base64.StdEncoding.DecodeString(decrypt.PlaintextBase64)
	if err != nil {
		return nil, err
	}
	cTemp[cTemp["need_kms_key"].GetStringValue()] = structpb.NewStringValue(string(decodeString))
	delete(cTemp, "need_kms_key")
	delete(cTemp, "need_kms_data")
	bytes, err := json.Marshal(cTemp)
	if err != nil {
		return nil, err
	}
	data.Config = string(bytes)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}

	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	return &pb.GetNotifyChannelEnabledResponse{Data: s.CovertToPbNotifyChannel(apis.Language(ctx), data)}, nil
}

func (s *notifyChannelService) UpdateNotifyChannelEnabled(ctx context.Context, req *pb.UpdateNotifyChannelEnabledRequest) (*pb.UpdateNotifyChannelEnabledResponse, error) {

	if req.Id == "" {
		return nil, pkgerrors.NewMissingParameterError("id")
	}
	channel, err := s.NotifyChannelDB.GetById(req.Id)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	if channel == nil {
		return nil, errors.New(fmt.Sprintf("not found channel by id (%s)", req.Id))
	}

	if channel.IsEnabled == false && req.Enable == true {
		enabledCount, err := s.NotifyChannelDB.GetCountByScopeAndType(channel.ScopeId, channel.ScopeType, channel.Type)
		if err != nil {
			return nil, pkgerrors.NewInternalServerError(err)
		}
		if enabledCount >= 1 {
			return nil, pkgerrors.NewWarnError(fmt.Sprintf(s.p.I18n.Text(apis.Language(ctx), "enabled_exception"), channel.Type))
		}
	}
	channel.IsEnabled = req.Enable
	channel.UpdatedAt = time.Now()
	updated, err := s.NotifyChannelDB.UpdateById(channel)
	if err != nil {
		return nil, pkgerrors.NewInternalServerError(err)
	}
	return &pb.UpdateNotifyChannelEnabledResponse{Id: updated.Id, Enable: updated.IsEnabled}, nil
}

func (s *notifyChannelService) ConfigValidate(channelType string, c map[string]*structpb.Value) (map[string]*structpb.Value, error) {
	switch channelType {
	case strings.ToLower(pb.ProviderType_ALIYUN_SMS.String()):
		bytes, err := json.Marshal(c)
		if err != nil {
			return nil, errors.New("Json parser failed.")
		}
		var asm kind.AliyunSMS
		err = json.Unmarshal(bytes, &asm)
		if err != nil {
			return nil, errors.New("Json parser failed.")
		}
		err = asm.Validate()
		if err != nil {
			return nil, err
		}
		c["need_kms_key"] = structpb.NewStringValue("accessKeySecret")
		c["need_kms_data"] = structpb.NewStringValue(asm.AccessKeySecret)
		return c, nil
	default:
		return nil, errors.New("Not support notify channel type")
	}
}

func (s *notifyChannelService) CovertToPbNotifyChannel(lang i18n.LanguageCodes, channel *model.NotifyChannel) *pb.NotifyChannel {
	if channel == nil {
		return nil
	}
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
	ncpb.Type = &pb.NotifyChannelType{
		Name:        channel.Type,
		DisplayName: s.p.I18n.Text(lang, channel.Type),
	}
	ncpb.ChannelProviderType = &pb.NotifyChannelProviderType{
		Name:        channel.ChannelProvider,
		DisplayName: s.p.I18n.Text(lang, channel.ChannelProvider),
	}
	user, err := s.p.uc.GetUser(channel.CreatorId)
	if user != nil && user.Name != "" {
		ncpb.CreatorName = user.Name
	}
	ncpb.Enable = channel.IsEnabled
	layout := "2006-01-02 15:04:05"
	ncpb.CreateAt = channel.CreatedAt.Format(layout)
	ncpb.UpdateAt = channel.UpdatedAt.Format(layout)
	return &ncpb
}

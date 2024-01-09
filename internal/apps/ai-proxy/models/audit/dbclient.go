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

package audit

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	metapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/pkg/common/pbutil"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.AuditGetRequest) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(req.AuditId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.AuditPagingRequest) (*pb.AuditPagingResponse, error) {
	c := &Audit{BizSource: req.Source}
	sql := dbClient.DB.Model(c).Where(c)
	if len(req.Ids) > 0 {
		sql = sql.Where("id IN (?)", req.Ids)
	}
	var (
		total int64
		list  Audits
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	err := sql.Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &pb.AuditPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}

func (dbClient *DBClient) CreateWhenReceived(ctx context.Context, req *pb.AuditCreateRequestWhenReceived) (*pb.Audit, error) {
	c := &Audit{}
	c.RequestAt = func() time.Time {
		t := pbutil.GetTimeInLocal(req.RequestAt)
		if t == nil {
			return time.Time{}
		}
		return *t
	}()
	c.AuthKey = req.AuthKey
	c.RequestBody = req.RequestBody
	c.UserAgent = req.UserAgent
	c.XRequestID = req.XRequestId
	c.Username = req.Username
	c.Email = req.Email
	c.BizSource = req.BizSource

	// metadata
	meta := metadata.AuditMetadata{
		Public: metadata.AuditMetadataPublic{
			RequestContentType: req.RequestContentType,
			RequestHeader:      req.RequestHeader,
		},
		Secret: metadata.AuditMetadataSecret{
			IdentityPhoneNumber: req.IdentityPhoneNumber,
			IdentityJobNumber:   req.IdentityJobNumber,
			DingtalkStaffId:     req.DingtalkStaffId,
			DingtalkChatType:    req.DingtalkChatType,
			DingtalkChatTitle:   req.DingtalkChatTitle,
			DingtalkChatId:      req.DingtalkChatId,
		},
	}
	var pbMeta metapb.Metadata
	cputil.MustObjJSONTransfer(&meta, &pbMeta)
	c.Metadata = metadata.FromProtobuf(&pbMeta)

	if err := dbClient.DB.Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) UpdateAfterBasicContextParsed(ctx context.Context, req *pb.AuditUpdateRequestAfterBasicContextParsed) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(req.AuditId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}

	c.ClientID = req.ClientId
	c.ModelID = req.ModelId
	c.SessionID = req.SessionId

	c.Username = req.Username
	c.Email = req.Email

	c.OperationID = req.OperationId
	if req.BizSource != "" {
		c.BizSource = req.BizSource
	}
	if req.Username != "" {
		c.Username = req.Username
	}
	if req.Email != "" {
		c.Email = req.Email
	}

	var auditMetadata metadata.AuditMetadata
	cputil.MustObjJSONTransfer(&c.Metadata, &auditMetadata)
	if req.DingtalkStaffId != "" {
		auditMetadata.Secret.DingtalkStaffId = req.DingtalkStaffId
	}
	if req.IdentityJobNumber != "" {
		auditMetadata.Secret.IdentityJobNumber = req.IdentityJobNumber
	}
	if req.IdentityPhoneNumber != "" {
		auditMetadata.Secret.IdentityPhoneNumber = req.IdentityPhoneNumber
	}

	cputil.MustObjJSONTransfer(&auditMetadata, &c.Metadata)

	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) UpdateAfterSpecificContextParsed(ctx context.Context, req *pb.AuditUpdateRequestAfterSpecificContextParsed) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(req.AuditId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}

	c.Prompt = req.Prompt

	var auditMetadata metadata.AuditMetadata
	cputil.MustObjJSONTransfer(&c.Metadata, &auditMetadata)
	auditMetadata.Public.RequestFunctionCallName = req.RequestFunctionCallName
	// audio info
	if req.AudioFileName != "" {
		auditMetadata.Public.AudioFileName = req.AudioFileName
	}
	if req.AudioFileSize != "" {
		auditMetadata.Public.AudioFileSize = req.AudioFileSize
	}
	if req.AudioFileHeaders != "" {
		auditMetadata.Public.AudioFileHeaders = req.AudioFileHeaders
	}
	// image info
	if req.ImageQuality != "" {
		auditMetadata.Public.ImageQuality = req.ImageQuality
	}
	if req.ImageSize != "" {
		auditMetadata.Public.ImageSize = req.ImageSize
	}
	if req.ImageStyle != "" {
		auditMetadata.Public.ImageStyle = req.ImageStyle
	}

	cputil.MustObjJSONTransfer(&auditMetadata, &c.Metadata)

	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) UpdateAfterLLMDirectorInvoke(ctx context.Context, req *pb.AuditUpdateRequestAfterLLMDirectorInvoke) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(req.AuditId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	c.ActualRequestBody = req.ActualRequestBody
	var auditMetadata metadata.AuditMetadata
	cputil.MustObjJSONTransfer(&c.Metadata, &auditMetadata)
	auditMetadata.Public.ActualRequestURL = req.ActualRequestURL
	auditMetadata.Public.ActualRequestHeader = req.ActualRequestHeader
	cputil.MustObjJSONTransfer(&auditMetadata, &c.Metadata)

	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) UpdateAfterLLMResponse(ctx context.Context, req *pb.AuditUpdateRequestAfterLLMResponse) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(req.AuditId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}

	c.ResponseAt = *pbutil.GetTimeInLocal(req.ResponseAt)
	c.Status = req.Status
	c.ActualResponseBody = req.ActualResponseBody

	var auditMetadata metadata.AuditMetadata
	cputil.MustObjJSONTransfer(&c.Metadata, &auditMetadata)
	auditMetadata.Public.ResponseContentType = req.ResponseContentType
	if req.ResponseStreamDoneAt != nil {
		auditMetadata.Public.ResponseStreamDoneAt = pbutil.GetTimeInLocal(req.ResponseStreamDoneAt).String()
	}
	// time cost
	if req.ResponseStreamDoneAt != nil {
		auditMetadata.Public.TimeCost = req.ResponseStreamDoneAt.AsTime().Sub(c.RequestAt).String()
	} else {
		auditMetadata.Public.TimeCost = req.ResponseAt.AsTime().Sub(c.RequestAt).String()
	}
	cputil.MustObjJSONTransfer(&auditMetadata, &c.Metadata)

	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) UpdateAfterLLMDirectorResponse(ctx context.Context, req *pb.AuditUpdateRequestAfterLLMDirectorResponse) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(req.AuditId)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}

	c.Completion = req.Completion
	c.ResponseBody = req.ResponseBody
	c.ResponseFunctionCallName = req.ResponseFunctionCallName

	var auditMetadata metadata.AuditMetadata
	cputil.MustObjJSONTransfer(&c.Metadata, &auditMetadata)
	auditMetadata.Public.ResponseHeader = req.ResponseHeader
	cputil.MustObjJSONTransfer(&auditMetadata, &c.Metadata)

	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

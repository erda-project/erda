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

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.AuditPagingRequest) (*pb.AuditPagingResponse, error) {
	c := &Audit{}
	// auth_key
	if req.AuthKey != "" {
		c.AuthKey = req.AuthKey
	}
	// x-request-id
	if req.XRequestId != "" {
		c.XRequestID = req.XRequestId
	}
	// call-id
	if req.CallId != "" {
		c.CallID = req.CallId
	}

	sql := dbClient.DB.Model(c)
	sql = sql.Where(c).Unscoped()

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
	if req.PageSize > 20 {
		req.PageSize = 20
	}
	offset := (req.PageNum - 1) * req.PageSize
	sql = sql.Count(&total)
	// order by
	sql = sql.Order("created_at DESC")
	err := sql.Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
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
	c.CallID = req.CallId
	c.Username = req.Username
	c.Email = req.Email
	c.BizSource = req.BizSource

	// metadata
	meta := metadata.AuditMetadata{
		Public: metadata.AuditMetadataPublic{
			RequestContentType: req.RequestContentType,
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

func (dbClient *DBClient) Update(ctx context.Context, rec *Audit) (*pb.Audit, error) {
	c := &Audit{BaseModel: common.BaseModelWithID(rec.ID.String)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	if err := dbClient.DB.Model(c).Updates(rec).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

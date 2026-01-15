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
	"fmt"
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
	c := &Audit{
		AuthKey:    req.AuthKey,
		XRequestID: req.XRequestId,
		CallID:     req.CallId,
		ClientID:   req.ClientId,
		ModelID:    req.ModelId,
		Username:   req.Username,
	}
	sql := dbClient.DB.Model(c)
	// prompt
	if req.Prompt != "" {
		sql = sql.Where("prompt LIKE ?", "%"+req.Prompt+"%")
	}
	// operation_id
	if req.OperationId != "" {
		sql = sql.Where("operation_id LIKE ?", "%"+req.OperationId+"%")
	}

	before, after, err := ValidateAndGetTimeRange(req)
	if err != nil {
		return nil, err
	}
	sql = sql.Where("created_at <= ?", before)
	sql = sql.Where("created_at >= ?", after)

	// status filter
	if len(req.StatusIn) > 0 {
		sql = sql.Where("status IN ?", req.StatusIn)
	}
	if len(req.StatusNotIn) > 0 {
		sql = sql.Where("status NOT IN ?", req.StatusNotIn)
	}

	sql = sql.WithContext(ctx).Where(c).Unscoped()

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
	err = sql.Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &pb.AuditPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}

func ValidateAndGetTimeRange(req *pb.AuditPagingRequest) (time.Time, time.Time, error) {
	// time range
	var before, after time.Time
	if req.TimeRangeBeforeMs == 0 && req.TimeRangeAfterMs == 0 {
		before = time.Now()
		after = before.AddDate(0, 0, -1)
	} else if req.TimeRangeBeforeMs != 0 && req.TimeRangeAfterMs != 0 {
		var ok bool
		before, ok = pbutil.TimeFromMillis(req.TimeRangeBeforeMs)
		if !ok {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid TimeRangeBeforeMs")
		}
		after, ok = pbutil.TimeFromMillis(req.TimeRangeAfterMs)
		if !ok {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid TimeRangeAfterMs")
		}
	} else {
		return time.Time{}, time.Time{}, fmt.Errorf("TimeRangeBeforeMs and TimeRangeAfterMs must be passed together")
	}
	if before.Sub(after) > 24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("time range span cannot exceed one day")
	}
	if before.Sub(after) < 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("TimeRangeBeforeMs must be after TimeRangeAfterMs")
	}
	return before, after, nil
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

	if err := dbClient.DB.WithContext(ctx).Create(c).Error; err != nil {
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

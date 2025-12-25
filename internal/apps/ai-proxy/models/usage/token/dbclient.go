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
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/pkg/common/pbutil"
)

type DBClient struct {
	DB *gorm.DB
}

func (db *DBClient) Create(ctx context.Context, req *usagepb.TokenUsageCreateRequest) (*usagepb.TokenUsage, error) {
	if req == nil {
		return nil, gorm.ErrInvalidData
	}
	record := createRequestToRecord(req)
	if err := db.DB.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return record.ToProtobuf(), nil
}

func (db *DBClient) Paging(ctx context.Context, req *usagepb.TokenUsagePagingRequest) (*usagepb.TokenUsagePagingResponse, error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	model := &TokenUsage{
		CallID:        req.CallId,
		XRequestID:    req.XRequestId,
		ClientID:      req.ClientId,
		ClientTokenID: req.ClientTokenId,
		ProviderID:    req.ProviderId,
		ModelID:       req.ModelId,
		IsEstimated: func() bool {
			if req.IsEstimated != nil {
				return *req.IsEstimated
			}
			return false
		}(),
	}
	query := db.DB.WithContext(ctx).Model(model).Where(model)

	if len(req.Ids) > 0 {
		query.Where("id in (?)", req.Ids)
	}

	// time range
	if before, ok := pbutil.TimeFromMillis(req.TimeRangeBeforeMs); ok {
		query = query.Where("created_at <= ?", before)
	}
	if after, ok := pbutil.TimeFromMillis(req.TimeRangeAfterMs); ok {
		query = query.Where("created_at >= ?", after)
	}

	var (
		total int64
		list  TokenUsages
	)

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (req.PageNum - 1) * req.PageSize
	if err := query.Order("id DESC").Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error; err != nil {
		return nil, err
	}

	return &usagepb.TokenUsagePagingResponse{
		Total: uint64(total),
		List:  list.ToProtobuf(),
	}, nil
}

func (db *DBClient) Aggregate(ctx context.Context, req *usagepb.TokenUsagePagingRequest) ([]*usagepb.TokenUsage, error) {
	// handle request
	if err := handleAggregateRequest(req); err != nil {
		return nil, err
	}

	pagingResp, err := db.Paging(ctx, req)
	if err != nil {
		return nil, err
	}
	return pagingResp.List, nil
}

func handleAggregateRequest(req *usagepb.TokenUsagePagingRequest) error {
	req.PageSize = 1000000
	req.PageNum = 1

	if err := enforceTimeRangeWithin(req, time.Hour*24*31); err != nil {
		return err
	}

	return nil
}

func enforceTimeRangeWithin(req *usagepb.TokenUsagePagingRequest, maxSpan time.Duration) error {
	if req == nil || maxSpan <= 0 {
		return nil
	}

	after, hasAfter := pbutil.TimeFromMillis(req.TimeRangeAfterMs)
	before, hasBefore := pbutil.TimeFromMillis(req.TimeRangeBeforeMs)

	if hasAfter && hasBefore {
		if before.Before(after) {
			return fmt.Errorf("invalid time range: before < after")
		}
		if before.Sub(after) > maxSpan {
			return fmt.Errorf("time range exceeds limit of %s", maxSpan)
		}
		return nil
	}

	if hasAfter || hasBefore {
		return fmt.Errorf("time range requires both start and end timestamps within %s", maxSpan)
	}

	return fmt.Errorf("time range must be specified")
}

func createRequestToRecord(req *usagepb.TokenUsageCreateRequest) *TokenUsage {
	usageDetails := strings.TrimSpace(req.UsageDetails)
	if usageDetails == "" {
		usageDetails = "{}"
	}

	record := &TokenUsage{
		CallID:        req.CallId,
		XRequestID:    req.XRequestId,
		ClientID:      req.ClientId,
		ClientTokenID: req.ClientTokenId,
		ProviderID:    req.ProviderId,
		ModelID:       req.ModelId,
		InputTokens:   req.InputTokens,
		OutputTokens:  req.OutputTokens,
		TotalTokens:   req.TotalTokens,
		IsEstimated:   req.IsEstimated,
		Metadata:      metadata.FromProtobuf(req.Metadata),
		UsageDetails:  usageDetails,
	}
	if record.TotalTokens == 0 {
		record.TotalTokens = record.InputTokens + record.OutputTokens
	}
	if req.CreatedAt != nil {
		record.CreatedAt = req.CreatedAt.AsTime().UTC()
	} else {
		record.CreatedAt = time.Now().UTC()
	}
	return record
}

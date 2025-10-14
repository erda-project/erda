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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type TokenUsage struct {
	ID            uint64            `gorm:"column:id" json:"id" yaml:"id"`
	CallID        string            `gorm:"column:call_id;type:varchar(64)" json:"call_id" yaml:"call_id"`
	XRequestID    string            `gorm:"column:x_request_id;type:varchar(64)" json:"x_request_id" yaml:"x_request_id"`
	ClientID      string            `gorm:"column:client_id;type:char(36)" json:"client_id" yaml:"client_id"`
	ClientTokenID string            `gorm:"column:client_token_id;type:char(36)" json:"client_token_id" yaml:"client_token_id"`
	ProviderID    string            `gorm:"column:provider_id;type:char(36)" json:"provider_id" yaml:"provider_id"`
	ModelID       string            `gorm:"column:model_id;type:char(36)" json:"model_id" yaml:"model_id"`
	CreatedAt     time.Time         `gorm:"column:created_at;type:datetime(3)" json:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time         `gorm:"column:updated_at;type:datetime(3)" json:"updated_at" yaml:"updated_at"`
	InputTokens   uint64            `gorm:"column:input_tokens;type:bigint(20) unsigned" json:"input_tokens" yaml:"input_tokens"`
	OutputTokens  uint64            `gorm:"column:output_tokens;type:bigint(20) unsigned" json:"output_tokens" yaml:"output_tokens"`
	TotalTokens   uint64            `gorm:"column:total_tokens;type:bigint(20) unsigned" json:"total_tokens" yaml:"total_tokens"`
	IsEstimated   bool              `gorm:"column:is_estimated;type:tinyint(1);not null;default:0" json:"is_estimated" yaml:"is_estimated"`
	Metadata      metadata.Metadata `gorm:"column:metadata;type:mediumtext" json:"metadata" yaml:"metadata"`
	UsageDetails  string            `gorm:"column:usage_details;type:text" json:"usage_details" yaml:"usage_details"`
}

func (*TokenUsage) TableName() string { return "ai_proxy_token_usage" }

func (u *TokenUsage) ToProtobuf() *usagepb.TokenUsage {
	if u == nil {
		return nil
	}
	return &usagepb.TokenUsage{
		Id:            u.ID,
		CallId:        u.CallID,
		XRequestId:    u.XRequestID,
		ClientId:      u.ClientID,
		ClientTokenId: u.ClientTokenID,
		ProviderId:    u.ProviderID,
		ModelId:       u.ModelID,
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
		InputTokens:   u.InputTokens,
		OutputTokens:  u.OutputTokens,
		TotalTokens:   u.TotalTokens,
		IsEstimated:   u.IsEstimated,
		Metadata:      u.Metadata.ToProtobuf(),
		UsageDetails:  u.UsageDetails,
	}
}

type TokenUsages []*TokenUsage

func (us TokenUsages) ToProtobuf() []*usagepb.TokenUsage {
	var list []*usagepb.TokenUsage
	for _, item := range us {
		list = append(list, item.ToProtobuf())
	}
	return list
}

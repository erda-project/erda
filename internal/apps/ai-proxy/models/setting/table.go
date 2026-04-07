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

package setting

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/setting/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

type Setting struct {
	common.BaseModel
	Namespace string `gorm:"column:namespace;type:varchar(191);not null;uniqueIndex:uk_namespace_key,priority:1" json:"namespace"`
	Key       string `gorm:"column:key;type:varchar(191);not null;uniqueIndex:uk_namespace_key,priority:2" json:"key"`
	Value     string `gorm:"column:value;type:text;not null" json:"value"`
}

func (*Setting) TableName() string { return "ai_proxy_setting" }

func (s *Setting) ToProtobuf() *pb.Setting {
	return &pb.Setting{
		Id:        s.ID.String,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
		DeletedAt: timestamppb.New(s.DeletedAt.Time),
		Namespace: s.Namespace,
		Key:       s.Key,
		Value:     s.Value,
	}
}

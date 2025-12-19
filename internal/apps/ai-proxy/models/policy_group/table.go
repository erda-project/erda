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

package policy_group

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

// PolicyGroup gorm model for ai_proxy_policy_group.
type PolicyGroup struct {
	common.BaseModel

	ClientID  string                       `gorm:"column:client_id;type:char(36)" json:"clientID" yaml:"clientID"`
	Name      string                       `gorm:"column:name;type:varchar(191)" json:"name" yaml:"name"`
	Desc      string                       `gorm:"column:desc;type:varchar(1024)" json:"desc" yaml:"desc"`
	Mode      common_types.PolicyGroupMode `gorm:"column:mode;type:varchar(32)" json:"mode" yaml:"mode"`
	StickyKey string                       `gorm:"column:sticky_key;type:varchar(191)" json:"stickyKey" yaml:"stickyKey"`
	Branches  []*pb.PolicyBranch           `gorm:"column:branches;type:json;serializer:json" json:"branches" yaml:"branches"`
	Source    string                       `gorm:"column:source;type:varchar(191)" json:"source,omitempty" yaml:"source,omitempty"`
}

func (*PolicyGroup) TableName() string { return "ai_proxy_policy_group" }

func (pg *PolicyGroup) ToProtobuf() *pb.PolicyGroup {
	return &pb.PolicyGroup{
		Id:        pg.ID.String,
		CreatedAt: timestamppb.New(pg.CreatedAt),
		UpdatedAt: timestamppb.New(pg.UpdatedAt),
		DeletedAt: timestamppb.New(pg.DeletedAt.Time),
		ClientId:  pg.ClientID,
		Name:      pg.Name,
		Desc:      pg.Desc,
		Mode:      pg.Mode.String(),
		StickyKey: pg.StickyKey,
		Branches:  pg.Branches,
		Source:    pg.Source,
	}
}

// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apistructs

type AppPublishItemRelation struct {
	Env             string   `json:"env"`
	OrgID           int64    `json:"-"`
	AppID           int64    `json:"appId"`
	PublishItemID   int64    `json:"publishItemId"`
	PublishItemName string   `json:"publishItemName"`
	PublisherID     int64    `json:"publisherId"`
	PublisherName   string   `json:"publisherName"`
	PublishItemNs   []string `json:"publishItemNs"` // 同步 nexus 配置至 pipeline cm
	AK              string   `json:"ak"`
	AI              string   `json:"ai"`
}

type QueryAppPublishItemRelationGroupByENVResponse struct {
	Header
	Data map[string]AppPublishItemRelation `json:"data"`
}

type QueryAppPublishItemRelationRequest struct {
	AppID         int64  `query:"appID"`
	PublishItemID int64  `query:"publishItemID"`
	AK            string `query:"ak"`
	AI            string `query:"ai"`
}

type QueryAppPublishItemRelationResponse struct {
	Header
	Data []AppPublishItemRelation
}

type UpdateAppPublishItemRelationRequest struct {
	AppID         int64                         `json:"-"`
	UserID        string                        `json:"-"`
	ProdItemID    int64                         `json:"PROD"`
	STAGINGItemID int64                         `json:"STAGING"`
	TESTItemID    int64                         `json:"TEST"`
	DEVItemID     int64                         `json:"DEV"`
	AKAIMap       map[DiceWorkspace]MonitorKeys `json:"-"` // 0是AK，1是AI
}

type MonitorKeys struct {
	AK    string `json:"ak"`
	AI    string `json:"ai"`
	Env   string `json:"env"`
	AppID int64  `json:"appId"`
}

// GetPublishItemIDByWorkspace 根据环境获取对应的 publishItem ID
func (u *UpdateAppPublishItemRelationRequest) GetPublishItemIDByWorkspace(workspace DiceWorkspace) int64 {
	switch workspace {
	case ProdWorkspace:
		return u.ProdItemID
	case StagingWorkspace:
		return u.STAGINGItemID
	case TestWorkspace:
		return u.TESTItemID
	case DevWorkspace:
		return u.DEVItemID
	default:
		return 0
	}
}

// SetPublishItemIDTo0ByWorkspace 根据环境获取对应的 设置publishItem ID 为 0
func (u *UpdateAppPublishItemRelationRequest) SetPublishItemIDTo0ByWorkspace(workspace DiceWorkspace) {
	switch workspace {
	case ProdWorkspace:
		u.ProdItemID = 0
	case StagingWorkspace:
		u.STAGINGItemID = 0
	case TestWorkspace:
		u.TESTItemID = 0
	case DevWorkspace:
		u.DEVItemID = 0
	}
}

type UpdateAppPublishItemRelationResponse struct {
	Header
}

type RemoveAppPublishItemRelationsRequest struct {
	PublishItemId int64 `json:"publishItemId"`
}

type RemoveAppPublishItemRelationsResponse struct {
	Header
}

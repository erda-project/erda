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

package dao

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// IssueStream 事件流水表
type IssueStream struct {
	dbengine.BaseModel

	IssueID      int64                      // issue id
	Operator     string                     // 操作人
	StreamType   apistructs.IssueStreamType // 通过事件流类型找到对应模板
	StreamParams apistructs.ISTParam        // 事件流模板值, 参数取值范围见: apistructs.ISTParam
}

// TableName 表名
func (IssueStream) TableName() string {
	return "dice_issue_streams"
}

// Convert 事件流格式转换
func (stream *IssueStream) Convert(content string) apistructs.IssueStream {
	return apistructs.IssueStream{
		ID:         int64(stream.ID),
		IssueID:    stream.IssueID,
		Operator:   stream.Operator,
		StreamType: stream.StreamType,
		Content:    content,
		CreatedAt:  stream.CreatedAt,
		UpdatedAt:  stream.UpdatedAt,
	}
}

// CreateIssueStream 创建 issue 事件流
func (client *DBClient) CreateIssueStream(is *IssueStream) error {
	return client.Create(is).Error
}

// UpdateIssueStream 更新 issue 事件流
func (client *DBClient) UpdateIssueStream(is *IssueStream) error {
	return client.Update(is).Error
}

// PagingIssueStream issue 分页查询事件流
func (client *DBClient) PagingIssueStream(req *apistructs.IssueStreamPagingRequest) (int64, []IssueStream, error) {
	var (
		total        int64
		issueStreams []IssueStream
	)

	if err := client.Where("issue_id = ?", req.IssueID).
		Order("id DESC").
		Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&issueStreams).
		// reset offset & limit before count
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, err
	}
	return total, issueStreams, nil
}

func (client *DBClient) FindIssueStream(issueId int) ([]IssueStream, error) {
	var issueStreams []IssueStream

	if err := client.Where("issue_id = ?", issueId).
		Order("id DESC").Find(&issueStreams).Error; err != nil {
		return nil, err
	}
	return issueStreams, nil
}

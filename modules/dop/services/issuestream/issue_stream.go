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

package issuestream

import (
	"bytes"
	"strconv"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

// IssueStream issue stream service 对象
type IssueStream struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 IssueStream 对象配置选项
type Option func(*IssueStream)

// New 新建 issue stream 对象
func New(options ...Option) *IssueStream {
	is := &IssueStream{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(is *IssueStream) {
		is.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(is *IssueStream) {
		is.bdl = bdl
	}
}

// Create 创建事件记录
func (s *IssueStream) Create(req *apistructs.IssueStreamCreateRequest) (int64, error) {
	// TODO 请求校验
	// TODO 鉴权
	is := &dao.IssueStream{
		IssueID:      req.IssueID,
		Operator:     req.Operator,
		StreamType:   req.StreamType,
		StreamParams: req.StreamParams,
	}
	if err := s.db.CreateIssueStream(is); err != nil {
		return 0, err
	}

	if req.StreamType == apistructs.ISTRelateMR {
		// 添加事件应用关联关系
		issueAppRel := dao.IssueAppRelation{
			IssueID:   req.IssueID,
			CommentID: int64(is.ID),
			AppID:     req.StreamParams.MRInfo.AppID,
			MRID:      req.StreamParams.MRInfo.MRID,
		}
		if err := s.db.CreateIssueAppRelation(&issueAppRel); err != nil {
			return 0, err
		}
		// Send issue create event
		go func() {
			content, _ := GetDefaultContent(is.StreamType, is.StreamParams)
			issue, _ := s.db.GetIssue(req.IssueID)
			projectModel, _ := s.bdl.GetProject(issue.ProjectID)
			ev := &apistructs.EventCreateRequest{
				EventHeader: apistructs.EventHeader{
					Event:         bundle.IssueEvent,
					Action:        bundle.CreateAction,
					OrgID:         strconv.FormatInt(int64(projectModel.OrgID), 10),
					ProjectID:     strconv.FormatUint(issue.ProjectID, 10),
					ApplicationID: "-1",
					TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
				},
				Sender: bundle.SenderDOP,
				Content: apistructs.IssueEventData{
					Title:     issue.Title,
					Content:   content,
					AtUserIDs: issue.Assignee,
				},
			}
			if err := s.bdl.CreateEvent(ev); err != nil {
				logrus.Warnf("failed to send issue mr relate event, (%v)", err)
			}
		}()
	}

	return int64(is.ID), nil
}

// Paging 事件流记录分页查询
func (s *IssueStream) Paging(req *apistructs.IssueStreamPagingRequest) (*apistructs.IssueStreamPagingResponseData, error) {
	// 请求校验
	if req.IssueID == 0 {
		return nil, apierrors.ErrPagingIssueStream.MissingParameter("missing issueID")
	}
	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	// 分页查询
	total, issueStreams, err := s.db.PagingIssueStream(req)
	if err != nil {
		return nil, err
	}
	iss := make([]apistructs.IssueStream, 0, len(issueStreams))
	for _, v := range issueStreams {
		is := apistructs.IssueStream{
			ID:         int64(v.ID),
			IssueID:    v.IssueID,
			Operator:   v.Operator,
			StreamType: v.StreamType,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
		}
		if v.StreamType == apistructs.ISTRelateMR {
			is.MRInfo = v.StreamParams.MRInfo
		} else {
			content, err := GetDefaultContent(v.StreamType, v.StreamParams)
			if err != nil {
				return nil, err
			}
			is.Content = content

		}
		iss = append(iss, is)
	}

	resp := &apistructs.IssueStreamPagingResponseData{
		Total: total,
		List:  iss,
	}
	return resp, nil
}

// GetDefaultContent 获取渲染后的事件流内容
func GetDefaultContent(ist apistructs.IssueStreamType, param apistructs.ISTParam) (string, error) {
	locale := "zh"
	ct, err := apistructs.GetIssueStreamTemplate(locale, ist)
	if err != nil {
		return "", err
	}
	tpl, err := template.New("c").Parse(ct)
	if err != nil {
		return "", err
	}

	var content bytes.Buffer
	if err := tpl.Execute(&content, param.Localize(locale)); err != nil {
		return "", err
	}

	return content.String(), nil
}

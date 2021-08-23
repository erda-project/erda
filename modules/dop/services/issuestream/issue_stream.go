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
	}

	// send issue create or update event when creat issue stream
	go func() {
		if err := s.CreateIssueEvent(req.IssueID, req.StreamType, req.StreamParams); err != nil {
			logrus.Errorf("create issue %d event err: %v", req.IssueID, err)
		}
	}()

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

// CreateIssueEvent create issue event
func (s *IssueStream) CreateIssueEvent(issueID int64, streamType apistructs.IssueStreamType,
	streamParams apistructs.ISTParam) error {
	content, err := GetDefaultContent(streamType, streamParams)
	if err != nil {
		logrus.Errorf("get issue %d content error: %v, content will be empty", issueID, err)
	}
	logrus.Debugf("old issue content is: %s", content)
	issue, err := s.db.GetIssue(issueID)
	if err != nil {
		return err
	}
	receivers, err := s.db.GetReceiversByIssueID(issueID)
	if err != nil {
		logrus.Errorf("get issue %d  recevier error: %v, recevicer will be empty", issueID, err)
		receivers = []string{}
	}
	projectModel, err := s.bdl.GetProject(issue.ProjectID)
	if err != nil {
		return err
	}
	orgModel, err := s.bdl.GetOrg(int64(projectModel.OrgID))
	if err != nil {
		return err
	}
	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.IssueEvent,
			Action:        streamType.GetEventAction(),
			OrgID:         strconv.FormatInt(int64(projectModel.OrgID), 10),
			ProjectID:     strconv.FormatUint(issue.ProjectID, 10),
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender: bundle.SenderDOP,
		Content: apistructs.IssueEventData{
			Title:        issue.Title,
			Content:      content,
			AtUserIDs:    issue.Assignee,
			IssueType:    issue.Type,
			StreamType:   streamType,
			StreamParams: streamParams,
			Receivers:    receivers,
			Params: map[string]string{
				"orgName":     orgModel.Name,
				"projectName": projectModel.Name,
				"issueID":     strconv.FormatInt(issueID, 10),
			},
		},
	}

	return s.bdl.CreateEvent(ev)
}

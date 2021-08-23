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

package issuepanel

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// IssuePanel issue panel 对象
type IssuePanel struct {
	db    *dao.DBClient
	bdl   *bundle.Bundle
	issue *issue.Issue
}

// Option 定义 IssuePanel 对象配置选项
type Option func(*IssuePanel)

// New 新建 IssuePanel 对象
func New(options ...Option) *IssuePanel {
	is := &IssuePanel{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(is *IssuePanel) {
		is.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(is *IssuePanel) {
		is.bdl = bdl
	}
}

// WithIssue 配置 issue service
func WithIssue(i *issue.Issue) Option {
	return func(is *IssuePanel) {
		is.issue = i
	}
}

// CreatePanel 增加自定义面板
func (ip *IssuePanel) CreatePanel(req *apistructs.IssuePanelRequest) (*dao.IssuePanel, error) {
	issuePanel := &dao.IssuePanel{
		ProjectID: req.ProjectID,
		PanelName: req.PanelName,
		IssueID:   0,
		Relation:  0,
	}
	if err := ip.db.CreatePanel(issuePanel); err != nil {
		return nil, err
	}
	return issuePanel, nil
}

// DeletePanel 删除自定义面板
func (ip *IssuePanel) DeletePanel(req *apistructs.IssuePanelRequest) (*apistructs.IssuePanel, error) {
	panel, err := ip.db.GetPanelByPanelID(req.PanelID)
	if err != nil {
		return nil, err
	}
	err = ip.db.DeletePanelByPanelID(req.PanelID)
	res := &apistructs.IssuePanel{
		PanelName: panel.PanelName,
		PanelID:   int64(panel.ID),
	}
	return res, nil
}

func (ip *IssuePanel) UpdatePanel(req *apistructs.IssuePanelRequest) (*dao.IssuePanel, error) {
	panel, err := ip.db.GetPanelByPanelID(req.PanelID)
	if err != nil {
		return nil, err
	}
	panel.PanelName = req.PanelName
	if err := ip.db.UpdatePanel(panel); err != nil {
		return nil, err
	}
	return panel, nil
}

func (ip *IssuePanel) UpdatePanelIssue(req *apistructs.IssuePanelRequest) (*dao.IssuePanel, error) {
	// 默认返回值
	defaultPanel := &dao.IssuePanel{
		BaseModel: dbengine.BaseModel{
			ID: 0,
		},
	}
	panel, err := ip.GetPanelIssue(req)
	if err != nil {
		return nil, err
	}
	if panel != nil {
		// 如果是将事件移出自定义看板
		if req.PanelID == 0 {
			if panel.ID != 0 {
				if err := ip.db.DeletePanelByPanelID(int64(panel.ID)); err != nil {
					return nil, err
				}
			}
			return defaultPanel, nil
		}
		// 如果事件有所属看板则更新
		panel.Relation = req.PanelID
		if err := ip.db.UpdatePanel(panel); err != nil {
			return nil, err
		}
		return panel, nil
	}
	// 如果事件无所属看板则新建
	if req.PanelID == 0 {
		return defaultPanel, nil
	}
	relatedPanel, err := ip.db.GetPanelByPanelID(req.PanelID)
	if err != nil {
		return nil, err
	}
	issuePanel := &dao.IssuePanel{
		ProjectID: relatedPanel.ProjectID,
		PanelName: relatedPanel.PanelName,
		IssueID:   req.IssueID,
		Relation:  int64(relatedPanel.ID),
	}
	err = ip.db.CreatePanel(issuePanel)
	if err != nil {
		return nil, err
	}
	return issuePanel, nil
}

func (ip *IssuePanel) GetPanelByProjectID(req *apistructs.IssuePanelRequest) ([]apistructs.IssuePanelIssues, error) {
	var res []apistructs.IssuePanelIssues
	res = append(res, apistructs.IssuePanelIssues{
		IssuePanel: apistructs.IssuePanel{
			PanelName: "默认",
			PanelID:   0,
		},
	})
	panels, err := ip.db.GetPanelsByProjectID(req.ProjectID)
	if err != nil {
		return nil, err
	}
	for _, p := range panels {
		res = append(res, apistructs.IssuePanelIssues{
			IssuePanel: apistructs.IssuePanel{
				PanelName: p.PanelName,
				PanelID:   int64(p.ID),
			},
		})
	}
	return res, nil
}

// GetPanelIssues 获取看板下的事件ID
func (ip *IssuePanel) GetPanelIssues(req *apistructs.IssuePanelRequest) ([]apistructs.Issue, uint64, error) {
	req.External = true
	// custom boards
	if req.PanelID != 0 {
		req.IssuePagingRequest.IssueListRequest.CustomPanelID = req.PanelID
		issues, total, err := ip.issue.Paging(req.IssuePagingRequest)
		if err != nil {
			return nil, 0, err
		}
		return issues, total, err
	} else {
		// default board
		ids, err := ip.db.GetPanelIssuesIDByProjectID(req.ProjectID)
		if err != nil {
			return nil, 0, err
		}
		for _, v := range ids {
			req.ExceptIDs = append(req.ExceptIDs, v.IssueID)
		}
		issues, total, err := ip.issue.Paging(req.IssuePagingRequest)
		if err != nil {
			return nil, 0, err
		}
		return issues, total, nil
	}
}

func (ip *IssuePanel) GetPanelIssue(req *apistructs.IssuePanelRequest) (*dao.IssuePanel, error) {
	panel, err := ip.db.GetPanelByIssueID(req.IssueID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return panel, nil
}

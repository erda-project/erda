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

package issuepanel

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/issue"
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
		PanelID:   panel.ID,
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
		BaseModel: dao.BaseModel{
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
				if err := ip.db.DeletePanelByPanelID(panel.ID); err != nil {
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
		Relation:  relatedPanel.ID,
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
				PanelID:   p.ID,
			},
		})
	}
	return res, nil
}

// GetPanelIssues 获取看板下的事件ID
func (ip *IssuePanel) GetPanelIssues(req *apistructs.IssuePanelRequest) ([]apistructs.Issue, uint64, error) {
	var res []int64
	req.External = true
	// 如果是自定义创建的看板
	if req.PanelID != 0 {
		panels, total, err := ip.db.GetPanelIssuesByPanel(req.PanelID, req.PageNo, req.PageSize)
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, 0, nil
		}
		for _, p := range panels {
			res = append(res, p.IssueID)
		}
		req.IDs = res
		issues, total, err := ip.issue.Paging(req.IssuePagingRequest)
		if err != nil {
			return nil, 0, err
		}
		return issues, total, nil
	} else {
		//不属于新创建看板的事件
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

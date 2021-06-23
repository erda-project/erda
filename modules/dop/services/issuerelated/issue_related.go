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

package issuerelated

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
)

type IssueRelated struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 IssueStream 对象配置选项
type Option func(*IssueRelated)

// New 新建 issue stream 对象
func New(options ...Option) *IssueRelated {
	is := &IssueRelated{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(is *IssueRelated) {
		is.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(is *IssueRelated) {
		is.bdl = bdl
	}
}

// AddRelatedIssue 添加issue关联关系
func (ir *IssueRelated) AddRelatedIssue(req *apistructs.IssueRelationCreateRequest) (*dao.IssueRelation, error) {
	issueRel := &dao.IssueRelation{
		IssueID:      req.IssueID,
		RelatedIssue: req.RelatedIssue,
		Comment:      req.Comment,
	}

	if err := ir.db.CreateIssueRelations(issueRel); err != nil {
		return nil, err
	}

	return issueRel, nil
}

// GetIssueRelationsByIssueIDs 获取issue的关联关系
func (ir *IssueRelated) GetIssueRelationsByIssueIDs(issueID uint64) ([]uint64, []uint64, error) {
	relatingIssueIDs, err := ir.db.GetRelatingIssues(issueID)
	if err != nil {
		return nil, nil, err
	}

	relatedIssueIDs, err := ir.db.GetRelatedIssues(issueID)
	if err != nil {
		return nil, nil, err
	}

	return relatingIssueIDs, relatedIssueIDs, nil
}

// DeleteIssueRelation 删除issue关联关系
func (ir *IssueRelated) DeleteIssueRelation(issueID, relatedIssueID uint64) error {
	if err := ir.db.DeleteIssueRelation(issueID, relatedIssueID); err != nil {
		return err
	}

	return nil
}

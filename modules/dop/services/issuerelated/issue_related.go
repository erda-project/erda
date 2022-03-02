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

package issuerelated

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
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

func (ir *IssueRelated) ValidIssueRelationType(id uint64, relationType apistructs.IssueType) error {
	issue, err := ir.db.GetIssue(int64(id))
	if err != nil {
		return err
	}
	if issue.Type != relationType {
		return fmt.Errorf("issue id %v type is %v, not %v", id, issue.Type, relationType)
	}
	return nil
}

func (ir *IssueRelated) ValidIssueRelationTypes(ids []uint64, relationType []apistructs.IssueType) error {
	issueIDs := make([]int64, 0, len(ids))
	for _, i := range ids {
		issueIDs = append(issueIDs, int64(i))
	}
	issues, err := ir.db.ListIssue(apistructs.IssueListRequest{IDs: issueIDs, Type: relationType})
	if err != nil {
		return err
	}
	if len(issues) > 0 {
		return fmt.Errorf("issue ids %v contains invalid type", ids)
	}
	return nil
}

// AddRelatedIssue 添加issue关联关系
func (ir *IssueRelated) AddRelatedIssue(req *apistructs.IssueRelationCreateRequest) ([]dao.IssueRelation, error) {
	if req.Type == apistructs.IssueRelationInclusion {
		if err := ir.ValidIssueRelationType(req.IssueID, apistructs.IssueTypeRequirement); err != nil {
			return nil, err
		}
		if err := ir.ValidIssueRelationTypes(req.RelatedIssue, []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeBug}); err != nil {
			return nil, err
		}
	}
	issueRels := make([]dao.IssueRelation, 0, len(req.RelatedIssue))
	for _, i := range req.RelatedIssue {
		issueRels = append(issueRels, dao.IssueRelation{
			IssueID:      req.IssueID,
			RelatedIssue: i,
			Comment:      req.Comment,
			Type:         req.Type,
		})
	}

	if err := ir.db.BatchCreateIssueRelations(issueRels); err != nil {
		return nil, err
	}

	if req.Type == apistructs.IssueRelationInclusion {
		if err := ir.AfterIssueInclusionRelationChange(req.IssueID); err != nil {
			return nil, err
		}
	}
	return issueRels, nil
}

// GetIssueRelationsByIssueIDs 获取issue的关联关系
func (ir *IssueRelated) GetIssueRelationsByIssueIDs(issueID uint64, relationType []string) ([]uint64, []uint64, error) {
	relatingIssueIDs, err := ir.db.GetRelatingIssues(issueID, relationType)
	if err != nil {
		return nil, nil, err
	}

	relatedIssueIDs, err := ir.db.GetRelatedIssues(issueID, relationType)
	if err != nil {
		return nil, nil, err
	}

	return relatingIssueIDs, relatedIssueIDs, nil
}

// DeleteIssueRelation 删除issue关联关系
func (ir *IssueRelated) DeleteIssueRelation(issueID, relatedIssueID uint64, relationTypes []string) error {
	if err := ir.db.DeleteIssueRelation(issueID, relatedIssueID, relationTypes); err != nil {
		return err
	}

	if strutil.Exist(relationTypes, apistructs.IssueRelationInclusion) {
		return ir.AfterIssueInclusionRelationChange(issueID)
	}
	return nil
}

func (ir *IssueRelated) AfterIssueInclusionRelationChange(id uint64) error {
	fields := make(map[string]interface{})
	start, end, err := ir.db.FindIssueChildrenTimeRange(id)
	if err != nil {
		return err
	}
	if start != nil {
		fields["plan_started_at"] = start
	}
	if end != nil {
		now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
		fields["expiry_status"] = dao.GetExpiryStatus(end, now)
		fields["plan_finished_at"] = end
	}
	if len(fields) > 0 {
		if err := ir.db.UpdateIssue(id, fields); err != nil {
			return apierrors.ErrUpdateIssue.InternalError(err)
		}
	}
	return nil
}

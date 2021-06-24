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

package issue

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func (svc *Issue) CreateIssueStage(req *apistructs.IssueStageRequest) error {
	err := svc.db.DeleteIssuesStage(req.OrgID, req.IssueType)
	if err != nil {
		return err
	}
	var stages []dao.IssueStage
	for _, v := range req.List {
		stage := dao.IssueStage{
			OrgID:     req.OrgID,
			IssueType: req.IssueType,
			Name:      v.Name,
			Value:     v.Value,
		}
		if stage.Value == "" {
			stage.Value = v.Name
		}
		stages = append(stages, stage)
	}
	return svc.db.CreateIssueStage(stages)
}

func (svc *Issue) GetIssueStage(req *apistructs.IssueStageRequest) ([]apistructs.IssueStage, error) {
	stages, err := svc.db.GetIssuesStage(req.OrgID, req.IssueType)
	if err != nil {
		return nil, err
	}
	var res []apistructs.IssueStage
	for _, v := range stages {
		res = append(res, apistructs.IssueStage{
			ID:    int64(v.ID),
			Name:  v.Name,
			Value: v.Value,
		})
	}
	return res, nil
}
